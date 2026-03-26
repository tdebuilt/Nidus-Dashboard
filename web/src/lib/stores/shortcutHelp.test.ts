import { describe, it, expect, beforeEach } from 'vitest'
import { get } from 'svelte/store'
import { shortcutHelpOpen, toggleShortcutHelp, closeShortcutHelp } from './shortcutHelp'

describe('Shortcut help store', () => {
  beforeEach(() => {
    shortcutHelpOpen.set(false)
  })

  it('starts closed', () => {
    expect(get(shortcutHelpOpen)).toBe(false)
  })

  it('toggles open', () => {
    toggleShortcutHelp()
    expect(get(shortcutHelpOpen)).toBe(true)
  })

  it('toggles closed', () => {
    toggleShortcutHelp()
    toggleShortcutHelp()
    expect(get(shortcutHelpOpen)).toBe(false)
  })

  it('closes explicitly', () => {
    shortcutHelpOpen.set(true)
    closeShortcutHelp()
    expect(get(shortcutHelpOpen)).toBe(false)
  })
})
