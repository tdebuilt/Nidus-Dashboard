import { writable, derived, get } from 'svelte/store'
import { localeMetadata } from './locales'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Messages = Record<string, any>

export interface LocaleDefinition {
  code: string
  label: string
  flag: string
  messages: Messages
}

const localeRegistry = new Map<string, LocaleDefinition>()

let initialized = false

export function registerLocale(def: LocaleDefinition) {
  localeRegistry.set(def.code, def)
  // Trigger reactivity for components using getAvailableLocales (only after init)
  if (initialized) locale.update(v => v)
}

export function getAvailableLocales(): LocaleDefinition[] {
  return Array.from(localeRegistry.values())
}

export function getLocaleMessages(code: string): Messages | undefined {
  return localeRegistry.get(code)?.messages
}

// Auto-discover and register all JSON locale files in this directory.
// Vite's import.meta.glob eagerly loads them at build time.
const localeModules = import.meta.glob('./*.json', { eager: true }) as Record<string, { default: Messages }>

for (const [path, mod] of Object.entries(localeModules)) {
  const code = path.replace('./', '').replace('.json', '')
  const meta = localeMetadata[code] || { label: code, flag: '' }
  registerLocale({ code, label: meta.label, flag: meta.flag, messages: mod.default })
}

const stored = typeof localStorage !== 'undefined' ? localStorage.getItem('nidus-locale') : null
const initialLocale = stored && localeRegistry.has(stored) ? stored : 'fr'

/** Detect the best locale from the browser's language preferences. */
export function detectBrowserLocale(): string {
  if (typeof navigator === 'undefined') return 'fr'
  for (const lang of navigator.languages ?? [navigator.language]) {
    const code = lang.split('-')[0].toLowerCase()
    if (localeRegistry.has(code)) return code
  }
  return 'fr'
}

/** Apply browser language detection if no preference has been stored yet. */
export function applyBrowserLocale() {
  const saved = typeof localStorage !== 'undefined' ? localStorage.getItem('nidus-locale') : null
  if (!saved) {
    setLocale(detectBrowserLocale())
  }
}
export const locale = writable<string>(initialLocale)
initialized = true

export const isRTL = derived(locale, ($l) => localeMetadata[$l]?.rtl === true)

locale.subscribe((value) => {
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('nidus-locale', value)
  }
  if (typeof document !== 'undefined') {
    document.documentElement.lang = value
    document.documentElement.dir = localeMetadata[value]?.rtl ? 'rtl' : 'ltr'
  }
})

export function setLocale(l: string) {
  if (localeRegistry.has(l)) {
    locale.set(l)
  }
}

function lookup(dict: Messages, parts: string[]): string | undefined {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let node: any = dict
  for (const part of parts) {
    if (node == null || typeof node !== 'object') return undefined
    node = node[part]
  }
  return typeof node === 'string' ? node : undefined
}

export const t = derived(locale, ($locale) => {
  return (key: string, params?: Record<string, string | number>): string => {
    const parts = key.split('.')

    const currentMessages = getLocaleMessages($locale)
    let value = currentMessages ? lookup(currentMessages, parts) : undefined

    // Fallback chain: current → en → fr → raw key
    if (value === undefined) {
      const enMessages = getLocaleMessages('en')
      if (enMessages) value = lookup(enMessages, parts)
    }
    if (value === undefined) {
      const frMessages = getLocaleMessages('fr')
      if (frMessages) value = lookup(frMessages, parts)
    }

    // Key not found
    if (value === undefined) return key

    // Replace {param} placeholders
    if (params) {
      for (const [k, v] of Object.entries(params)) {
        value = value.replace(`{${k}}`, String(v))
      }
    }

    return value
  }
})

// Convenience function for use outside Svelte components (e.g. in stores)
export function translate(key: string, params?: Record<string, string | number>): string {
  return get(t)(key, params)
}
