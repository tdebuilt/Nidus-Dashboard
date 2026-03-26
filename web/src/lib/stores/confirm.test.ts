import { describe, it, expect, afterEach } from 'vitest'
import { get } from 'svelte/store'
import { confirmState, confirm, resolveConfirm } from './confirm'

describe('Confirm store', () => {
  afterEach(() => {
    // Reset state
    confirmState.set({ open: false, options: { title: '', message: '' }, resolve: null })
  })

  it('starts closed', () => {
    const state = get(confirmState)
    expect(state.open).toBe(false)
    expect(state.resolve).toBeNull()
  })

  it('confirm() opens dialog with options', () => {
    confirm({ title: 'Test', message: 'Are you sure?' })
    const state = get(confirmState)
    expect(state.open).toBe(true)
    expect(state.options.title).toBe('Test')
    expect(state.options.message).toBe('Are you sure?')
    expect(state.resolve).toBeTypeOf('function')
  })

  it('resolveConfirm(true) resolves promise with true', async () => {
    const result = confirm({ title: 'T', message: 'M' })
    resolveConfirm(true)
    expect(await result).toBe(true)
    expect(get(confirmState).open).toBe(false)
  })

  it('resolveConfirm(false) resolves promise with false', async () => {
    const result = confirm({ title: 'T', message: 'M' })
    resolveConfirm(false)
    expect(await result).toBe(false)
    expect(get(confirmState).open).toBe(false)
  })

  it('supports destructive option', () => {
    confirm({ title: 'Delete', message: 'Sure?', destructive: true, confirmLabel: 'Delete' })
    const state = get(confirmState)
    expect(state.options.destructive).toBe(true)
    expect(state.options.confirmLabel).toBe('Delete')
  })
})
