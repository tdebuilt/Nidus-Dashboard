import { describe, it, expect } from 'vitest'
import { getResponsiveColumns, clampResponsiveColumns } from './responsiveColumns'

describe('getResponsiveColumns', () => {
  it('returns desktop columns on desktop', () => {
    expect(getResponsiveColumns({ columns: 3 }, 'desktop', 2)).toBe(3)
  })

  it('uses default when columns not set', () => {
    expect(getResponsiveColumns({}, 'desktop', 2)).toBe(2)
  })

  it('returns columnsTablet on tablet', () => {
    expect(getResponsiveColumns({ columns: 4, columnsTablet: 2 }, 'tablet', 1)).toBe(2)
  })

  it('falls back to min(columns, 2) for tablet when not set', () => {
    expect(getResponsiveColumns({ columns: 3 }, 'tablet', 1)).toBe(2)
    expect(getResponsiveColumns({ columns: 1 }, 'tablet', 1)).toBe(1)
  })

  it('returns columnsMobile on mobile', () => {
    expect(getResponsiveColumns({ columns: 4, columnsMobile: 2 }, 'mobile', 1)).toBe(2)
  })

  it('falls back to 1 for mobile when not set', () => {
    expect(getResponsiveColumns({ columns: 3 }, 'mobile', 1)).toBe(1)
  })

  it('uses undefined columnsTablet and falls back', () => {
    expect(getResponsiveColumns({ columns: 3, columnsTablet: undefined }, 'tablet', 2)).toBe(2)
  })
})

describe('clampResponsiveColumns', () => {
  it('returns values unchanged when valid', () => {
    expect(clampResponsiveColumns(3, 2, 1)).toEqual({ desktop: 3, tablet: 2, mobile: 1 })
  })

  it('clamps tablet to desktop', () => {
    expect(clampResponsiveColumns(2, 3, 1)).toEqual({ desktop: 2, tablet: 2, mobile: 1 })
  })

  it('clamps mobile to tablet', () => {
    expect(clampResponsiveColumns(3, 2, 4)).toEqual({ desktop: 3, tablet: 2, mobile: 2 })
  })

  it('cascades clamping', () => {
    expect(clampResponsiveColumns(1, 3, 4)).toEqual({ desktop: 1, tablet: 1, mobile: 1 })
  })

  it('clamps minimum to 1', () => {
    expect(clampResponsiveColumns(0, 0, 0)).toEqual({ desktop: 1, tablet: 1, mobile: 1 })
  })

  it('clamps maximum to 4', () => {
    expect(clampResponsiveColumns(6, 5, 5)).toEqual({ desktop: 4, tablet: 4, mobile: 4 })
  })
})
