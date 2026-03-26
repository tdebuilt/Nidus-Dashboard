import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import WidgetGrid from './WidgetGrid.svelte'
import { editMode } from '../stores/editMode'

vi.mock('../widgetRegistry', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../widgetRegistry')>()
  return {
    ...actual,
    loadWidgetComponent: vi.fn().mockResolvedValue(undefined),
    loadConfigComponent: vi.fn().mockResolvedValue(undefined),
  }
})

describe('WidgetGrid', () => {
  beforeEach(() => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({}),
    })
    editMode.set(true)
  })

  afterEach(() => {
    cleanup()
    editMode.set(false)
    vi.restoreAllMocks()
  })

  it('renders the widget grid', () => {
    render(WidgetGrid, { props: { categoryId: 1 } })
    expect(screen.getByTestId('widget-grid')).toBeTruthy()
  })

  it('shows empty state when no widgets', () => {
    render(WidgetGrid, { props: { categoryId: 1, widgets: [] } })
    expect(screen.getByTestId('widget-grid-empty')).toBeTruthy()
    expect(screen.getByText('Aucun widget dans cette catégorie')).toBeTruthy()
  })

  it('shows add button in empty state', () => {
    render(WidgetGrid, { props: { categoryId: 1, widgets: [] } })
    expect(screen.getByTestId('widget-add-empty')).toBeTruthy()
  })

  it('renders widget cards', () => {
    render(WidgetGrid, {
      props: {
        categoryId: 1,
        widgets: [
          { id: 1, category_id: 1, type: 'docker', title: 'Docker', config: '{}', pos_x: 0, pos_y: 0, width: 2, height: 2, collapsed: false },
          { id: 2, category_id: 1, type: 'proxmox', title: 'Proxmox', config: '{}', pos_x: 2, pos_y: 0, width: 2, height: 2, collapsed: false },
        ],
      },
    })
    const cards = screen.getAllByTestId('widget-card')
    expect(cards).toHaveLength(2)
    expect(screen.getByText('Docker')).toBeTruthy()
    expect(screen.getByText('Proxmox')).toBeTruthy()
  })

  it('has delete button and resize handle on widgets', () => {
    render(WidgetGrid, {
      props: {
        categoryId: 1,
        widgets: [
          { id: 1, category_id: 1, type: 'docker', title: 'Docker', config: '{}', pos_x: 0, pos_y: 0, width: 2, height: 2, collapsed: false },
        ],
      },
    })
    expect(screen.getByTestId('widget-delete')).toBeTruthy()
    expect(screen.getByTestId('widget-resize')).toBeTruthy()
  })

  it('shows add button when widgets exist', () => {
    render(WidgetGrid, {
      props: {
        categoryId: 1,
        widgets: [
          { id: 1, category_id: 1, type: 'docker', title: 'Docker', config: '{}', pos_x: 0, pos_y: 0, width: 2, height: 2, collapsed: false },
        ],
      },
    })
    expect(screen.getByTestId('widget-add-btn')).toBeTruthy()
  })
})
