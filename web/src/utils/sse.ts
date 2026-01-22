export type SSEParams = Record<
  string,
  string | number | boolean | Array<string | number | boolean> | null | undefined
>

export const buildSSEUrl = (path: string, params?: SSEParams) => {
  if (!params) return path
  const search = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value === null || value === undefined || value === '') return
    if (Array.isArray(value)) {
      const filtered = value
        .map(item => String(item))
        .filter(item => item.trim().length > 0)
      if (filtered.length > 0) {
        search.set(key, filtered.join(','))
      }
      return
    }
    search.set(key, String(value))
  })
  const query = search.toString()
  return query ? `${path}?${query}` : path
}

type FallbackStarter = () => void | (() => void)

export const subscribeSSE = <T>(options: {
  url: string
  event: string
  onMessage: (payload: T) => void
  onOpen?: () => void
  onError?: (err: Event) => void
  fallback?: FallbackStarter
  timeoutMs?: number
}) => {
  const { url, event, onMessage, onOpen, onError, fallback, timeoutMs = 12000 } = options
  let source: EventSource | null = null
  let opened = false
  let fallbackCleanup: (() => void) | null = null
  let fallbackStarted = false

  const startFallback = () => {
    if (fallbackStarted) return
    fallbackStarted = true
    if (fallback) {
      const cleanup = fallback()
      if (typeof cleanup === 'function') {
        fallbackCleanup = cleanup
      }
    }
  }

  const stopFallback = () => {
    if (fallbackCleanup) {
      fallbackCleanup()
      fallbackCleanup = null
    }
    fallbackStarted = false
  }

  const closeSource = () => {
    if (source) {
      source.close()
      source = null
    }
  }

  const cleanup = () => {
    closeSource()
    stopFallback()
  }

  if (typeof EventSource === 'undefined') {
    startFallback()
    return { close: cleanup }
  }

  source = new EventSource(url)

  const openTimeout = window.setTimeout(() => {
    if (!opened) {
      startFallback()
    }
  }, timeoutMs)

  source.onopen = () => {
    opened = true
    window.clearTimeout(openTimeout)
    stopFallback()
    if (onOpen) onOpen()
  }

  source.onerror = (err) => {
    if (onError) onError(err)
    if (!fallbackStarted) {
      startFallback()
    }
  }

  source.addEventListener(event, (evt) => {
    try {
      const data = JSON.parse((evt as MessageEvent).data) as T
      onMessage(data)
    } catch {
      // ignore invalid payload
    }
  })

  return { close: cleanup }
}
