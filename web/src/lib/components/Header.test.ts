import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach } from 'vitest'
import Header from './Header.svelte'

describe('Header (mobile)', () => {
  afterEach(() => cleanup())

  it('renders the mobile header', () => {
    render(Header)
    expect(screen.getByTestId('mobile-header')).toBeTruthy()
  })

  it('renders the burger button', () => {
    render(Header)
    expect(screen.getByTestId('burger-button')).toBeTruthy()
  })

  it('renders Nidus title', () => {
    render(Header)
    expect(screen.getByText('Nidus')).toBeTruthy()
  })
})
