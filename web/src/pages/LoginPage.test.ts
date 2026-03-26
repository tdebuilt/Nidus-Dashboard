import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor, cleanup } from '@testing-library/svelte';
import LoginPage from './LoginPage.svelte';

describe('LoginPage', () => {
  beforeEach(() => {
    window.history.replaceState({}, '', '/login');
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it('renders login form with username and password fields', () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('{}', { status: 200 }));
    render(LoginPage);

    expect(screen.getByText('Connexion')).toBeTruthy();
    expect(screen.getByLabelText("Nom d'utilisateur")).toBeTruthy();
    expect(screen.getByLabelText('Mot de passe')).toBeTruthy();
    expect(screen.getByText('Se connecter')).toBeTruthy();
  });

  it('disables submit button when fields are empty', () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('{}', { status: 200 }));
    render(LoginPage);

    const button = screen.getByText('Se connecter') as HTMLButtonElement;
    expect(button.disabled).toBe(true);
  });

  it('submits login and redirects on success', async () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ message: 'login successful', user: { id: 1, username: 'admin' } }), { status: 200 })
    );
    render(LoginPage);

    await fireEvent.input(screen.getByLabelText("Nom d'utilisateur"), { target: { value: 'admin' } });
    await fireEvent.input(screen.getByLabelText('Mot de passe'), { target: { value: 'password123' } });
    await fireEvent.click(screen.getByText('Se connecter'));

    await waitFor(() => {
      expect(fetchSpy).toHaveBeenCalledWith('/api/auth/login', expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ username: 'admin', password: 'password123' }),
      }));
    });
  });

  it('shows error message on invalid credentials', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ error: 'invalid credentials' }), { status: 401 })
    );
    render(LoginPage);

    await fireEvent.input(screen.getByLabelText("Nom d'utilisateur"), { target: { value: 'admin' } });
    await fireEvent.input(screen.getByLabelText('Mot de passe'), { target: { value: 'wrong' } });
    await fireEvent.click(screen.getByText('Se connecter'));

    await waitFor(() => {
      expect(screen.getByText('Identifiants incorrects')).toBeTruthy();
    });
  });

  it('shows TOTP field when totp_required', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ error: 'totp_required' }), { status: 401 })
    );
    render(LoginPage);

    expect(screen.queryByLabelText('Code 2FA')).toBeNull();

    await fireEvent.input(screen.getByLabelText("Nom d'utilisateur"), { target: { value: 'admin' } });
    await fireEvent.input(screen.getByLabelText('Mot de passe'), { target: { value: 'password123' } });
    await fireEvent.click(screen.getByText('Se connecter'));

    await waitFor(() => {
      expect(screen.getByLabelText('Code 2FA')).toBeTruthy();
    });
  });

  it('shows rate limit error on 429', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ error: 'too many requests' }), { status: 429 })
    );
    render(LoginPage);

    await fireEvent.input(screen.getByLabelText("Nom d'utilisateur"), { target: { value: 'admin' } });
    await fireEvent.input(screen.getByLabelText('Mot de passe'), { target: { value: 'pass1234' } });
    await fireEvent.click(screen.getByText('Se connecter'));

    await waitFor(() => {
      expect(screen.getByText('Trop de tentatives, réessayez plus tard')).toBeTruthy();
    });
  });

  it('shows network error message', async () => {
    vi.spyOn(globalThis, 'fetch').mockRejectedValue(new TypeError('Failed to fetch'));
    render(LoginPage);

    await fireEvent.input(screen.getByLabelText("Nom d'utilisateur"), { target: { value: 'admin' } });
    await fireEvent.input(screen.getByLabelText('Mot de passe'), { target: { value: 'pass1234' } });
    await fireEvent.click(screen.getByText('Se connecter'));

    await waitFor(() => {
      expect(screen.getByText('Impossible de contacter le serveur')).toBeTruthy();
    });
  });
});
