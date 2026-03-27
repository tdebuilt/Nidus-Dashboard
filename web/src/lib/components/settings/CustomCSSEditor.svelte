<script lang="ts">
  import { Code } from 'lucide-svelte'
  import { t } from '../../i18n'

  interface Props {
    customCSS: string
    onChange: (e: Event) => void
    onClear: () => void
  }

  const { customCSS, onChange, onClear }: Props = $props()
</script>

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
    onchange={onChange}
    placeholder={$t('settings.customCssPlaceholder')}
    rows="6"
    spellcheck="false"
    class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 font-mono text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
    data-testid="settings-custom-css-textarea"
  ></textarea>
  {#if customCSS}
    <div class="mt-2 flex justify-end">
      <button onclick={onClear}
        class="rounded-lg bg-[var(--color-bg-tertiary)] px-3 py-1.5 text-sm text-[var(--color-text)] hover:bg-[var(--color-border)]"
        data-testid="settings-custom-css-clear">
        {$t('settings.customCssClear')}
      </button>
    </div>
  {/if}
</section>
