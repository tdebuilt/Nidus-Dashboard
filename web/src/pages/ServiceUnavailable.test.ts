import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach } from 'vitest'
import ServiceUnavailable from './ServiceUnavailable.svelte'

describe('ServiceUnavailable page', () => {
  afterEach(() => cleanup())

  it('renders the page', () => {
    render(ServiceUnavailable)
    expect(screen.getByTestId('service-unavailable-page')).toBeTruthy()
  })

  it('displays title', () => {
    render(ServiceUnavailable)
    expect(screen.getByText('Service indisponible')).toBeTruthy()
  })

  it('displays default message', () => {
    render(ServiceUnavailable)
    expect(screen.getByText('Le service est temporairement indisponible')).toBeTruthy()
  })

  it('shows retry button when onRetry provided', () => {
    render(ServiceUnavailable, { props: { onRetry: () => {} } })
    expect(screen.getByTestId('retry-button')).toBeTruthy()
  })

  it('hides retry button when no onRetry', () => {
    render(ServiceUnavailable)
    expect(screen.queryByTestId('retry-button')).toBeNull()
  })
})
