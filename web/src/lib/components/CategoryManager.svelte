<script lang="ts">
  import { Plus, Pencil, Trash2, GripVertical } from 'lucide-svelte'
  import { api } from '../api/client'
  import { toasts } from '../stores/toast'
  import { confirm } from '../stores/confirm'
  import { t, translate } from '../i18n'
  import IconPicker from './IconPicker.svelte'

  interface Category {
    id: number
    name: string
    icon: string
    sort_order: number
  }

  interface Props {
    categories?: Category[]
    onUpdate?: () => void
  }

  const { categories = [], onUpdate }: Props = $props()

  let showCreate = $state(false)
  let editingId = $state<number | null>(null)
  let newName = $state('')
  let newIcon = $state('folder')
  let editName = $state('')
  let editIcon = $state('')

  async function handleCreate() {
    if (!newName.trim()) return
    try {
      await api.post('/api/categories', { name: newName.trim(), icon: newIcon })
      toasts.success(translate('category.created'))
      newName = ''
      newIcon = 'folder'
      showCreate = false
      onUpdate?.()
    } catch {
      toasts.error(translate('category.createError'))
    }
  }

  function startEdit(cat: Category) {
    editingId = cat.id
    editName = cat.name
    editIcon = cat.icon
  }

  async function handleEdit() {
    if (!editName.trim() || editingId === null) return
    try {
      await api.put(`/api/categories/${editingId}`, { name: editName.trim(), icon: editIcon })
      toasts.success(translate('category.updated'))
      editingId = null
      onUpdate?.()
    } catch {
      toasts.error(translate('category.updateError'))
    }
  }

  async function handleDelete(cat: Category) {
    const confirmed = await confirm({
      title: translate('category.deleteTitle'),
      message: translate('category.deleteMessage', { name: cat.name }),
      confirmLabel: translate('common.delete'),
      destructive: true,
    })
    if (!confirmed) return
    try {
      await api.delete(`/api/categories/${cat.id}`)
      toasts.success(translate('category.deleted'))
      onUpdate?.()
    } catch {
      toasts.error(translate('category.deleteError'))
    }
  }

  async function handleReorder(fromIndex: number, toIndex: number) {
    if (fromIndex === toIndex) return
    const reordered = [...categories]
    const [moved] = reordered.splice(fromIndex, 1)
    reordered.splice(toIndex, 0, moved)
    const ids = reordered.map((c) => c.id)
    try {
      await api.put('/api/categories/reorder', { ids })
      onUpdate?.()
    } catch {
      toasts.error(translate('category.reorderError'))
    }
  }

  let dragIndex = $state<number | null>(null)

  function handleDragStart(index: number) {
    dragIndex = index
  }

  function handleDragOver(e: DragEvent, index: number) {
    e.preventDefault()
    if (dragIndex !== null && dragIndex !== index) {
      handleReorder(dragIndex, index)
      dragIndex = index
    }
  }

  function handleDragEnd() {
    dragIndex = null
  }
</script>

<div data-testid="category-manager">
  <!-- Category list -->
  <div class="mb-4 space-y-1">
    {#each categories as cat, i (cat.id)}
      {#if editingId === cat.id}
        <div class="rounded-lg border border-[var(--color-primary)] bg-[var(--color-bg)] p-3" data-testid="category-edit-form">
          <input type="text" bind:value={editName}
            class="mb-2 w-full rounded border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-2 py-1 text-sm text-[var(--color-text)] outline-none"
            data-testid="category-edit-name" />
          <IconPicker selected={editIcon} onSelect={(icon) => editIcon = icon} />
          <div class="mt-2 flex justify-end gap-2">
            <button onclick={() => editingId = null}
              class="rounded px-3 py-1 text-xs text-[var(--color-text-muted)] hover:bg-[var(--color-bg-tertiary)]"
              data-testid="category-edit-cancel">{$t('common.cancel')}</button>
            <button onclick={handleEdit}
              class="rounded bg-[var(--color-primary)] px-3 py-1 text-xs text-white"
              data-testid="category-edit-save">{$t('common.save')}</button>
          </div>
        </div>
      {:else}
        <div
          class="flex items-center gap-2 rounded-lg px-2 py-1.5 transition-colors hover:bg-[var(--color-bg-tertiary)]"
          draggable="true"
          ondragstart={() => handleDragStart(i)}
          ondragover={(e) => handleDragOver(e, i)}
          ondragend={handleDragEnd}
          role="listitem"
          data-testid="category-item"
        >
          <GripVertical size={14} class="cursor-grab text-[var(--color-text-muted)]" />
          <span class="flex-1 text-sm text-[var(--color-text)]">{cat.name}</span>
          <button onclick={() => startEdit(cat)} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-text)]" data-testid="category-edit-btn">
            <Pencil size={14} />
          </button>
          <button onclick={() => handleDelete(cat)} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]" data-testid="category-delete-btn">
            <Trash2 size={14} />
          </button>
        </div>
      {/if}
    {/each}
  </div>

  <!-- Create form -->
  {#if showCreate}
    <div class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] p-3" data-testid="category-create-form">
      <input type="text" bind:value={newName} placeholder={$t('category.namePlaceholder')}
        class="mb-2 w-full rounded border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-2 py-1 text-sm text-[var(--color-text)] outline-none"
        data-testid="category-create-name" />
      <IconPicker selected={newIcon} onSelect={(icon) => newIcon = icon} />
      <div class="mt-2 flex justify-end gap-2">
        <button onclick={() => { showCreate = false; newName = ''; newIcon = 'folder' }}
          class="rounded px-3 py-1 text-xs text-[var(--color-text-muted)] hover:bg-[var(--color-bg-tertiary)]"
          data-testid="category-create-cancel">{$t('common.cancel')}</button>
        <button onclick={handleCreate}
          class="rounded bg-[var(--color-primary)] px-3 py-1 text-xs text-white"
          data-testid="category-create-save">{$t('common.add')}</button>
      </div>
    </div>
  {:else}
    <button onclick={() => showCreate = true}
      class="flex w-full items-center justify-center gap-2 rounded-lg border border-dashed border-[var(--color-border)] px-3 py-2 text-sm text-[var(--color-text-muted)] transition-colors hover:border-[var(--color-primary)] hover:text-[var(--color-primary)]"
      data-testid="category-add-btn">
      <Plus size={16} /> {$t('category.addCategory')}
    </button>
  {/if}
</div>
