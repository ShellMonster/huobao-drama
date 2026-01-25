import enUS from 'element-plus/es/locale/lang/en'
import zhCN from 'element-plus/es/locale/lang/zh-cn'
import jaJP from 'element-plus/es/locale/lang/ja'
import koKR from 'element-plus/es/locale/lang/ko'

export type ElementLocale = typeof zhCN

export const getElementLocale = (lang: string): ElementLocale => {
  const normalized = lang.toLowerCase()
  if (normalized.startsWith('zh')) {
    return zhCN
  }
  if (normalized.startsWith('ja')) {
    return jaJP
  }
  if (normalized.startsWith('ko')) {
    return koKR
  }
  return enUS
}
