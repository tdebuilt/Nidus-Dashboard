import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import QBittorrentWidget from './QBittorrentWidget.svelte'
import { setLocale } from '../../i18n'

describe('QBittorrentWidget', () => {
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
    render(QBittorrentWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('qbittorrent-widget')).toBeTruthy()
  })

  it('shows loading state when inactive', () => {
    render(QBittorrentWidget, { props: { config: '{}', active: false } })
    expect(screen.getByText('Loading torrents...')).toBeTruthy()
  })

  it('accepts config prop as JSON string', () => {
    render(QBittorrentWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('qbittorrent-widget')).toBeTruthy()
  })
})
