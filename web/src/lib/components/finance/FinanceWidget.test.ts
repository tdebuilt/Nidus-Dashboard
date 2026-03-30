import { render, screen, cleanup, waitFor } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import FinanceWidget from './FinanceWidget.svelte'
import { setLocale } from '../../i18n'

describe('FinanceWidget', () => {
  beforeEach(() => {
    setLocale('en')
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          quotes: [],
          fetched_at: Math.floor(Date.now() / 1000),
        }),
    })
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('shows not configured when no symbols are set', async () => {
    render(FinanceWidget, { props: { config: '{}', active: true } })

    await waitFor(() => {
      expect(screen.getByText('Finance widget not configured')).toBeTruthy()
    })
  })

  it('shows not configured with empty symbols array', async () => {
    const config = JSON.stringify({ symbols: [] })
    render(FinanceWidget, { props: { config, active: true } })

    await waitFor(() => {
      expect(screen.getByText('Finance widget not configured')).toBeTruthy()
    })
  })

  it('renders stock quotes table when data loads', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          quotes: [
            {
              symbol: 'AAPL',
              short_name: 'Apple Inc.',
              quote_type: 'EQUITY',
              currency: 'USD',
              price: 178.5,
              change: 2.3,
              change_percent: 1.31,
              volume: 55000000,
              open: 176.2,
              day_high: 179.0,
              day_low: 175.8,
              market_cap: 2800000000000,
              market_state: 'REGULAR',
            },
          ],
          fetched_at: Math.floor(Date.now() / 1000),
        }),
    })

    const config = JSON.stringify({ symbols: ['AAPL'] })
    render(FinanceWidget, { props: { config, active: true } })

    await waitFor(() => {
      expect(screen.getByText('AAPL')).toBeTruthy()
      expect(screen.getByText('Apple Inc.')).toBeTruthy()
    })
  })

  it('renders table headers', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          quotes: [
            {
              symbol: 'BTC-USD',
              short_name: 'Bitcoin',
              quote_type: 'CRYPTO',
              currency: 'USD',
              price: 64000,
              change: -500,
              change_percent: -0.78,
              volume: 30000000000,
              open: 64500,
              day_high: 65000,
              day_low: 63500,
              market_cap: 0,
              market_state: 'REGULAR',
            },
          ],
          fetched_at: Math.floor(Date.now() / 1000),
        }),
    })

    const config = JSON.stringify({ symbols: ['BTC-USD'] })
    render(FinanceWidget, { props: { config, active: true } })

    await waitFor(() => {
      expect(screen.getByText('Symbol')).toBeTruthy()
      expect(screen.getByText('Price')).toBeTruthy()
      expect(screen.getByText('Change')).toBeTruthy()
    })
  })
})
