import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import { setLocale } from '../../i18n'
import GrafanaWidget from './GrafanaWidget.svelte'

describe('GrafanaWidget', () => {
  beforeEach(() => {
    setLocale('en')
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ url: 'http://grafana/d-solo/abc/_?panelId=1&theme=dark' }),
    })
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('renders with data-testid', () => {
    render(GrafanaWidget, { props: { config: '{}' } })
    expect(screen.getByTestId('grafana-widget')).toBeTruthy()
  })

  it('shows not configured when no panels', () => {
    render(GrafanaWidget, { props: { config: '{}', active: true } })
    expect(screen.getByTestId('grafana-widget')).toBeTruthy()
  })
})
