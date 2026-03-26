import { render, screen, cleanup, fireEvent } from '@testing-library/svelte'
import { describe, it, expect, afterEach, beforeEach } from 'vitest'
import { get } from 'svelte/store'
import ThemeToggle from './ThemeToggle.svelte'
import { theme } from '../stores/theme'

describe('ThemeToggle', () => {
  beforeEach(() => {
    theme.set('dark')
  })

  afterEach(() => {
    cleanup()
    localStorage.removeItem('nidus-theme')
    document.documentElement.classList.remove('dark', 'light')
  })

  it('renders the toggle button', () => {
    render(ThemeToggle)
    expect(screen.getByTestId('theme-toggle')).toBeTruthy()
  })

  it('shows Sun icon in dark mode', () => {
    render(ThemeToggle)
    const button = screen.getByTestId('theme-toggle')
    // Sun icon should be visible, check aria-label
    expect(button.getAttribute('aria-label')).toBe('Passer en mode clair')
  })

  it('toggles theme on click', async () => {
    render(ThemeToggle)
    const button = screen.getByTestId('theme-toggle')
    await fireEvent.click(button)
    expect(get(theme)).toBe('light')
    expect(button.getAttribute('aria-label')).toBe('Passer en mode sombre')
  })
})
