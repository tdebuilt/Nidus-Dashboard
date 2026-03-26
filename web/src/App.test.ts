import { render, screen, cleanup, waitFor } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import App from './App.svelte'
import { currentPath } from './lib/stores/router'
import { auth } from './lib/stores/auth'

describe('App', () => {
  beforeEach(() => {
    currentPath.set('/')
    // Mock fetch: auth status returns setup_completed, other endpoints return arrays
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      if (typeof url === 'string' && url.includes('/api/auth/status')) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: () => Promise.resolve({ setup_completed: true }),
        })
      }
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve([]),
      })
    })
  })

  afterEach(() => {
    cleanup()
    vi.restoreAllMocks()
  })

  it('renders loading state initially', () => {
    auth.set({ authenticated: false, setupCompleted: false, loading: true })
    render(App)
    expect(screen.getByTestId('loading')).toBeTruthy()
  })

  it('renders main layout when authenticated', async () => {
    render(App)
    await waitFor(() => {
      expect(screen.getByTestId('main-content')).toBeTruthy()
    })
  })

  it('renders sidebar in main layout', async () => {
    render(App)
    await waitFor(() => {
      expect(screen.getByTestId('sidebar')).toBeTruthy()
    })
  })

  it('renders dashboard page on / route', async () => {
    render(App)
    await waitFor(() => {
      expect(screen.getByTestId('dashboard-page')).toBeTruthy()
    })
  })

  it('renders login page on /login route', async () => {
    currentPath.set('/login')
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ error: 'unauthorized' }),
    })
    render(App)
    await waitFor(() => {
      expect(screen.getByTestId('login-page')).toBeTruthy()
    })
  })

  it('renders setup page on /setup route', async () => {
    currentPath.set('/setup')
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ setup_completed: false }),
    })
    render(App)
    await waitFor(() => {
      expect(screen.getByTestId('setup-page')).toBeTruthy()
    })
  })

  it('renders not-found for unknown routes', async () => {
    currentPath.set('/some/unknown/route')
    render(App)
    await waitFor(() => {
      expect(screen.getByTestId('not-found-page')).toBeTruthy()
    })
  })

  it('renders network error page on fetch failure', async () => {
    globalThis.fetch = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'))
    render(App)
    await waitFor(() => {
      expect(screen.getByTestId('network-error-page')).toBeTruthy()
    })
  })

  it('shows retry button on network error', async () => {
    globalThis.fetch = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'))
    render(App)
    await waitFor(() => {
      expect(screen.getByTestId('retry-button')).toBeTruthy()
    })
  })
})
