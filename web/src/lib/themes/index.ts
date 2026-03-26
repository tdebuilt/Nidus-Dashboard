/**
 * Theme engine — loads, validates, and applies themes as CSS custom properties.
 */

import type { ThemeColors, ThemeDefinition } from './themes'
import { builtinThemes, parseThemeJSON } from './themes'

export { type ThemeColors, type ThemeDefinition, builtinThemes, getAvailableThemes, darkTheme, lightTheme, nordTheme, draculaTheme, themeColorKeys, parseThemeJSON } from './themes'

/** Registry of all themes (built-in + custom) */
const themeRegistry = new Map<string, ThemeDefinition>(
  Object.entries(builtinThemes)
)

/** Register a custom theme */
export function registerTheme(theme: ThemeDefinition): void {
  themeRegistry.set(theme.id, theme)
}

/** Get a theme by ID from the full registry (built-in + custom) */
export function getTheme(id: string): ThemeDefinition | undefined {
  return themeRegistry.get(id)
}

/** Get all registered themes (built-in + custom) */
export function getAllThemes(): ThemeDefinition[] {
  return Array.from(themeRegistry.values())
}

/** Remove a custom theme from the registry */
export function unregisterTheme(id: string): boolean {
  return themeRegistry.delete(id)
}

/**
 * Apply a theme's CSS variables to the document root.
 * Also sets the `dark` / `light` class for any CSS that depends on it.
 */
export function applyTheme(theme: ThemeDefinition): void {
  if (typeof document === 'undefined') return
  const root = document.documentElement

  // Set CSS custom properties
  for (const [key, value] of Object.entries(theme.colors) as [keyof ThemeColors, string][]) {
    root.style.setProperty(`--${key}`, value)
  }

  // Toggle mode class
  root.classList.toggle('dark', theme.mode === 'dark')
  root.classList.toggle('light', theme.mode === 'light')
}

/**
 * Load a theme from a JSON object or string, validate it, register it,
 * and optionally apply it immediately.
 * Returns the parsed ThemeDefinition on success, or an error string.
 */
export function loadThemeFromJSON(
  json: unknown,
  options: { apply?: boolean } = {}
): ThemeDefinition | string {
  // Parse string input
  let parsed = json
  if (typeof json === 'string') {
    try {
      parsed = JSON.parse(json)
    } catch {
      return 'Invalid JSON string'
    }
  }

  const result = parseThemeJSON(parsed)
  if (typeof result === 'string') return result

  // Register in the theme registry
  registerTheme(result)

  // Apply if requested
  if (options.apply) applyTheme(result)

  return result
}

/**
 * Clear all theme CSS variables from inline styles.
 * Useful when switching back to CSS-only defaults.
 */
export function clearThemeStyles(): void {
  if (typeof document === 'undefined') return
  const root = document.documentElement
  const style = root.style
  for (let i = style.length - 1; i >= 0; i--) {
    const prop = style[i]
    if (prop.startsWith('--color-')) {
      style.removeProperty(prop)
    }
  }
}
