import { render, screen, cleanup, fireEvent } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import Login from './Login.svelte'

describe('Login page', () => {
  beforeEach(() => {
    globalThis.fetch = vi.fn()
  })

  afterEach(() => {
    cleanup()
    vi.restoreAllMocks()
  })

  it('renders the login form', () => {
    render(Login)
    expect(screen.getByTestId('login-form')).toBeTruthy()
    expect(screen.getByText('Nidus')).toBeTruthy()
  })

  it('has username and password fields', () => {
    render(Login)
    expect(screen.getByTestId('login-username')).toBeTruthy()
    expect(screen.getByTestId('login-password')).toBeTruthy()
  })

  it('has a submit button', () => {
    render(Login)
    expect(screen.getByTestId('login-submit')).toBeTruthy()
    expect(screen.getByText('Se connecter')).toBeTruthy()
  })

  it('does not show TOTP field by default', () => {
    render(Login)
    expect(screen.queryByTestId('totp-field')).toBeNull()
  })

  it('shows error on failed login', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 400,
      json: () => Promise.resolve({ error: 'Identifiants invalides' }),
    })

    render(Login)
    const username = screen.getByTestId('login-username') as HTMLInputElement
    const password = screen.getByTestId('login-password') as HTMLInputElement

    await fireEvent.input(username, { target: { value: 'admin' } })
    await fireEvent.input(password, { target: { value: 'wrongpass' } })
    await fireEvent.submit(screen.getByTestId('login-form'))

    // Wait for async
    await vi.waitFor(() => {
      expect(screen.getByTestId('login-error')).toBeTruthy()
    })
  })

  it('shows TOTP field when 2FA required', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 400,
      json: () => Promise.resolve({ error: 'TOTP code required' }),
    })

    render(Login)
    const username = screen.getByTestId('login-username') as HTMLInputElement
    const password = screen.getByTestId('login-password') as HTMLInputElement

    await fireEvent.input(username, { target: { value: 'admin' } })
    await fireEvent.input(password, { target: { value: 'password' } })
    await fireEvent.submit(screen.getByTestId('login-form'))

    await vi.waitFor(() => {
      expect(screen.getByTestId('totp-field')).toBeTruthy()
    })
  })
})
