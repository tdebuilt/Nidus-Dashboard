<script lang="ts">
  import type { BaseColors } from '../themes/color-utils'
  import { t } from '../i18n'

  const colorFields = [
    { key: 'bg', labelKey: 'theme.colorBackground' },
    { key: 'text', labelKey: 'theme.colorText' },
    { key: 'primary', labelKey: 'theme.colorPrimary' },
    { key: 'accent', labelKey: 'theme.colorAccent' },
    { key: 'danger', labelKey: 'theme.colorDanger' },
    { key: 'success', labelKey: 'theme.colorSuccess' },
    { key: 'warning', labelKey: 'theme.colorWarning' },
  ] as const

  interface Props {
    colors: BaseColors
    onColorChange: (key: string, value: string) => void
  }

  const { colors, onColorChange }: Props = $props()
</script>

<div class="mb-5 space-y-2">
  {#each colorFields as field (field.key)}
    <div class="flex items-center gap-3">
      <span class="w-24 text-sm text-[var(--color-text-secondary)]">{$t(field.labelKey)}</span>
      <input
        type="color"
        value={colors[field.key]}
        oninput={(e) => onColorChange(field.key, (e.target as HTMLInputElement).value)}
        class="h-8 w-10 cursor-pointer rounded border border-[var(--color-border)] bg-transparent"
      />
      <input
        type="text"
        value={colors[field.key]}
        oninput={(e) => onColorChange(field.key, (e.target as HTMLInputElement).value)}
        class="w-24 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1 font-mono text-xs text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
        maxlength="7"
      />
    </div>
  {/each}
</div>
