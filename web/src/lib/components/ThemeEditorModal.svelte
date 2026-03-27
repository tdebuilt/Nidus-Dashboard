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
  import { focusTrap } from '../actions/focusTrap'
  import ThemeColorPickers from './ThemeColorPickers.svelte'
  import ThemeSettingsPanel from './ThemeSettingsPanel.svelte'
  import ThemeJSONEditor from './ThemeJSONEditor.svelte'

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

  function slugify(text: string): string {
    return text.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '').slice(0, 30)
  }

  function buildThemeDefinition(): ThemeDefinition {
    const colors = deriveFullTheme(baseColors)
    const id = $themeEditorState.editingTheme?.id ?? `custom-${slugify(themeName)}`
    return { id, name: themeName, author: 'Custom', mode, colors }
  }

  function applyPreview() {
    applyTheme(buildThemeDefinition())
  }

  function handleColorChange(key: string, value: string) {
    baseColors = { ...baseColors, [key]: value, mode }
    applyPreview()
  }

  function handleModeChange(newMode: 'dark' | 'light') {
    mode = newMode
    baseColors = { ...baseColors, mode: newMode }
    applyPreview()
  }

  function handleBaseThemeChange(themeId: string) {
    baseThemeId = themeId
    const base = getAllThemes().find(t => t.id === themeId)
    if (base) {
      baseColors = extractBaseColors(base.colors, base.mode)
      mode = base.mode
      applyPreview()
    }
  }

  function handleNameChange(name: string) {
    themeName = name
  }

  function toggleJsonMode() {
    if (!jsonMode) {
      jsonText = JSON.stringify(buildThemeDefinition(), null, 2)
      jsonError = ''
    } else {
      try {
        const parsed = JSON.parse(jsonText)
        const result = parseThemeJSON(parsed)
        if (typeof result === 'string') { jsonError = result; return }
        themeName = result.name
        mode = result.mode
        baseColors = extractBaseColors(result.colors, result.mode)
        jsonError = ''
      } catch (e) { jsonError = (e as Error).message; return }
    }
    jsonMode = !jsonMode
  }

  function handleJsonInput(e: Event) {
    jsonText = (e.target as HTMLTextAreaElement).value
    try {
      const parsed = JSON.parse(jsonText)
      const result = parseThemeJSON(parsed)
      if (typeof result === 'string') { jsonError = result; return }
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

  function validateAndParseTheme(): ThemeDefinition | null {
    if (!themeName.trim()) {
      toasts.error(translate('theme.themeNameRequired'))
      return null
    }
    if (!jsonMode) return buildThemeDefinition()
    try {
      const parsed = JSON.parse(jsonText)
      const result = parseThemeJSON(parsed)
      if (typeof result === 'string') {
        toasts.error(translate('theme.jsonError', { message: result }))
        return null
      }
      return result
    } catch (e) {
      toasts.error(translate('theme.jsonError', { message: (e as Error).message }))
      return null
    }
  }

  async function persistTheme(def: ThemeDefinition, dbId: number | undefined) {
    const themeJSON = JSON.stringify(def)
    if (dbId) {
      const ok = await customThemes.update(dbId, themeName, themeJSON)
      if (ok) { setTheme(def.id); toasts.success(translate('theme.themeUpdated')); closeThemeEditor() }
      else { toasts.error(translate('theme.themeError')) }
    } else {
      const record = await customThemes.create(themeName, themeJSON)
      if (record) { setTheme(def.id); toasts.success(translate('theme.themeCreated')); closeThemeEditor() }
      else { toasts.error(translate('theme.themeError')) }
    }
  }

  async function handleSave() {
    const def = validateAndParseTheme()
    if (!def) return
    saving = true
    await persistTheme(def, $themeEditorState.editingDbId)
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
        const darkTheme = getAllThemes().find(t => t.id === 'dark')
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
    <div class="w-full max-w-2xl max-h-[90vh] overflow-y-auto rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-6 shadow-2xl animate-[dialogIn_0.2s_ease-out]" role="dialog" aria-modal="true" use:focusTrap={{ onClose: handleCancel }}>
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

      <ThemeSettingsPanel
        {themeName}
        {baseThemeId}
        {mode}
        themes={getAllThemes()}
        onNameChange={handleNameChange}
        onBaseThemeChange={handleBaseThemeChange}
        onModeChange={handleModeChange}
      />

      {#if !jsonMode}
        <ThemeColorPickers colors={baseColors} onColorChange={handleColorChange} />
      {:else}
        <ThemeJSONEditor {jsonText} {jsonError} onInput={handleJsonInput} />
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
