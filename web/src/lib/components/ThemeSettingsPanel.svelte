<script lang="ts">
  import { t } from '../i18n'
  import type { ThemeDefinition } from '../themes'

  interface Props {
    themeName: string
    baseThemeId: string
    mode: 'dark' | 'light'
    themes: ThemeDefinition[]
    onNameChange: (name: string) => void
    onBaseThemeChange: (themeId: string) => void
    onModeChange: (mode: 'dark' | 'light') => void
  }

  const { themeName, baseThemeId, mode, themes, onNameChange, onBaseThemeChange, onModeChange }: Props = $props()
</script>

<div class="mb-5 space-y-3">
  <div>
    <label for="theme-name" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('theme.themeName')}</label>
    <input id="theme-name" type="text" value={themeName} oninput={(e) => onNameChange((e.target as HTMLInputElement).value)} placeholder={$t('theme.namePlaceholder')}
      class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]" />
  </div>
  <div class="flex gap-3">
    <div class="flex-1">
      <label for="base-theme" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('theme.baseTheme')}</label>
      <select id="base-theme" value={baseThemeId}
        onchange={(e) => onBaseThemeChange((e.target as HTMLSelectElement).value)}
        class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]">
        {#each themes as thm (thm.id)}
          <option value={thm.id}>{thm.name}</option>
        {/each}
      </select>
    </div>
    <div>
      <span class="mb-1 block text-xs text-[var(--color-text-secondary)]">Mode</span>
      <div class="flex overflow-hidden rounded-lg border border-[var(--color-border)]">
        <button
          onclick={() => onModeChange('dark')}
          class="px-3 py-2 text-sm transition-colors {mode === 'dark' ? 'bg-[var(--color-primary)] text-white' : 'bg-[var(--color-bg)] text-[var(--color-text-secondary)]'}"
        >{$t('theme.dark')}</button>
        <button
          onclick={() => onModeChange('light')}
          class="px-3 py-2 text-sm transition-colors {mode === 'light' ? 'bg-[var(--color-primary)] text-white' : 'bg-[var(--color-bg)] text-[var(--color-text-secondary)]'}"
        >{$t('theme.light')}</button>
      </div>
    </div>
  </div>
</div>
