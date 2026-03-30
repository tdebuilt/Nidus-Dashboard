import { render, screen, cleanup, waitFor } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import AppLinkWidget from './AppLinkWidget.svelte'
import { setLocale } from '../../i18n'

describe('AppLinkWidget', () => {
  beforeEach(() => {
    setLocale('en')
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ status: 'up' }),
    })
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('renders with data-testid', () => {
    render(AppLinkWidget, { props: { config: '{}' } })
    expect(screen.getByTestId('applink-widget')).toBeTruthy()
  })

  it('shows no links configured when config has no links', () => {
    render(AppLinkWidget, { props: { config: '{}', active: true } })
    expect(screen.getByText('No links configured')).toBeTruthy()
  })

  it('shows no links configured with empty links array', () => {
    const config = JSON.stringify({ links: [] })
    render(AppLinkWidget, { props: { config, active: true } })
    expect(screen.getByText('No links configured')).toBeTruthy()
  })

  it('renders link cards when links are configured', async () => {
    const config = JSON.stringify({
      links: [
        { name: 'Home Assistant', url: 'https://ha.local:8123' },
        { name: 'Portainer', url: 'https://portainer.local:9443' },
      ],
    })

    render(AppLinkWidget, { props: { config, active: true } })

    await waitFor(() => {
      expect(screen.getByText('Home Assistant')).toBeTruthy()
      expect(screen.getByText('Portainer')).toBeTruthy()
    })
  })

  it('renders group labels when links have groups', async () => {
    const config = JSON.stringify({
      links: [
        { name: 'Grafana', url: 'https://grafana.local', group: 'Monitoring' },
        { name: 'Prometheus', url: 'https://prom.local', group: 'Monitoring' },
        { name: 'Plex', url: 'https://plex.local', group: 'Media' },
      ],
    })

    render(AppLinkWidget, { props: { config, active: true } })

    await waitFor(() => {
      expect(screen.getByText('Monitoring')).toBeTruthy()
      expect(screen.getByText('Media')).toBeTruthy()
      expect(screen.getByText('Grafana')).toBeTruthy()
      expect(screen.getByText('Plex')).toBeTruthy()
    })
  })
})
