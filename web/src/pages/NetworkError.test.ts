import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach } from 'vitest'
import NetworkError from './NetworkError.svelte'

describe('NetworkError page', () => {
  afterEach(() => cleanup())

  it('renders the network error page', () => {
    render(NetworkError)
    expect(screen.getByTestId('network-error-page')).toBeTruthy()
  })

  it('displays error title', () => {
    render(NetworkError)
    expect(screen.getByText('Erreur réseau')).toBeTruthy()
  })

  it('displays default message', () => {
    render(NetworkError)
    expect(screen.getByText('Impossible de contacter le serveur')).toBeTruthy()
  })

  it('displays custom message', () => {
    render(NetworkError, { props: { message: 'Timeout' } })
    expect(screen.getByText('Timeout')).toBeTruthy()
  })

  it('shows retry button when onRetry provided', () => {
    render(NetworkError, { props: { onRetry: () => {} } })
    expect(screen.getByTestId('retry-button')).toBeTruthy()
    expect(screen.getByText('Réessayer')).toBeTruthy()
  })

  it('hides retry button when no onRetry', () => {
    render(NetworkError)
    expect(screen.queryByTestId('retry-button')).toBeNull()
  })
})
