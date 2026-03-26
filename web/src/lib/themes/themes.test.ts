import { describe, it, expect } from 'vitest'
import { darkTheme, lightTheme, nordTheme, draculaTheme, builtinThemes, getBuiltinTheme, getAvailableThemes, parseThemeJSON, themeColorKeys } from './themes'
import type { ThemeColors } from './themes'

describe('Theme definitions', () => {
  const requiredKeys: (keyof ThemeColors)[] = [
    'color-bg', 'color-bg-primary', 'color-bg-secondary', 'color-bg-tertiary',
    'color-border',
    'color-text', 'color-text-primary', 'color-text-secondary', 'color-text-muted',
    'color-primary', 'color-primary-hover',
    'color-accent', 'color-accent-hover',
    'color-danger', 'color-danger-hover',
    'color-success', 'color-warning',
    'color-sidebar-bg', 'color-sidebar-hover',
    'color-error-text', 'color-error-border', 'color-error-bg',
    'color-success-text', 'color-success-border', 'color-success-bg',
    'color-info-text', 'color-info-border', 'color-info-bg',
  ]

  it('dark theme has all required color keys', () => {
    for (const key of requiredKeys) {
      expect(darkTheme.colors[key], `missing ${key}`).toBeDefined()
      expect(darkTheme.colors[key]).not.toBe('')
    }
  })

  it('all built-in themes have all required color keys', () => {
    for (const theme of [darkTheme, lightTheme, nordTheme, draculaTheme]) {
      for (const key of requiredKeys) {
        expect(theme.colors[key], `${theme.id} missing ${key}`).toBeDefined()
        expect(theme.colors[key]).not.toBe('')
      }
    }
  })

  it('all built-in themes have the same keys', () => {
    const darkKeys = Object.keys(darkTheme.colors).sort()
    for (const theme of [lightTheme, nordTheme, draculaTheme]) {
      expect(Object.keys(theme.colors).sort()).toEqual(darkKeys)
    }
  })

  it('themes have correct mode', () => {
    expect(darkTheme.mode).toBe('dark')
    expect(lightTheme.mode).toBe('light')
    expect(nordTheme.mode).toBe('dark')
    expect(draculaTheme.mode).toBe('dark')
  })

  it('builtinThemes contains all 4 themes', () => {
    expect(builtinThemes).toHaveProperty('dark')
    expect(builtinThemes).toHaveProperty('light')
    expect(builtinThemes).toHaveProperty('nord')
    expect(builtinThemes).toHaveProperty('dracula')
  })

  it('getBuiltinTheme returns the correct theme', () => {
    expect(getBuiltinTheme('dark')).toBe(darkTheme)
    expect(getBuiltinTheme('light')).toBe(lightTheme)
    expect(getBuiltinTheme('nord')).toBe(nordTheme)
    expect(getBuiltinTheme('dracula')).toBe(draculaTheme)
    expect(getBuiltinTheme('nonexistent')).toBeUndefined()
  })

  it('getAvailableThemes returns all built-in themes', () => {
    const themes = getAvailableThemes()
    expect(themes).toHaveLength(4)
    expect(themes).toContain(darkTheme)
    expect(themes).toContain(lightTheme)
    expect(themes).toContain(nordTheme)
    expect(themes).toContain(draculaTheme)
  })

  it('themeColorKeys lists all required keys', () => {
    expect(themeColorKeys).toHaveLength(Object.keys(darkTheme.colors).length)
  })
})

describe('parseThemeJSON', () => {
  function validJSON() {
    return JSON.parse(JSON.stringify(darkTheme))
  }

  it('parses a valid theme JSON', () => {
    const result = parseThemeJSON(validJSON())
    expect(typeof result).toBe('object')
    expect((result as unknown as Record<string, unknown>).id).toBe('dark')
  })

  it('rejects null', () => {
    expect(parseThemeJSON(null)).toBe('Theme must be a JSON object')
  })

  it('rejects non-object', () => {
    expect(parseThemeJSON('string')).toBe('Theme must be a JSON object')
  })

  it('rejects missing id', () => {
    const json = validJSON()
    delete json.id
    expect(parseThemeJSON(json)).toContain('id')
  })

  it('rejects missing name', () => {
    const json = validJSON()
    json.name = ''
    expect(parseThemeJSON(json)).toContain('name')
  })

  it('rejects invalid mode', () => {
    const json = validJSON()
    json.mode = 'neon'
    expect(parseThemeJSON(json)).toContain('mode')
  })

  it('rejects missing colors', () => {
    const json = validJSON()
    delete json.colors
    expect(parseThemeJSON(json)).toContain('colors')
  })

  it('reports missing color keys', () => {
    const json = validJSON()
    delete json.colors['color-bg']
    delete json.colors['color-text']
    const result = parseThemeJSON(json)
    expect(typeof result).toBe('string')
    expect(result).toContain('color-bg')
    expect(result).toContain('color-text')
  })

  it('roundtrips built-in themes through JSON', () => {
    const json = JSON.parse(JSON.stringify(lightTheme))
    const result = parseThemeJSON(json)
    expect(typeof result).toBe('object')
    expect((result as unknown as Record<string, unknown>).id).toBe('light')
    expect((result as unknown as Record<string, Record<string, string>>).colors['color-bg']).toBe('#f8fafc')
  })
})
