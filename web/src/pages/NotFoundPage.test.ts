import { render, screen, cleanup, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, afterEach, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import { currentRoute, setAuthenticated } from '../lib/router';
import NotFoundPage from './NotFoundPage.svelte';

describe('NotFoundPage', () => {
  beforeEach(() => {
    setAuthenticated(true);
  });

  afterEach(() => {
    cleanup();
  });

  it('shows 404 heading', () => {
    render(NotFoundPage);
    expect(screen.getByText('404')).toBeTruthy();
  });

  it('shows page not found message', () => {
    render(NotFoundPage);
    expect(screen.getByText('Page introuvable')).toBeTruthy();
  });

  it('shows FileQuestion icon', () => {
    const { container } = render(NotFoundPage);
    const svg = container.querySelector('svg.lucide-icon');
    expect(svg).toBeTruthy();
  });

  it('has a back to dashboard button', () => {
    render(NotFoundPage);
    expect(screen.getByText('Retour au dashboard')).toBeTruthy();
  });

  it('navigates to dashboard on button click', async () => {
    render(NotFoundPage);
    await fireEvent.click(screen.getByText('Retour au dashboard'));
    expect(get(currentRoute)).toBe('dashboard');
  });
});
