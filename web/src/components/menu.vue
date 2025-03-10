<script setup>
import { ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Setting, Odometer } from '@element-plus/icons-vue'

const menuItems = [
    {
        index: '/',
        title: '仪表盘',
        icon: Odometer,
    },
    {
        index: '/settings',
        title: '设置',
        icon: Setting,
    },
];
const router = useRouter()
const route = useRoute()
const activeIndex = ref(route.path)

watch(() => route.path, (newPath) => {
    activeIndex.value = newPath
})

const handleSelect = (index) => {
    router.push(index)
}

</script>

<template>
    <div class="wrapper">
        <el-menu :default-active="activeIndex" mode="horizontal" @select="handleSelect">
            <el-menu-item v-for="item in menuItems" :key="item.index" :index="item.index">
                <el-icon>
                    <component :is="item.icon" />
                </el-icon>
                {{ item.title }}
            </el-menu-item>
        </el-menu>
    </div>
</template>