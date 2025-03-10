<script setup>
import { ref, onMounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { LegendComponent } from 'echarts/components';
import { GridComponent, TooltipComponent, TitleComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { get, pick, isArray } from 'lodash';
import { DefaultServerSettings } from '../models/server'
import dayjs from 'dayjs';


use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, TitleComponent, LegendComponent])

const ChartHistoryData = ref([])
const GPUProcessData = ref([])
const maxHistory = 60 // 保留60个数据点（1分钟）
const GPUList = ref([])

const serverSettings = ref({ ...DefaultServerSettings })

const loadSettingsFromLocalStorage = () => {
    // 从本地存储加载设置
    const settings = localStorage.getItem('watchdog_api_settings')
    if (settings) {
        try {
            serverSettings.value = Object.assign({}, DefaultServerSettings, JSON.parse(settings))
        } catch (e) {
            console.error('Error parsing settings:', e)
        }
    }
    return settings
}

onMounted(() => {
    loadSettingsFromLocalStorage();
    const ws = new WebSocket(`${location.protocol === 'https' ? 'wss' : 'ws'}://${serverSettings.value.apiHost}${serverSettings.value.apiBasePath}/ws`)

    ws.onmessage = (event) => {
        const newData = JSON.parse(event.data)

        const GPUListData = get(newData, 'gpu_info', [])
        const GPUProcessData = get(newData, 'gpu_processes', [])
        // 更新历史数据
        ChartHistoryData.value.push({
            list: GPUListData,
            timestamp: get(newData, 'timestamp', 0)
        })
        GPUProcessData.value = GPUProcessData

        if (ChartHistoryData.value.length > maxHistory) {
            ChartHistoryData.value.shift()
        }

        GPUList.value = GPUListData.map((item, index) => {
            return pick(item, ['device_id', 'bus_id', 'name'])
        });
    }
})

const customToFix = (value, fixed) => {
    const fixStr = value.toFixed(fixed);
    return fixStr.replace(/\.0+$/, '')
}

const metricMap = {
    temperature: {
        key: 'temperature',
        name: '温度',
        proc: (info) => {
            return info.temperature;
        },
        unit: '℃',
    },
    usaged: [
        {
            key: 'mem_used',
            name: '显存使用率',
            proc: (info) => {
                return (info.mem_used / info.mem_total) * 100;
            },
            unit: '%',
        },
        {
            key: 'gpu_used',
            name: 'GPU使用率',
            proc: (info) => {
                return info.gpu_used;
            },
            unit: '%',
        }
    ],
    power_usage: {
        key: 'power_usage',
        name: '功耗',
        proc: (info) => {
            return info.power_usage;
        },
        unit: 'W',
    }
}

const getCurrentValue = (metric) => {
    let currentVal = '';
    try {
        const gpuInfo = ChartHistoryData.value[0].list.find(item => item.device_id === GPUList.value[gpuIndex].device_id);
        currentVal = metric.proc(gpuInfo);
    } catch (error) {}
}

const getChartOption = (gpuIndex, metricKey) => {
   
    const metricsItem = metricMap[metricKey];
    if (!metricsItem) return {};
    let metrics = metricsItem;
    if (!isArray(metricsItem)){
        metrics = [metricsItem];
    }
    return {
        tooltip: {
            trigger: 'axis',
            formatter: function (params) {
                return metrics.map((metric, index) => {
                    const args = params[index].value;
                    var date = new Date(args[0]* 1000) ;
                    return `${dayjs(date).format('YYYY-MM-DD HH:mm:ss')}<br /><strong>${metric.name}</strong>：${customToFix(args[1])} ${metric.unit || ''}`;
                }).join('<br />');
            },
            axisPointer: {
                animation: false
            }
        },
        legend: { data: metrics.map(metric => metric.name) },
        xAxis: {
            type: 'time',
            splitLine: {
                show: false
            },
            axisLabel: {
                formatter: function (value) {
                    // 将时间戳转换为日期字符串
                    return dayjs(value*1000).format('HH:mm:ss');
                }
            },
        },
        yAxis: metrics.map((metric) => ({
            type: 'value',
            boundaryGap: [0, '100%'],
            splitLine: {
                show: false
            },
            name: `${metric.name} (单位：${metric.unit || ''}）`,
            ...(metric.unit === '%' ? {max: 100, min: 0} : {}),
        })),
        series: metrics.map((metric) => ({
            name: metric.name,
            type: 'line',
            showSymbol: false,
            data: ChartHistoryData.value.map(record => {
                const gpuInfo = record.list.find(item => item.device_id === GPUList.value[gpuIndex].device_id);
                return {
                    name: metric.name,
                    value: [record.timestamp, metric.proc(gpuInfo)],
                };
            })
        })),
    }
}
</script>

<template>
    <div class="monitor-container">
        <div v-for="(gpu, index) in GPUList" :key="index" class="gpu-panel">
            <h3>GPU {{ index + 1 }}: {{ gpu.name }}</h3>

            <div class="chart-container">
                <v-chart class="chart" :option="getChartOption(index, 'usaged')" :update-options="{ notMerge: true }" autoresize />
                <v-chart class="chart" :option="getChartOption(index, 'temperature')" :update-options="{ notMerge: true }" autoresize />
                <v-chart class="chart" :option="getChartOption(index, 'power_usage')" :update-options="{ notMerge: true }" autoresize />
            </div>

            <div class="processes">
                <div class="process" v-for="(proc, pIndex) in gpu.processes" :key="pIndex">
                    <span class="pid">{{ proc.pid }}</span>
                    <span class="name">{{ proc.name }}</span>
                    <span class="memory">{{ proc.memory_used }} MB</span>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.monitor-container {
    padding: 20px;
}

.gpu-panel {
    margin: 20px;
    padding: 20px;
    border: 1px solid #eee;
    border-radius: 8px;
}

.chart-container {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;
}

.chart {
    height: 300px;
}

.processes {
    margin-top: 20px;
}

.process {
    display: flex;
    padding: 8px;
    border-bottom: 1px solid #eee;
}

.pid {
    width: 80px;
}

.name {
    flex: 1;
}

.memory {
    width: 100px;
    text-align: right;
}
</style>