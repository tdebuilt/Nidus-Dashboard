import { render, screen, cleanup, waitFor } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import MediaServerWidget from './MediaServerWidget.svelte'
import { setLocale } from '../../i18n'

describe('MediaServerWidget', () => {
  beforeEach(() => {
    setLocale('en')
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          sessions: [],
          session_count: 0,
          server_name: 'Test Server',
          server_type: 'jellyfin',
        }),
    })
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('renders with data-testid', () => {
    render(MediaServerWidget, { props: { config: '{}' } })
    expect(screen.getByTestId('mediaserver-widget')).toBeTruthy()
  })

  it('shows no active streams when session count is zero', async () => {
    render(MediaServerWidget, { props: { config: '{}', active: true } })

    await waitFor(() => {
      expect(screen.getByText('No active streams')).toBeTruthy()
    })
  })

  it('shows server name and type when data loads', async () => {
    render(MediaServerWidget, { props: { config: '{}', active: true } })

    await waitFor(() => {
      expect(screen.getByText('Test Server')).toBeTruthy()
      expect(screen.getByText('Jellyfin')).toBeTruthy()
    })
  })

  it('renders session cards when sessions exist', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          sessions: [
            {
              id: 'sess-1',
              user_name: 'Alice',
              title: 'Big Buck Bunny',
              media_type: 'movie',
              progress: 0.45,
              state: 'playing',
              player: 'Web',
              duration: 3600,
              position: 1620,
            },
          ],
          session_count: 1,
          server_name: 'My Plex',
          server_type: 'plex',
        }),
    })

    const config = JSON.stringify({ server_type: 'plex' })
    render(MediaServerWidget, { props: { config, active: true } })

    await waitFor(() => {
      expect(screen.getByText('Big Buck Bunny')).toBeTruthy()
      expect(screen.getByText('My Plex')).toBeTruthy()
    })
  })

  it('shows not configured state on 404 response', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 404,
      json: () => Promise.resolve({ error: 'not found' }),
    })

    render(MediaServerWidget, { props: { config: '{}', active: true } })

    await waitFor(() => {
      expect(screen.getByText('Media server not configured')).toBeTruthy()
    })
  })
})
