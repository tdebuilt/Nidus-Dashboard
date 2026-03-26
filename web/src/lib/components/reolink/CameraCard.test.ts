import { render, screen, cleanup, waitFor } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import CameraCard from './CameraCard.svelte'
import { auth } from '../../stores/auth'
import { setLocale } from '../../i18n'

describe('CameraCard', () => {
  beforeEach(() => {
    // Set auth to editor role so live toggle button is visible
    auth.set({
      authenticated: true,
      setupCompleted: true,
      loading: false,
      role: 'editor',
      userId: 1,
      username: 'test',
    })

    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      if (typeof url === 'string' && url.includes('/stream')) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: () => Promise.resolve({ go2rtc: null }),
        })
      }
      // Snapshot requests
      return Promise.resolve({
        ok: true,
        status: 200,
        blob: () => Promise.resolve(new Blob()),
      })
    })

    setLocale('en')

    // Mock Image constructor for snapshot preloading
    globalThis.Image = class MockImage {
      onload: (() => void) | null = null
      onerror: (() => void) | null = null
      set src(_: string) {
        setTimeout(() => this.onload?.(), 0)
      }
    } as unknown as typeof Image
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('renders camera name', () => {
    render(CameraCard, { props: { id: 'abc', name: 'Test Cam' } })
    expect(screen.getByText('Test Cam')).toBeTruthy()
  })

  it('renders snapshot image', () => {
    render(CameraCard, { props: { id: 'abc', name: 'Test Cam' } })
    const img = screen.getByAltText('Test Cam') as HTMLImageElement
    expect(img).toBeTruthy()
    expect(img.src).toContain('/api/reolink/cameras/abc/snapshot')
  })

  it('fetches stream info on mount', async () => {
    render(CameraCard, { props: { id: 'abc', name: 'Test Cam' } })

    await waitFor(() => {
      expect(globalThis.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/reolink/cameras/abc/stream'),
        expect.any(Object),
      )
    })
  })

  it('shows live toggle button', () => {
    render(CameraCard, { props: { id: 'abc', name: 'Test Cam' } })
    // The live toggle button contains a Video icon; title is 'Fast snapshot' when no go2rtc URL
    const button = screen.getByTitle('Fast snapshot')
    expect(button).toBeTruthy()
    // Should contain an SVG (Video icon)
    expect(button.querySelector('svg')).toBeTruthy()
  })
})
