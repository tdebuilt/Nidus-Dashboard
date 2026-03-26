<script lang="ts">
  import { EllipsisVertical, Trash2 } from 'lucide-svelte'
  import { t } from '../../i18n'

  interface Props {
    onDelete: () => void
  }

  const { onDelete }: Props = $props()
  let open = $state(false)

  function handleClickOutside(e: MouseEvent) {
    const target = e.target as HTMLElement
    if (!target.closest('[data-overflow-menu]')) {
      open = false
    }
  }
</script>

<svelte:window onclick={open ? handleClickOutside : undefined} />

<div class="relative" data-overflow-menu>
  <button
    onclick={(e) => { e.stopPropagation(); open = !open }}
    class="rounded border border-[var(--color-border)] px-1.5 py-1 text-xs text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]"
    title={$t('common.moreActions')}
    data-testid="service-overflow-btn"
  >
    <EllipsisVertical size={14} />
  </button>
  {#if open}
    <div class="absolute right-0 top-full z-10 mt-1 min-w-[140px] rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-secondary)] py-1 shadow-lg">
      <button
        onclick={() => { onDelete(); open = false }}
        class="flex w-full items-center gap-2 px-3 py-1.5 text-xs text-[var(--color-danger)] hover:bg-[var(--color-error-bg)]"
        data-testid="service-delete-btn"
      >
        <Trash2 size={12} />
        {$t('common.delete')}
      </button>
    </div>
  {/if}
</div>
