import { writable } from 'svelte/store'

export const keyboardShortcutsEnabled = writable(true)

export function setKeyboardShortcutsEnabled(enabled: boolean) {
  keyboardShortcutsEnabled.set(enabled)
}
