type CacheEntry<T> = {
  timestamp: number
  data: T
}

const safeParse = <T>(raw: string): CacheEntry<T> | null => {
  try {
    return JSON.parse(raw) as CacheEntry<T>
  } catch {
    return null
  }
}

export const getCache = <T>(key: string, maxAgeMs: number): T | null => {
  const raw = localStorage.getItem(key)
  if (!raw) return null
  const parsed = safeParse<T>(raw)
  if (!parsed || typeof parsed.timestamp !== 'number') {
    localStorage.removeItem(key)
    return null
  }
  if (maxAgeMs > 0 && Date.now() - parsed.timestamp > maxAgeMs) {
    localStorage.removeItem(key)
    return null
  }
  return parsed.data
}

export const setCache = <T>(key: string, data: T) => {
  const entry: CacheEntry<T> = {
    timestamp: Date.now(),
    data
  }
  localStorage.setItem(key, JSON.stringify(entry))
}

export const clearCache = (key: string) => {
  localStorage.removeItem(key)
}
