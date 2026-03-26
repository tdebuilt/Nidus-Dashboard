#!/usr/bin/env node

/**
 * Validates that all i18n locale files have the same keys as fr.json (reference).
 * Usage: node scripts/validate-i18n.js
 *
 * Exit code 0 = all valid, 1 = missing or extra keys found.
 */

import { readFileSync, readdirSync } from 'fs'
import { join, basename } from 'path'

const i18nDir = join(import.meta.dirname, '../src/lib/i18n')
const files = readdirSync(i18nDir).filter(f => f.endsWith('.json'))

if (!files.includes('fr.json')) {
  console.error('Error: fr.json (reference file) not found in', i18nDir)
  process.exit(1)
}

const reference = JSON.parse(readFileSync(join(i18nDir, 'fr.json'), 'utf-8'))

function getKeys(obj, prefix = '') {
  const keys = []
  for (const [key, value] of Object.entries(obj)) {
    const fullKey = prefix ? `${prefix}.${key}` : key
    if (typeof value === 'object' && value !== null) {
      keys.push(...getKeys(value, fullKey))
    } else {
      keys.push(fullKey)
    }
  }
  return keys.sort()
}

const refKeys = getKeys(reference)
let hasErrors = false

for (const file of files) {
  if (file === 'fr.json') continue

  const code = basename(file, '.json')
  const locale = JSON.parse(readFileSync(join(i18nDir, file), 'utf-8'))
  const localeKeys = getKeys(locale)

  const missing = refKeys.filter(k => !localeKeys.includes(k))
  const extra = localeKeys.filter(k => !refKeys.includes(k))

  if (missing.length === 0 && extra.length === 0) {
    console.log(`  ${code}.json: OK (${localeKeys.length} keys)`)
  } else {
    hasErrors = true
    console.log(`  ${code}.json: ISSUES FOUND`)
    if (missing.length > 0) {
      console.log(`    Missing (${missing.length}):`)
      for (const k of missing) console.log(`      - ${k}`)
    }
    if (extra.length > 0) {
      console.log(`    Extra (${extra.length}):`)
      for (const k of extra) console.log(`      + ${k}`)
    }
  }
}

console.log(`\nReference: fr.json (${refKeys.length} keys)`)
console.log(`Checked: ${files.length - 1} locale(s)`)

if (hasErrors) {
  console.log('\nValidation FAILED — fix missing/extra keys above.')
  process.exit(1)
} else {
  console.log('\nValidation PASSED — all locales match the reference.')
  process.exit(0)
}
