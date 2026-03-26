import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { api, ApiError, NetworkError } from './client'

describe('API client', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    // Reset location mock
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { pathname: '/dashboard', href: '' },
    })
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
  })

  it('makes GET request with correct options', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ data: 'test' }),
    })

    const result = await api.get<{ data: string }>('/api/test')
    expect(result.data).toBe('test')
    expect(globalThis.fetch).toHaveBeenCalledWith('/api/test', expect.objectContaining({
      method: 'GET',
      credentials: 'include',
    }))
  })

  it('makes POST request with body', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ success: true }),
    })

    await api.post('/api/test', { name: 'test' })
    expect(globalThis.fetch).toHaveBeenCalledWith('/api/test', expect.objectContaining({
      method: 'POST',
      body: JSON.stringify({ name: 'test' }),
    }))
  })

  it('makes PUT request', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({}),
    })

    await api.put('/api/test', { value: 1 })
    expect(globalThis.fetch).toHaveBeenCalledWith('/api/test', expect.objectContaining({
      method: 'PUT',
    }))
  })

  it('makes DELETE request', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({}),
    })

    await api.delete('/api/test')
    expect(globalThis.fetch).toHaveBeenCalledWith('/api/test', expect.objectContaining({
      method: 'DELETE',
    }))
  })

  it('throws NetworkError on fetch failure', async () => {
    globalThis.fetch = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'))

    await expect(api.get('/api/test')).rejects.toThrow(NetworkError)
  })

  it('throws ApiError on non-ok response', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      json: () => Promise.resolve({ error: 'server error' }),
    })

    try {
      await api.get('/api/test')
      expect.fail('should have thrown')
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError)
      expect((e as ApiError).status).toBe(500)
      expect((e as ApiError).message).toBe('server error')
    }
  })

  it('redirects to /login on 401', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ error: 'unauthorized' }),
    })

    await expect(api.get('/api/test')).rejects.toThrow(ApiError)
    expect(window.location.href).toBe('/login')
  })

  it('does not redirect to /login if already on /login', async () => {
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { pathname: '/login', href: '' },
    })

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ error: 'unauthorized' }),
    })

    await expect(api.get('/api/test')).rejects.toThrow(ApiError)
    expect(window.location.href).toBe('')
  })

  it('does not redirect to /login on 401 when on /register', async () => {
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { pathname: '/register', href: '' },
    })

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ error: 'unauthorized' }),
    })

    await expect(api.get('/api/test')).rejects.toThrow(ApiError)
    expect(window.location.href).toBe('')
  })

  it('handles 204 No Content', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
    })

    const result = await api.delete('/api/test')
    expect(result).toBeUndefined()
  })
})
