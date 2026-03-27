<script lang="ts">
  import { Palette, Plus, Pencil, Trash2 } from 'lucide-svelte'
  import { isAdmin } from '../../stores/auth'
  import { t } from '../../i18n'
  import type { ThemeDefinition } from '../../themes'

  interface Props {
    themes: ThemeDefinition[]
    currentTheme: string
    isCustomTheme: (id: string) => boolean
    onSelect: (themeId: string) => void
    onEdit: (themeId: string) => void
    onDelete: (themeId: string, themeName: string) => void
    onCreate: () => void
  }

  const { themes, currentTheme, isCustomTheme, onSelect, onEdit, onDelete, onCreate }: Props = $props()
</script>

<section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-theme">
  <div class="mb-3 flex items-center gap-2">
    <Palette size={18} class="text-[var(--color-text-secondary)]" />
    <h3 class="font-semibold text-[var(--color-text)]">{$t('settings.themeSection')}</h3>
  </div>
  <div class="grid grid-cols-2 gap-3 sm:grid-cols-4" data-testid="settings-theme-select">
    {#each themes as thm (thm.id)}
      <div
        class="group relative flex cursor-pointer flex-col items-center gap-2 rounded-lg border-2 p-3 transition-all {currentTheme === thm.id ? 'border-[var(--color-primary)] shadow-md' : 'border-[var(--color-border)] hover:border-[var(--color-text-muted)]'}"
        style="background-color: {thm.colors['color-bg-primary']}"
        data-testid="theme-card-{thm.id}"
        role="button"
        tabindex="0"
        onclick={() => onSelect(thm.id)}
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
              onclick={(e) => { e.stopPropagation(); onEdit(thm.id) }}
              class="rounded p-1 transition-colors hover:bg-[var(--color-bg-tertiary)]"
              title={$t('theme.editTheme')}
            >
              <Pencil size={12} style="color: {thm.colors['color-text-secondary']}" />
            </button>
            <button
              onclick={(e) => { e.stopPropagation(); onDelete(thm.id, thm.name) }}
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
      onclick={onCreate}
      class="mt-3 flex w-full items-center justify-center gap-1.5 rounded-lg border border-dashed border-[var(--color-border)] px-4 py-2.5 text-sm text-[var(--color-text-secondary)] transition-colors hover:border-[var(--color-primary)] hover:text-[var(--color-primary)]"
      data-testid="create-theme-btn"
    >
      <Plus size={16} />
      {$t('theme.createTheme')}
    </button>
  {/if}
</section>
