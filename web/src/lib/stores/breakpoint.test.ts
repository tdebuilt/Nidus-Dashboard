import { describe, it, expect } from 'vitest'
import { get } from 'svelte/store'
import { breakpoint } from './breakpoint'

describe('breakpoint store', () => {
  it('defaults to desktop when no matchMedia available', () => {
    expect(get(breakpoint)).toBe('desktop')
  })
})
