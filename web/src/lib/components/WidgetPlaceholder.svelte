<script lang="ts">
  import { Plus } from 'lucide-svelte'
  import { editMode } from '../stores/editMode'
  import { t } from '../i18n'

  interface Props {
    onAddWidget: () => void
  }

  const { onAddWidget }: Props = $props()
</script>

<div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-[var(--color-border)] py-16 text-center" data-testid="widget-grid-empty">
  <p class="mb-4 text-[var(--color-text-muted)]">{$t('widget.emptyState')}</p>
  {#if $editMode}
    <button onclick={onAddWidget}
      class="flex items-center gap-2 rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white hover:bg-[var(--color-primary-hover)]"
      data-testid="widget-add-empty">
      <Plus size={16} /> {$t('widget.addWidget')}
    </button>
  {/if}
</div>
