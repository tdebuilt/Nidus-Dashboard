import { writable } from 'svelte/store'

export const shortcutHelpOpen = writable(false)

export function toggleShortcutHelp() {
  shortcutHelpOpen.update((v) => !v)
}

export function closeShortcutHelp() {
  shortcutHelpOpen.set(false)
}
