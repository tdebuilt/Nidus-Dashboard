import { render, screen, cleanup, waitFor } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import { setLocale } from '../../i18n'
import AdGuardWidget from './AdGuardWidget.svelte'

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

describe('AdGuardWidget', () => {
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
      avg_response_time: 0, filtering_enabled: true, active_filters: 0, total_rules: 0,
    })
    render(AdGuardWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('adguard-widget')).toBeTruthy()
  })

  it('shows not configured when API returns 404', async () => {
    vi.mocked(api.get).mockRejectedValue({ status: 404 })
    render(AdGuardWidget, { props: { config: '{}', active: true } })
    await waitFor(() => {
      expect(screen.getByText('AdGuard not configured')).toBeTruthy()
    })
  })

  it('displays stats with mock data', async () => {
    vi.mocked(api.get).mockResolvedValue({
      total_queries: 54321,
      blocked_queries: 12345,
      blocked_percent: 22.7,
      avg_response_time: 0.015,
      filtering_enabled: true,
      active_filters: 3,
      total_rules: 150000,
    })
    render(AdGuardWidget, { props: { config: '{}', active: true } })
    await waitFor(() => {
      expect(screen.getByText('54.3k')).toBeTruthy()
      expect(screen.getByText('12.3k')).toBeTruthy()
      expect(screen.getByText('22.7%')).toBeTruthy()
      expect(screen.getByText('DNS Queries')).toBeTruthy()
    })
  })

  it('shows filtering status when protection is active', async () => {
    vi.mocked(api.get).mockResolvedValue({
      total_queries: 100,
      blocked_queries: 10,
      blocked_percent: 10.0,
      avg_response_time: 0.005,
      filtering_enabled: true,
      active_filters: 2,
      total_rules: 50000,
    })
    render(AdGuardWidget, { props: { config: '{}', active: true } })
    await waitFor(() => {
      expect(screen.getByText('Protection active')).toBeTruthy()
    })
  })
})
