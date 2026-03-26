import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach, vi } from 'vitest'
import { get } from 'svelte/store'
import { toasts } from '../stores/toast'
import ToastContainer from './ToastContainer.svelte'

describe('ToastContainer', () => {
  afterEach(() => {
    cleanup()
    const current = get(toasts)
    current.forEach((t) => toasts.remove(t.id))
  })

  it('renders the container', () => {
    render(ToastContainer)
    expect(screen.getByTestId('toast-container')).toBeTruthy()
  })

  it('shows no toasts initially', () => {
    render(ToastContainer)
    expect(screen.queryAllByTestId('toast')).toHaveLength(0)
  })

  it('displays a toast when added', async () => {
    render(ToastContainer)
    toasts.success('Test message', 0)
    await vi.waitFor(() => {
      expect(screen.getAllByTestId('toast')).toHaveLength(1)
      expect(screen.getByText('Test message')).toBeTruthy()
    })
  })

  it('displays multiple toasts', async () => {
    render(ToastContainer)
    toasts.success('A', 0)
    toasts.error('B', 0)
    await vi.waitFor(() => {
      expect(screen.getAllByTestId('toast')).toHaveLength(2)
    })
  })

  it('has close buttons on toasts', async () => {
    render(ToastContainer)
    toasts.info('Close me', 0)
    await vi.waitFor(() => {
      expect(screen.getByTestId('toast-close')).toBeTruthy()
    })
  })
})
