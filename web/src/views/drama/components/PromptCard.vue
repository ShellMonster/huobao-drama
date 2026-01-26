<template>
  <el-card
    shadow="hover"
    class="prompt-card"
    :class="{ selected }"
    @click="emit('preview')"
  >
    <el-checkbox
      class="prompt-checkbox"
      :model-value="selected"
      @change="emit('select')"
      @click.stop
    />
    <div class="prompt-preview">
      <el-icon :size="24" class="preview-icon">
        <Document />
      </el-icon>
      <span class="preview-label">提示词</span>
    </div>
    <div class="prompt-body">
      <div class="prompt-text">{{ prompt }}</div>
    </div>
    <div class="prompt-tools">
      <el-button text size="small" @click.stop="emit('edit')">编辑</el-button>
      <el-button text size="small" @click.stop="emit('copy')">复制</el-button>
      <el-button text size="small" type="danger" @click.stop="emit('delete')">删除</el-button>
    </div>
    <el-button
      type="primary"
      size="small"
      class="prompt-generate"
      :loading="loading"
      :disabled="disabled"
      @click.stop="emit('generate')"
    >
      生成图
    </el-button>
  </el-card>
</template>

<script setup lang="ts">
import { Document } from '@element-plus/icons-vue'

withDefaults(defineProps<{
  prompt: string
  selected?: boolean
  loading?: boolean
  disabled?: boolean
}>(), {
  selected: false,
  loading: false,
  disabled: false
})

const emit = defineEmits<{
  (e: 'select'): void
  (e: 'preview'): void
  (e: 'edit'): void
  (e: 'copy'): void
  (e: 'delete'): void
  (e: 'generate'): void
}>()
</script>

<style scoped>
.prompt-card {
  position: relative;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
  cursor: pointer;
}

.prompt-card.selected {
  border-color: #409eff;
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.3);
}

.prompt-card :deep(.el-card__body) {
  padding: 12px;
}

.prompt-checkbox {
  position: absolute;
  top: 10px;
  right: 10px;
  z-index: 1;
}

.prompt-preview {
  height: 96px;
  border-radius: 8px;
  background: #f5f7fa;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  color: #909399;
}

.preview-label {
  font-size: 12px;
}

.prompt-body {
  min-height: 48px;
}

.prompt-text {
  font-size: 12px;
  color: #303133;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.prompt-tools {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 4px;
}

.prompt-generate {
  width: 100%;
}
</style>
