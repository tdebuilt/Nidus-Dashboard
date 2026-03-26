import { render, screen, cleanup, waitFor } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import Sidebar from './Sidebar.svelte'

vi.mock('../api/client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
  ApiError: class extends Error {
    constructor(public status: number, message: string) {
      super(message)
    }
  },
  NetworkError: class extends Error {},
}))

import { api } from '../api/client'
import { categories } from '../stores/categories'

describe('Sidebar', () => {
  beforeEach(() => {
    vi.mocked(api.get).mockResolvedValue([])
  })

  afterEach(() => {
    cleanup()
    vi.clearAllMocks()
  })

  it('renders the sidebar element', () => {
    render(Sidebar)
    expect(screen.getByTestId('sidebar')).toBeTruthy()
  })

  it('renders the Nidus brand', () => {
    render(Sidebar)
    expect(screen.getByText('Nidus')).toBeTruthy()
  })

  it('renders Dashboard link', () => {
    render(Sidebar)
    expect(screen.getByText('Tableau de bord')).toBeTruthy()
  })

  it('renders settings link', () => {
    render(Sidebar)
    expect(screen.getByTestId('settings-link')).toBeTruthy()
    expect(screen.getByText('Paramètres')).toBeTruthy()
  })

  it('renders logout button', () => {
    render(Sidebar)
    expect(screen.getByTestId('logout-button')).toBeTruthy()
    expect(screen.getByText('Déconnexion')).toBeTruthy()
  })

  it('renders categories when loaded from API', async () => {
    vi.mocked(api.get).mockResolvedValue([
      { id: 1, name: 'Infra', icon: 'Server' },
      { id: 2, name: 'Monitoring', icon: 'Monitor' },
    ])
    await categories.load()
    render(Sidebar)
    await waitFor(() => {
      expect(screen.getByText('Infra')).toBeTruthy()
      expect(screen.getByText('Monitoring')).toBeTruthy()
      expect(screen.getByText('Catégories')).toBeTruthy()
    })
  })

  it('does not show categories heading when empty', async () => {
    vi.mocked(api.get).mockResolvedValue([])
    await categories.load()
    render(Sidebar)
    await waitFor(() => {
      expect(screen.queryByText('Catégories')).toBeNull()
    })
  })

  it('renders theme toggle', () => {
    render(Sidebar)
    expect(screen.getByTestId('theme-toggle')).toBeTruthy()
  })
})
