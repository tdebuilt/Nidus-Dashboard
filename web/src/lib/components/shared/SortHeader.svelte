<script lang="ts">
  import { ArrowUp, ArrowDown } from 'lucide-svelte'
  import { t } from '../../i18n'

  interface Column {
    field: string
    labelKey: string
    class?: string
  }

  interface Props {
    columns: Column[]
    sortField: string
    sortDirection: 'asc' | 'desc'
    onToggleSort: (field: string) => void
  }

  const { columns, sortField, sortDirection, onToggleSort }: Props = $props()
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
