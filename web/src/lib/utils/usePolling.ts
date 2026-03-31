import { untrack } from 'svelte'
import { pollingInterval as defaultPollingStore } from '../stores/polling'

type Subscribable = {
  subscribe: (cb: (ms: number) => void) => () => void
  get: () => number
}

export interface UsePollingOptions {
  fetchFn: () => Promise<void>
  active: () => boolean
  pollingStore?: Subscribable
  intervalTransform?: (ms: number) => number
  fixedIntervalMs?: number
}

export function usePolling(opts: UsePollingOptions) {
  const store = opts.pollingStore ?? defaultPollingStore
  const transform = opts.intervalTransform ?? ((ms: number) => ms)

  let interval: ReturnType<typeof setInterval> | null = null
  let unsub: (() => void) | null = null
  let running = false

  function start() {
    if (running) return
    running = true

    untrack(() => { opts.fetchFn() })

    if (opts.fixedIntervalMs != null) {
      interval = setInterval(opts.fetchFn, opts.fixedIntervalMs)
    } else {
      const ms = transform(store.get())
      interval = setInterval(opts.fetchFn, ms)

      unsub = store.subscribe((newMs: number) => {
        if (interval) clearInterval(interval)
        interval = setInterval(opts.fetchFn, transform(newMs))
      })
    }
  }

  function stop() {
    if (!running) return
    running = false
    if (interval) {
      clearInterval(interval)
      interval = null
    }
    if (unsub) {
      unsub()
      unsub = null
    }
  }

  return { start, stop }
}
