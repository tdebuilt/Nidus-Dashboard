import { writable } from 'svelte/store';

export interface ConfirmState {
  open: boolean;
  title: string;
  message: string;
  confirmLabel: string;
  cancelLabel: string;
  destructive: boolean;
  resolve: ((value: boolean) => void) | null;
}

const initialState: ConfirmState = {
  open: false,
  title: '',
  message: '',
  confirmLabel: 'Confirmer',
  cancelLabel: 'Annuler',
  destructive: false,
  resolve: null,
};

export const confirmState = writable<ConfirmState>({ ...initialState });

export interface ConfirmOptions {
  title?: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  destructive?: boolean;
}

export function confirm(options: ConfirmOptions): Promise<boolean> {
  return new Promise((resolve) => {
    confirmState.set({
      open: true,
      title: options.title || 'Confirmation',
      message: options.message,
      confirmLabel: options.confirmLabel || 'Confirmer',
      cancelLabel: options.cancelLabel || 'Annuler',
      destructive: options.destructive ?? false,
      resolve,
    });
  });
}

export function resolveConfirm(value: boolean): void {
  confirmState.update((state) => {
    state.resolve?.(value);
    return { ...initialState };
  });
}
