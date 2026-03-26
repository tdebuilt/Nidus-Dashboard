/**
 * Central theme definitions.
 * Each theme provides a complete set of CSS custom properties.
 */

export interface ThemeColors {
  // Backgrounds
  'color-bg': string
  'color-bg-primary': string
  'color-bg-secondary': string
  'color-bg-tertiary': string
  'color-border': string

  // Text
  'color-text': string
  'color-text-primary': string
  'color-text-secondary': string
  'color-text-muted': string

  // Accent
  'color-primary': string
  'color-primary-hover': string
  'color-accent': string
  'color-accent-hover': string

  // Status
  'color-danger': string
  'color-danger-hover': string
  'color-success': string
  'color-warning': string

  // Sidebar
  'color-sidebar-bg': string
  'color-sidebar-hover': string

  // Feedback — error
  'color-error-text': string
  'color-error-border': string
  'color-error-bg': string

  // Feedback — success
  'color-success-text': string
  'color-success-border': string
  'color-success-bg': string

  // Feedback — info
  'color-info-text': string
  'color-info-border': string
  'color-info-bg': string
}

export interface ThemeDefinition {
  id: string
  name: string
  author: string
  mode: 'dark' | 'light'
  colors: ThemeColors
}

// ── Built-in themes ──

export const darkTheme: ThemeDefinition = {
  id: 'dark',
  name: 'Dark',
  author: 'Nidus',
  mode: 'dark',
  colors: {
    'color-bg': '#0f172a',
    'color-bg-primary': '#111827',
    'color-bg-secondary': '#1e293b',
    'color-bg-tertiary': '#334155',
    'color-border': '#334155',
    'color-text': '#f1f5f9',
    'color-text-primary': '#f3f4f6',
    'color-text-secondary': '#94a3b8',
    'color-text-muted': '#64748b',
    'color-primary': '#3b82f6',
    'color-primary-hover': '#2563eb',
    'color-accent': '#6366f1',
    'color-accent-hover': '#818cf8',
    'color-danger': '#ef4444',
    'color-danger-hover': '#dc2626',
    'color-success': '#22c55e',
    'color-warning': '#eab308',
    'color-sidebar-bg': '#1e293b',
    'color-sidebar-hover': '#334155',
    'color-error-text': '#f87171',
    'color-error-border': '#ef4444',
    'color-error-bg': 'rgba(239, 68, 68, 0.1)',
    'color-success-text': '#4ade80',
    'color-success-border': '#22c55e',
    'color-success-bg': 'rgba(34, 197, 94, 0.1)',
    'color-info-text': '#60a5fa',
    'color-info-border': '#3b82f6',
    'color-info-bg': 'rgba(59, 130, 246, 0.1)',
  },
}

export const lightTheme: ThemeDefinition = {
  id: 'light',
  name: 'Light',
  author: 'Nidus',
  mode: 'light',
  colors: {
    'color-bg': '#f8fafc',
    'color-bg-primary': '#ffffff',
    'color-bg-secondary': '#f9fafb',
    'color-bg-tertiary': '#e2e8f0',
    'color-border': '#cbd5e1',
    'color-text': '#0f172a',
    'color-text-primary': '#111827',
    'color-text-secondary': '#475569',
    'color-text-muted': '#94a3b8',
    'color-primary': '#3b82f6',
    'color-primary-hover': '#2563eb',
    'color-accent': '#4f46e5',
    'color-accent-hover': '#6366f1',
    'color-danger': '#dc2626',
    'color-danger-hover': '#b91c1c',
    'color-success': '#16a34a',
    'color-warning': '#ca8a04',
    'color-sidebar-bg': '#ffffff',
    'color-sidebar-hover': '#f1f5f9',
    'color-error-text': '#dc2626',
    'color-error-border': '#f87171',
    'color-error-bg': 'rgba(239, 68, 68, 0.08)',
    'color-success-text': '#16a34a',
    'color-success-border': '#4ade80',
    'color-success-bg': 'rgba(34, 197, 94, 0.08)',
    'color-info-text': '#2563eb',
    'color-info-border': '#60a5fa',
    'color-info-bg': 'rgba(59, 130, 246, 0.08)',
  },
}

export const nordTheme: ThemeDefinition = {
  id: 'nord',
  name: 'Nord',
  author: 'Nidus',
  mode: 'dark',
  colors: {
    // Nord Polar Night
    'color-bg': '#2e3440',
    'color-bg-primary': '#3b4252',
    'color-bg-secondary': '#434c5e',
    'color-bg-tertiary': '#4c566a',
    'color-border': '#4c566a',
    // Nord Snow Storm
    'color-text': '#eceff4',
    'color-text-primary': '#e5e9f0',
    'color-text-secondary': '#d8dee9',
    'color-text-muted': '#7b88a1',
    // Nord Frost
    'color-primary': '#88c0d0',
    'color-primary-hover': '#81a1c1',
    'color-accent': '#5e81ac',
    'color-accent-hover': '#81a1c1',
    // Nord Aurora
    'color-danger': '#bf616a',
    'color-danger-hover': '#a5545c',
    'color-success': '#a3be8c',
    'color-warning': '#ebcb8b',
    // Sidebar
    'color-sidebar-bg': '#3b4252',
    'color-sidebar-hover': '#434c5e',
    // Feedback
    'color-error-text': '#bf616a',
    'color-error-border': '#bf616a',
    'color-error-bg': 'rgba(191, 97, 106, 0.1)',
    'color-success-text': '#a3be8c',
    'color-success-border': '#a3be8c',
    'color-success-bg': 'rgba(163, 190, 140, 0.1)',
    'color-info-text': '#88c0d0',
    'color-info-border': '#88c0d0',
    'color-info-bg': 'rgba(136, 192, 208, 0.1)',
  },
}

export const draculaTheme: ThemeDefinition = {
  id: 'dracula',
  name: 'Dracula',
  author: 'Nidus',
  mode: 'dark',
  colors: {
    'color-bg': '#282a36',
    'color-bg-primary': '#21222c',
    'color-bg-secondary': '#343746',
    'color-bg-tertiary': '#44475a',
    'color-border': '#44475a',
    'color-text': '#f8f8f2',
    'color-text-primary': '#f8f8f2',
    'color-text-secondary': '#bd93f9',
    'color-text-muted': '#6272a4',
    'color-primary': '#bd93f9',
    'color-primary-hover': '#9f7aea',
    'color-accent': '#ff79c6',
    'color-accent-hover': '#ff92d0',
    'color-danger': '#ff5555',
    'color-danger-hover': '#e04848',
    'color-success': '#50fa7b',
    'color-warning': '#f1fa8c',
    'color-sidebar-bg': '#21222c',
    'color-sidebar-hover': '#343746',
    'color-error-text': '#ff5555',
    'color-error-border': '#ff5555',
    'color-error-bg': 'rgba(255, 85, 85, 0.1)',
    'color-success-text': '#50fa7b',
    'color-success-border': '#50fa7b',
    'color-success-bg': 'rgba(80, 250, 123, 0.1)',
    'color-info-text': '#8be9fd',
    'color-info-border': '#8be9fd',
    'color-info-bg': 'rgba(139, 233, 253, 0.1)',
  },
}

/** All built-in themes, keyed by ID */
export const builtinThemes: Record<string, ThemeDefinition> = {
  dark: darkTheme,
  light: lightTheme,
  nord: nordTheme,
  dracula: draculaTheme,
}

/** Get a built-in theme by ID */
export function getBuiltinTheme(id: string): ThemeDefinition | undefined {
  return builtinThemes[id]
}

/** Get all available theme definitions */
export function getAvailableThemes(): ThemeDefinition[] {
  return Object.values(builtinThemes)
}

/** All required color keys for a valid theme */
export const themeColorKeys: (keyof ThemeColors)[] = [
  'color-bg', 'color-bg-primary', 'color-bg-secondary', 'color-bg-tertiary',
  'color-border',
  'color-text', 'color-text-primary', 'color-text-secondary', 'color-text-muted',
  'color-primary', 'color-primary-hover',
  'color-accent', 'color-accent-hover',
  'color-danger', 'color-danger-hover',
  'color-success', 'color-warning',
  'color-sidebar-bg', 'color-sidebar-hover',
  'color-error-text', 'color-error-border', 'color-error-bg',
  'color-success-text', 'color-success-border', 'color-success-bg',
  'color-info-text', 'color-info-border', 'color-info-bg',
]

/**
 * Parse and validate a JSON object as a ThemeDefinition.
 * Returns the theme if valid, or an error string if not.
 */
export function parseThemeJSON(json: unknown): ThemeDefinition | string {
  if (!json || typeof json !== 'object') return 'Theme must be a JSON object'
  const obj = json as Record<string, unknown>

  if (typeof obj.id !== 'string' || !obj.id) return 'Missing or invalid "id"'
  if (typeof obj.name !== 'string' || !obj.name) return 'Missing or invalid "name"'
  if (typeof obj.author !== 'string') return 'Missing or invalid "author"'
  if (obj.mode !== 'dark' && obj.mode !== 'light') return '"mode" must be "dark" or "light"'
  if (!obj.colors || typeof obj.colors !== 'object') return 'Missing or invalid "colors"'

  const colors = obj.colors as Record<string, unknown>
  const missing: string[] = []
  for (const key of themeColorKeys) {
    if (typeof colors[key] !== 'string') missing.push(key)
  }
  if (missing.length > 0) return `Missing color keys: ${missing.join(', ')}`

  return {
    id: obj.id,
    name: obj.name,
    author: obj.author as string,
    mode: obj.mode as 'dark' | 'light',
    colors: colors as unknown as ThemeColors,
  }
}
