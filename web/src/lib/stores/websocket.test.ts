import { describe, it, expect, vi } from 'vitest'
import { get } from 'svelte/store'
import { ws } from './websocket'

describe('websocket store', () => {
  it('starts disconnected', () => {
    const state = get(ws)
    expect(state.status).toBe('disconnected')
    expect(state.reconnectAttempts).toBe(0)
  })

  it('subscribes and unsubscribes handlers', () => {
    const handler = vi.fn()
    const unsubscribe = ws.on('test', handler)
    expect(typeof unsubscribe).toBe('function')
    unsubscribe()
  })

  it('supports multiple handlers for same type', () => {
    const handler1 = vi.fn()
    const handler2 = vi.fn()
    const unsub1 = ws.on('multi', handler1)
    const unsub2 = ws.on('multi', handler2)
    unsub1()
    unsub2()
  })

  it('supports wildcard handlers', () => {
    const handler = vi.fn()
    const unsub = ws.on('*', handler)
    unsub()
  })

  it('disconnect sets status to disconnected', () => {
    ws.disconnect()
    const state = get(ws)
    expect(state.status).toBe('disconnected')
  })
})
