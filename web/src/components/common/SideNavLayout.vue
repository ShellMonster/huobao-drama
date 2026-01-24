<template>
  <div class="side-layout">
    <aside class="side-nav">
      <div class="side-brand">
        <span class="brand-mark">🎬</span>
        <span class="brand-text">HuoBao Drama</span>
      </div>
      <el-menu
        class="side-menu"
        :default-active="activeMenu"
        :router="true"
      >
        <el-menu-item index="/">
          <el-icon><Document /></el-icon>
          <span>项目管理</span>
        </el-menu-item>
        <el-menu-item index="/settings/styles">
          <el-icon><Brush /></el-icon>
          <span>风格管理</span>
        </el-menu-item>
        <el-menu-item index="/settings/brands">
          <el-icon><Setting /></el-icon>
          <span>品牌管理</span>
        </el-menu-item>
      </el-menu>
    </aside>
    <div class="side-content">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Brush, Document, Setting } from '@element-plus/icons-vue'

const route = useRoute()

const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/settings/styles')) return '/settings/styles'
  if (path.startsWith('/settings/brands')) return '/settings/brands'
  return '/'
})
</script>

<style scoped>
.side-layout {
  min-height: 100vh;
  display: flex;
  background: var(--bg);
}

.side-nav {
  width: 220px;
  background: var(--bg-card);
  border-right: 1px solid var(--border-primary);
  padding: 0 var(--space-3) var(--space-4);
  position: sticky;
  top: 0;
  height: 100vh;
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.side-brand {
  height: 70px;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: 0 var(--space-2);
  border-bottom: 1px solid var(--border-primary);
  font-weight: 700;
  color: var(--text-primary);
}

.brand-mark {
  font-size: 1.1rem;
  line-height: 1;
}

.brand-text {
  font-size: 1rem;
  letter-spacing: 0.2px;
}

.side-menu {
  border-right: none;
  background: transparent;
}

.side-menu :deep(.el-menu-item) {
  border-radius: var(--radius-lg);
  height: 44px;
  line-height: 44px;
  margin-bottom: 6px;
}

.side-menu :deep(.el-menu-item.is-active) {
  background: var(--accent-light);
  color: var(--accent);
}

.side-content {
  flex: 1;
  min-width: 0;
  background: var(--bg);
}
</style>
