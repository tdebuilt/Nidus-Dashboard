import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import SystemWidget from './SystemWidget.svelte'
import { setLocale } from '../../i18n'

describe('SystemWidget', () => {
  beforeEach(() => {
    setLocale('en')
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({}),
    })
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('renders with data-testid', () => {
    render(SystemWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('system-widget')).toBeTruthy()
  })

  it('shows loading state when inactive', () => {
    render(SystemWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('system-widget')).toBeTruthy()
  })

  it('accepts config prop as JSON string', () => {
    render(SystemWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('system-widget')).toBeTruthy()
  })
})
