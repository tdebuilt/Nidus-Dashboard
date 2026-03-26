import { writable } from 'svelte/store'

export type ConfirmOptions = {
  title: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  destructive?: boolean
}

type ConfirmState = {
  open: boolean
  options: ConfirmOptions
  resolve: ((confirmed: boolean) => void) | null
}

export const confirmState = writable<ConfirmState>({
  open: false,
  options: { title: '', message: '' },
  resolve: null,
})

export function confirm(options: ConfirmOptions): Promise<boolean> {
  return new Promise((resolve) => {
    confirmState.set({
      open: true,
      options,
      resolve,
    })
  })
}

export function resolveConfirm(confirmed: boolean) {
  confirmState.update((state) => {
    state.resolve?.(confirmed)
    return { open: false, options: { title: '', message: '' }, resolve: null }
  })
}
