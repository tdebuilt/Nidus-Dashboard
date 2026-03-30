import { render, screen, cleanup, waitFor } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import WeatherWidget from './WeatherWidget.svelte'
import { setLocale } from '../../i18n'

describe('WeatherWidget', () => {
  beforeEach(() => {
    setLocale('en')
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(null),
    })
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('renders with data-testid', () => {
    render(WeatherWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('weather-widget')).toBeTruthy()
  })

  it('shows not-configured state when no api_key', async () => {
    render(WeatherWidget, { props: { config: '{}' } })

    await waitFor(() => {
      expect(screen.getByText('Weather widget not configured')).toBeTruthy()
    })
  })

  it('shows not-configured state when api_key but no city or coords', async () => {
    const config = JSON.stringify({ api_key: 'test-key' })
    render(WeatherWidget, { props: { config } })

    await waitFor(() => {
      expect(screen.getByText('Weather widget not configured')).toBeTruthy()
    })
  })

  it('shows loading state while fetching', () => {
    globalThis.fetch = vi.fn().mockImplementation(() => new Promise(() => {}))
    const config = JSON.stringify({ api_key: 'test-key', city: 'Paris,FR' })
    render(WeatherWidget, { props: { config } })
    expect(screen.getByText(/Loading weather/)).toBeTruthy()
  })

  it('renders weather data when available', async () => {
    const weatherData = {
      current: {
        temp: 18.5,
        feels_like: 17.2,
        temp_min: 15,
        temp_max: 22,
        humidity: 65,
        pressure: 1013,
        wind_speed: 3.5,
        wind_deg: 180,
        description: 'Partly cloudy',
        icon: '02d',
        city: 'Paris',
        country: 'FR',
        sunrise: 1700000000,
        sunset: 1700040000,
      },
      forecast: [],
    }

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(weatherData),
    })

    const config = JSON.stringify({ api_key: 'test-key', city: 'Paris,FR' })
    render(WeatherWidget, { props: { config } })

    await waitFor(() => {
      expect(screen.getByText(/19°C/)).toBeTruthy()
      expect(screen.getByText('Partly cloudy')).toBeTruthy()
      expect(screen.getByText(/Paris/)).toBeTruthy()
    })
  })
})
