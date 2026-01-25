import { getCurrentLanguage } from '@/locales'

export const formatDateTime = (value?: string | number | Date): string => {
  if (!value) return ''
  const date = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const lang = getCurrentLanguage() as string
  return new Intl.DateTimeFormat(lang, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}
