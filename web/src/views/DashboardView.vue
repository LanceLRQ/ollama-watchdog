<script setup>
import { ref, onMounted } from "vue";
import { use } from "echarts/core";
import { CanvasRenderer } from "echarts/renderers";
import { LineChart } from "echarts/charts";
import { LegendComponent } from "echarts/components";
import {
    GridComponent,
    TooltipComponent,
    TitleComponent,
} from "echarts/components";
import {
    ElMessage,
    ElMessageBox
} from 'element-plus';
import VChart from "vue-echarts";
import { get, pick, isArray } from "lodash";
import { DefaultServerSettings } from "../models/server";
import dayjs from "dayjs";
import axios from "axios";

use([
    CanvasRenderer,
    LineChart,
    GridComponent,
    TooltipComponent,
    TitleComponent,
    LegendComponent,
]);

const wsWorker = ref(null)
const wsReconnectAttempts = ref(0);
const wsMaxReconnectAttempts = ref(50);


const ChartHistoryData = ref([]);
const GPUProcessData = ref([]);
const maxHistory = ref(120); // 保留60个数据点（1分钟）
const GPUList = ref([]);
const CurrentGPUInfo = ref([]);
const OllamaPSList = ref([]);

const serverSettings = ref({ ...DefaultServerSettings });

const loadSettingsFromLocalStorage = () => {
    // 从本地存储加载设置
    const settings = localStorage.getItem("watchdog_api_settings");
    if (settings) {
        try {
            serverSettings.value = Object.assign(
                {},
                DefaultServerSettings,
                JSON.parse(settings)
            );
        } catch (e) {
            console.error("Error parsing settings:", e);
        }
    }
    return settings;
};



const connectWebSocket = () => {
    const url = `${location.protocol === "https" ? "wss" : "ws"}://${serverSettings.value.apiHost}${serverSettings.value.apiBasePath}/realtime`;
    console.log(url)
    wsWorker.value = new WebSocket(url);


    let timeoutId;

    // 设置超时定时器
    timeoutId = setTimeout(() => {
        console.log('WebSocket 连接超时');
        wsWorker.value.close(); // 关闭连接
    }, 2000);

    // WebSocket 打开事件
    wsWorker.value.onopen = () => {
        clearTimeout(timeoutId);
        wsReconnectAttempts.value = 0; // 重置重连次数
        console.log('WebSocket 连接成功');
    };

    // WebSocket 关闭事件
    wsWorker.value.onclose = () => {
        clearTimeout(timeoutId);
        console.log('WebSocket 连接关闭');
        handleReconnect();
    };

    // WebSocket 错误事件
    wsWorker.value.onerror = (error) => {
        wsWorker.value.close(); // 关闭连接
        clearTimeout(timeoutId);
        console.error('WebSocket 错误:', error);
        handleReconnect();
    };

    wsWorker.value.onmessage = (event) => {
        const newData = JSON.parse(event.data);

        // 更新历史数据
        const GPUListData = get(newData, "nvidia.gpu_info", []);
        ChartHistoryData.value.push({
            list: GPUListData,
            timestamp: get(newData, "nvidia.timestamp", 0),
        });
        CurrentGPUInfo.value = GPUListData;
        GPUProcessData.value = get(newData, "nvidia.gpu_processes", []);

        if (ChartHistoryData.value.length > maxHistory.value) {
            ChartHistoryData.value.shift();
        }

        GPUList.value = GPUListData.map((item, index) => {
            return pick(item, ["device_id", "bus_id", "name"]);
        });

        // 更新Ollama数据
        if (get(newData, "ollama.status")) {
            OllamaPSList.value = get(newData, "ollama.data.models", []);
        }
    };
}

const handleReconnect = () => {
    if (wsReconnectAttempts.value < wsMaxReconnectAttempts.value) {
        wsReconnectAttempts.value++;
        setTimeout(() => {
            console.log(`尝试重连，第 ${wsReconnectAttempts.value} 次`);
            connectWebSocket();
        }, 5000);
    } else {
        console.log('已达到最大重连次数，停止重连');
    }
}

// 加载GPU采样历史数据
const LoadGPUSampleHistoryData = (callback) => {
    axios.get(`//${serverSettings.value.apiHost}${serverSettings.value.apiBasePath}/nvidia/history?range=${maxHistory.value}`)
        .then(response => {
            const resp = response.data;
            if (resp.status) {
                const list = get(resp, 'data', []);
                list.forEach(item => {
                    ChartHistoryData.value.push({
                        list: get(item, 'gpu_info', []),
                        timestamp: get(item, "timestamp", 0),
                    })
                });
            }
            callback()
        })
}

onMounted(() => {
    loadSettingsFromLocalStorage();
    LoadGPUSampleHistoryData(() => {
        connectWebSocket();
    });
});


const customToFix = (value, fixed) => {
    const fixStr = value.toFixed(fixed);
    return fixStr.replace(/\.0+$/, "");
};

const metricMap = {
    temperature: {
        key: "temperature",
        name: "温度",
        proc: (info) => {
            return info.temperature;
        },
        unit: "℃",
    },
    usaged: [
        {
            key: "mem_used",
            name: "显存使用率",
            proc: (info) => {
                return (info.mem_used / info.mem_total) * 100;
            },
            unit: "%",
        },
        {
            key: "gpu_used",
            name: "GPU使用率",
            proc: (info) => {
                return info.gpu_used;
            },
            unit: "%",
        },
    ],
    power_usage: {
        key: "power_usage",
        name: "功耗",
        proc: (info) => {
            return info.power_usage;
        },
        unit: "W",
    },
    power_limit: {
        key: "power_limit",
        name: "功耗",
        proc: (info) => {
            return info.power_limit;
        },
        unit: "W",
    },
    mem_used: {
        key: "mem_used",
        name: "显存占用",
        proc: (info) => {
            return info.mem_used;
        },
        unit: "MB",
    },
    mem_total: {
        key: "mem_total",
        name: "总显存",
        proc: (info) => {
            return info.mem_total;
        },
        unit: "MB",
    },
};

function formatBytes(bytes, decimals = 1, base = 1024) {
    if (bytes === 0) return '0 B';

    const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    let value = bytes;
    let unitIndex = 0;

    while (value >= base && unitIndex < units.length - 1) {
        value /= base;
        unitIndex += 1;
    }

    // 四舍五入并保留指定小数位
    return `${value.toFixed(decimals)} ${units[unitIndex]}`;
}

const handleKillProcess = (pid) => {
    ElMessageBox.confirm(
        `确定要结束进程 ${pid} 吗？`,
        'Warning',
        {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning',
        }
    )
        .then(() => {
            axios.post(`//${serverSettings.value.apiHost}${serverSettings.value.apiBasePath}/kill`, {
                type: 'process',
                pid,
            }).then((resp) => {
                if (resp.data.status) {
                    ElMessage({
                        type: 'success',
                        message: '结束进程成功',
                    });
                } else {
                    ElMessage({
                        type: 'error',
                        message: '结束进程失败',
                    });
                    console.log(resp.data.message);
                }
            }).catch((error) => {
                ElMessage.error('结束进程失败');
                console.log(error);
            });
        })
        .catch(() => { })
}

const handleOllamaKillProcess = (name) => {
    ElMessageBox.confirm(
        `确定要停止模型 ${name} 吗？`,
        'Warning',
        {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning',
        }
    )
        .then(() => {
            axios.post(`//${serverSettings.value.apiHost}${serverSettings.value.apiBasePath}/kill`, {
                type: 'ollama',
                name,
            }).then((resp) => {
                if (resp.data.status) {
                    ElMessage({
                        type: 'success',
                        message: '停止模型成功',
                    });
                } else {
                    ElMessage({
                        type: 'error',
                        message: '停止模型失败',
                    });
                    console.log(resp.data.message);
                }
            }).catch((error) => {
                ElMessage.error('停止模型失败');
                console.log(error);
            });
        })
        .catch(() => { })
}

const getCurrentValue = (metric, gpuIndex) => {
    let currentVal = "";
    try {
        const gpuInfo = CurrentGPUInfo.value.find(
            (item) => item.device_id === GPUList.value[gpuIndex].device_id
        );
        currentVal = metric.proc(gpuInfo);
    } catch (error) { }
    return customToFix(currentVal);
};

const getChartOption = (gpuIndex, metricKey) => {
    const metricsItem = metricMap[metricKey];
    if (!metricsItem) return {};
    let metrics = metricsItem;
    if (!isArray(metricsItem)) {
        metrics = [metricsItem];
    }
    return {
        tooltip: {
            trigger: "axis",
            formatter: function (params) {
                return metrics
                    .map((metric, index) => {
                        const args = params[index].value;
                        var date = new Date(args[0] * 1000);
                        return `${dayjs(date).format("YYYY-MM-DD HH:mm:ss")}<br /><strong>${metric.name
                            }</strong>：${customToFix(args[1])} ${metric.unit || ""}`;
                    })
                    .join("<br />");
            },
            axisPointer: {
                animation: false,
            },
        },
        legend:
            metrics.length > 1
                ? { data: metrics.map((metric) => metric.name) }
                : false,
        xAxis: {
            type: "time",
            splitLine: {
                show: false,
            },
            axisLabel: {
                formatter: function (value) {
                    // 将时间戳转换为日期字符串
                    return dayjs(value * 1000).format("HH:mm:ss");
                },
            },
        },
        yAxis: metrics.map((metric) => ({
            type: "value",
            boundaryGap: [0, "100%"],
            splitLine: {
                show: false,
            },
            name: `${metric.name} (${metric.unit || ""}）`,
            ...(metric.unit === "%" ? { max: 100, min: 0 } : {}),
        })),
        series: metrics.map((metric) => ({
            name: metric.name,
            type: "line",
            showSymbol: false,
            data: ChartHistoryData.value.map((record) => {
                const gpuInfo = record.list.find(
                    (item) => item.device_id === GPUList.value[gpuIndex].device_id
                );
                return {
                    name: metric.name,
                    value: [record.timestamp, metric.proc(gpuInfo)],
                };
            }),
        })),
    };
};
</script>

<template>
    <el-row :gutter="200" class="monitor-container">
        <el-col>
            <el-card v-for="(gpu, index) in GPUList">
                <template #header>
                    <div class="card-header">
                        <el-space>
                            <el-tag>GPU {{ index + 1 }}</el-tag>
                            <el-tag type="info">{{ gpu.bus_id }}</el-tag>
                            <span>{{ gpu.name }}</span>
                        </el-space>
                    </div>
                </template>
                <el-row :gutter="20" :key="index">
                    <el-col :span="8" :xl="6">
                        <v-chart class="chart" :option="getChartOption(index, 'temperature')"
                            :update-options="{ notMerge: true }" autoresize />
                    </el-col>
                    <el-col :span="8" :xl="12">
                        <v-chart class="chart" :option="getChartOption(index, 'usaged')"
                            :update-options="{ notMerge: true }" autoresize />
                    </el-col>
                    <el-col :span="8" :xl="6">
                        <v-chart class="chart" :option="getChartOption(index, 'power_usage')"
                            :update-options="{ notMerge: true }" autoresize />
                    </el-col>
                </el-row>
                <template #footer>
                    <el-descriptions title="当前数据">
                        <el-descriptions-item :label="metricMap.temperature.name">
                            {{ getCurrentValue(metricMap.temperature, index) }}
                            {{ metricMap.temperature.unit }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="metricMap.usaged[0].name">
                            {{ getCurrentValue(metricMap.usaged[0], index) }}
                            {{ metricMap.usaged[0].unit }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="metricMap.usaged[1].name">
                            {{ getCurrentValue(metricMap.usaged[1], index) }}
                            {{ metricMap.usaged[1].unit }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="metricMap.power_usage.name">
                            {{ getCurrentValue(metricMap.power_usage, index) }}
                            {{ metricMap.power_usage.unit }} /
                            {{ getCurrentValue(metricMap.power_limit, index) }}
                            {{ metricMap.power_limit.unit }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="metricMap.mem_used.name">
                            {{ getCurrentValue(metricMap.mem_used, index) }}
                            {{ metricMap.mem_used.unit }} /
                            {{ getCurrentValue(metricMap.mem_total, index) }}
                            {{ metricMap.mem_total.unit }}
                        </el-descriptions-item>
                    </el-descriptions>
                    <el-table :data="GPUProcessData" style="width: 100%">
                        <el-table-column prop="bus_id" label="总线ID" width="180">
                            <template #default="scope">
                                <el-tag type="info">{{ scope.row.bus_id }}</el-tag>
                            </template>
                        </el-table-column>
                        <el-table-column prop="pid" label="进程ID" width="180" />
                        <el-table-column prop="name" label="进程名称" />
                        <el-table-column label="显存占用">
                            <template #default="scope">
                                {{ scope.row.mem_used }} MB
                            </template>
                        </el-table-column>
                        <!-- <el-table-column label="Operations">
                            <template #default="scope">
                                <el-button size="small" type="danger" @click="handleKillProcess(scope.row.pid)">
                                    结束进程
                                </el-button>
                            </template>
                        </el-table-column> -->
                    </el-table>
                </template>
            </el-card>
        </el-col>
        <el-col style="margin-top:20px">
            <el-card>
                <template #header>
                    <div class="card-header">
                        Ollama 进程管理
                    </div>
                </template>
                <template #default>
                    <el-table :data="OllamaPSList" style="width: 100%">
                        <el-table-column prop="bus_id" label="Hash" width="150">
                            <template #default="scope">
                                <el-tag type="info">{{ scope.row.digest.substring(0, 8) }}</el-tag>
                            </template>
                        </el-table-column>
                        <el-table-column prop="name" label="模型名称" />
                        <el-table-column label="资源占用 (GPU/CPU)">
                            <template #default="scope">
                                {{ formatBytes(scope.row.size_vram) }}
                                <span v-if="scope.row.size - scope.row.size_vram > 0"> / {{
                                    formatBytes(scope.row.size -
                                        scope.row.size_vram) }}</span>

                                ( {{ (scope.row.size_vram / scope.row.size * 100).toFixed(0) }}% GPU
                                <span v-if="scope.row.size - scope.row.size_vram > 0"> / {{
                                    ((scope.row.size - scope.row.size_vram) / scope.row.size * 100).toFixed(0) }}%
                                    CPU</span>)
                            </template>
                        </el-table-column>
                        <el-table-column prop="size" label="过期时间" width="200">
                            <template #default="scope">
                                <el-tag type="info">{{ dayjs(scope.row.expires_at).format('YYYY-MM-DD HH:mm:ss')
                                }}</el-tag>
                            </template>
                        </el-table-column>
                        <el-table-column label="操作">
                            <template #default="scope">
                                <el-button size="small" type="danger" @click="handleOllamaKillProcess(scope.row.name)">
                                    停止
                                </el-button>
                            </template>
                        </el-table-column>
                    </el-table>
                </template>
            </el-card>
        </el-col>
    </el-row>
</template>

<style scoped>
.monitor-container {
    padding: 20px;
}

.chart {
    height: 300px;
}
</style>
