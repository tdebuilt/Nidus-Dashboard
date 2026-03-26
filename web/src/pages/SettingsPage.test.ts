import { render, screen, cleanup, fireEvent } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import SettingsPage from './SettingsPage.svelte'
import { auth } from '../lib/stores/auth'

describe('SettingsPage', () => {
  beforeEach(() => {
    // Set admin role so all sections are visible
    auth.set({ authenticated: true, setupCompleted: true, loading: false, role: 'admin' })

    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      if (typeof url === 'string' && url.includes('/api/services')) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: () => Promise.resolve([]),
        })
      }
      if (typeof url === 'string' && url.includes('/api/users')) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: () => Promise.resolve([]),
        })
      }
      if (typeof url === 'string' && url.includes('/api/invites')) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: () => Promise.resolve([]),
        })
      }
      if (typeof url === 'string' && url.includes('/api/notifications/')) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: () => Promise.resolve([]),
        })
      }
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ theme: 'dark', language: 'fr', refresh_interval: 30 }),
      })
    })
  })

  afterEach(() => {
    cleanup()
    vi.restoreAllMocks()
  })

  it('renders the settings page', () => {
    render(SettingsPage)
    expect(screen.getByTestId('settings-page')).toBeTruthy()
    expect(screen.getByText('Paramètres')).toBeTruthy()
  })

  it('has pill navigation', () => {
    render(SettingsPage)
    expect(screen.getByTestId('settings-nav')).toBeTruthy()
    expect(screen.getByTestId('settings-pill-preferences')).toBeTruthy()
    expect(screen.getByTestId('settings-pill-appearance')).toBeTruthy()
  })

  it('shows preferences tab by default', () => {
    render(SettingsPage)
    expect(screen.getByTestId('settings-language')).toBeTruthy()
    expect(screen.getByTestId('settings-language-select')).toBeTruthy()
  })

  it('has language section', () => {
    render(SettingsPage)
    expect(screen.getByTestId('settings-language')).toBeTruthy()
    expect(screen.getByTestId('settings-language-select')).toBeTruthy()
  })

  it('has refresh interval section', () => {
    render(SettingsPage)
    expect(screen.getByTestId('settings-refresh')).toBeTruthy()
    expect(screen.getByTestId('settings-refresh-select')).toBeTruthy()
  })

  it('has theme section in appearance tab', async () => {
    render(SettingsPage)
    await fireEvent.click(screen.getByTestId('settings-pill-appearance'))
    expect(screen.getByTestId('settings-theme')).toBeTruthy()
    expect(screen.getByTestId('settings-theme-select')).toBeTruthy()
  })

  it('has services section in services tab', async () => {
    render(SettingsPage)
    await fireEvent.click(screen.getByTestId('settings-pill-services'))
    expect(screen.getByTestId('settings-services')).toBeTruthy()
  })

  it('shows empty state when no services configured', async () => {
    render(SettingsPage)
    await fireEvent.click(screen.getByTestId('settings-pill-services'))
    expect(screen.getByTestId('services-empty')).toBeTruthy()
    expect(screen.getByTestId('service-add-btn')).toBeTruthy()
  })

  it('has 2FA section in account tab', async () => {
    render(SettingsPage)
    await fireEvent.click(screen.getByTestId('settings-pill-account'))
    expect(screen.getByTestId('settings-2fa')).toBeTruthy()
    expect(screen.getByTestId('totp-setup-btn')).toBeTruthy()
  })

  it('has export/import section in backup tab', async () => {
    render(SettingsPage)
    await fireEvent.click(screen.getByTestId('settings-pill-backup'))
    expect(screen.getByTestId('settings-export-import')).toBeTruthy()
    expect(screen.getByTestId('settings-export-btn')).toBeTruthy()
    expect(screen.getByTestId('settings-import-btn')).toBeTruthy()
  })

  it('has notifications section in notifications tab', async () => {
    render(SettingsPage)
    await fireEvent.click(screen.getByTestId('settings-pill-notifications'))
    expect(screen.getByTestId('settings-notifications')).toBeTruthy()
  })

  it('hides admin pills for non-admin users', () => {
    auth.set({ authenticated: true, setupCompleted: true, loading: false, role: 'viewer' })
    render(SettingsPage)
    expect(screen.queryByTestId('settings-pill-services')).toBeNull()
    expect(screen.queryByTestId('settings-pill-notifications')).toBeNull()
    expect(screen.queryByTestId('settings-pill-webhooks')).toBeNull()
    expect(screen.queryByTestId('settings-pill-backup')).toBeNull()
    expect(screen.queryByTestId('settings-pill-users')).toBeNull()
    // Non-admin tabs should still be visible
    expect(screen.getByTestId('settings-pill-appearance')).toBeTruthy()
    expect(screen.getByTestId('settings-pill-preferences')).toBeTruthy()
    expect(screen.getByTestId('settings-pill-account')).toBeTruthy()
  })
})
