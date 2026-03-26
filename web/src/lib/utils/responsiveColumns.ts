import type { Breakpoint } from '../stores/breakpoint'

interface ColumnsConfig {
  columns?: number
  columnsTablet?: number
  columnsMobile?: number
}

export function getResponsiveColumns(
  config: ColumnsConfig,
  bp: Breakpoint,
  defaultDesktop: number
): number {
  const desktop = config.columns ?? defaultDesktop
  const tablet = config.columnsTablet ?? Math.min(desktop, 2)
  const mobile = config.columnsMobile ?? 1

  if (bp === 'mobile') return mobile
  if (bp === 'tablet') return tablet
  return desktop
}

export function clampResponsiveColumns(
  desktop: number,
  tablet: number,
  mobile: number
): { desktop: number; tablet: number; mobile: number } {
  const d = Math.max(1, Math.min(4, desktop))
  const t = Math.max(1, Math.min(d, tablet))
  const m = Math.max(1, Math.min(t, mobile))
  return { desktop: d, tablet: t, mobile: m }
}
