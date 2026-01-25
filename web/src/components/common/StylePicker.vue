<template>
  <div class="style-grid" :style="gridStyle">
    <button
      v-for="style in styles"
      :key="style.key"
      type="button"
      class="style-card"
      :class="{ selected: currentValue === style.key }"
      :aria-pressed="currentValue === style.key"
      @click="selectStyle(style.key)"
    >
      <div class="style-thumb">
        <img
          v-if="style.preview_url && !failedImages[style.key]"
          :src="style.preview_url"
          :alt="style.name"
          @error="markFailed(style.key)"
        />
        <div v-else class="style-thumb-placeholder">
          <span>{{ $t('style.previewPlaceholder') }}</span>
        </div>
        <div class="style-label">
          <span>{{ style.name }}</span>
        </div>
      </div>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import type { StyleOption } from '@/types/style'

const props = defineProps<{ modelValue?: string; styles: StyleOption[]; columns?: number }>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const { t: $t } = useI18n()

const failedImages = reactive<Record<string, boolean>>({})

const currentValue = computed(() => props.modelValue || '')
const gridStyle = computed(() => {
  const columns = props.columns || 3
  const maxWidth = columns * 120 + (columns - 1) * 12
  return {
    '--style-columns': String(columns),
    '--style-grid-max': `${maxWidth}px`
  }
})

const selectStyle = (key: string) => {
  emit('update:modelValue', key)
}

const markFailed = (key: string) => {
  failedImages[key] = true
}
</script>

<style scoped>
.style-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 12px;
  width: 100%;
  max-width: var(--style-grid-max, 100%);
  justify-content: start;
}

.style-card {
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  padding: 0;
  background: var(--bg-secondary);
  cursor: pointer;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast), transform var(--transition-fast);
  width: 100%;
}

.style-card:hover {
  border-color: var(--border-secondary);
  transform: translateY(-2px);
}

.style-card.selected {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px var(--accent);
}

.style-card:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.style-thumb {
  position: relative;
  width: 100%;
  aspect-ratio: 1 / 1;
  border-radius: var(--radius-lg);
  overflow: hidden;
  background: linear-gradient(135deg, rgba(231, 236, 242, 0.9), rgba(205, 214, 226, 0.9));
}

.style-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.style-thumb-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #5f6b7a;
  font-size: 12px;
  letter-spacing: 0.4px;
}

.style-label {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 8px 10px;
  background: rgba(0, 0, 0, 0.35);
  backdrop-filter: blur(8px);
  color: #fff;
  font-size: 12px;
  font-weight: 500;
  text-shadow: 0 2px 6px rgba(0, 0, 0, 0.4);
}
</style>
