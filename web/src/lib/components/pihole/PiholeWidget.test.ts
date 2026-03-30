import { render, screen, cleanup, waitFor } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import { setLocale } from '../../i18n'
import PiholeWidget from './PiholeWidget.svelte'

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

describe('PiholeWidget', () => {
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
      total_queries: 0, blocked_queries: 0, blocked_percent: 0,
      unique_domains: 0, cached_queries: 0, forwarded_queries: 0, blocking_enabled: true,
    })
    render(PiholeWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('pihole-widget')).toBeTruthy()
  })

  it('shows not configured when API returns 404', async () => {
    vi.mocked(api.get).mockRejectedValue({ status: 404 })
    render(PiholeWidget, { props: { config: '{}', active: true } })
    await waitFor(() => {
      expect(screen.getByText('Pi-hole not configured')).toBeTruthy()
    })
  })

  it('displays stats with mock data', async () => {
    vi.mocked(api.get).mockResolvedValue({
      total_queries: 87654,
      blocked_queries: 23456,
      blocked_percent: 26.8,
      unique_domains: 5432,
      cached_queries: 34567,
      forwarded_queries: 29641,
      blocking_enabled: true,
    })
    render(PiholeWidget, { props: { config: '{}', active: true } })
    await waitFor(() => {
      expect(screen.getByText('87.7k')).toBeTruthy()
      expect(screen.getByText('23.5k')).toBeTruthy()
      expect(screen.getByText('26.8%')).toBeTruthy()
      expect(screen.getByText('Total queries')).toBeTruthy()
    })
  })

  it('shows blocking enabled status', async () => {
    vi.mocked(api.get).mockResolvedValue({
      total_queries: 100,
      blocked_queries: 10,
      blocked_percent: 10.0,
      unique_domains: 50,
      cached_queries: 30,
      forwarded_queries: 60,
      blocking_enabled: true,
    })
    render(PiholeWidget, { props: { config: '{}', active: true } })
    await waitFor(() => {
      expect(screen.getByText('Protection active')).toBeTruthy()
    })
  })

  it('shows additional stats (unique domains, cached, forwarded)', async () => {
    vi.mocked(api.get).mockResolvedValue({
      total_queries: 1000,
      blocked_queries: 200,
      blocked_percent: 20.0,
      unique_domains: 350,
      cached_queries: 500,
      forwarded_queries: 300,
      blocking_enabled: true,
    })
    render(PiholeWidget, { props: { config: '{}', active: true } })
    await waitFor(() => {
      expect(screen.getByText('350')).toBeTruthy()
      expect(screen.getByText('500')).toBeTruthy()
      expect(screen.getByText('300')).toBeTruthy()
    })
  })
})
