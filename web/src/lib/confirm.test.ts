import { describe, it, expect, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import { confirmState, confirm, resolveConfirm } from './stores/confirm';

describe('Confirm', () => {
  beforeEach(() => {
    resolveConfirm(false);
  });

  it('opens confirm dialog with options', () => {
    confirm({ title: 'Confirmation', message: 'Are you sure?' });
    const state = get(confirmState);
    expect(state.open).toBe(true);
    expect(state.options.message).toBe('Are you sure?');
    expect(state.options.title).toBe('Confirmation');
  });

  it('uses custom title and labels', () => {
    confirm({
      title: 'Delete item',
      message: 'This cannot be undone',
      confirmLabel: 'Delete',
      cancelLabel: 'Keep',
      destructive: true,
    });
    const state = get(confirmState);
    expect(state.options.title).toBe('Delete item');
    expect(state.options.confirmLabel).toBe('Delete');
    expect(state.options.cancelLabel).toBe('Keep');
    expect(state.options.destructive).toBe(true);
  });

  it('resolves true when confirmed', async () => {
    const promise = confirm({ title: 'Confirm', message: 'Confirm?' });
    resolveConfirm(true);
    const result = await promise;
    expect(result).toBe(true);
    expect(get(confirmState).open).toBe(false);
  });

  it('resolves false when cancelled', async () => {
    const promise = confirm({ title: 'Confirm', message: 'Confirm?' });
    resolveConfirm(false);
    const result = await promise;
    expect(result).toBe(false);
    expect(get(confirmState).open).toBe(false);
  });
});
