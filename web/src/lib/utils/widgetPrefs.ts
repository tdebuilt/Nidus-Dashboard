/**
 * Load a persisted widget preference from localStorage.
 */
export function loadPref<T>(prefix: string, key: string, fallback: T, valid?: T[]): T {
  if (typeof localStorage === 'undefined') return fallback
  const v = localStorage.getItem(prefix + key)
  if (v === null) return fallback
  if (valid && !valid.includes(v as T)) return fallback
  return v as T
}

/**
 * Save a widget preference to localStorage.
 */
export function savePref(prefix: string, key: string, value: string): void {
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem(prefix + key, value)
  }
}
