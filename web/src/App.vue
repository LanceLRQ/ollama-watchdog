<script setup>
import { RouterView } from 'vue-router'
import AppMenu from './components/menu.vue'
</script>

<template>
  <header>
    <app-menu></app-menu>
  </header>
  <RouterView />
</template>

<style scoped></style>
