import { describe, it, expect, afterEach } from 'vitest'
import { applyTheme, clearThemeStyles, darkTheme, lightTheme, getTheme, getAllThemes, registerTheme, loadThemeFromJSON } from './index'
import type { ThemeDefinition } from './index'

describe('Theme engine', () => {
  afterEach(() => {
    clearThemeStyles()
    document.documentElement.classList.remove('dark', 'light')
  })

  it('applyTheme sets CSS custom properties on :root', () => {
    applyTheme(darkTheme)
    const root = document.documentElement
    expect(root.style.getPropertyValue('--color-bg')).toBe('#0f172a')
    expect(root.style.getPropertyValue('--color-text')).toBe('#f1f5f9')
    expect(root.style.getPropertyValue('--color-primary')).toBe('#3b82f6')
  })

  it('applyTheme sets dark class for dark theme', () => {
    applyTheme(darkTheme)
    expect(document.documentElement.classList.contains('dark')).toBe(true)
    expect(document.documentElement.classList.contains('light')).toBe(false)
  })

  it('applyTheme sets light class for light theme', () => {
    applyTheme(lightTheme)
    expect(document.documentElement.classList.contains('light')).toBe(true)
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('applyTheme switches from dark to light correctly', () => {
    applyTheme(darkTheme)
    expect(document.documentElement.style.getPropertyValue('--color-bg')).toBe('#0f172a')

    applyTheme(lightTheme)
    expect(document.documentElement.style.getPropertyValue('--color-bg')).toBe('#f8fafc')
    expect(document.documentElement.classList.contains('light')).toBe(true)
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('clearThemeStyles removes all --color-* inline styles', () => {
    applyTheme(darkTheme)
    expect(document.documentElement.style.getPropertyValue('--color-bg')).toBe('#0f172a')

    clearThemeStyles()
    expect(document.documentElement.style.getPropertyValue('--color-bg')).toBe('')
  })

  it('applyTheme sets all color properties', () => {
    applyTheme(darkTheme)
    const root = document.documentElement
    for (const [key, value] of Object.entries(darkTheme.colors)) {
      expect(root.style.getPropertyValue(`--${key}`), `--${key}`).toBe(value)
    }
  })
})

describe('Theme registry', () => {
  it('getTheme returns built-in themes', () => {
    expect(getTheme('dark')).toBe(darkTheme)
    expect(getTheme('light')).toBe(lightTheme)
  })

  it('getTheme returns undefined for unknown theme', () => {
    expect(getTheme('nonexistent')).toBeUndefined()
  })

  it('getAllThemes returns at least dark and light', () => {
    const themes = getAllThemes()
    const ids = themes.map(t => t.id)
    expect(ids).toContain('dark')
    expect(ids).toContain('light')
  })

  it('registerTheme adds a custom theme', () => {
    const custom: ThemeDefinition = { ...darkTheme, id: 'test-custom', name: 'Test Custom' }
    registerTheme(custom)
    expect(getTheme('test-custom')).toBe(custom)
    expect(getAllThemes().map(t => t.id)).toContain('test-custom')
  })
})

describe('loadThemeFromJSON', () => {
  afterEach(() => {
    clearThemeStyles()
    document.documentElement.classList.remove('dark', 'light')
  })

  it('loads a valid JSON object and registers it', () => {
    const json = JSON.parse(JSON.stringify({ ...darkTheme, id: 'json-test' }))
    const result = loadThemeFromJSON(json)
    expect(typeof result).toBe('object')
    expect((result as ThemeDefinition).id).toBe('json-test')
    expect(getTheme('json-test')).toBeDefined()
  })

  it('loads from a JSON string', () => {
    const str = JSON.stringify({ ...lightTheme, id: 'json-str-test' })
    const result = loadThemeFromJSON(str)
    expect(typeof result).toBe('object')
    expect((result as ThemeDefinition).id).toBe('json-str-test')
  })

  it('returns error for invalid JSON string', () => {
    expect(loadThemeFromJSON('not valid json{')).toBe('Invalid JSON string')
  })

  it('returns error for invalid theme structure', () => {
    const result = loadThemeFromJSON({ id: 'bad' })
    expect(typeof result).toBe('string')
  })

  it('applies theme immediately when apply: true', () => {
    const json = JSON.parse(JSON.stringify({ ...lightTheme, id: 'apply-test' }))
    loadThemeFromJSON(json, { apply: true })
    expect(document.documentElement.style.getPropertyValue('--color-bg')).toBe('#f8fafc')
    expect(document.documentElement.classList.contains('light')).toBe(true)
  })

  it('does not apply theme by default', () => {
    const json = JSON.parse(JSON.stringify({ ...darkTheme, id: 'no-apply-test' }))
    loadThemeFromJSON(json)
    // No inline style set (previous clearThemeStyles in afterEach)
    expect(document.documentElement.style.getPropertyValue('--color-bg')).toBe('')
  })
})
