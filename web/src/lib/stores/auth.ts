import { writable, derived } from 'svelte/store'

export type UserRole = 'admin' | 'editor' | 'viewer'

export type AuthState = {
  authenticated: boolean
  setupCompleted: boolean
  loading: boolean
  totpEnabled?: boolean
  role?: UserRole
  userId?: number
  username?: string
}

export const auth = writable<AuthState>({
  authenticated: false,
  setupCompleted: false,
  loading: true,
})

/** Whether the current user has at least admin role */
export const isAdmin = derived(auth, ($auth) => $auth.role === 'admin')

/** Whether the current user has at least editor role */
export const isEditor = derived(
  auth,
  ($auth) => $auth.role === 'admin' || $auth.role === 'editor',
)

/** Whether the current user is viewer only */
export const isViewer = derived(auth, ($auth) => $auth.role === 'viewer')
