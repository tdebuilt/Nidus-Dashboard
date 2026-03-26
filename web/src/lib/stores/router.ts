import { writable, derived } from 'svelte/store'

export const currentPath = writable(typeof window !== 'undefined' ? window.location.pathname : '/')

// Navigate to a new path
export function navigate(path: string) {
  if (typeof window !== 'undefined') {
    window.history.pushState({}, '', path)
    currentPath.set(path)
  }
}

// Handle browser back/forward
if (typeof window !== 'undefined') {
  window.addEventListener('popstate', () => {
    currentPath.set(window.location.pathname)
  })
}

// Route matching helper
export const currentRoute = derived(currentPath, ($path) => {
  if ($path === '/login') return 'login'
  if ($path === '/setup') return 'setup'
  if ($path === '/register') return 'register'
  if ($path === '/reset-password') return 'reset-password'
  if ($path === '/settings') return 'settings'
  if ($path === '/help') return 'help'
  if ($path === '/kiosk') return 'kiosk'
  if ($path === '/' || $path.startsWith('/dashboard')) return 'dashboard'
  return 'not-found'
})
