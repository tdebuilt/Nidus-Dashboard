import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import ArrWidget from './ArrWidget.svelte'
import { setLocale } from '../../i18n'

describe('ArrWidget', () => {
  beforeEach(() => {
    setLocale('en')
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve([]),
    })
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('renders with data-testid', () => {
    render(ArrWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('arr-widget')).toBeTruthy()
  })

  it('shows loading state when inactive', () => {
    render(ArrWidget, { props: { config: '{}', active: false } })
    expect(screen.getByText('Loading media data...')).toBeTruthy()
  })

  it('accepts config prop as JSON string', () => {
    render(ArrWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('arr-widget')).toBeTruthy()
  })
})
