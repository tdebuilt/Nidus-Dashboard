<script lang="ts">
  import { Palette } from 'lucide-svelte'
  import { t } from '../../i18n'

  interface Props {
    accentColor: string
    defaultColor: string
    onChange: (e: Event) => void
    onReset: () => void
  }

  const { accentColor, defaultColor, onChange, onReset }: Props = $props()
</script>

<section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-accent">
  <div class="mb-3 flex items-center gap-2">
    <Palette size={18} class="text-[var(--color-text-secondary)]" />
    <h3 class="font-semibold text-[var(--color-text)]">{$t('settings.accentSection')}</h3>
  </div>
  <div class="flex items-center gap-4">
    <input type="color"
      value={accentColor || defaultColor}
      oninput={onChange}
      aria-label={$t('settings.accentSection')}
      class="h-10 w-14 cursor-pointer rounded border border-[var(--color-border)] bg-transparent"
      data-testid="settings-accent-picker"
    />
    <span class="text-sm font-mono text-[var(--color-text-secondary)]">{accentColor || $t('settings.accentDefault')}</span>
    {#if accentColor}
      <button onclick={onReset}
        class="rounded-lg bg-[var(--color-bg-tertiary)] px-3 py-1.5 text-sm text-[var(--color-text)] hover:bg-[var(--color-border)]"
        data-testid="settings-accent-reset">
        {$t('settings.accentReset')}
      </button>
    {/if}
  </div>
</section>
