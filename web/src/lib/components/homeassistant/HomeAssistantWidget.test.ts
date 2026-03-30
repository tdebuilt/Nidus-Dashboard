import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import HomeAssistantWidget from './HomeAssistantWidget.svelte'
import { setLocale } from '../../i18n'

describe('HomeAssistantWidget', () => {
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
    render(HomeAssistantWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('homeassistant-widget')).toBeTruthy()
  })

  it('shows loading state when inactive', () => {
    render(HomeAssistantWidget, { props: { config: '{}', active: false } })
    expect(screen.getByText('Loading entities...')).toBeTruthy()
  })

  it('accepts config prop as JSON string', () => {
    render(HomeAssistantWidget, { props: { config: '{"entity_ids":["light.living"]}', active: false } })
    expect(screen.getByTestId('homeassistant-widget')).toBeTruthy()
  })
})
