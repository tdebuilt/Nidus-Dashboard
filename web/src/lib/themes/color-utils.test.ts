import { describe, it, expect } from 'vitest'
import { hexToRgb, rgbToHex, lightenColor, darkenColor, muteColor, hexToRgba, deriveFullTheme } from './color-utils'
import { themeColorKeys } from './themes'
import type { BaseColors } from './color-utils'

describe('color-utils', () => {
  describe('hexToRgb', () => {
    it('converts black', () => {
      expect(hexToRgb('#000000')).toEqual([0, 0, 0])
    })
    it('converts white', () => {
      expect(hexToRgb('#ffffff')).toEqual([255, 255, 255])
    })
    it('converts blue', () => {
      expect(hexToRgb('#3b82f6')).toEqual([59, 130, 246])
    })
  })

  describe('rgbToHex', () => {
    it('converts black', () => {
      expect(rgbToHex(0, 0, 0)).toBe('#000000')
    })
    it('converts white', () => {
      expect(rgbToHex(255, 255, 255)).toBe('#ffffff')
    })
    it('round-trips with hexToRgb', () => {
      const hex = '#3b82f6'
      const [r, g, b] = hexToRgb(hex)
      expect(rgbToHex(r, g, b)).toBe(hex)
    })
    it('clamps values', () => {
      expect(rgbToHex(300, -10, 128)).toBe('#ff0080')
    })
  })

  describe('lightenColor', () => {
    it('lightens black toward white', () => {
      const result = lightenColor('#000000', 0.5)
      expect(hexToRgb(result)).toEqual([128, 128, 128])
    })
    it('does not change white', () => {
      expect(lightenColor('#ffffff', 0.5)).toBe('#ffffff')
    })
  })

  describe('darkenColor', () => {
    it('darkens white toward black', () => {
      const result = darkenColor('#ffffff', 0.5)
      expect(hexToRgb(result)).toEqual([128, 128, 128])
    })
    it('does not change black', () => {
      expect(darkenColor('#000000', 0.5)).toBe('#000000')
    })
  })

  describe('muteColor', () => {
    it('blends toward gray', () => {
      const result = muteColor('#ffffff', 0.5)
      const [r, _g, _b] = hexToRgb(result)
      expect(r).toBeGreaterThan(128)
      expect(r).toBeLessThan(255)
    })
  })

  describe('hexToRgba', () => {
    it('creates rgba string', () => {
      expect(hexToRgba('#ff0000', 0.1)).toBe('rgba(255, 0, 0, 0.1)')
    })
  })

  describe('deriveFullTheme', () => {
    const darkBase: BaseColors = {
      bg: '#0f172a',
      text: '#f1f5f9',
      primary: '#3b82f6',
      accent: '#6366f1',
      danger: '#ef4444',
      success: '#22c55e',
      warning: '#eab308',
      mode: 'dark',
    }

    const lightBase: BaseColors = {
      ...darkBase,
      bg: '#f8fafc',
      text: '#0f172a',
      mode: 'light',
    }

    it('produces all 28 color keys', () => {
      const result = deriveFullTheme(darkBase)
      for (const key of themeColorKeys) {
        expect(result[key]).toBeDefined()
      }
    })

    it('preserves base colors', () => {
      const result = deriveFullTheme(darkBase)
      expect(result['color-bg']).toBe(darkBase.bg)
      expect(result['color-text']).toBe(darkBase.text)
      expect(result['color-primary']).toBe(darkBase.primary)
    })

    it('dark and light modes produce different bg-secondary', () => {
      const dark = deriveFullTheme(darkBase)
      const light = deriveFullTheme(lightBase)
      expect(dark['color-bg-secondary']).not.toBe(light['color-bg-secondary'])
    })
  })
})
