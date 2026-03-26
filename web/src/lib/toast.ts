import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'info';

export interface Toast {
  id: number;
  type: ToastType;
  message: string;
}

let nextId = 0;

export const toasts = writable<Toast[]>([]);

function addToast(type: ToastType, message: string, duration = 4000): void {
  const id = nextId++;
  toasts.update((all) => [...all, { id, type, message }]);

  setTimeout(() => {
    removeToast(id);
  }, duration);
}

function removeToast(id: number): void {
  toasts.update((all) => all.filter((t) => t.id !== id));
}

export const toast = {
  success: (message: string) => addToast('success', message),
  error: (message: string) => addToast('error', message),
  info: (message: string) => addToast('info', message),
  remove: removeToast,
};
