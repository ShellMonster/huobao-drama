<template>
  <el-dropdown @command="handleCommand">
    <span class="language-switcher">
      <el-icon><Switch /></el-icon>
      <span class="lang-text">{{ currentLangText }}</span>
    </span>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          v-for="lang in languageOptions"
          :key="lang.value"
          :command="lang.value"
          :disabled="currentLang === lang.value"
        >
          {{ lang.label }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLanguage } from '@/locales'
import { ElMessage, ElMessageBox } from 'element-plus'
import { settingsAPI } from '@/api/settings'

const { locale, t } = useI18n()

const currentLang = computed(() => locale.value as string)
const loading = ref(false)

const languageOptions = computed(() => [
  { value: 'zh-CN', label: `🇨🇳 ${t('settings.languages.zh')}`, backend: 'zh' },
  { value: 'en-US', label: `🇺🇸 ${t('settings.languages.en')}`, backend: 'en' },
  { value: 'ja-JP', label: `🇯🇵 ${t('settings.languages.ja')}`, backend: null },
  { value: 'ko-KR', label: `🇰🇷 ${t('settings.languages.ko')}`, backend: null }
])

const currentLangText = computed(() => {
  const current = languageOptions.value.find(item => item.value === currentLang.value)
  return current?.label || currentLang.value
})

const handleCommand = async (lang: string) => {
  if (loading.value) return
  
  const targetOption = languageOptions.value.find(option => option.value === lang)
  const backendLang = targetOption?.backend || null
  const shouldUpdateBackend = Boolean(backendLang)
  const languageLabel = targetOption?.label || lang
  const confirmMessage = t('settings.switchConfirmMessage', { language: languageLabel })
  
  try {
    if (shouldUpdateBackend) {
      await ElMessageBox.confirm(
        confirmMessage,
        t('settings.switchConfirmTitle'),
        {
          confirmButtonText: t('common.confirm'),
          cancelButtonText: t('common.cancel'),
          type: 'warning',
          dangerouslyUseHTMLString: false
        }
      )
    }

    loading.value = true
    
    let res: any
    if (shouldUpdateBackend && backendLang) {
      res = await settingsAPI.updateLanguage(backendLang)
      console.log('Backend language updated:', res)
    }
    
    // 更新前端语言
    setLanguage(lang)
    
    const message = res?.message || t('settings.switchSuccess', { language: languageLabel })
    ElMessage.success({ message, duration: 3000 })
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('Failed to switch language:', error)
      
      // 安全获取错误消息
      let errorMessage = t('common.unknownError')
      if (error?.message) {
        errorMessage = error.message
      } else if (error?.response?.data?.error?.message) {
        errorMessage = error.response.data.error.message
      } else if (typeof error === 'string') {
        errorMessage = error
      }
      
      ElMessage.error({
        message: t('settings.switchFailed', { error: errorMessage }),
        duration: 5000
      })
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.language-switcher {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 6px;
  transition: all 0.2s;
}

.language-switcher:hover {
  background-color: rgba(0, 0, 0, 0.05);
}

.lang-text {
  font-size: 14px;
  color: #606266;
}
</style>
