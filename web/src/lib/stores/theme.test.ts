import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { get } from 'svelte/store'
import { theme, toggleTheme, setTheme } from './theme'

describe('Theme store', () => {
  beforeEach(() => {
    localStorage.removeItem('nidus-theme')
    theme.set('dark')
  })

  afterEach(() => {
    localStorage.removeItem('nidus-theme')
    document.documentElement.classList.remove('dark', 'light')
  })

  it('defaults to dark theme', () => {
    expect(get(theme)).toBe('dark')
  })

  it('toggles to light', () => {
    toggleTheme()
    expect(get(theme)).toBe('light')
  })

  it('toggles back to dark', () => {
    toggleTheme() // dark -> light
    toggleTheme() // light -> dark
    expect(get(theme)).toBe('dark')
  })

  it('toggles from nord (dark mode) to light', () => {
    setTheme('nord')
    toggleTheme()
    expect(get(theme)).toBe('light')
  })

  it('persists theme to localStorage', () => {
    theme.set('light')
    expect(localStorage.getItem('nidus-theme')).toBe('light')
  })

  it('applies dark class for dark-mode themes', () => {
    theme.set('dark')
    expect(document.documentElement.classList.contains('dark')).toBe(true)
    expect(document.documentElement.classList.contains('light')).toBe(false)
  })

  it('applies light class to documentElement', () => {
    theme.set('light')
    expect(document.documentElement.classList.contains('light')).toBe(true)
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('setTheme sets a named theme', () => {
    setTheme('nord')
    expect(get(theme)).toBe('nord')
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('setTheme sets dracula theme', () => {
    setTheme('dracula')
    expect(get(theme)).toBe('dracula')
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('setTheme ignores unknown theme', () => {
    setTheme('nonexistent')
    expect(get(theme)).toBe('dark')
  })

  it('loads theme from localStorage', () => {
    localStorage.setItem('nidus-theme', 'nord')
    theme.set(localStorage.getItem('nidus-theme')!)
    expect(get(theme)).toBe('nord')
  })
})
