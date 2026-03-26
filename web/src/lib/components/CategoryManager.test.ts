import { render, screen, cleanup, fireEvent } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import CategoryManager from './CategoryManager.svelte'

describe('CategoryManager', () => {
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

  it('renders the category manager', () => {
    render(CategoryManager)
    expect(screen.getByTestId('category-manager')).toBeTruthy()
  })

  it('shows add category button', () => {
    render(CategoryManager)
    expect(screen.getByTestId('category-add-btn')).toBeTruthy()
    expect(screen.getByText('Ajouter une catégorie')).toBeTruthy()
  })

  it('shows create form when add button clicked', async () => {
    render(CategoryManager)
    await fireEvent.click(screen.getByTestId('category-add-btn'))
    expect(screen.getByTestId('category-create-form')).toBeTruthy()
    expect(screen.getByTestId('category-create-name')).toBeTruthy()
  })

  it('renders category items when provided', () => {
    render(CategoryManager, {
      props: {
        categories: [
          { id: 1, name: 'Serveurs', icon: 'server', sort_order: 0 },
          { id: 2, name: 'Réseau', icon: 'network', sort_order: 1 },
        ],
      },
    })
    const items = screen.getAllByTestId('category-item')
    expect(items).toHaveLength(2)
    expect(screen.getByText('Serveurs')).toBeTruthy()
    expect(screen.getByText('Réseau')).toBeTruthy()
  })

  it('has edit and delete buttons on category items', () => {
    render(CategoryManager, {
      props: {
        categories: [{ id: 1, name: 'Test', icon: 'folder', sort_order: 0 }],
      },
    })
    expect(screen.getByTestId('category-edit-btn')).toBeTruthy()
    expect(screen.getByTestId('category-delete-btn')).toBeTruthy()
  })

  it('shows edit form when edit button clicked', async () => {
    render(CategoryManager, {
      props: {
        categories: [{ id: 1, name: 'Test', icon: 'folder', sort_order: 0 }],
      },
    })
    await fireEvent.click(screen.getByTestId('category-edit-btn'))
    expect(screen.getByTestId('category-edit-form')).toBeTruthy()
    expect(screen.getByTestId('category-edit-name')).toBeTruthy()
  })
})
