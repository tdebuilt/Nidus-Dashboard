export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

export class NetworkError extends Error {
  constructor(message = 'Network error') {
    super(message)
    this.name = 'NetworkError'
  }
}

type RequestOptions = {
  method?: string
  body?: unknown
  headers?: Record<string, string>
}

async function request<T>(url: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, headers = {} } = options

  const config: RequestInit = {
    method,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...headers,
    },
  }

  if (body !== undefined) {
    config.body = JSON.stringify(body)
  }

  let response: Response
  try {
    response = await fetch(url, config)
  } catch {
    throw new NetworkError()
  }

  if (response.status === 401) {
    let message = 'Unauthorized'
    try {
      const data = await response.json()
      if (data.error) message = data.error
    } catch {
      // ignore parse errors
    }
    // Redirect to login on auth failure (but not during login/setup/register)
    if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/login') && !window.location.pathname.startsWith('/setup') && !window.location.pathname.startsWith('/register') && !window.location.pathname.startsWith('/reset-password')) {
      window.location.href = '/login'
    }
    throw new ApiError(401, message)
  }

  if (!response.ok) {
    let message = `HTTP ${response.status}`
    try {
      const data = await response.json()
      if (data.error) message = data.error
    } catch {
      // ignore parse errors
    }
    throw new ApiError(response.status, message)
  }

  // Handle 204 No Content
  if (response.status === 204) {
    return undefined as T
  }

  return response.json()
}

export const api = {
  get: <T>(url: string) => request<T>(url),
  post: <T>(url: string, body?: unknown) => request<T>(url, { method: 'POST', body }),
  put: <T>(url: string, body?: unknown) => request<T>(url, { method: 'PUT', body }),
  patch: <T>(url: string, body?: unknown) => request<T>(url, { method: 'PATCH', body }),
  delete: <T>(url: string) => request<T>(url, { method: 'DELETE' }),
}
