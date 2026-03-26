import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach } from 'vitest'
import NotFound from './NotFound.svelte'

describe('NotFound page', () => {
  afterEach(() => cleanup())

  it('renders the 404 page', () => {
    render(NotFound)
    expect(screen.getByTestId('not-found-page')).toBeTruthy()
  })

  it('displays 404 code', () => {
    render(NotFound)
    expect(screen.getByText('404')).toBeTruthy()
  })

  it('displays message in French', () => {
    render(NotFound)
    expect(screen.getByText('Page introuvable')).toBeTruthy()
  })

  it('has a back to dashboard button', () => {
    render(NotFound)
    expect(screen.getByText('Retour au dashboard')).toBeTruthy()
  })
})
