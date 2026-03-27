<script lang="ts">
  import { ChevronDown, ChevronUp, Pencil, Trash2, Maximize2 } from 'lucide-svelte'
  import { editMode } from '../stores/editMode'
  import { translate } from '../i18n'
  import type { GridWidget } from './widgets/gridEngine'

  interface Props {
    widget: GridWidget
    editingTitleId: number | null
    editingTitleValue: string
    draggingId: number | null
    collapsed: boolean
    onCollapse: () => void
    onTitleStartEdit: () => void
    onTitleSave: () => void
    onTitleValueChange: (value: string) => void
    onTitleKeydown: (e: KeyboardEvent) => void
    onEdit: () => void
    onDelete: () => void
    onAutoHeight: () => void
    onDragStart: (e: PointerEvent) => void
  }

  const {
    widget, editingTitleId, editingTitleValue, draggingId,
    collapsed, onCollapse, onTitleStartEdit, onTitleSave,
    onTitleValueChange, onTitleKeydown,
    onEdit, onDelete, onAutoHeight, onDragStart,
  }: Props = $props()

  function autoFocus(node: HTMLElement) {
    node.focus()
  }

  function handlePointerDown(e: PointerEvent) {
    if (!$editMode) return
    if ((e.target as HTMLElement).closest('button, input')) return
    onDragStart(e)
  }
</script>

<div
  class="flex items-center justify-between select-none"
  class:mb-2={!collapsed}
  class:touch-none={$editMode}
  class:cursor-grab={$editMode && draggingId !== widget.id}
  class:cursor-grabbing={$editMode && draggingId === widget.id}
  role="toolbar"
  tabindex="-1"
  onpointerdown={handlePointerDown}
>
  <div class="flex items-center gap-2 min-w-0 flex-1">
    {#if $editMode}
      <button
        onclick={onCollapse}
        class="shrink-0 rounded p-0.5 text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
        title={collapsed ? translate('widget.expand') : translate('widget.collapse')}
        data-testid="widget-collapse"
      >
        {#if collapsed}
          <ChevronDown size={14} />
        {:else}
          <ChevronUp size={14} />
        {/if}
      </button>
    {/if}
    {#if editingTitleId === widget.id && $editMode}
      <input
        type="text"
        value={editingTitleValue}
        oninput={(e) => onTitleValueChange((e.target as HTMLInputElement).value)}
        onblur={onTitleSave}
        onkeydown={onTitleKeydown}
        use:autoFocus
        class="min-w-0 flex-1 rounded border border-[var(--color-primary)] bg-[var(--color-bg)] px-1 py-0.5 text-sm font-semibold text-[var(--color-text)] outline-none"
        data-testid="widget-title-edit"
      />
    {:else}
      <h3
        class="min-w-0 flex-1 truncate text-sm font-semibold text-[var(--color-text)]"
        ondblclick={() => { if ($editMode) onTitleStartEdit() }}
        title={$editMode ? translate('widget.renameHint') : ''}
        data-testid="widget-title"
      >{widget.title}</h3>
    {/if}
  </div>
  {#if $editMode}
    <div class="flex shrink-0 items-center gap-1">
      <button onclick={onAutoHeight}
        class="touch-action-btn rounded p-2 sm:p-1 transition-colors"
        class:text-[var(--color-primary)]={widget.height === 0}
        class:text-[var(--color-text-muted)]={widget.height !== 0}
        title={translate(widget.height === 0 ? 'widget.autoHeightOn' : 'widget.autoHeightOff')}
        data-testid="widget-auto-height">
        <Maximize2 size={12} />
      </button>
      <button onclick={onEdit} class="touch-action-btn rounded p-2 sm:p-1 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]" title={translate('widget.edit')} data-testid="widget-edit">
        <Pencil size={12} />
      </button>
      <button onclick={onDelete} class="touch-action-btn rounded p-2 sm:p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]" title={translate('common.delete')} aria-label={translate('common.delete')} data-testid="widget-delete">
        <Trash2 size={12} />
      </button>
    </div>
  {/if}
</div>
