import { render, screen, cleanup, waitFor } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import { setLocale } from '../../i18n'
import JDownloaderWidget from './JDownloaderWidget.svelte'

vi.mock('../../api/client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
  ApiError: class extends Error {
    constructor(public status: number, message: string) {
      super(message)
    }
  },
  NetworkError: class extends Error {},
}))

import { api } from '../../api/client'

describe('JDownloaderWidget', () => {
  beforeEach(() => {
    setLocale('en')
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('renders the widget container', () => {
    vi.mocked(api.get).mockResolvedValue({ packages: [], total_speed: 0, running: false })
    render(JDownloaderWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('jdownloader-widget')).toBeTruthy()
  })

  it('shows not configured when API returns 404', async () => {
    vi.mocked(api.get).mockRejectedValue({ status: 404 })
    render(JDownloaderWidget, { props: { config: '{}', active: true } })
    await waitFor(() => {
      expect(screen.getByText('JDownloader not configured')).toBeTruthy()
    })
  })

  it('shows empty queue message when no packages', async () => {
    vi.mocked(api.get).mockResolvedValue({
      packages: [],
      total_speed: 0,
      running: false,
    })
    render(JDownloaderWidget, { props: { config: '{}', active: true } })
    await waitFor(() => {
      expect(screen.getByText('No downloads in queue')).toBeTruthy()
    })
  })

  it('shows idle status when queue is not running', async () => {
    vi.mocked(api.get).mockResolvedValue({
      packages: [
        {
          uuid: 1,
          name: 'test-file.zip',
          status: 'queued',
          progress: 0,
          size: 1048576,
          downloaded: 0,
          speed: 0,
          eta: 0,
          finished: false,
          link_count: 1,
        },
      ],
      total_speed: 0,
      running: false,
    })
    render(JDownloaderWidget, { props: { config: '{}', active: true } })
    await waitFor(() => {
      expect(screen.getByText('Idle')).toBeTruthy()
    })
  })
})
