import { render, screen, cleanup, fireEvent } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest'
import AddWidgetDialog from './AddWidgetDialog.svelte'

const mockServices = [
  { type: 'portainer', enabled: true },
  { type: 'proxmox', enabled: true },
  { type: 'homeassistant', enabled: true },
  { type: 'adguard', enabled: true },
  { type: 'jdownloader', enabled: true },
  { type: 'transmission', enabled: true },
  { type: 'uptimekuma', enabled: true },
]

describe('AddWidgetDialog', () => {
  beforeEach(() => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockServices),
    })
  })

  afterEach(() => {
    cleanup()
    vi.restoreAllMocks()
  })

  it('does not render when closed', () => {
    render(AddWidgetDialog, { props: { categoryId: 1 } })
    expect(screen.queryByTestId('add-widget-dialog')).toBeNull()
  })

  it('renders when open', () => {
    render(AddWidgetDialog, { props: { categoryId: 1, open: true } })
    expect(screen.getByTestId('add-widget-dialog')).toBeTruthy()
    expect(screen.getByText('Ajouter un widget')).toBeTruthy()
  })

  it('shows widget type grid', async () => {
    render(AddWidgetDialog, { props: { categoryId: 1, open: true } })
    await vi.waitFor(() => {
      expect(screen.getByTestId('widget-type-grid')).toBeTruthy()
    })
  })

  it('shows 7 widget types', async () => {
    render(AddWidgetDialog, { props: { categoryId: 1, open: true } })
    await vi.waitFor(() => {
      expect(screen.getByTestId('widget-type-docker')).toBeTruthy()
      expect(screen.getByTestId('widget-type-proxmox')).toBeTruthy()
      expect(screen.getByTestId('widget-type-homeassistant')).toBeTruthy()
      expect(screen.getByTestId('widget-type-adguard')).toBeTruthy()
      expect(screen.getByTestId('widget-type-jdownloader')).toBeTruthy()
      expect(screen.getByTestId('widget-type-transmission')).toBeTruthy()
      expect(screen.getByTestId('widget-type-uptimekuma')).toBeTruthy()
    })
  })

  it('shows config form after selecting type', async () => {
    render(AddWidgetDialog, { props: { categoryId: 1, open: true } })
    await vi.waitFor(() => {
      expect(screen.getByTestId('widget-type-docker')).toBeTruthy()
    })
    await fireEvent.click(screen.getByTestId('widget-type-docker'))
    expect(screen.getByTestId('widget-title-input')).toBeTruthy()
    expect(screen.getByTestId('widget-create')).toBeTruthy()
    expect(screen.getByTestId('widget-back')).toBeTruthy()
  })

  it('has back button on config step', async () => {
    render(AddWidgetDialog, { props: { categoryId: 1, open: true } })
    await vi.waitFor(() => {
      expect(screen.getByTestId('widget-type-proxmox')).toBeTruthy()
    })
    await fireEvent.click(screen.getByTestId('widget-type-proxmox'))
    expect(screen.getByTestId('widget-back')).toBeTruthy()
  })

  it('has close button', () => {
    render(AddWidgetDialog, { props: { categoryId: 1, open: true } })
    expect(screen.getByTestId('add-widget-close')).toBeTruthy()
  })
})
