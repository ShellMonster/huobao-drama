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

export const buildUnifiedEventUrl = (params?: SSEParams) =>
  buildSSEUrl('/api/v1/events/stream', params)

export const subscribeUnifiedSSE = <T>(options: {
  params?: SSEParams
  types: string[]
  onMessage: (payload: T, type: string) => void
  onOpen?: () => void
  onError?: (err: Event) => void
  fallback?: FallbackStarter
  timeoutMs?: number
}) => {
  const { params, types, onMessage, onOpen, onError, fallback, timeoutMs } = options
  const url = buildUnifiedEventUrl({
    ...(params || {}),
    types: types.join(',')
  })
  return subscribeSSE<{
    type: string
    data: T
  }>({
    url,
    event: 'event',
    onMessage: (payload) => {
      onMessage(payload.data, payload.type)
    },
    onOpen,
    onError,
    fallback,
    timeoutMs
  })
}

type SSESubscriber<T> = {
  id: number
  event: string
  onMessage: (payload: T) => void
  onOpen?: () => void
  onError?: (err: Event) => void
  fallback?: FallbackStarter
  fallbackStarted: boolean
  fallbackCleanup: (() => void) | null
}

type SSEConnection = {
  url: string
  source: EventSource | null
  opened: boolean
  needsFallback: boolean
  openTimeout: number | null
  timeoutMs: number
  subscribers: Map<number, SSESubscriber<any>>
  eventHandlers: Map<string, (evt: MessageEvent) => void>
}

const connections = new Map<string, SSEConnection>()
let nextSubscriberId = 1

const startSubscriberFallback = (subscriber: SSESubscriber<any>) => {
  if (subscriber.fallbackStarted) return
  subscriber.fallbackStarted = true
  if (subscriber.fallback) {
    const cleanup = subscriber.fallback()
    if (typeof cleanup === 'function') {
      subscriber.fallbackCleanup = cleanup
    }
  }
}

const stopSubscriberFallback = (subscriber: SSESubscriber<any>) => {
  if (subscriber.fallbackCleanup) {
    subscriber.fallbackCleanup()
    subscriber.fallbackCleanup = null
  }
  subscriber.fallbackStarted = false
}

const startConnectionFallback = (connection: SSEConnection) => {
  connection.subscribers.forEach((subscriber) => {
    if (typeof subscriber.fallback === 'function') {
      startSubscriberFallback(subscriber)
    }
  })
}

const stopConnectionFallback = (connection: SSEConnection) => {
  connection.subscribers.forEach(stopSubscriberFallback)
}

const scheduleOpenTimeout = (connection: SSEConnection) => {
  if (connection.openTimeout !== null) {
    window.clearTimeout(connection.openTimeout)
    connection.openTimeout = null
  }
  connection.openTimeout = window.setTimeout(() => {
    if (!connection.opened) {
      connection.needsFallback = true
      startConnectionFallback(connection)
    }
  }, connection.timeoutMs)
}

const ensureConnection = (url: string, timeoutMs: number) => {
  const existing = connections.get(url)
  if (existing) {
    if (!existing.opened && timeoutMs < existing.timeoutMs) {
      existing.timeoutMs = timeoutMs
      if (existing.source) {
        scheduleOpenTimeout(existing)
      }
    }
    return existing
  }

  const connection: SSEConnection = {
    url,
    source: null,
    opened: false,
    needsFallback: false,
    openTimeout: null,
    timeoutMs,
    subscribers: new Map(),
    eventHandlers: new Map()
  }

  connections.set(url, connection)

  if (typeof EventSource === 'undefined') {
    connection.needsFallback = true
    return connection
  }

  connection.source = new EventSource(url)
  scheduleOpenTimeout(connection)

  connection.source.onopen = () => {
    connection.opened = true
    connection.needsFallback = false
    if (connection.openTimeout !== null) {
      window.clearTimeout(connection.openTimeout)
      connection.openTimeout = null
    }
    stopConnectionFallback(connection)
    connection.subscribers.forEach((subscriber) => {
      if (subscriber.onOpen) {
        subscriber.onOpen()
      }
    })
  }

  connection.source.onerror = (err) => {
    connection.opened = false
    connection.needsFallback = true
    connection.subscribers.forEach((subscriber) => {
      if (subscriber.onError) {
        subscriber.onError(err)
      }
    })
    startConnectionFallback(connection)
  }

  return connection
}

const cleanupConnection = (connection: SSEConnection) => {
  if (connection.openTimeout !== null) {
    window.clearTimeout(connection.openTimeout)
    connection.openTimeout = null
  }
  connection.subscribers.forEach(stopSubscriberFallback)
  connection.subscribers.clear()
  connection.eventHandlers.forEach((handler, event) => {
    connection.source?.removeEventListener(event, handler)
  })
  connection.eventHandlers.clear()
  if (connection.source) {
    connection.source.close()
    connection.source = null
  }
  connections.delete(connection.url)
}

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
  const connection = ensureConnection(url, timeoutMs)
  const subscriberId = nextSubscriberId++
  const subscriber: SSESubscriber<T> = {
    id: subscriberId,
    event,
    onMessage,
    onOpen,
    onError,
    fallback,
    fallbackStarted: false,
    fallbackCleanup: null
  }

  connection.subscribers.set(subscriberId, subscriber)

  if (connection.source && !connection.eventHandlers.has(event)) {
    const handler = (evt: MessageEvent) => {
      let payload: T
      try {
        payload = JSON.parse(evt.data) as T
      } catch {
        return
      }
      connection.subscribers.forEach((sub) => {
        if (sub.event === event) {
          sub.onMessage(payload)
        }
      })
    }
    connection.eventHandlers.set(event, handler)
    connection.source.addEventListener(event, handler)
  }

  if (connection.opened && subscriber.onOpen) {
    subscriber.onOpen()
  }
  if (connection.needsFallback) {
    startSubscriberFallback(subscriber)
  }

  const close = () => {
    const current = connection.subscribers.get(subscriberId)
    if (!current) return
    connection.subscribers.delete(subscriberId)

    if (connection.source && connection.eventHandlers.has(event)) {
      const hasOtherSubscribers = Array.from(connection.subscribers.values()).some(
        (sub) => sub.event === event
      )
      if (!hasOtherSubscribers) {
        const handler = connection.eventHandlers.get(event)
        if (handler) {
          connection.source.removeEventListener(event, handler)
        }
        connection.eventHandlers.delete(event)
      }
    }

    stopSubscriberFallback(subscriber)

    if (connection.subscribers.size === 0) {
      cleanupConnection(connection)
    }
  }

  return { close }
}
