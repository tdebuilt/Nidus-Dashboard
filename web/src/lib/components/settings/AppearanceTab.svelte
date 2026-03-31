<script lang="ts">
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { confirm } from '../../stores/confirm'
  import { isAdmin } from '../../stores/auth'
  import { setTheme } from '../../stores/theme'
  import { getAllThemes, getTheme, builtinThemes } from '../../themes'
  import { openThemeEditor } from '../../stores/themeEditor'
  import { customThemes } from '../../stores/customThemes'
  import { get } from 'svelte/store'
  import { t, translate } from '../../i18n'
  import ThemeSelector from './ThemeSelector.svelte'
  import AccentColorPicker from './AccentColorPicker.svelte'
  import CustomCSSEditor from './CustomCSSEditor.svelte'

  const builtinIds = new Set(Object.keys(builtinThemes))

  function isCustomTheme(id: string): boolean {
    return !builtinIds.has(id)
  }

  function getCustomDbId(themeId: string): number | null {
    const list = get(customThemes)
    const record = list.find(r => {
      try {
        const parsed = JSON.parse(r.theme_json)
        return parsed.id === themeId
      } catch { return false }
    })
    return record?.id ?? null
  }

  async function handleDeleteCustomTheme(themeId: string, themeName: string) {
    const ok = await confirm({
      title: translate('theme.deleteTheme'),
      message: translate('theme.deleteThemeConfirm', { name: themeName }),
      confirmLabel: translate('common.delete'),
      destructive: true,
    })
    if (!ok) return
    const dbId = getCustomDbId(themeId)
    if (dbId === null) return
    const success = await customThemes.remove(dbId, themeId)
    if (success) {
      if (currentTheme === themeId) {
        currentTheme = 'dark'
        setTheme('dark')
        updatePreference('theme', 'dark')
      }
      toasts.success(translate('theme.themeDeleted'))
    } else {
      toasts.error(translate('theme.themeError'))
    }
  }

  function handleEditCustomTheme(themeId: string) {
    const def = getTheme(themeId)
    if (!def) return
    const dbId = getCustomDbId(themeId)
    openThemeEditor(def, dbId ?? undefined)
  }

  let currentTheme = $state('dark')
  let accentColor = $state('')
  let customCSS = $state('')

  async function loadAppearance() {
    try {
      const prefs = await api.get<{ theme: string; accent_color: string }>('/api/preferences')
      currentTheme = prefs.theme
      setTheme(prefs.theme)
      accentColor = prefs.accent_color || ''
      if (accentColor) applyAccentColor(accentColor)

      const global = await api.get<{ custom_css: string }>('/api/settings')
      customCSS = global.custom_css || ''
    } catch {
      toasts.error(translate('settings.loadError'))
    }
  }

  $effect(() => {
    loadAppearance()
  })

  async function updatePreference(key: string, value: unknown) {
    try {
      await api.put('/api/preferences', { [key]: value })
      toasts.success(translate('settings.updated'))
    } catch {
      toasts.error(translate('settings.updateError'))
    }
  }

  async function updateGlobalSetting(key: string, value: unknown) {
    try {
      await api.put('/api/settings', { [key]: value })
      toasts.success(translate('settings.updated'))
    } catch {
      toasts.error(translate('settings.updateError'))
    }
  }

  function applyAccentColor(color: string) {
    if (typeof document === 'undefined') return
    const root = document.documentElement
    if (color) {
      root.style.setProperty('--color-primary', color)
      root.style.setProperty('--color-primary-hover', darkenColor(color, 0.15))
    } else {
      root.style.removeProperty('--color-primary')
      root.style.removeProperty('--color-primary-hover')
      setTheme(currentTheme)
    }
  }

  function darkenColor(hex: string, amount: number): string {
    const r = parseInt(hex.slice(1, 3), 16)
    const g = parseInt(hex.slice(3, 5), 16)
    const b = parseInt(hex.slice(5, 7), 16)
    const dr = Math.round(r * (1 - amount))
    const dg = Math.round(g * (1 - amount))
    const db = Math.round(b * (1 - amount))
    return `#${dr.toString(16).padStart(2, '0')}${dg.toString(16).padStart(2, '0')}${db.toString(16).padStart(2, '0')}`
  }

  function handleThemeSelect(themeId: string) {
    currentTheme = themeId
    setTheme(themeId)
    updatePreference('theme', themeId)
    accentColor = ''
    applyAccentColor('')
    updatePreference('accent_color', '')
  }

  function handleAccentColorChange(e: Event) {
    const value = (e.target as HTMLInputElement).value
    accentColor = value
    applyAccentColor(value)
    updatePreference('accent_color', value)
  }

  function resetAccentColor() {
    accentColor = ''
    applyAccentColor('')
    updatePreference('accent_color', '')
  }

  function applyCustomCSS(css: string) {
    if (typeof document === 'undefined') return
    let el = document.getElementById('nidus-custom-css')
    if (!css) {
      el?.remove()
      return
    }
    if (!el) {
      el = document.createElement('style')
      el.id = 'nidus-custom-css'
      document.head.appendChild(el)
    }
    el.textContent = css
  }

  function handleCustomCSSChange(e: Event) {
    const value = (e.target as HTMLTextAreaElement).value
    customCSS = value
    applyCustomCSS(value)
    updateGlobalSetting('custom_css', value)
  }

  function clearCustomCSS() {
    customCSS = ''
    applyCustomCSS('')
    updateGlobalSetting('custom_css', '')
  }

  const defaultPrimaryColor = $derived(getTheme(currentTheme)?.colors['color-primary'] || '#3b82f6')
</script>

<div class="space-y-6">
  <h3 class="text-lg font-semibold text-[var(--color-text)]">{$t('settings.tabs.appearance')}</h3>

  <ThemeSelector
    themes={getAllThemes()}
    {currentTheme}
    {isCustomTheme}
    onSelect={handleThemeSelect}
    onEdit={handleEditCustomTheme}
    onDelete={handleDeleteCustomTheme}
    onCreate={() => openThemeEditor()}
  />

  <AccentColorPicker
    {accentColor}
    defaultColor={defaultPrimaryColor}
    onChange={handleAccentColorChange}
    onReset={resetAccentColor}
  />

  {#if $isAdmin}
    <CustomCSSEditor
      {customCSS}
      onChange={handleCustomCSSChange}
      onClear={clearCustomCSS}
    />
  {/if}
</div>
