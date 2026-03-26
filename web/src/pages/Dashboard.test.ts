import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import Dashboard from './Dashboard.svelte'
import { editMode } from '../lib/stores/editMode'

describe('Dashboard page', () => {
  beforeEach(() => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve([]),
    })
    editMode.set(true)
  })

  afterEach(() => {
    cleanup()
    editMode.set(false)
    vi.restoreAllMocks()
  })

  it('renders the dashboard page', () => {
    render(Dashboard)
    expect(screen.getByTestId('dashboard-page')).toBeTruthy()
  })

  it('shows empty state when no categories', async () => {
    render(Dashboard)
    await vi.waitFor(() => {
      expect(screen.getByText('Aucune catégorie')).toBeTruthy()
      expect(screen.getByTestId('create-first-category')).toBeTruthy()
    })
  })

  it('shows loading initially', () => {
    // Delay the fetch response
    globalThis.fetch = vi.fn().mockReturnValue(new Promise(() => {}))
    render(Dashboard)
    expect(screen.getByText('Chargement...')).toBeTruthy()
  })

  it('shows category tabs when categories exist', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve([
          { id: 1, name: 'Serveurs', slug: 'serveurs', icon: 'server', sort_order: 0 },
          { id: 2, name: 'Réseau', slug: 'reseau', icon: 'network', sort_order: 1 },
        ]),
    })
    render(Dashboard)
    await vi.waitFor(() => {
      const tabs = screen.getAllByTestId('category-tab')
      expect(tabs).toHaveLength(2)
      expect(screen.getByText('Serveurs')).toBeTruthy()
    })
  })

  it('has add category button when categories exist', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve([{ id: 1, name: 'Test', slug: 'test', icon: 'folder', sort_order: 0 }]),
    })
    render(Dashboard)
    await vi.waitFor(() => {
      expect(screen.getByTestId('category-add-btn')).toBeTruthy()
    })
  })
})
