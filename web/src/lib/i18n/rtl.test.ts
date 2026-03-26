import { describe, it, expect } from 'vitest'
import { localeMetadata } from './locales'

describe('RTL locale metadata', () => {
  it('marks Arabic as RTL', () => {
    expect(localeMetadata['ar'].rtl).toBe(true)
  })

  it('does not mark other locales as RTL', () => {
    const nonRtlLocales = ['fr', 'en', 'es', 'de', 'pt', 'it', 'nl', 'ru', 'zh', 'ja']
    for (const code of nonRtlLocales) {
      expect(localeMetadata[code].rtl, `${code} should not be RTL`).toBeFalsy()
    }
  })

  it('has rtl property in type definition', () => {
    // Ensure the type allows rtl property
    const arMeta = localeMetadata['ar']
    expect(arMeta).toHaveProperty('rtl')
    expect(arMeta).toHaveProperty('label')
    expect(arMeta).toHaveProperty('flag')
  })
})
