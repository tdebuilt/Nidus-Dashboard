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

async function parseErrorBody(response: Response, fallback: string): Promise<string> {
  try {
    const data = await response.json()
    if (data.error) return data.error
  } catch {
    // ignore parse errors
  }
  return fallback
}

function handleUnauthorizedRedirect() {
  if (typeof window === 'undefined') return
  const path = window.location.pathname
  const authPaths = ['/login', '/setup', '/register', '/reset-password']
  if (!authPaths.some((p) => path.startsWith(p))) {
    window.location.href = '/login'
  }
}

function buildRequestInit(options: RequestOptions): RequestInit {
  const { method = 'GET', body, headers = {} } = options
  const config: RequestInit = {
    method,
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...headers },
  }
  if (body !== undefined) config.body = JSON.stringify(body)
  return config
}

async function fetchWithTimeout(url: string, config: RequestInit): Promise<Response> {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 30_000)
  config.signal = controller.signal
  try {
    return await fetch(url, config)
  } catch (err) {
    clearTimeout(timeoutId)
    if (err instanceof DOMException && err.name === 'AbortError') {
      throw new NetworkError('Request timeout')
    }
    throw new NetworkError()
  } finally {
    clearTimeout(timeoutId)
  }
}

async function request<T>(url: string, options: RequestOptions = {}): Promise<T> {
  const config = buildRequestInit(options)
  const response = await fetchWithTimeout(url, config)

  if (response.status === 401) {
    const message = await parseErrorBody(response, 'Unauthorized')
    handleUnauthorizedRedirect()
    throw new ApiError(401, message)
  }

  if (!response.ok) {
    const message = await parseErrorBody(response, `HTTP ${response.status}`)
    throw new ApiError(response.status, message)
  }

  if (response.status === 204) return undefined as T
  return response.json()
}

export const api = {
  get: <T>(url: string) => request<T>(url),
  post: <T>(url: string, body?: unknown) => request<T>(url, { method: 'POST', body }),
  put: <T>(url: string, body?: unknown) => request<T>(url, { method: 'PUT', body }),
  patch: <T>(url: string, body?: unknown) => request<T>(url, { method: 'PATCH', body }),
  delete: <T>(url: string) => request<T>(url, { method: 'DELETE' }),
}
