<script lang="ts">
  import { Globe, RefreshCw, Keyboard, ToggleLeft, ToggleRight } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { pollingInterval } from '../../stores/polling'
  import { setKeyboardShortcutsEnabled } from '../../stores/keyboardShortcuts'
  import { t, translate, setLocale, getAvailableLocales } from '../../i18n'
  import { onMount } from 'svelte'

  const refreshOptions = [10, 15, 30, 60, 120, 300]

  let currentLanguage = $state('fr')
  let refreshInterval = $state(30)
  let enableShortcuts = $state(true)

  onMount(async () => {
    try {
      const prefs = await api.get<{ language: string; refresh_interval: number; enable_keyboard_shortcuts: boolean }>('/api/preferences')
      currentLanguage = prefs.language
      setLocale(prefs.language)
      refreshInterval = prefs.refresh_interval
      enableShortcuts = prefs.enable_keyboard_shortcuts !== false
      setKeyboardShortcutsEnabled(enableShortcuts)
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

  function handleLanguageChange(e: Event) {
    const value = (e.target as HTMLSelectElement).value
    currentLanguage = value
    setLocale(value)
    updatePreference('language', value)
  }

  function handleRefreshChange(e: Event) {
    const value = parseInt((e.target as HTMLSelectElement).value, 10)
    if (value >= 5 && value <= 300) {
      refreshInterval = value
      pollingInterval.update(value)
    }
  }
</script>

<div class="space-y-6">
  <h3 class="text-lg font-semibold text-[var(--color-text)]">{$t('settings.tabs.preferences')}</h3>

  <!-- Language -->
  <section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-language">
    <div class="mb-3 flex items-center gap-2">
      <Globe size={18} class="text-[var(--color-text-secondary)]" />
      <h3 class="font-semibold text-[var(--color-text)]">{$t('settings.languageSection')}</h3>
    </div>
    <select onchange={handleLanguageChange} value={currentLanguage}
      class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none"
      data-testid="settings-language-select">
      {#each getAvailableLocales() as loc (loc.code)}
        <option value={loc.code}>{loc.flag} {loc.label}</option>
      {/each}
    </select>
  </section>

  <!-- Refresh interval -->
  <section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-refresh">
    <div class="mb-3 flex items-center gap-2">
      <RefreshCw size={18} class="text-[var(--color-text-secondary)]" />
      <h3 class="font-semibold text-[var(--color-text)]">{$t('settings.refreshSection')}</h3>
    </div>
    <select onchange={handleRefreshChange} value={refreshInterval}
      class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none"
      data-testid="settings-refresh-select">
      {#each refreshOptions as seconds (seconds)}
        <option value={seconds}>{$t('settings.refreshSeconds', { seconds: String(seconds) })}</option>
      {/each}
    </select>
  </section>

  <!-- Keyboard shortcuts -->
  <section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-keyboard-shortcuts">
    <div class="mb-3 flex items-center gap-2">
      <Keyboard size={18} class="text-[var(--color-text-secondary)]" />
      <h3 class="font-semibold text-[var(--color-text)]">{$t('settings.keyboardShortcuts')}</h3>
    </div>
    <div class="flex items-center justify-between">
      <div>
        <p class="text-sm text-[var(--color-text)]">{$t('settings.keyboardShortcutsEnabled')}</p>
        <p class="text-xs text-[var(--color-text-muted)]">{$t('settings.keyboardShortcutsHint')}</p>
      </div>
      <button
        onclick={() => {
          enableShortcuts = !enableShortcuts
          setKeyboardShortcutsEnabled(enableShortcuts)
          updatePreference('enable_keyboard_shortcuts', enableShortcuts)
        }}
        class="text-[var(--color-text-secondary)] transition-colors hover:text-[var(--color-primary)]"
        data-testid="settings-keyboard-toggle"
        aria-label={$t('settings.keyboardShortcutsEnabled')}
      >
        {#if enableShortcuts}
          <ToggleRight size={32} class="text-[var(--color-primary)]" />
        {:else}
          <ToggleLeft size={32} />
        {/if}
      </button>
    </div>
  </section>
</div>
