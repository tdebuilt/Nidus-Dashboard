import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import KioskPage from './KioskPage.svelte'

const mockCategories = [
  { id: 1, name: 'Home', slug: 'home', icon: 'home', sort_order: 0 },
  { id: 2, name: 'Media', slug: 'media', icon: 'film', sort_order: 1 },
]

describe('KioskPage', () => {
  beforeEach(() => {
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      if (typeof url === 'string' && url.includes('/api/categories')) {
        // Categories list or widgets for a category
        if (url.match(/\/api\/categories\/\d+\/widgets/)) {
          return Promise.resolve({
            ok: true,
            status: 200,
            json: () => Promise.resolve([]),
          })
        }
        return Promise.resolve({
          ok: true,
          status: 200,
          json: () => Promise.resolve(mockCategories),
        })
      }
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve([]),
      })
    })

    // Mock fullscreen API
    Object.defineProperty(document, 'fullscreenElement', { value: null, writable: true, configurable: true })
    Object.defineProperty(HTMLElement.prototype, 'requestFullscreen', {
      value: vi.fn().mockResolvedValue(undefined),
      writable: true,
      configurable: true,
    })
  })

  afterEach(() => {
    cleanup()
    vi.restoreAllMocks()
  })

  it('renders the kiosk page', () => {
    render(KioskPage)
    expect(screen.getByTestId('kiosk-page')).toBeTruthy()
  })

  it('has rotation toggle button', () => {
    render(KioskPage)
    expect(screen.getByTestId('kiosk-rotation-toggle')).toBeTruthy()
  })

  it('has exit button', () => {
    render(KioskPage)
    expect(screen.getByTestId('kiosk-exit')).toBeTruthy()
  })
})
