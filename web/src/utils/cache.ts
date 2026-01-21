type CacheEntry<T> = {
  timestamp: number
  data: T
}

const MAX_CACHE_BYTES = 512 * 1024
const CACHE_PREFIXES = ['drama:', 'storyboards:']

const safeParse = <T>(raw: string): CacheEntry<T> | null => {
  try {
    return JSON.parse(raw) as CacheEntry<T>
  } catch {
    return null
  }
}

export const getCache = <T>(key: string, maxAgeMs: number): T | null => {
  let raw: string | null = null
  try {
    raw = localStorage.getItem(key)
  } catch {
    return null
  }
  if (!raw) return null
  const parsed = safeParse<T>(raw)
  if (!parsed || typeof parsed.timestamp !== 'number') {
    try {
      localStorage.removeItem(key)
    } catch {
      return null
    }
    return null
  }
  if (maxAgeMs > 0 && Date.now() - parsed.timestamp > maxAgeMs) {
    try {
      localStorage.removeItem(key)
    } catch {
      return null
    }
    return null
  }
  return parsed.data
}

const isQuotaError = (error: unknown) => {
  if (!(error instanceof DOMException)) return false
  return error.name === 'QuotaExceededError' || error.name === 'NS_ERROR_DOM_QUOTA_REACHED'
}

const evictCacheEntries = () => {
  try {
    const keysToRemove: string[] = []
    for (let i = 0; i < localStorage.length; i += 1) {
      const key = localStorage.key(i)
      if (!key) continue
      if (CACHE_PREFIXES.some(prefix => key.startsWith(prefix))) {
        keysToRemove.push(key)
      }
    }
    keysToRemove.forEach(key => localStorage.removeItem(key))
  } catch {
    // ignore eviction errors
  }
}

export const setCache = <T>(key: string, data: T) => {
  const entry: CacheEntry<T> = {
    timestamp: Date.now(),
    data
  }
  const raw = JSON.stringify(entry)
  if (raw.length > MAX_CACHE_BYTES) {
    return
  }
  try {
    localStorage.setItem(key, raw)
  } catch (error) {
    if (isQuotaError(error)) {
      evictCacheEntries()
      try {
        localStorage.setItem(key, raw)
      } catch {
        // ignore secondary failures
      }
    }
  }
}

export const clearCache = (key: string) => {
  try {
    localStorage.removeItem(key)
  } catch {
    // ignore
  }
}
