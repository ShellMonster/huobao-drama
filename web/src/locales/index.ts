import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN'
import enUS from './en-US'
import jaJP from './ja-JP'
import koKR from './ko-KR'

const isPlainObject = (value: unknown): value is Record<string, any> => {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value)
}

const deepMerge = <T extends Record<string, any>>(base: T, override: Record<string, any>) => {
  const result: Record<string, any> = { ...base }
  Object.keys(override).forEach(key => {
    const baseValue = result[key]
    const overrideValue = override[key]
    if (isPlainObject(baseValue) && isPlainObject(overrideValue)) {
      result[key] = deepMerge(baseValue, overrideValue)
      return
    }
    result[key] = overrideValue
  })
  return result as T
}

// 从 localStorage 获取保存的语言，默认为中文
const getStoredLanguage = (): string => {
  const stored = localStorage.getItem('language')
  if (stored) return stored
  
  // 自动检测浏览器语言
  const browserLang = navigator.language.toLowerCase()
  if (browserLang.startsWith('zh')) return 'zh-CN'
  if (browserLang.startsWith('ja')) return 'ja-JP'
  if (browserLang.startsWith('ko')) return 'ko-KR'
  return 'en-US'
}

const i18n = createI18n({
  legacy: false, // 使用 Composition API 模式
  locale: getStoredLanguage(),
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS,
    'ja-JP': deepMerge(enUS, jaJP),
    'ko-KR': deepMerge(enUS, koKR)
  }
})

export default i18n

// 导出语言切换函数
export const setLanguage = (lang: string) => {
  i18n.global.locale.value = lang as any
  localStorage.setItem('language', lang)
}

export const getCurrentLanguage = () => {
  return i18n.global.locale.value
}
