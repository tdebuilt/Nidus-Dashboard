import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { get } from 'svelte/store';
import { toasts, toast } from './toast';

describe('Toast', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    toasts.set([]);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('adds success toast', () => {
    toast.success('Operation successful');
    const items = get(toasts);
    expect(items).toHaveLength(1);
    expect(items[0].type).toBe('success');
    expect(items[0].message).toBe('Operation successful');
  });

  it('adds error toast', () => {
    toast.error('Something failed');
    const items = get(toasts);
    expect(items).toHaveLength(1);
    expect(items[0].type).toBe('error');
    expect(items[0].message).toBe('Something failed');
  });

  it('adds info toast', () => {
    toast.info('FYI');
    const items = get(toasts);
    expect(items).toHaveLength(1);
    expect(items[0].type).toBe('info');
  });

  it('auto-removes toast after timeout', () => {
    toast.success('Temporary');
    expect(get(toasts)).toHaveLength(1);

    vi.advanceTimersByTime(4000);
    expect(get(toasts)).toHaveLength(0);
  });

  it('manually removes toast', () => {
    toast.success('To remove');
    const items = get(toasts);
    expect(items).toHaveLength(1);

    toast.remove(items[0].id);
    expect(get(toasts)).toHaveLength(0);
  });

  it('supports multiple toasts simultaneously', () => {
    toast.success('First');
    toast.error('Second');
    toast.info('Third');
    expect(get(toasts)).toHaveLength(3);
  });

  it('assigns unique IDs', () => {
    toast.success('A');
    toast.success('B');
    const items = get(toasts);
    expect(items[0].id).not.toBe(items[1].id);
  });
});
