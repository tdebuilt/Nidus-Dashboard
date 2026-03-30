import { render, screen, cleanup, waitFor } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import { setLocale } from '../../i18n'
import CalendarWidget from './CalendarWidget.svelte'

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

describe('CalendarWidget', () => {
  beforeEach(() => {
    setLocale('en')
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('renders the widget container', () => {
    vi.mocked(api.get).mockResolvedValue({ events: [] })
    render(CalendarWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('calendar-widget')).toBeTruthy()
  })

  it('shows not configured state when no URLs provided', async () => {
    render(CalendarWidget, { props: { config: '{}', active: true } })
    await waitFor(() => {
      expect(screen.getByText('Calendar widget not configured')).toBeTruthy()
    })
  })

  it('shows empty state when no events returned', async () => {
    vi.mocked(api.get).mockResolvedValue({ events: [] })
    const config = JSON.stringify({ urls: ['https://example.com/cal.ics'] })
    render(CalendarWidget, { props: { config, active: true } })
    await waitFor(() => {
      expect(screen.getByText('No upcoming events')).toBeTruthy()
    })
  })

  it('renders events from mock data', async () => {
    const tomorrow = new Date()
    tomorrow.setDate(tomorrow.getDate() + 1)
    tomorrow.setHours(10, 0, 0, 0)

    vi.mocked(api.get).mockResolvedValue({
      events: [
        {
          uid: 'evt-1',
          summary: 'Team meeting',
          start: tomorrow.toISOString(),
          end: tomorrow.toISOString(),
          all_day: false,
        },
      ],
    })
    const config = JSON.stringify({ urls: ['https://example.com/cal.ics'] })
    render(CalendarWidget, { props: { config, active: true } })
    await waitFor(() => {
      expect(screen.getByText('Team meeting')).toBeTruthy()
    })
  })
})
