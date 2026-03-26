import { describe, it, expect, beforeEach } from 'vitest'
import { get } from 'svelte/store'
import { keyboardShortcutsEnabled, setKeyboardShortcutsEnabled } from './keyboardShortcuts'

describe('Keyboard shortcuts store', () => {
  beforeEach(() => {
    keyboardShortcutsEnabled.set(true)
  })

  it('defaults to enabled', () => {
    expect(get(keyboardShortcutsEnabled)).toBe(true)
  })

  it('can be disabled', () => {
    setKeyboardShortcutsEnabled(false)
    expect(get(keyboardShortcutsEnabled)).toBe(false)
  })

  it('can be re-enabled', () => {
    setKeyboardShortcutsEnabled(false)
    setKeyboardShortcutsEnabled(true)
    expect(get(keyboardShortcutsEnabled)).toBe(true)
  })
})
