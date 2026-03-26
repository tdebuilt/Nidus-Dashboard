export interface ApiError {
  status: number;
  message: string;
  isNetwork: boolean;
}

export class ApiRequestError extends Error {
  status: number;
  isNetwork: boolean;

  constructor(status: number, message: string, isNetwork = false) {
    super(message);
    this.status = status;
    this.isNetwork = isNetwork;
  }
}

let onUnauthorized: (() => void) | null = null;

export function setUnauthorizedHandler(handler: () => void): void {
  onUnauthorized = handler;
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const url = path.startsWith('/') ? path : `/${path}`;
  const options: RequestInit = {
    method,
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
  };

  if (body !== undefined) {
    options.body = JSON.stringify(body);
  }

  let response: Response;
  try {
    response = await fetch(url, options);
  } catch {
    throw new ApiRequestError(0, 'Network error', true);
  }

  if (response.status === 401) {
    let message = 'Unauthorized';
    try {
      const data = await response.json();
      if (data.error) message = data.error;
    } catch {
      // ignore parse errors
    }
    if (message !== 'totp_required') {
      onUnauthorized?.();
    }
    throw new ApiRequestError(401, message);
  }

  if (!response.ok) {
    let message = `HTTP ${response.status}`;
    try {
      const data = await response.json();
      if (data.error) message = data.error;
    } catch {
      // ignore parse errors
    }
    throw new ApiRequestError(response.status, message);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, body),
  delete: <T>(path: string) => request<T>('DELETE', path),
};
