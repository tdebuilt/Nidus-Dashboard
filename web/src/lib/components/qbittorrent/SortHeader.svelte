<script lang="ts">
  import { ArrowUp, ArrowDown } from 'lucide-svelte'
  import { t } from '../../i18n'

  type SortField = 'name' | 'progress' | 'size' | 'speed_down' | 'speed_up' | 'eta' | 'ratio' | 'added_on' | 'status'

  interface Props {
    sortField: SortField
    sortDirection: 'asc' | 'desc'
    onToggleSort: (field: SortField) => void
  }

  const { sortField, sortDirection, onToggleSort }: Props = $props()

  const columns: { field: SortField; labelKey: string; class?: string }[] = [
    { field: 'name', labelKey: 'qbittorrent.sortName', class: 'flex-1 text-left' },
    { field: 'progress', labelKey: 'qbittorrent.sortProgress' },
    { field: 'size', labelKey: 'qbittorrent.sortSize' },
    { field: 'speed_down', labelKey: 'qbittorrent.sortSpeedDown' },
    { field: 'speed_up', labelKey: 'qbittorrent.sortSpeedUp' },
    { field: 'ratio', labelKey: 'qbittorrent.sortRatio' },
    { field: 'added_on', labelKey: 'qbittorrent.sortAdded' },
  ]
</script>

<div class="flex items-center gap-1 text-[10px] text-[var(--color-text-muted)] px-1">
  {#each columns as col (col.field)}
    <button
      onclick={() => onToggleSort(col.field)}
      class="flex items-center gap-0.5 hover:text-[var(--color-text)] transition-colors {col.class ?? ''}"
    >
      {$t(col.labelKey)}
      {#if sortField === col.field}
        {#if sortDirection === 'asc'}
          <ArrowUp size={10} />
        {:else}
          <ArrowDown size={10} />
        {/if}
      {/if}
    </button>
  {/each}
</div>
