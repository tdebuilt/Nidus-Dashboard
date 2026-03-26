import { render, screen, cleanup, waitFor } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import ReolinkWidget from './ReolinkWidget.svelte'
import { setLocale } from '../../i18n'

describe('ReolinkWidget', () => {
  beforeEach(() => {
    Object.defineProperty(globalThis, 'crypto', {
      value: {
        subtle: {
          digest: vi.fn().mockResolvedValue(
            new Uint8Array([
              0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
              0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
              0, 0, 0, 0, 0, 0, 0, 0,
            ]).buffer,
          ),
        },
      },
      writable: true,
    })

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ go2rtc: null }),
    })

    setLocale('en')
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('renders empty state when no cameras configured', () => {
    render(ReolinkWidget, { props: { config: '{}' } })
    expect(screen.getByTestId('reolink-widget')).toBeTruthy()
    expect(screen.getByText('No cameras configured')).toBeTruthy()
  })

  it('renders empty state with invalid JSON config', () => {
    render(ReolinkWidget, { props: { config: 'invalid' } })
    expect(screen.getByTestId('reolink-widget')).toBeTruthy()
    expect(screen.getByText('No cameras configured')).toBeTruthy()
  })

  it('renders cameras when configured', async () => {
    const config = JSON.stringify({
      cameras: [
        { name: 'Front Door', ip: '192.168.1.100', username: 'admin', password: 'pass', channel: 0, source: 'direct' },
      ],
      columns: 2,
    })

    render(ReolinkWidget, { props: { config, active: true } })

    await waitFor(() => {
      expect(screen.getByText('Front Door')).toBeTruthy()
    })
  })

  it('uses configured columns', async () => {
    const config = JSON.stringify({
      cameras: [
        { name: 'Cam1', ip: '192.168.1.100', username: 'admin', password: 'pass', channel: 0, source: 'direct' },
      ],
      columns: 3,
    })

    render(ReolinkWidget, { props: { config, active: true } })

    await waitFor(() => {
      expect(screen.getByText('Cam1')).toBeTruthy()
    })

    const grid = screen.getByTestId('reolink-widget').querySelector('.grid')
    expect(grid).toBeTruthy()
    expect(grid!.getAttribute('style')).toContain('repeat(3, 1fr)')
  })
})
