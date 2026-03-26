<script lang="ts">
  import { Palette, Code, Plus, Pencil, Trash2 } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { confirm } from '../../stores/confirm'
  import { isAdmin } from '../../stores/auth'
  import { setTheme } from '../../stores/theme'
  import { getAllThemes, getTheme, builtinThemes } from '../../themes'
  import { openThemeEditor } from '../../stores/themeEditor'
  import { customThemes } from '../../stores/customThemes'
  import { t, translate } from '../../i18n'
  import { onMount } from 'svelte'

  const builtinIds = new Set(Object.keys(builtinThemes))

  function isCustomTheme(id: string): boolean {
    return !builtinIds.has(id)
  }

  function getCustomDbId(themeId: string): number | null {
    let dbId: number | null = null
    customThemes.subscribe(list => {
      const record = list.find(r => {
        try {
          const parsed = JSON.parse(r.theme_json)
          return parsed.id === themeId
        } catch { return false }
      })
      if (record) dbId = record.id
    })()
    return dbId
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

  onMount(async () => {
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
</script>

<div class="space-y-6">
  <h3 class="text-lg font-semibold text-[var(--color-text)]">{$t('settings.tabs.appearance')}</h3>

  <!-- Theme -->
  <section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-theme">
    <div class="mb-3 flex items-center gap-2">
      <Palette size={18} class="text-[var(--color-text-secondary)]" />
      <h3 class="font-semibold text-[var(--color-text)]">{$t('settings.themeSection')}</h3>
    </div>
    <div class="grid grid-cols-2 gap-3 sm:grid-cols-4" data-testid="settings-theme-select">
      {#each getAllThemes() as thm (thm.id)}
        <div
          class="group relative flex cursor-pointer flex-col items-center gap-2 rounded-lg border-2 p-3 transition-all {currentTheme === thm.id ? 'border-[var(--color-primary)] shadow-md' : 'border-[var(--color-border)] hover:border-[var(--color-text-muted)]'}"
          style="background-color: {thm.colors['color-bg-primary']}"
          data-testid="theme-card-{thm.id}"
          role="button"
          tabindex="0"
          onclick={() => {
            currentTheme = thm.id;
            setTheme(thm.id);
            updatePreference('theme', thm.id);
            accentColor = '';
            applyAccentColor('');
            updatePreference('accent_color', '');
          }}
          onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') e.currentTarget.click() }}
        >
          <div class="flex w-full gap-1 rounded-md overflow-hidden h-6">
            <div class="flex-1" style="background-color: {thm.colors['color-bg']}"></div>
            <div class="flex-1" style="background-color: {thm.colors['color-primary']}"></div>
            <div class="flex-1" style="background-color: {thm.colors['color-accent']}"></div>
            <div class="flex-1" style="background-color: {thm.colors['color-success']}"></div>
            <div class="flex-1" style="background-color: {thm.colors['color-danger']}"></div>
          </div>
          <div class="w-full rounded-md p-2" style="background-color: {thm.colors['color-bg']}">
            <div class="mb-1 h-1.5 w-3/4 rounded" style="background-color: {thm.colors['color-text']}"></div>
            <div class="mb-1 h-1.5 w-1/2 rounded" style="background-color: {thm.colors['color-text-secondary']}"></div>
            <div class="flex gap-1">
              <div class="h-2 w-6 rounded" style="background-color: {thm.colors['color-primary']}"></div>
              <div class="h-2 w-4 rounded" style="background-color: {thm.colors['color-accent']}"></div>
            </div>
          </div>
          <span class="text-xs font-medium" style="color: {thm.colors['color-text']}">{thm.name}</span>
          {#if $isAdmin && isCustomTheme(thm.id)}
            <div class="absolute top-1 end-1 flex gap-0.5">
              <button
                onclick={(e) => { e.stopPropagation(); handleEditCustomTheme(thm.id) }}
                class="rounded p-1 transition-colors hover:bg-[var(--color-bg-tertiary)]"
                title={$t('theme.editTheme')}
              >
                <Pencil size={12} style="color: {thm.colors['color-text-secondary']}" />
              </button>
              <button
                onclick={(e) => { e.stopPropagation(); handleDeleteCustomTheme(thm.id, thm.name) }}
                class="rounded p-1 transition-colors hover:bg-[var(--color-bg-tertiary)]"
                title={$t('theme.deleteTheme')}
              >
                <Trash2 size={12} style="color: {thm.colors['color-danger']}" />
              </button>
            </div>
          {/if}
        </div>
      {/each}
    </div>

    {#if $isAdmin}
      <button
        onclick={() => openThemeEditor()}
        class="mt-3 flex w-full items-center justify-center gap-1.5 rounded-lg border border-dashed border-[var(--color-border)] px-4 py-2.5 text-sm text-[var(--color-text-secondary)] transition-colors hover:border-[var(--color-primary)] hover:text-[var(--color-primary)]"
        data-testid="create-theme-btn"
      >
        <Plus size={16} />
        {$t('theme.createTheme')}
      </button>
    {/if}
  </section>

  <!-- Accent Color -->
  <section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-accent">
    <div class="mb-3 flex items-center gap-2">
      <Palette size={18} class="text-[var(--color-text-secondary)]" />
      <h3 class="font-semibold text-[var(--color-text)]">{$t('settings.accentSection')}</h3>
    </div>
    <div class="flex items-center gap-4">
      <input type="color"
        value={accentColor || getTheme(currentTheme)?.colors['color-primary'] || '#3b82f6'}
        oninput={handleAccentColorChange}
        class="h-10 w-14 cursor-pointer rounded border border-[var(--color-border)] bg-transparent"
        data-testid="settings-accent-picker"
      />
      <span class="text-sm font-mono text-[var(--color-text-secondary)]">{accentColor || $t('settings.accentDefault')}</span>
      {#if accentColor}
        <button onclick={resetAccentColor}
          class="rounded-lg bg-[var(--color-bg-tertiary)] px-3 py-1.5 text-sm text-[var(--color-text)] hover:bg-[var(--color-border)]"
          data-testid="settings-accent-reset">
          {$t('settings.accentReset')}
        </button>
      {/if}
    </div>
  </section>

  <!-- Custom CSS (admin only) -->
  {#if $isAdmin}
  <section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-custom-css">
    <div class="mb-3 flex items-center gap-2">
      <Code size={18} class="text-[var(--color-text-secondary)]" />
      <h3 class="font-semibold text-[var(--color-text)]">{$t('settings.customCssSection')}</h3>
    </div>
    <div class="mb-2 rounded-lg border border-[var(--color-warning)] bg-[var(--color-warning-bg)] px-3 py-2">
      <p class="text-xs text-[var(--color-warning)]">{$t('settings.customCssWarning')}</p>
    </div>
    <textarea
      value={customCSS}
      onchange={handleCustomCSSChange}
      placeholder={$t('settings.customCssPlaceholder')}
      rows="6"
      spellcheck="false"
      class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 font-mono text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
      data-testid="settings-custom-css-textarea"
    ></textarea>
    {#if customCSS}
      <div class="mt-2 flex justify-end">
        <button onclick={clearCustomCSS}
          class="rounded-lg bg-[var(--color-bg-tertiary)] px-3 py-1.5 text-sm text-[var(--color-text)] hover:bg-[var(--color-border)]"
          data-testid="settings-custom-css-clear">
          {$t('settings.customCssClear')}
        </button>
      </div>
    {/if}
  </section>
  {/if}
</div>
