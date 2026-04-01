<script lang="ts">
  import { ChevronLeft, ChevronRight } from 'lucide-svelte'
  import { t } from '../../i18n'

  interface Props {
    page: number
    totalPages: number
    totalItems: number
    pageSize: number
    onPageChange: (page: number) => void
    onPageSizeChange: (size: number) => void
  }

  const { page, totalPages, totalItems, pageSize, onPageChange, onPageSizeChange }: Props = $props()

  const pageSizes = [10, 20, 50, 100]
</script>

{#if totalPages > 1 || totalItems > 10}
  <div class="flex items-center justify-between text-xs text-[var(--color-text-muted)] pt-1">
    <div class="flex items-center gap-1">
      <select
        value={pageSize}
        onchange={(e) => onPageSizeChange(Number((e.target as HTMLSelectElement).value))}
        class="rounded border border-[var(--color-border)] bg-[var(--color-bg)] px-1 py-0.5 text-xs text-[var(--color-text)]"
        aria-label={$t('qbittorrent.perPage')}
      >
        {#each pageSizes as size (size)}
          <option value={size}>{size}</option>
        {/each}
      </select>
      <span>{$t('qbittorrent.perPage')}</span>
    </div>

    <div class="flex items-center gap-1">
      <button
        onclick={() => onPageChange(page - 1)}
        disabled={page <= 1}
        class="rounded p-0.5 hover:text-[var(--color-text)] disabled:opacity-30"
        aria-label="Previous"
      >
        <ChevronLeft size={14} />
      </button>
      <span>{page} / {totalPages}</span>
      <button
        onclick={() => onPageChange(page + 1)}
        disabled={page >= totalPages}
        class="rounded p-0.5 hover:text-[var(--color-text)] disabled:opacity-30"
        aria-label="Next"
      >
        <ChevronRight size={14} />
      </button>
      <span class="ml-1">({totalItems})</span>
    </div>
  </div>
{/if}
