import { render, screen, cleanup } from '@testing-library/svelte'
import { describe, it, expect, afterEach } from 'vitest'
import { confirmState } from '../stores/confirm'
import ConfirmDialog from './ConfirmDialog.svelte'

describe('ConfirmDialog', () => {
  afterEach(() => {
    cleanup()
    confirmState.set({ open: false, options: { title: '', message: '' }, resolve: null })
  })

  it('does not render when closed', () => {
    render(ConfirmDialog)
    expect(screen.queryByTestId('confirm-dialog')).toBeNull()
  })

  it('renders when open', () => {
    confirmState.set({
      open: true,
      options: { title: 'Confirmer', message: 'Êtes-vous sûr ?' },
      resolve: () => {},
    })
    render(ConfirmDialog)
    expect(screen.getByTestId('confirm-dialog')).toBeTruthy()
    expect(screen.getByText('Êtes-vous sûr ?')).toBeTruthy()
  })

  it('has cancel and confirm buttons', () => {
    confirmState.set({
      open: true,
      options: { title: 'T', message: 'M' },
      resolve: () => {},
    })
    render(ConfirmDialog)
    expect(screen.getByTestId('confirm-cancel')).toBeTruthy()
    expect(screen.getByTestId('confirm-ok')).toBeTruthy()
  })

  it('shows custom button labels', () => {
    confirmState.set({
      open: true,
      options: { title: 'T', message: 'M', confirmLabel: 'Supprimer', cancelLabel: 'Non' },
      resolve: () => {},
    })
    render(ConfirmDialog)
    expect(screen.getByText('Supprimer')).toBeTruthy()
    expect(screen.getByText('Non')).toBeTruthy()
  })
})
