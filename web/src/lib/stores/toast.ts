import { writable } from 'svelte/store'

export type ToastType = 'success' | 'error' | 'info'

export type Toast = {
  id: number
  type: ToastType
  message: string
}

let nextId = 0

function createToastStore() {
  const { subscribe, update } = writable<Toast[]>([])

  function add(type: ToastType, message: string, duration = 4000) {
    const id = nextId++
    update((toasts) => [...toasts, { id, type, message }])
    if (duration > 0) {
      setTimeout(() => remove(id), duration)
    }
    return id
  }

  function remove(id: number) {
    update((toasts) => toasts.filter((t) => t.id !== id))
  }

  return {
    subscribe,
    success: (message: string, duration?: number) => add('success', message, duration),
    error: (message: string, duration?: number) => add('error', message, duration),
    info: (message: string, duration?: number) => add('info', message, duration),
    remove,
  }
}

export const toasts = createToastStore()
