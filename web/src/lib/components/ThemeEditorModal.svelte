<script lang="ts">
  import { X, ChevronDown } from 'lucide-svelte'
  import { themeEditorState, closeThemeEditor } from '../stores/themeEditor'
  import { customThemes } from '../stores/customThemes'
  import { toasts } from '../stores/toast'
  import { theme, setTheme } from '../stores/theme'
  import { getAllThemes, applyTheme } from '../themes'
  import { parseThemeJSON } from '../themes'
  import type { ThemeDefinition } from '../themes'
  import { deriveFullTheme, extractBaseColors } from '../themes/color-utils'
  import type { BaseColors } from '../themes/color-utils'
  import { t, translate } from '../i18n'
  import { get } from 'svelte/store'

  let previousThemeId = $state('')
  let themeName = $state('')
  let baseThemeId = $state('dark')
  let mode = $state<'dark' | 'light'>('dark')
  let baseColors = $state<BaseColors>({
    bg: '#0f172a', text: '#f1f5f9', primary: '#3b82f6',
    accent: '#6366f1', danger: '#ef4444', success: '#22c55e',
    warning: '#eab308', mode: 'dark',
  })
  let jsonMode = $state(false)
  let jsonText = $state('')
  let jsonError = $state('')
  let saving = $state(false)

  const colorFields = [
    { key: 'bg', labelKey: 'theme.colorBackground' },
    { key: 'text', labelKey: 'theme.colorText' },
    { key: 'primary', labelKey: 'theme.colorPrimary' },
    { key: 'accent', labelKey: 'theme.colorAccent' },
    { key: 'danger', labelKey: 'theme.colorDanger' },
    { key: 'success', labelKey: 'theme.colorSuccess' },
    { key: 'warning', labelKey: 'theme.colorWarning' },
  ] as const

  function slugify(text: string): string {
    return text.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '').slice(0, 30)
  }

  function buildThemeDefinition(): ThemeDefinition {
    const colors = deriveFullTheme(baseColors)
    const id = $themeEditorState.editingTheme?.id ?? `custom-${slugify(themeName)}`
    return {
      id,
      name: themeName,
      author: 'Custom',
      mode,
      colors,
    }
  }

  function applyPreview() {
    const tempTheme = buildThemeDefinition()
    applyTheme(tempTheme)
  }

  function handleColorChange(key: string, value: string) {
    baseColors = { ...baseColors, [key]: value }
    baseColors.mode = mode
    applyPreview()
  }

  function handleModeChange(newMode: 'dark' | 'light') {
    mode = newMode
    baseColors = { ...baseColors, mode: newMode }
    applyPreview()
  }

  function handleBaseThemeChange(themeId: string) {
    baseThemeId = themeId
    const themes = getAllThemes()
    const base = themes.find(t => t.id === themeId)
    if (base) {
      const extracted = extractBaseColors(base.colors, base.mode)
      baseColors = extracted
      mode = base.mode
      applyPreview()
    }
  }

  function toggleJsonMode() {
    if (!jsonMode) {
      const def = buildThemeDefinition()
      jsonText = JSON.stringify(def, null, 2)
      jsonError = ''
    } else {
      // Parse JSON back to visual mode
      try {
        const parsed = JSON.parse(jsonText)
        const result = parseThemeJSON(parsed)
        if (typeof result === 'string') {
          jsonError = result
          return
        }
        themeName = result.name
        mode = result.mode
        baseColors = extractBaseColors(result.colors, result.mode)
        jsonError = ''
      } catch (e) {
        jsonError = (e as Error).message
        return
      }
    }
    jsonMode = !jsonMode
  }

  function handleJsonInput(e: Event) {
    jsonText = (e.target as HTMLTextAreaElement).value
    try {
      const parsed = JSON.parse(jsonText)
      const result = parseThemeJSON(parsed)
      if (typeof result === 'string') {
        jsonError = result
        return
      }
      jsonError = ''
      applyTheme(result)
    } catch {
      jsonError = 'Invalid JSON'
    }
  }

  function handleCancel() {
    setTheme(previousThemeId)
    closeThemeEditor()
  }

  async function handleSave() {
    if (!themeName.trim()) {
      toasts.error(translate('theme.themeNameRequired'))
      return
    }

    saving = true
    let def: ThemeDefinition

    if (jsonMode) {
      try {
        const parsed = JSON.parse(jsonText)
        const result = parseThemeJSON(parsed)
        if (typeof result === 'string') {
          toasts.error(translate('theme.jsonError', { message: result }))
          saving = false
          return
        }
        def = result
      } catch (e) {
        toasts.error(translate('theme.jsonError', { message: (e as Error).message }))
        saving = false
        return
      }
    } else {
      def = buildThemeDefinition()
    }

    const themeJSON = JSON.stringify(def)
    const dbId = $themeEditorState.editingDbId

    if (dbId) {
      const ok = await customThemes.update(dbId, themeName, themeJSON)
      if (ok) {
        setTheme(def.id)
        toasts.success(translate('theme.themeUpdated'))
        closeThemeEditor()
      } else {
        toasts.error(translate('theme.themeError'))
      }
    } else {
      const record = await customThemes.create(themeName, themeJSON)
      if (record) {
        setTheme(def.id)
        toasts.success(translate('theme.themeCreated'))
        closeThemeEditor()
      } else {
        toasts.error(translate('theme.themeError'))
      }
    }
    saving = false
  }

  // React to modal open/close
  $effect(() => {
    if ($themeEditorState.open) {
      previousThemeId = get(theme)
      jsonMode = false
      jsonError = ''
      saving = false

      if ($themeEditorState.editingTheme) {
        const t = $themeEditorState.editingTheme
        themeName = t.name
        mode = t.mode
        baseThemeId = t.id
        baseColors = extractBaseColors(t.colors, t.mode)
      } else {
        themeName = ''
        baseThemeId = 'dark'
        mode = 'dark'
        const themes = getAllThemes()
        const darkTheme = themes.find(t => t.id === 'dark')
        if (darkTheme) {
          baseColors = extractBaseColors(darkTheme.colors, 'dark')
        }
      }
    }
  })
</script>

{#if $themeEditorState.open}
  <!-- Backdrop -->
  <button
    class="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm"
    onclick={handleCancel}
    aria-label={$t('common.close')}
  ></button>

  <!-- Dialog -->
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <div class="w-full max-w-2xl max-h-[90vh] overflow-y-auto rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-6 shadow-2xl animate-[dialogIn_0.2s_ease-out]">
      <!-- Header -->
      <div class="mb-5 flex items-center justify-between">
        <h3 class="text-lg font-semibold text-[var(--color-text)]">
          {$themeEditorState.editingTheme ? $t('theme.editTheme') : $t('theme.createTheme')}
        </h3>
        <button
          onclick={handleCancel}
          class="rounded-lg p-1 text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
          aria-label={$t('common.close')}
        >
          <X size={18} />
        </button>
      </div>

      <!-- Theme name + base + mode -->
      <div class="mb-5 space-y-3">
        <div>
          <label for="theme-name" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('theme.themeName')}</label>
          <input id="theme-name" type="text" bind:value={themeName} placeholder={$t('theme.namePlaceholder')}
            class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]" />
        </div>
        <div class="flex gap-3">
          <div class="flex-1">
            <label for="base-theme" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('theme.baseTheme')}</label>
            <select id="base-theme" value={baseThemeId}
              onchange={(e) => handleBaseThemeChange((e.target as HTMLSelectElement).value)}
              class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]">
              {#each getAllThemes() as thm (thm.id)}
                <option value={thm.id}>{thm.name}</option>
              {/each}
            </select>
          </div>
          <div>
            <span class="mb-1 block text-xs text-[var(--color-text-secondary)]">Mode</span>
            <div class="flex overflow-hidden rounded-lg border border-[var(--color-border)]">
              <button
                onclick={() => handleModeChange('dark')}
                class="px-3 py-2 text-sm transition-colors {mode === 'dark' ? 'bg-[var(--color-primary)] text-white' : 'bg-[var(--color-bg)] text-[var(--color-text-secondary)]'}"
              >{$t('theme.dark')}</button>
              <button
                onclick={() => handleModeChange('light')}
                class="px-3 py-2 text-sm transition-colors {mode === 'light' ? 'bg-[var(--color-primary)] text-white' : 'bg-[var(--color-bg)] text-[var(--color-text-secondary)]'}"
              >{$t('theme.light')}</button>
            </div>
          </div>
        </div>
      </div>

      {#if !jsonMode}
        <!-- Color pickers -->
        <div class="mb-5 space-y-2">
          {#each colorFields as field (field.key)}
            <div class="flex items-center gap-3">
              <span class="w-24 text-sm text-[var(--color-text-secondary)]">{$t(field.labelKey)}</span>
              <input
                type="color"
                value={baseColors[field.key]}
                oninput={(e) => handleColorChange(field.key, (e.target as HTMLInputElement).value)}
                class="h-8 w-10 cursor-pointer rounded border border-[var(--color-border)] bg-transparent"
              />
              <input
                type="text"
                value={baseColors[field.key]}
                oninput={(e) => handleColorChange(field.key, (e.target as HTMLInputElement).value)}
                class="w-24 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1 font-mono text-xs text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                maxlength="7"
              />
            </div>
          {/each}
        </div>
      {:else}
        <!-- JSON editor -->
        <div class="mb-5">
          <textarea
            value={jsonText}
            oninput={handleJsonInput}
            rows="16"
            spellcheck="false"
            class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 font-mono text-xs text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
          ></textarea>
          {#if jsonError}
            <p class="mt-1 text-xs text-[var(--color-danger)]">{jsonError}</p>
          {/if}
        </div>
      {/if}

      <!-- Toggle JSON mode -->
      <button
        onclick={toggleJsonMode}
        class="mb-5 flex items-center gap-1 text-xs text-[var(--color-primary)] hover:underline"
      >
        <ChevronDown size={12} class={jsonMode ? 'rotate-180' : ''} />
        {$t('theme.advancedJson')}
      </button>

      <!-- Footer -->
      <div class="flex justify-end gap-2">
        <button onclick={handleCancel}
          class="rounded-lg border border-[var(--color-border)] px-4 py-2 text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]">
          {$t('common.cancel')}
        </button>
        <button onclick={handleSave}
          disabled={saving}
          class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white hover:bg-[var(--color-primary-hover)] disabled:opacity-50">
          {$t('common.save')}
        </button>
      </div>
    </div>
  </div>
{/if}
