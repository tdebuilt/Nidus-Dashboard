import { writable } from 'svelte/store'
import { applyTheme, getTheme, darkTheme } from '../themes'
import type { ThemeDefinition } from '../themes'

function getInitialTheme(): string {
  if (typeof window === 'undefined') return 'dark'
  const stored = localStorage.getItem('nidus-theme')
  if (stored && getTheme(stored)) return stored
  return 'dark'
}

export const theme = writable<string>(getInitialTheme())

// Apply theme to DOM and persist
theme.subscribe((value) => {
  if (typeof document === 'undefined') return
  const def: ThemeDefinition = getTheme(value) ?? darkTheme
  applyTheme(def)
  localStorage.setItem('nidus-theme', value)
})

/** Set theme by ID */
export function setTheme(id: string) {
  if (getTheme(id)) theme.set(id)
}

/** Toggle between dark and light (for quick toggle button) */
export function toggleTheme() {
  theme.update((current) => {
    const currentDef = getTheme(current)
    return currentDef?.mode === 'dark' ? 'light' : 'dark'
  })
}
