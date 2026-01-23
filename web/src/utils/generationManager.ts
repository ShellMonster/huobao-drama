import { subscribeUnifiedSSE, type SSEParams } from '@/utils/sse'

type WatchCallbacks<T> = {
  onComplete: (payload: T) => Promise<void> | void
  onFailed?: (payload: T) => Promise<void> | void
}

type StatusWatcherOptions<T> = {
  types: string[]
  params?: SSEParams | (() => SSEParams)
  fetchById: (id: number) => Promise<T>
  getId: (payload: T) => number | null | undefined
  isCompleted: (payload: T) => boolean
  isFailed: (payload: T) => boolean
  onUpdate?: (payload: T) => void
  onFailedGlobal?: (payload: T) => void
  onTimeout?: (id: number) => void
  pollIntervalMs?: number
  timeoutMs?: number
}

type StatusWatcherState<T> = {
  done: boolean
  onComplete: (payload: T) => Promise<void> | void
  onFailed?: (payload: T) => Promise<void> | void
  resolve: () => void
  timeoutTimer: number | null
}

export const createStatusWatcher = <T>(options: StatusWatcherOptions<T>) => {
  const watchers = new Map<number, StatusWatcherState<T>>()
  let streamStop: (() => void) | null = null
  let pollTimer: number | null = null
  let pollInFlight = false
  const pollIntervalMs = options.pollIntervalMs ?? 6000
  const timeoutMs = options.timeoutMs ?? pollIntervalMs * 100

  const resolveParams = () =>
    typeof options.params === 'function' ? options.params() : options.params

  const stop = () => {
    if (streamStop) {
      streamStop()
      streamStop = null
    }
    if (pollTimer) {
      window.clearInterval(pollTimer)
      pollTimer = null
    }
    pollInFlight = false
    watchers.forEach((watcher) => {
      if (watcher.timeoutTimer) {
        window.clearTimeout(watcher.timeoutTimer)
      }
    })
    watchers.clear()
  }

  const checkEmpty = () => {
    if (watchers.size === 0) {
      stop()
    }
  }

  const handleUpdate = async (payload: T) => {
    const id = options.getId(payload)
    if (id === null || id === undefined) return
    const watcher = watchers.get(id)
    if (!watcher || watcher.done) return

    if (options.onUpdate) {
      options.onUpdate(payload)
    }

    if (options.isCompleted(payload)) {
      watcher.done = true
      if (watcher.timeoutTimer) {
        window.clearTimeout(watcher.timeoutTimer)
        watcher.timeoutTimer = null
      }
      watchers.delete(id)
      try {
        await watcher.onComplete(payload)
      } catch (error) {
        console.error('[status-watcher] 完成回调失败:', error)
      }
      watcher.resolve()
      checkEmpty()
      return
    }

    if (options.isFailed(payload)) {
      watcher.done = true
      if (watcher.timeoutTimer) {
        window.clearTimeout(watcher.timeoutTimer)
        watcher.timeoutTimer = null
      }
      watchers.delete(id)
      if (options.onFailedGlobal) {
        options.onFailedGlobal(payload)
      }
      if (watcher.onFailed) {
        try {
          await watcher.onFailed(payload)
        } catch (error) {
          console.error('[status-watcher] 失败回调失败:', error)
        }
      }
      watcher.resolve()
      checkEmpty()
    }
  }

  const pollOnce = async () => {
    if (pollInFlight) return
    const ids = Array.from(watchers.keys())
    if (ids.length === 0) {
      checkEmpty()
      return
    }
    pollInFlight = true
    const results = await Promise.allSettled(ids.map(id => options.fetchById(id)))
    results.forEach((result) => {
      if (result.status === 'fulfilled') {
        void handleUpdate(result.value)
      }
    })
    pollInFlight = false
    checkEmpty()
  }

  const startFallback = () => {
    if (pollTimer) return
    pollTimer = window.setInterval(() => {
      void pollOnce()
    }, pollIntervalMs)
    return () => {
      if (pollTimer) {
        window.clearInterval(pollTimer)
        pollTimer = null
      }
    }
  }

  const ensureStream = () => {
    if (streamStop || watchers.size === 0) return
    streamStop = subscribeUnifiedSSE<T>({
      types: options.types,
      params: resolveParams(),
      onMessage: (payload) => {
        void handleUpdate(payload)
      },
      fallback: startFallback
    }).close
  }

  const watch = (id: number, callbacks: WatchCallbacks<T>) => {
    return new Promise<void>((resolve) => {
      const existing = watchers.get(id)
      if (existing && existing.timeoutTimer) {
        window.clearTimeout(existing.timeoutTimer)
      }
      const timeoutTimer = window.setTimeout(() => {
        const watcher = watchers.get(id)
        if (!watcher || watcher.done) return
        watcher.done = true
        watcher.timeoutTimer = null
        watchers.delete(id)
        if (options.onTimeout) {
          options.onTimeout(id)
        }
        watcher.resolve()
        checkEmpty()
      }, timeoutMs)

      watchers.set(id, {
        done: false,
        onComplete: callbacks.onComplete,
        onFailed: callbacks.onFailed,
        resolve,
        timeoutTimer
      })

      ensureStream()
      void pollOnce()
    })
  }

  return {
    watch,
    stopAll: stop,
    hasActive: () => watchers.size > 0
  }
}

type ListStreamOptions<T> = {
  types: string[]
  getParams: () => SSEParams
  onMessage: (payload: T) => void
  poll: () => void
  shouldPoll: () => boolean
  pollIntervalMs?: number
  scheduleDelayMs?: number
}

export const createListStream = <T>(options: ListStreamOptions<T>) => {
  let streamStop: (() => void) | null = null
  let pollTimer: number | null = null
  let reloadTimer: number | null = null
  const pollIntervalMs = options.pollIntervalMs ?? 5000
  const scheduleDelayMs = options.scheduleDelayMs ?? 500

  const stop = () => {
    if (streamStop) {
      streamStop()
      streamStop = null
    }
    if (pollTimer) {
      window.clearInterval(pollTimer)
      pollTimer = null
    }
    if (reloadTimer) {
      window.clearTimeout(reloadTimer)
      reloadTimer = null
    }
  }

  const scheduleReload = () => {
    if (reloadTimer) return
    reloadTimer = window.setTimeout(() => {
      reloadTimer = null
      options.poll()
    }, scheduleDelayMs)
  }

  const startFallback = () => {
    if (pollTimer) return
    pollTimer = window.setInterval(() => {
      if (options.shouldPoll()) {
        options.poll()
      }
    }, pollIntervalMs)
    return () => {
      if (pollTimer) {
        window.clearInterval(pollTimer)
        pollTimer = null
      }
    }
  }

  const start = () => {
    stop()
    streamStop = subscribeUnifiedSSE<T>({
      types: options.types,
      params: options.getParams(),
      onMessage: options.onMessage,
      fallback: startFallback
    }).close
  }

  return {
    start,
    stop,
    scheduleReload
  }
}
