import { render, screen, cleanup, fireEvent } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import Setup from './Setup.svelte'

describe('Setup page', () => {
  beforeEach(() => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({}),
    })
  })

  afterEach(() => {
    cleanup()
    vi.restoreAllMocks()
  })

  it('renders the setup page', () => {
    render(Setup)
    expect(screen.getByTestId('setup-page')).toBeTruthy()
  })

  it('shows 2-step progress indicator', () => {
    render(Setup)
    expect(screen.getByTestId('setup-progress')).toBeTruthy()
  })

  it('starts on admin account step', () => {
    render(Setup)
    expect(screen.getByTestId('setup-step-admin')).toBeTruthy()
    expect(screen.getByText('Compte admin')).toBeTruthy()
  })

  it('has username and password fields on step 1', () => {
    render(Setup)
    expect(screen.getByTestId('setup-username')).toBeTruthy()
    expect(screen.getByTestId('setup-password')).toBeTruthy()
    expect(screen.getByTestId('setup-password-confirm')).toBeTruthy()
  })

  it('has next button', () => {
    render(Setup)
    expect(screen.getByTestId('setup-next')).toBeTruthy()
    expect(screen.getByText('Suivant')).toBeTruthy()
  })

  it('shows error on password mismatch', async () => {
    render(Setup)
    const username = screen.getByTestId('setup-username') as HTMLInputElement
    const password = screen.getByTestId('setup-password') as HTMLInputElement
    const confirm = screen.getByTestId('setup-password-confirm') as HTMLInputElement

    await fireEvent.input(username, { target: { value: 'admin' } })
    await fireEvent.input(password, { target: { value: 'password123' } })
    await fireEvent.input(confirm, { target: { value: 'different' } })
    await fireEvent.click(screen.getByTestId('setup-next'))

    await vi.waitFor(() => {
      expect(screen.getByTestId('setup-error')).toBeTruthy()
      expect(screen.getByText('Les mots de passe ne correspondent pas')).toBeTruthy()
    })
  })

  it('shows error on short password', async () => {
    render(Setup)
    const username = screen.getByTestId('setup-username') as HTMLInputElement
    const password = screen.getByTestId('setup-password') as HTMLInputElement
    const confirm = screen.getByTestId('setup-password-confirm') as HTMLInputElement

    await fireEvent.input(username, { target: { value: 'admin' } })
    await fireEvent.input(password, { target: { value: 'short' } })
    await fireEvent.input(confirm, { target: { value: 'short' } })
    await fireEvent.click(screen.getByTestId('setup-next'))

    await vi.waitFor(() => {
      expect(screen.getByTestId('setup-error')).toBeTruthy()
    })
  })

  it('advances to category step on valid admin', async () => {
    render(Setup)
    const username = screen.getByTestId('setup-username') as HTMLInputElement
    const password = screen.getByTestId('setup-password') as HTMLInputElement
    const confirm = screen.getByTestId('setup-password-confirm') as HTMLInputElement

    await fireEvent.input(username, { target: { value: 'admin' } })
    await fireEvent.input(password, { target: { value: 'password123' } })
    await fireEvent.input(confirm, { target: { value: 'password123' } })
    await fireEvent.click(screen.getByTestId('setup-next'))

    await vi.waitFor(() => {
      expect(screen.getByTestId('setup-step-category')).toBeTruthy()
    })
  })

  it('has back button on steps > 0', async () => {
    render(Setup)
    const username = screen.getByTestId('setup-username') as HTMLInputElement
    const password = screen.getByTestId('setup-password') as HTMLInputElement
    const confirm = screen.getByTestId('setup-password-confirm') as HTMLInputElement

    await fireEvent.input(username, { target: { value: 'admin' } })
    await fireEvent.input(password, { target: { value: 'password123' } })
    await fireEvent.input(confirm, { target: { value: 'password123' } })
    await fireEvent.click(screen.getByTestId('setup-next'))

    await vi.waitFor(() => {
      expect(screen.getByTestId('setup-back')).toBeTruthy()
    })
  })
})
