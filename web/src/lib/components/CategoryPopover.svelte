<script lang="ts">
  import { Trash2, X } from 'lucide-svelte'
  import { t, translate } from '../i18n'
  import IconPicker from './IconPicker.svelte'

  interface Props {
    open: boolean
    position: { top: number; left: number }
    title: string
    name: string
    icon: string
    showDelete: boolean
    onSave: () => void
    onDelete?: () => void
    onClose: () => void
  }

  let {
    open = $bindable(),
    position,
    title,
    name = $bindable(),
    icon = $bindable(),
    showDelete,
    onSave,
    onDelete,
    onClose,
  }: Props = $props()

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault()
      onSave()
    } else if (e.key === 'Escape') {
      onClose()
    }
  }
</script>

{#if open}
  <button
    class="fixed inset-0 z-40"
    onclick={onClose}
    aria-label="close"
  ></button>
  <div
    class="fixed z-50 w-72 rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-4 shadow-xl"
    style="top: {position.top}px; left: {position.left}px;"
  >
    <div class="mb-3 flex items-center justify-between">
      <span class="text-sm font-semibold text-[var(--color-text)]"
        >{title}</span
      >
      {#if showDelete && onDelete}
        <button
          onclick={onDelete}
          class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"
          title={translate('common.delete')}
          aria-label={translate('common.delete')}
        >
          <Trash2 size={14} />
        </button>
      {:else}
        <button
          onclick={onClose}
          class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
          aria-label={translate('common.close')}
        >
          <X size={14} />
        </button>
      {/if}
    </div>
    <input
      type="text"
      bind:value={name}
      onkeydown={handleKeydown}
      placeholder={translate('category.addCategory')}
      class="mb-3 w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
      data-testid={showDelete ? 'tab-edit-input' : 'category-create-input'}
    />
    <div class="mb-3">
      <IconPicker
        selected={icon}
        onSelect={(i) => (icon = i)}
      />
    </div>
    <div class="flex justify-end gap-2">
      <button
        onclick={onClose}
        class="rounded-lg border border-[var(--color-border)] px-3 py-1.5 text-xs text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]"
      >
        {$t('common.cancel')}
      </button>
      <button
        onclick={onSave}
        disabled={!name.trim()}
        class="rounded-lg bg-[var(--color-primary)] px-3 py-1.5 text-xs text-white hover:bg-[var(--color-primary-hover)] disabled:opacity-50"
      >
        {$t(showDelete ? 'common.save' : 'common.add')}
      </button>
    </div>
  </div>
{/if}
