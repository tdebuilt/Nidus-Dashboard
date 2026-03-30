import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import TransmissionWidget from './TransmissionWidget.svelte'
import { setLocale } from '../../i18n'

describe('TransmissionWidget', () => {
  beforeEach(() => {
    setLocale('en')
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ torrents: [], download_speed: 0, upload_speed: 0, total_count: 0, active_count: 0 }),
    })
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('renders with data-testid', () => {
    render(TransmissionWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('transmission-widget')).toBeTruthy()
  })

  it('shows loading state when inactive', () => {
    render(TransmissionWidget, { props: { config: '{}', active: false } })
    expect(screen.getByText('Loading torrents...')).toBeTruthy()
  })

  it('accepts config prop as JSON string', () => {
    render(TransmissionWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('transmission-widget')).toBeTruthy()
  })
})
