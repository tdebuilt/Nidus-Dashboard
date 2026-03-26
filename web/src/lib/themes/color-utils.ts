import type { ThemeColors } from './themes'

export interface BaseColors {
  bg: string
  text: string
  primary: string
  accent: string
  danger: string
  success: string
  warning: string
  mode: 'dark' | 'light'
}

export function hexToRgb(hex: string): [number, number, number] {
  const h = hex.replace('#', '')
  return [
    parseInt(h.slice(0, 2), 16),
    parseInt(h.slice(2, 4), 16),
    parseInt(h.slice(4, 6), 16),
  ]
}

export function rgbToHex(r: number, g: number, b: number): string {
  return '#' + [r, g, b].map(v => Math.max(0, Math.min(255, Math.round(v))).toString(16).padStart(2, '0')).join('')
}

export function lightenColor(hex: string, amount: number): string {
  const [r, g, b] = hexToRgb(hex)
  return rgbToHex(
    r + (255 - r) * amount,
    g + (255 - g) * amount,
    b + (255 - b) * amount,
  )
}

export function darkenColor(hex: string, amount: number): string {
  const [r, g, b] = hexToRgb(hex)
  return rgbToHex(
    r * (1 - amount),
    g * (1 - amount),
    b * (1 - amount),
  )
}

export function muteColor(hex: string, amount: number): string {
  const [r, g, b] = hexToRgb(hex)
  const gray = 128
  return rgbToHex(
    r + (gray - r) * amount,
    g + (gray - g) * amount,
    b + (gray - b) * amount,
  )
}

export function hexToRgba(hex: string, alpha: number): string {
  const [r, g, b] = hexToRgb(hex)
  return `rgba(${r}, ${g}, ${b}, ${alpha})`
}

export function deriveFullTheme(base: BaseColors): ThemeColors {
  const isDark = base.mode === 'dark'

  return {
    'color-bg': base.bg,
    'color-bg-primary': isDark ? lightenColor(base.bg, 0.02) : darkenColor(base.bg, 0.02),
    'color-bg-secondary': isDark ? lightenColor(base.bg, 0.08) : darkenColor(base.bg, 0.04),
    'color-bg-tertiary': isDark ? lightenColor(base.bg, 0.15) : darkenColor(base.bg, 0.08),
    'color-border': isDark ? lightenColor(base.bg, 0.15) : darkenColor(base.bg, 0.12),
    'color-text': base.text,
    'color-text-primary': base.text,
    'color-text-secondary': muteColor(base.text, 0.35),
    'color-text-muted': muteColor(base.text, 0.55),
    'color-primary': base.primary,
    'color-primary-hover': darkenColor(base.primary, 0.15),
    'color-accent': base.accent,
    'color-accent-hover': isDark ? lightenColor(base.accent, 0.15) : darkenColor(base.accent, 0.15),
    'color-danger': base.danger,
    'color-danger-hover': darkenColor(base.danger, 0.15),
    'color-success': base.success,
    'color-warning': base.warning,
    'color-sidebar-bg': isDark ? lightenColor(base.bg, 0.08) : darkenColor(base.bg, 0.02),
    'color-sidebar-hover': isDark ? lightenColor(base.bg, 0.15) : darkenColor(base.bg, 0.06),
    'color-error-text': isDark ? lightenColor(base.danger, 0.20) : darkenColor(base.danger, 0.10),
    'color-error-border': base.danger,
    'color-error-bg': hexToRgba(base.danger, 0.1),
    'color-success-text': isDark ? lightenColor(base.success, 0.30) : darkenColor(base.success, 0.10),
    'color-success-border': base.success,
    'color-success-bg': hexToRgba(base.success, 0.1),
    'color-info-text': isDark ? lightenColor(base.primary, 0.30) : darkenColor(base.primary, 0.10),
    'color-info-border': base.primary,
    'color-info-bg': hexToRgba(base.primary, 0.1),
  }
}

export function extractBaseColors(colors: ThemeColors, mode: 'dark' | 'light'): BaseColors {
  return {
    bg: colors['color-bg'],
    text: colors['color-text'],
    primary: colors['color-primary'],
    accent: colors['color-accent'],
    danger: colors['color-danger'],
    success: colors['color-success'],
    warning: colors['color-warning'],
    mode,
  }
}
