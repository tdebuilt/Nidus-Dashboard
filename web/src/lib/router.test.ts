import { describe, it, expect, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import {
  currentRoute,
  navigate,
  setAuthenticated,
  redirectToLogin,
  redirectToDashboard,
  initRouter,
} from './router';

describe('Router', () => {
  beforeEach(() => {
    setAuthenticated(false);
    window.history.replaceState({}, '', '/');
    currentRoute.set('dashboard');
  });

  it('navigates to login when not authenticated', () => {
    navigate('dashboard');
    expect(get(currentRoute)).toBe('login');
  });

  it('navigates to protected routes when authenticated', () => {
    setAuthenticated(true);
    navigate('settings');
    expect(get(currentRoute)).toBe('settings');
  });

  it('allows navigation to public routes without auth', () => {
    navigate('login');
    expect(get(currentRoute)).toBe('login');

    navigate('setup');
    expect(get(currentRoute)).toBe('setup');
  });

  it('updates browser URL on navigate', () => {
    setAuthenticated(true);
    navigate('settings');
    expect(window.location.pathname).toBe('/settings');
  });

  it('uses / for dashboard route', () => {
    setAuthenticated(true);
    navigate('dashboard');
    expect(window.location.pathname).toBe('/');
  });

  it('redirectToLogin sets route to login', () => {
    redirectToLogin();
    expect(get(currentRoute)).toBe('login');
  });

  it('redirectToDashboard navigates to dashboard when authenticated', () => {
    setAuthenticated(true);
    redirectToDashboard();
    expect(get(currentRoute)).toBe('dashboard');
  });

  it('handles popstate event (browser back/forward)', () => {
    setAuthenticated(true);
    const cleanup = initRouter();

    navigate('settings');
    expect(get(currentRoute)).toBe('settings');

    // Simulate browser back
    window.history.replaceState({}, '', '/');
    window.dispatchEvent(new PopStateEvent('popstate'));
    expect(get(currentRoute)).toBe('dashboard');

    cleanup();
  });

  it('redirects to login on popstate for protected route when not authenticated', () => {
    setAuthenticated(false);
    const cleanup = initRouter();

    window.history.replaceState({}, '', '/settings');
    window.dispatchEvent(new PopStateEvent('popstate'));
    expect(get(currentRoute)).toBe('login');

    cleanup();
  });

  it('initRouter redirects unauthenticated user to login', () => {
    currentRoute.set('dashboard');
    setAuthenticated(false);
    const cleanup = initRouter();
    expect(get(currentRoute)).toBe('login');
    cleanup();
  });

  it('resolves unknown paths to not-found route', () => {
    setAuthenticated(true);
    window.history.replaceState({}, '', '/totally-unknown');
    const cleanup = initRouter();
    window.dispatchEvent(new PopStateEvent('popstate'));
    expect(get(currentRoute)).toBe('not-found');
    cleanup();
  });
});
