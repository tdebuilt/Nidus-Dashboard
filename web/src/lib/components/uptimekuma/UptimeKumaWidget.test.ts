import { render, screen, cleanup, waitFor } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import { setLocale } from '../../i18n'
import UptimeKumaWidget from './UptimeKumaWidget.svelte'

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

describe('UptimeKumaWidget', () => {
  beforeEach(() => {
    setLocale('en')
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('renders the widget container', () => {
    vi.mocked(api.get).mockResolvedValue({
      monitors: [], total_up: 0, total_down: 0, total_count: 0, status_page: '',
    })
    render(UptimeKumaWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('uptimekuma-widget')).toBeTruthy()
  })

  it('shows not configured when API returns 404', async () => {
    vi.mocked(api.get).mockRejectedValue({ status: 404 })
    render(UptimeKumaWidget, { props: { config: '{}', active: true } })
    await waitFor(() => {
      expect(screen.getByText('Uptime Kuma not configured')).toBeTruthy()
    })
  })

  it('shows no monitors message when list is empty', async () => {
    vi.mocked(api.get).mockResolvedValue({
      monitors: [],
      total_up: 0,
      total_down: 0,
      total_count: 0,
      status_page: '',
    })
    render(UptimeKumaWidget, { props: { config: '{}', active: true } })
    await waitFor(() => {
      expect(screen.getByText('No monitors found on this status page')).toBeTruthy()
    })
  })

  it('displays monitor summary counts', async () => {
    vi.mocked(api.get).mockResolvedValue({
      monitors: [
        { id: 1, name: 'Web Server', type: 'http', status: 1, uptime_24h: 99.9, latency: 45, message: '' },
        { id: 2, name: 'Database', type: 'tcp', status: 1, uptime_24h: 100, latency: 12, message: '' },
        { id: 3, name: 'Mail Server', type: 'smtp', status: 0, uptime_24h: 85.2, latency: 0, message: 'Connection refused' },
      ],
      total_up: 2,
      total_down: 1,
      total_count: 3,
      status_page: 'default',
    })
    render(UptimeKumaWidget, { props: { config: '{}', active: true } })
    await waitFor(() => {
      expect(screen.getByText('2')).toBeTruthy()
      expect(screen.getByText('1')).toBeTruthy()
      expect(screen.getByText('3 monitors')).toBeTruthy()
    })
  })
})
