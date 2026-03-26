import { describe, it, expect, beforeEach, vi } from 'vitest';
import { api, ApiRequestError, setUnauthorizedHandler } from './api';

describe('API client', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    setUnauthorizedHandler(null as unknown as () => void);
  });

  it('makes GET requests with correct options', async () => {
    const mockResponse = { data: 'test' };
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify(mockResponse), { status: 200 })
    );

    const result = await api.get('/api/test');
    expect(result).toEqual(mockResponse);
    expect(fetch).toHaveBeenCalledWith('/api/test', {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
    });
  });

  it('makes POST requests with body', async () => {
    const body = { username: 'admin', password: 'secret' };
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), { status: 200 })
    );

    await api.post('/api/auth/login', body);
    expect(fetch).toHaveBeenCalledWith('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(body),
    });
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

  it('throws ApiRequestError on 401 and calls unauthorized handler', async () => {
    const handler = vi.fn();
    setUnauthorizedHandler(handler);
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ error: 'Unauthorized' }), { status: 401 })
    );

    await expect(api.get('/api/settings')).rejects.toThrow(ApiRequestError);
    await expect(api.get('/api/settings')).rejects.toThrow('Unauthorized');
    expect(handler).toHaveBeenCalled();
  });

  it('throws ApiRequestError with server error message', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ error: 'name is required' }), { status: 400 })
    );

    try {
      await api.post('/api/categories', {});
      expect.fail('should have thrown');
    } catch (e) {
      expect(e).toBeInstanceOf(ApiRequestError);
      expect((e as ApiRequestError).message).toBe('name is required');
      expect((e as ApiRequestError).status).toBe(400);
      expect((e as ApiRequestError).isNetwork).toBe(false);
    }
  });

  it('throws network error when fetch fails', async () => {
    vi.spyOn(globalThis, 'fetch').mockRejectedValue(new TypeError('Failed to fetch'));

    try {
      await api.get('/api/health');
      expect.fail('should have thrown');
    } catch (e) {
      expect(e).toBeInstanceOf(ApiRequestError);
      expect((e as ApiRequestError).isNetwork).toBe(true);
      expect((e as ApiRequestError).status).toBe(0);
    }
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
