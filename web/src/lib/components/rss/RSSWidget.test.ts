import { render, screen, cleanup, waitFor } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import RSSWidget from './RSSWidget.svelte'
import { setLocale } from '../../i18n'

describe('RSSWidget', () => {
  beforeEach(() => {
    setLocale('en')
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ items: [] }),
    })
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('renders with data-testid', () => {
    render(RSSWidget, { props: { config: '{}' } })
    expect(screen.getByTestId('rss-widget')).toBeTruthy()
  })

  it('shows not configured when no URLs are set', async () => {
    render(RSSWidget, { props: { config: '{}', active: true } })

    await waitFor(() => {
      expect(screen.getByText('RSS widget not configured')).toBeTruthy()
    })
  })

  it('shows no articles when feed returns empty items', async () => {
    const config = JSON.stringify({ urls: ['https://example.com/feed.xml'] })
    render(RSSWidget, { props: { config, active: true } })

    await waitFor(() => {
      expect(screen.getByText('No articles')).toBeTruthy()
    })
  })

  it('renders feed items when data loads', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          items: [
            {
              title: 'Breaking News: Test Passes',
              link: 'https://example.com/article-1',
              source: 'Example Blog',
              published: new Date().toISOString(),
              author: 'Jane Doe',
            },
            {
              title: 'Second Article',
              link: 'https://example.com/article-2',
              source: 'Tech News',
            },
          ],
        }),
    })

    const config = JSON.stringify({ urls: ['https://example.com/feed.xml'] })
    render(RSSWidget, { props: { config, active: true } })

    await waitFor(() => {
      expect(screen.getByText('Breaking News: Test Passes')).toBeTruthy()
      expect(screen.getByText('Second Article')).toBeTruthy()
      expect(screen.getByText('Example Blog')).toBeTruthy()
    })
  })

  it('renders article links with correct href', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          items: [
            {
              title: 'Linked Article',
              link: 'https://example.com/linked',
            },
          ],
        }),
    })

    const config = JSON.stringify({ urls: ['https://example.com/feed.xml'] })
    render(RSSWidget, { props: { config, active: true } })

    await waitFor(() => {
      const link = screen.getByText('Linked Article').closest('a')
      expect(link).toBeTruthy()
      expect(link!.getAttribute('href')).toBe('https://example.com/linked')
      expect(link!.getAttribute('target')).toBe('_blank')
    })
  })
})
