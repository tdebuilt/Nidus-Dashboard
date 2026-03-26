<script lang="ts">
  import { Monitor, Tablet, Smartphone } from 'lucide-svelte'
  import { t } from '../../i18n'
  import { clampResponsiveColumns } from '../../utils/responsiveColumns'

  interface Props {
    columns: number
    columnsTablet: number
    columnsMobile: number
    onchange?: (desktop: number, tablet: number, mobile: number) => void
  }

  const { columns, columnsTablet, columnsMobile, onchange }: Props = $props()

  function select(bp: 'desktop' | 'tablet' | 'mobile', n: number) {
    let d = columns
    let tab = columnsTablet
    let m = columnsMobile

    if (bp === 'desktop') d = n
    else if (bp === 'tablet') tab = n
    else m = n

    const clamped = clampResponsiveColumns(d, tab, m)
    onchange?.(clamped.desktop, clamped.tablet, clamped.mobile)
  }

  const rows: { bp: 'desktop' | 'tablet' | 'mobile'; icon: typeof Monitor; label: string; value: number }[] = $derived([
    { bp: 'desktop', icon: Monitor, label: $t('common.desktop'), value: columns },
    { bp: 'tablet', icon: Tablet, label: $t('common.tablet'), value: columnsTablet },
    { bp: 'mobile', icon: Smartphone, label: $t('common.mobile'), value: columnsMobile },
  ])
</script>

<div class="space-y-1.5">
  <span class="text-sm text-[var(--color-text-secondary)]">{$t('common.columns')}</span>
  {#each rows as row (row.bp)}
    <div class="flex items-center gap-2">
      <row.icon size={14} class="shrink-0 text-[var(--color-text-muted)]" />
      <span class="w-16 shrink-0 text-xs text-[var(--color-text-muted)]">{row.label}</span>
      {#each [1, 2, 3, 4] as n (n)}
        <button
          onclick={() => select(row.bp, n)}
          class="rounded-lg px-3 py-1 text-sm transition-colors"
          class:bg-[var(--color-primary)]={row.value === n}
          class:text-white={row.value === n}
          class:bg-[var(--color-bg-tertiary)]={row.value !== n}
          class:text-[var(--color-text-secondary)]={row.value !== n}
        >{n}</button>
      {/each}
    </div>
  {/each}
</div>
