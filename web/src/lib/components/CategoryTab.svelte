<script lang="ts">
  import DynamicIcon from './DynamicIcon.svelte'

  interface Category {
    id: number
    name: string
    icon: string
    slug: string
  }

  interface Props {
    category: Category
    selected: boolean
    editMode: boolean
    editingTabId: number | null
    dragOverIndex: boolean
    draggedIndex: boolean
    onClick: (catId: number) => void
    onDblClick: (e: MouseEvent, catId: number, name: string, icon: string) => void
    onDragStart: (e: DragEvent) => void
    onDragOver: (e: DragEvent) => void
    onDrop: (e: DragEvent) => void
    onDragEnd: () => void
  }

  const {
    category, selected, editMode, editingTabId,
    dragOverIndex, draggedIndex,
    onClick, onDblClick, onDragStart, onDragOver, onDrop, onDragEnd,
  }: Props = $props()
</script>

<div
  class="group relative flex shrink-0 items-center"
  role="listitem"
  draggable={editMode && editingTabId !== category.id}
  ondragstart={(e) => {
    if (editMode) onDragStart(e)
    else e.preventDefault()
  }}
  ondragover={(e) => {
    if (editMode) onDragOver(e)
  }}
  ondrop={(e) => {
    if (editMode) onDrop(e)
  }}
  ondragend={() => {
    if (editMode) onDragEnd()
  }}
>
  <button
    onclick={() => onClick(category.id)}
    ondblclick={(e) => {
      if (editMode) onDblClick(e, category.id, category.name, category.icon)
    }}
    class="flex items-center gap-2 whitespace-nowrap rounded-full px-4 py-2 text-sm font-medium transition-colors
            {editMode ? 'cursor-grab' : 'cursor-pointer'}
            {selected
      ? 'bg-[var(--color-primary)] text-white shadow-sm'
      : 'border border-[var(--color-border)] text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]'}
            {dragOverIndex && !draggedIndex ? 'ring-2 ring-[var(--color-primary)]' : ''}"
    class:opacity-50={draggedIndex}
    data-testid="category-tab"
  >
    <DynamicIcon name={category.icon} size={16} />
    {category.name}
  </button>
</div>
