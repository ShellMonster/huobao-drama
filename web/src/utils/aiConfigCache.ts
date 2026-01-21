import type { AIServiceConfig, AIServiceType } from '@/types/ai'

type CacheEntry = {
  timestamp: number
  data: AIServiceConfig[]
}

const cache = new Map<AIServiceType, CacheEntry>()

export const getAIConfigCache = (serviceType: AIServiceType, maxAgeMs: number) => {
  const entry = cache.get(serviceType)
  if (!entry) return null
  if (maxAgeMs > 0 && Date.now() - entry.timestamp > maxAgeMs) {
    cache.delete(serviceType)
    return null
  }
  return entry.data
}

export const setAIConfigCache = (serviceType: AIServiceType, data: AIServiceConfig[]) => {
  cache.set(serviceType, { timestamp: Date.now(), data })
}

export const clearAIConfigCache = (serviceType?: AIServiceType) => {
  if (serviceType) {
    cache.delete(serviceType)
    return
  }
  cache.clear()
}
