<template>
  <el-icon class="loading-icon" :size="size">
    <Loading />
  </el-icon>
</template>

<script setup lang="ts">
import { Loading } from '@element-plus/icons-vue'

defineProps<{
  size?: number
}>()
</script>

<style scoped>
.loading-icon {
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
