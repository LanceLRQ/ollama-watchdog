<script setup>
import { ElMessage } from 'element-plus';
import { DefaultServerSettings } from '@/models/server'
import { onMounted, reactive, ref } from 'vue'
import { saveSettingsToLocalStorage, loadSettingsFromLocalStorage } from '@/utils/config'

const settingsFormRef = ref()

const checkApiHost = (rule, value, callback) => {
    if (!value) {
        return callback(new Error('请输入服务器地址'))
    }
    const serverRegex = /^(localhost|([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}|(\d{1,3}\.){3}\d{1,3}):\d{1,5}$/;
    if (!serverRegex.test(value)) {
        return callback(new Error('请输入有效的服务器地址'))
    }
    callback()
}

const checkApiBase = (rule, value, callback) => {
    if (!value) {
        return callback(new Error('请输入API基础路径'))
    }
    const basePathRegex = /^\/[a-zA-Z0-9-_]+$/;
    if (!basePathRegex.test(value)) {
        return callback(new Error('请输入有效的API基础路径'))
    }
    callback()
}

const settingsForm = reactive({
    ...DefaultServerSettings
})

onMounted(() => {
    const settings = loadSettingsFromLocalStorage()
    console.log(settings)
    Object.keys(settings).forEach(key => {
        settingsForm[key] = settings[key]
    })
})

const rules = reactive({
    apiHost: [{ validator: checkApiHost, trigger: 'blur' }],
    apiBasePath: [{ validator: checkApiBase, trigger: 'blur' }],
})

const submitForm = (formEl) => {
    if (!formEl) return
    formEl.validate((valid) => {
        if (valid) {
            saveSettingsToLocalStorage(settingsForm);
            ElMessage.success('保存成功，刷新页面后生效');
        } else {
            ElMessage.error('表单填写不正确，无法保存')
        }
    })
}

const resetForm = (formEl) => {
    if (!formEl) return
    formEl.resetFields()
}
</script>

<template>
    <div class="settings-view">
        <el-card>
        <template #header>
            设置
        </template>
        <el-form ref="settingsFormRef" style="max-width: 600px" :model="settingsForm" status-icon :rules="rules"
            label-width="auto">
            <el-form-item label="服务器" prop="apiHost">
                <el-input v-model="settingsForm.apiHost" />
            </el-form-item>
            <el-form-item label="API路径" prop="apiBasePath">
                <el-input v-model="settingsForm.apiBasePath" />
            </el-form-item>
            <el-form-item>
                <el-button type="primary" @click="submitForm(settingsFormRef)">
                    保存
                </el-button>
            </el-form-item>
        </el-form>
    </el-card>
    </div>
</template>


<style scoped>
.settings-view {
    margin: 20px;
}
</style>