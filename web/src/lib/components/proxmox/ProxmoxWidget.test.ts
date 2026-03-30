import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import ProxmoxWidget from './ProxmoxWidget.svelte'
import { setLocale } from '../../i18n'

describe('ProxmoxWidget', () => {
  beforeEach(() => {
    setLocale('en')
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve([]),
    })
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('renders with data-testid', () => {
    render(ProxmoxWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('proxmox-widget')).toBeTruthy()
  })

  it('shows loading state when inactive', () => {
    render(ProxmoxWidget, { props: { config: '{}', active: false } })
    expect(screen.getByText('Loading VMs...')).toBeTruthy()
  })

  it('accepts config prop as JSON string', () => {
    render(ProxmoxWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('proxmox-widget')).toBeTruthy()
  })
})
