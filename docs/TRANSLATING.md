# Translating Nidus

Thanks for helping translate Nidus! This guide explains everything you need to add or improve a translation.

## Quick start

1. Copy the template file:
   ```bash
   cp web/src/lib/i18n/template.json web/src/lib/i18n/<code>.json
   ```
   Replace `<code>` with the [ISO 639-1 language code](https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes) (e.g. `ko` for Korean).

2. Register your locale in `web/src/lib/i18n/locales.ts`:
   ```ts
   ko: { label: '한국어', flag: '🇰🇷' },
   ```

3. Translate all values in your JSON file (see format below).

4. Validate your file:
   ```bash
   cd web && npm run i18n:validate
   ```

5. Open a pull request.

That's it — the i18n system auto-discovers JSON files and registers them automatically.

## File format

Translation files are JSON with **one level of nesting** — sections containing key-value pairs:

```json
{
  "section": {
    "key": "Translated text",
    "keyWithParam": "Hello {name}, you have {count} items"
  }
}
```

### Rules

- **Keys** must be identical to `fr.json` (the reference file) — don't add, remove, or rename keys
- **Values** are the translated strings in your target language
- **Placeholders** use `{paramName}` syntax — keep them exactly as-is, only translate the surrounding text
- **JSON special characters** must be escaped: use `\"` for literal quotes inside values, `\\` for backslashes
- For CJK languages, prefer corner brackets `「」` over quotes `""` to avoid JSON parsing issues

### Example

From `fr.json`:
```json
{
  "setup": {
    "progress": "Étape {current} sur {total}"
  }
}
```

Korean translation:
```json
{
  "setup": {
    "progress": "{total}단계 중 {current}단계"
  }
}
```

## Reference file

The reference file is `web/src/lib/i18n/fr.json` (French). All other locale files must have the **exact same keys**. English (`en.json`) is also available for reference if you don't read French.

## Template

`web/src/lib/i18n/template.json` contains all keys with empty values — use it as a starting point.

## Validation

Run the validation script to check your file against the reference:

```bash
cd web && npm run i18n:validate
```

It reports:
- **Missing keys** — keys present in `fr.json` but not in your file
- **Extra keys** — keys in your file that don't exist in `fr.json`

All keys must match for the validation to pass.

## Fallback chain

When a key is missing in the current locale, Nidus falls back:

1. Selected language
2. English (`en`)
3. French (`fr`)
4. Raw key (e.g. `common.loading`)

This means partially translated files still work — missing strings fall back to English.

## Currently supported languages

| Code | Language | Flag |
|------|----------|------|
| `fr` | Français | 🇫🇷 |
| `en` | English | 🇬🇧 |
| `es` | Español | 🇪🇸 |
| `de` | Deutsch | 🇩🇪 |
| `pt` | Português | 🇵🇹 |
| `it` | Italiano | 🇮🇹 |
| `nl` | Nederlands | 🇳🇱 |
| `ru` | Русский | 🇷🇺 |
| `zh` | 中文 | 🇨🇳 |
| `ja` | 日本語 | 🇯🇵 |
| `ar` | العربية | 🇸🇦 |

## Tips

- **Test your translation** by running `make dev` or `cd web && npm run dev`, then switching to your language in Settings
- **Keep it natural** — don't translate word-for-word, adapt to how the concept is expressed in your language
- **Be consistent** — use the same term for the same concept throughout (e.g. always "container" or always "conteneur")
- **UI space** — some languages are more verbose; keep labels concise where possible
