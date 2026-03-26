import { describe, it, expect, vi, afterEach } from 'vitest'
import { get } from 'svelte/store'
import { toasts } from './toast'

describe('Toast store', () => {
  afterEach(() => {
    // Clear all toasts
    const current = get(toasts)
    current.forEach((t) => toasts.remove(t.id))
  })

  it('starts empty', () => {
    expect(get(toasts)).toEqual([])
  })

  it('adds a success toast', () => {
    toasts.success('Done!', 0)
    const all = get(toasts)
    expect(all.length).toBe(1)
    expect(all[0].type).toBe('success')
    expect(all[0].message).toBe('Done!')
  })

  it('adds an error toast', () => {
    toasts.error('Failed!', 0)
    const all = get(toasts)
    expect(all.length).toBe(1)
    expect(all[0].type).toBe('error')
  })

  it('adds an info toast', () => {
    toasts.info('Info', 0)
    const all = get(toasts)
    expect(all.length).toBe(1)
    expect(all[0].type).toBe('info')
  })

  it('removes a toast by id', () => {
    const id = toasts.success('Test', 0)
    expect(get(toasts).length).toBe(1)
    toasts.remove(id)
    expect(get(toasts).length).toBe(0)
  })

  it('auto-removes after duration', async () => {
    vi.useFakeTimers()
    toasts.success('Auto', 100)
    expect(get(toasts).length).toBe(1)
    vi.advanceTimersByTime(150)
    expect(get(toasts).length).toBe(0)
    vi.useRealTimers()
  })

  it('supports multiple toasts', () => {
    toasts.success('A', 0)
    toasts.error('B', 0)
    toasts.info('C', 0)
    expect(get(toasts).length).toBe(3)
  })
})
