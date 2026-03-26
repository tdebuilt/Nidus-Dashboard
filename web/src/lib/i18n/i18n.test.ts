import { describe, it, expect, afterEach } from 'vitest'
import { get } from 'svelte/store'
import { locale, setLocale, t, translate, registerLocale, getAvailableLocales, getLocaleMessages, detectBrowserLocale, applyBrowserLocale } from './index'
import fr from './fr.json'
import en from './en.json'

describe('i18n', () => {
  afterEach(() => {
    setLocale('fr')
  })

  it('defaults to French locale', () => {
    expect(get(locale)).toBe('fr')
  })

  it('translates a key in French', () => {
    const $t = get(t)
    expect($t('common.loading')).toBe('Chargement...')
    expect($t('login.title')).toBe('Connexion')
  })

  it('translates a key in English', () => {
    setLocale('en')
    const $t = get(t)
    expect($t('common.loading')).toBe('Loading...')
    expect($t('login.title')).toBe('Login')
  })

  it('switches locale dynamically', () => {
    expect(get(t)('sidebar.settings')).toBe('Paramètres')
    setLocale('en')
    expect(get(t)('sidebar.settings')).toBe('Settings')
    setLocale('fr')
    expect(get(t)('sidebar.settings')).toBe('Paramètres')
  })

  it('returns key for missing translation', () => {
    const $t = get(t)
    expect($t('nonexistent.key')).toBe('nonexistent.key')
  })

  it('falls back to English then French for missing keys', () => {
    registerLocale({ code: 'test', label: 'Test', flag: '🏴', messages: {} })
    setLocale('test')
    const $t = get(t)
    // Should fall back to English
    expect($t('common.loading')).toBe('Loading...')
    // Bad key still returns key
    expect($t('bad.key')).toBe('bad.key')
  })

  it('supports parameter substitution', () => {
    const $t = get(t)
    expect($t('setup.progress', { current: 2, total: 5 })).toBe('Étape 2 sur 5')
    setLocale('en')
    expect(get(t)('setup.progress', { current: 2, total: 5 })).toBe('Step 2 of 5')
  })

  it('translate() convenience function works', () => {
    expect(translate('common.cancel')).toBe('Annuler')
    setLocale('en')
    expect(translate('common.cancel')).toBe('Cancel')
  })

  it('has all keys present in both languages', () => {
    const frSections = Object.keys(fr) as (keyof typeof fr)[]
    const enSections = Object.keys(en) as (keyof typeof en)[]

    // Same sections
    expect(frSections.sort()).toEqual(enSections.sort())

    // Same keys within each section
    for (const section of frSections) {
      const frKeys = Object.keys(fr[section]).sort()
      const enKeys = Object.keys((en as unknown as Record<string, Record<string, string>>)[section]).sort()
      expect(frKeys).toEqual(enKeys)
    }
  })

  it('returns raw key for invalid format', () => {
    const $t = get(t)
    expect($t('noDot')).toBe('noDot')
    expect($t('too.many.dots')).toBe('too.many.dots')
  })

  it('ignores setLocale for unregistered locale', () => {
    setLocale('xx')
    expect(get(locale)).toBe('fr')
  })

  it('registerLocale adds a new locale', () => {
    registerLocale({ code: 'de', label: 'Deutsch', flag: '🇩🇪', messages: { common: { loading: 'Laden...' } } })
    setLocale('de')
    expect(get(t)('common.loading')).toBe('Laden...')
    // Falls back to en for missing keys
    expect(get(t)('login.title')).toBe('Login')
  })

  it('getAvailableLocales returns all registered locales', () => {
    const locales = getAvailableLocales()
    const codes = locales.map(l => l.code)
    expect(codes).toContain('fr')
    expect(codes).toContain('en')
  })

  it('getLocaleMessages returns messages for a registered locale', () => {
    const msgs = getLocaleMessages('fr')
    expect(msgs).toBeDefined()
    expect(msgs?.common).toBeDefined()
  })

  it('getLocaleMessages returns undefined for unregistered locale', () => {
    expect(getLocaleMessages('zz')).toBeUndefined()
  })

  it('detectBrowserLocale returns a matching locale from navigator.languages', () => {
    const detected = detectBrowserLocale()
    // jsdom sets navigator.language to 'en-US', so should match 'en'
    expect(detected).toBe('en')
  })

  it('applyBrowserLocale sets locale from browser when no stored preference', () => {
    localStorage.removeItem('nidus-locale')
    setLocale('fr')
    applyBrowserLocale()
    // jsdom navigator.language is 'en-US' → 'en'
    expect(get(locale)).toBe('en')
  })

  it('applyBrowserLocale does not override stored preference', () => {
    localStorage.setItem('nidus-locale', 'fr')
    setLocale('fr')
    applyBrowserLocale()
    expect(get(locale)).toBe('fr')
  })
})
