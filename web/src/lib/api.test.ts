import { describe, it, expect, beforeEach, vi } from 'vitest';
import { api, ApiError, NetworkError } from './api/client';

describe('API client', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('makes GET requests with correct options', async () => {
    const mockResponse = { data: 'test' };
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify(mockResponse), { status: 200 })
    );

    const result = await api.get('/api/test');
    expect(result).toEqual(mockResponse);
    expect(fetch).toHaveBeenCalledWith('/api/test', expect.objectContaining({
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
    }));
  });

  it('makes POST requests with body', async () => {
    const body = { username: 'admin', password: 'secret' };
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), { status: 200 })
    );

    await api.post('/api/auth/login', body);
    expect(fetch).toHaveBeenCalledWith('/api/auth/login', expect.objectContaining({
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(body),
    }));
  });

  it('makes PUT requests with body', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), { status: 200 })
    );

    await api.put('/api/settings', { theme: 'dark' });
    expect(fetch).toHaveBeenCalledWith('/api/settings', expect.objectContaining({
      method: 'PUT',
      body: JSON.stringify({ theme: 'dark' }),
    }));
  });

  it('makes DELETE requests', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ message: 'deleted' }), { status: 200 })
    );

    await api.delete('/api/categories/1');
    expect(fetch).toHaveBeenCalledWith('/api/categories/1', expect.objectContaining({
      method: 'DELETE',
    }));
  });

  it('throws ApiError on 401', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ error: 'Unauthorized' }), { status: 401 })
    );

    await expect(api.get('/api/settings')).rejects.toThrow(ApiError);
    await expect(api.get('/api/settings')).rejects.toThrow('Unauthorized');
  });

  it('throws ApiError with server error message', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ error: 'name is required' }), { status: 400 })
    );

    try {
      await api.post('/api/categories', {});
      expect.fail('should have thrown');
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError);
      expect((e as ApiError).message).toBe('name is required');
      expect((e as ApiError).status).toBe(400);
    }
  });

  it('throws NetworkError when fetch fails', async () => {
    vi.spyOn(globalThis, 'fetch').mockRejectedValue(new TypeError('Failed to fetch'));

    await expect(api.get('/api/health')).rejects.toThrow(NetworkError);
  });

  it('includes credentials for cookie-based auth', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({}), { status: 200 })
    );

    await api.get('/api/categories');
    const callArgs = vi.mocked(fetch).mock.calls[0];
    expect((callArgs[1] as RequestInit).credentials).toBe('include');
  });
});
