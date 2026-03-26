import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { usePolling } from './usePolling'

function createMockStore(initialValue = 30000) {
  let value = initialValue
  const subscribers: ((ms: number) => void)[] = []
  return {
    get: () => value,
    subscribe: (cb: (ms: number) => void) => {
      subscribers.push(cb)
      return () => {
        const idx = subscribers.indexOf(cb)
        if (idx >= 0) subscribers.splice(idx, 1)
      }
    },
    set: (ms: number) => {
      value = ms
      subscribers.forEach((cb) => cb(ms))
    },
    _subscribers: subscribers,
  }
}

describe('usePolling', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('calls fetchFn immediately on start', () => {
    const fetchFn = vi.fn().mockResolvedValue(undefined)
    const store = createMockStore()
    const polling = usePolling({
      fetchFn,
      active: () => true,
      pollingStore: store,
    })
    polling.start()
    expect(fetchFn).toHaveBeenCalledTimes(1)
    polling.stop()
  })

  it('polls at the store interval', () => {
    const fetchFn = vi.fn().mockResolvedValue(undefined)
    const store = createMockStore(5000)
    const polling = usePolling({
      fetchFn,
      active: () => true,
      pollingStore: store,
    })
    polling.start()
    expect(fetchFn).toHaveBeenCalledTimes(1)

    vi.advanceTimersByTime(5000)
    expect(fetchFn).toHaveBeenCalledTimes(2)

    vi.advanceTimersByTime(5000)
    expect(fetchFn).toHaveBeenCalledTimes(3)

    polling.stop()
  })

  it('stops polling on stop()', () => {
    const fetchFn = vi.fn().mockResolvedValue(undefined)
    const store = createMockStore(5000)
    const polling = usePolling({
      fetchFn,
      active: () => true,
      pollingStore: store,
    })
    polling.start()
    expect(fetchFn).toHaveBeenCalledTimes(1)

    polling.stop()
    vi.advanceTimersByTime(10000)
    expect(fetchFn).toHaveBeenCalledTimes(1)
  })

  it('unsubscribes from store on stop', () => {
    const fetchFn = vi.fn().mockResolvedValue(undefined)
    const store = createMockStore(5000)
    const polling = usePolling({
      fetchFn,
      active: () => true,
      pollingStore: store,
    })
    polling.start()
    expect(store._subscribers).toHaveLength(1)

    polling.stop()
    expect(store._subscribers).toHaveLength(0)
  })

  it('applies intervalTransform', () => {
    const fetchFn = vi.fn().mockResolvedValue(undefined)
    const store = createMockStore(10000)
    const polling = usePolling({
      fetchFn,
      active: () => true,
      pollingStore: store,
      intervalTransform: (ms) => ms * 2,
    })
    polling.start()
    expect(fetchFn).toHaveBeenCalledTimes(1)

    // With transform, interval should be 20000ms
    vi.advanceTimersByTime(10000)
    expect(fetchFn).toHaveBeenCalledTimes(1)

    vi.advanceTimersByTime(10000)
    expect(fetchFn).toHaveBeenCalledTimes(2)

    polling.stop()
  })

  it('re-adjusts interval when store changes', () => {
    const fetchFn = vi.fn().mockResolvedValue(undefined)
    const store = createMockStore(10000)
    const polling = usePolling({
      fetchFn,
      active: () => true,
      pollingStore: store,
    })
    polling.start()
    expect(fetchFn).toHaveBeenCalledTimes(1)

    // Change interval to 2000ms
    store.set(2000)

    vi.advanceTimersByTime(2000)
    expect(fetchFn).toHaveBeenCalledTimes(2)

    polling.stop()
  })

  it('uses fixedIntervalMs instead of store', () => {
    const fetchFn = vi.fn().mockResolvedValue(undefined)
    const store = createMockStore(5000)
    const polling = usePolling({
      fetchFn,
      active: () => true,
      pollingStore: store,
      fixedIntervalMs: 60000,
    })
    polling.start()
    expect(fetchFn).toHaveBeenCalledTimes(1)

    // Should NOT fire at store interval
    vi.advanceTimersByTime(5000)
    expect(fetchFn).toHaveBeenCalledTimes(1)

    // Should fire at fixed interval
    vi.advanceTimersByTime(55000)
    expect(fetchFn).toHaveBeenCalledTimes(2)

    // Store subscribers should not be set up
    expect(store._subscribers).toHaveLength(0)

    polling.stop()
  })

  it('does not double-start', () => {
    const fetchFn = vi.fn().mockResolvedValue(undefined)
    const store = createMockStore(5000)
    const polling = usePolling({
      fetchFn,
      active: () => true,
      pollingStore: store,
    })
    polling.start()
    polling.start()
    expect(fetchFn).toHaveBeenCalledTimes(1)
    expect(store._subscribers).toHaveLength(1)

    polling.stop()
  })

  it('can restart after stop', () => {
    const fetchFn = vi.fn().mockResolvedValue(undefined)
    const store = createMockStore(5000)
    const polling = usePolling({
      fetchFn,
      active: () => true,
      pollingStore: store,
    })
    polling.start()
    expect(fetchFn).toHaveBeenCalledTimes(1)

    polling.stop()
    polling.start()
    expect(fetchFn).toHaveBeenCalledTimes(2)

    vi.advanceTimersByTime(5000)
    expect(fetchFn).toHaveBeenCalledTimes(3)

    polling.stop()
  })
})
