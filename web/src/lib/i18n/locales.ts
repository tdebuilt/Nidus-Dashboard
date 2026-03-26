// Metadata for known locales (label and flag).
// To add a new language:
// 1. Create <code>.json in this directory with translations
// 2. Add an entry here with the locale code, label, and flag
export const localeMetadata: Record<string, { label: string; flag: string; rtl?: boolean }> = {
  fr: { label: 'Français', flag: '🇫🇷' },
  en: { label: 'English', flag: '🇬🇧' },
  es: { label: 'Español', flag: '🇪🇸' },
  de: { label: 'Deutsch', flag: '🇩🇪' },
  pt: { label: 'Português', flag: '🇵🇹' },
  it: { label: 'Italiano', flag: '🇮🇹' },
  nl: { label: 'Nederlands', flag: '🇳🇱' },
  ru: { label: 'Русский', flag: '🇷🇺' },
  zh: { label: '中文', flag: '🇨🇳' },
  ja: { label: '日本語', flag: '🇯🇵' },
  ar: { label: 'العربية', flag: '🇸🇦', rtl: true },
}
