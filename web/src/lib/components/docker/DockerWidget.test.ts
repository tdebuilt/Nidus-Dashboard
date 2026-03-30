import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import DockerWidget from './DockerWidget.svelte'
import { setLocale } from '../../i18n'

describe('DockerWidget', () => {
  beforeEach(() => {
    setLocale('en')
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ stacks: [], standalone: [] }),
    })
  })

  afterEach(() => {
    cleanup()
    setLocale('fr')
    vi.restoreAllMocks()
  })

  it('renders with data-testid', () => {
    render(DockerWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('docker-widget')).toBeTruthy()
  })

  it('shows loading state when inactive', () => {
    render(DockerWidget, { props: { config: '{}', active: false } })
    expect(screen.getByTestId('docker-widget')).toBeTruthy()
    expect(screen.getByText('Loading containers...')).toBeTruthy()
  })

  it('accepts config prop as JSON string', () => {
    render(DockerWidget, { props: { config: '{"env_id":1}', active: false } })
    expect(screen.getByTestId('docker-widget')).toBeTruthy()
  })
})
