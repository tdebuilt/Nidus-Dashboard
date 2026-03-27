<script lang="ts">
  import { api } from '../api/client'
  import { t, translate } from '../i18n'
  import { categories } from '../stores/categories'
  import { editMode } from '../stores/editMode'
  import { toasts } from '../stores/toast'
  import { confirm } from '../stores/confirm'
  import { navigate } from '../stores/router'
  import { Plus, SquarePlus, Trash2, X } from 'lucide-svelte'
  import IconPicker from './IconPicker.svelte'
  import DynamicIcon from './DynamicIcon.svelte'

  interface Props {
    selectedCategoryId: number | null
    onSelect: (catId: number) => void
    onCategoryUpdate: () => void
    onShowAddWidget: () => void
  }

  const { selectedCategoryId, onSelect, onCategoryUpdate, onShowAddWidget }: Props = $props()

  // Inline tab editing
  let editingTabId = $state<number | null>(null)
  let editingTabName = $state('')
  let editingTabIcon = $state('')
  let editPanelPos = $state({ top: 0, left: 0 })

  // Tab drag state
  let dragTabIndex = $state<number | null>(null)
  let dragOverTabIndex = $state<number | null>(null)

  // Create category
  let showCreateInput = $state(false)
  let newCategoryName = $state('')
  let newCategoryIcon = $state('folder')
  let addBtnRef = $state<HTMLElement | null>(null)
  let createPanelPos = $state({ top: 0, left: 0 })

  export function openCreatePanel() {
    showCreateInput = true
  }

  // --- Tab interactions ---

  function handleTabClick(catId: number) {
    if (editingTabId !== null) return
    onSelect(catId)
    const cat = $categories.find((c) => c.id === catId)
    if (cat) navigate(`/dashboard/${cat.slug}`)
  }

  function handleTabDblClick(
    e: MouseEvent,
    catId: number,
    catName: string,
    catIcon: string,
  ) {
    const target = e.currentTarget as HTMLElement
    const rect = target.getBoundingClientRect()
    const panelWidth = 288
    let left = rect.left
    if (left + panelWidth > window.innerWidth - 8)
      left = window.innerWidth - panelWidth - 8
    if (left < 8) left = 8
    editPanelPos = { top: rect.bottom + 4, left }
    editingTabId = catId
    editingTabName = catName
    editingTabIcon = catIcon
  }

  async function saveTabEdit() {
    if (editingTabId === null || !editingTabName.trim()) {
      editingTabId = null
      return
    }
    const cat = $categories.find((c) => c.id === editingTabId)
    if (
      !cat ||
      (editingTabName.trim() === cat.name && editingTabIcon === cat.icon)
    ) {
      editingTabId = null
      return
    }
    try {
      await api.put(`/api/categories/${editingTabId}`, {
        name: editingTabName.trim(),
        icon: editingTabIcon,
      })
      toasts.success(translate('category.updated'))
      editingTabId = null
      onCategoryUpdate()
    } catch {
      toasts.error(translate('category.updateError'))
      editingTabId = null
    }
  }

  function closeEditPanel() {
    editingTabId = null
    editingTabName = ''
    editingTabIcon = ''
  }

  function handleTabEditKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault()
      saveTabEdit()
    } else if (e.key === 'Escape') {
      closeEditPanel()
    }
  }

  async function handleTabDelete(catId: number) {
    editingTabId = null
    const cat = $categories.find((c) => c.id === catId)
    if (!cat) return
    const confirmed = await confirm({
      title: translate('category.deleteTitle'),
      message: translate('category.deleteMessage', { name: cat.name }),
      confirmLabel: translate('common.delete'),
      destructive: true,
    })
    if (!confirmed) return
    try {
      await api.delete(`/api/categories/${catId}`)
      toasts.success(translate('category.deleted'))
      onCategoryUpdate()
    } catch {
      toasts.error(translate('category.deleteError'))
    }
  }

  // Tab drag & drop for reorder
  function handleTabDragStart(e: DragEvent, index: number) {
    dragTabIndex = index
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move'
    }
  }

  function handleTabDragOver(e: DragEvent, index: number) {
    e.preventDefault()
    dragOverTabIndex = index
  }

  async function handleTabDrop(e: DragEvent, toIndex: number) {
    e.preventDefault()
    if (dragTabIndex === null || dragTabIndex === toIndex) {
      dragTabIndex = null
      dragOverTabIndex = null
      return
    }
    const reordered = [...$categories]
    const [moved] = reordered.splice(dragTabIndex, 1)
    reordered.splice(toIndex, 0, moved)
    try {
      await api.put('/api/categories/reorder', {
        ids: reordered.map((c) => c.id),
      })
      onCategoryUpdate()
    } catch {
      toasts.error(translate('category.reorderError'))
    }
    dragTabIndex = null
    dragOverTabIndex = null
  }

  function handleTabDragEnd() {
    dragTabIndex = null
    dragOverTabIndex = null
  }

  // Create category
  async function handleCreateCategory() {
    if (!newCategoryName.trim()) return
    try {
      await api.post('/api/categories', {
        name: newCategoryName.trim(),
        icon: newCategoryIcon,
      })
      toasts.success(translate('category.created'))
      newCategoryName = ''
      newCategoryIcon = 'folder'
      showCreateInput = false
      onCategoryUpdate()
    } catch {
      toasts.error(translate('category.createError'))
    }
  }

  function handleCreateKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleCreateCategory()
    } else if (e.key === 'Escape') {
      showCreateInput = false
      newCategoryName = ''
      newCategoryIcon = 'folder'
    }
  }

  function toggleCreatePanel() {
    if (showCreateInput) {
      closeCreatePanel()
      return
    }
    if (addBtnRef) {
      const rect = addBtnRef.getBoundingClientRect()
      const panelWidth = 288
      let left = rect.right - panelWidth
      if (left < 8) left = 8
      createPanelPos = { top: rect.bottom + 4, left }
    }
    showCreateInput = true
  }

  function closeCreatePanel() {
    showCreateInput = false
    newCategoryName = ''
    newCategoryIcon = 'folder'
  }
</script>

<!-- Category tabs -->
<div class="tab-bar mb-6 flex items-center gap-2">
  <div
    class="flex min-w-0 flex-1 flex-wrap items-center gap-2 overflow-x-auto"
  >
    {#each $categories as cat, i (cat.id)}
      <div
        class="group relative flex shrink-0 items-center"
        role="listitem"
        draggable={$editMode && editingTabId !== cat.id}
        ondragstart={(e) => {
          if ($editMode) handleTabDragStart(e, i)
          else e.preventDefault()
        }}
        ondragover={(e) => {
          if ($editMode) handleTabDragOver(e, i)
        }}
        ondrop={(e) => {
          if ($editMode) handleTabDrop(e, i)
        }}
        ondragend={() => {
          if ($editMode) handleTabDragEnd()
        }}
      >
        <button
          onclick={() => handleTabClick(cat.id)}
          ondblclick={(e) => {
            if ($editMode) handleTabDblClick(e, cat.id, cat.name, cat.icon)
          }}
          class="flex items-center gap-2 whitespace-nowrap rounded-full px-4 py-2 text-sm font-medium transition-colors
                  {$editMode ? 'cursor-grab' : 'cursor-pointer'}
                  {selectedCategoryId === cat.id
            ? 'bg-[var(--color-primary)] text-white shadow-sm'
            : 'border border-[var(--color-border)] text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]'}
                  {dragOverTabIndex === i && dragTabIndex !== i ? 'ring-2 ring-[var(--color-primary)]' : ''}"
          class:opacity-50={dragTabIndex === i}
          data-testid="category-tab"
        >
          <DynamicIcon name={cat.icon} size={16} />
          {cat.name}
        </button>
      </div>
    {/each}
    <!-- Add category button -->
    {#if $editMode}
      <button
        bind:this={addBtnRef}
        onclick={toggleCreatePanel}
        class="shrink-0 rounded p-2 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]"
        title={translate('category.addCategory')}
        data-testid="category-add-btn"
      >
        <Plus size={16} />
      </button>
    {/if}
  </div>

  <!-- Add widget button -->
  {#if $editMode && selectedCategoryId !== null}
    <button
      onclick={onShowAddWidget}
      class="shrink-0 rounded p-2 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]"
      title={translate('widget.addWidget')}
      data-testid="widget-add-header-btn"
    >
      <SquarePlus size={16} />
    </button>
  {/if}
</div>

<!-- Add category popover (fixed, outside overflow) -->
{#if showCreateInput}
  <button
    class="fixed inset-0 z-40"
    onclick={closeCreatePanel}
    aria-label="close"
  ></button>
  <div
    class="fixed z-50 w-72 rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-4 shadow-xl"
    style="top: {createPanelPos.top}px; left: {createPanelPos.left}px;"
  >
    <div class="mb-3 flex items-center justify-between">
      <span class="text-sm font-semibold text-[var(--color-text)]"
        >{$t('category.addCategory')}</span
      >
      <button
        onclick={closeCreatePanel}
        class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
      >
        <X size={14} />
      </button>
    </div>
    <input
      type="text"
      bind:value={newCategoryName}
      onkeydown={handleCreateKeydown}
      placeholder={translate('category.addCategory')}
      class="mb-3 w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
      data-testid="category-create-input"
    />
    <div class="mb-3">
      <IconPicker
        selected={newCategoryIcon}
        onSelect={(icon) => (newCategoryIcon = icon)}
      />
    </div>
    <div class="flex justify-end gap-2">
      <button
        onclick={closeCreatePanel}
        class="rounded-lg border border-[var(--color-border)] px-3 py-1.5 text-xs text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]"
      >
        {$t('common.cancel')}
      </button>
      <button
        onclick={handleCreateCategory}
        disabled={!newCategoryName.trim()}
        class="rounded-lg bg-[var(--color-primary)] px-3 py-1.5 text-xs text-white hover:bg-[var(--color-primary-hover)] disabled:opacity-50"
      >
        {$t('common.add')}
      </button>
    </div>
  </div>
{/if}

<!-- Edit category popover -->
{#if editingTabId !== null}
  <button
    class="fixed inset-0 z-40"
    onclick={closeEditPanel}
    aria-label="close"
  ></button>
  <div
    class="fixed z-50 w-72 rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-4 shadow-xl"
    style="top: {editPanelPos.top}px; left: {editPanelPos.left}px;"
  >
    <div class="mb-3 flex items-center justify-between">
      <span class="text-sm font-semibold text-[var(--color-text)]"
        >{$t('category.editCategory')}</span
      >
      <button
        onclick={() => {
          const id = editingTabId!
          closeEditPanel()
          handleTabDelete(id)
        }}
        class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"
        title={translate('common.delete')}
      >
        <Trash2 size={14} />
      </button>
    </div>
    <input
      type="text"
      bind:value={editingTabName}
      onkeydown={handleTabEditKeydown}
      class="mb-3 w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
      data-testid="tab-edit-input"
    />
    <div class="mb-3">
      <IconPicker
        selected={editingTabIcon}
        onSelect={(icon) => (editingTabIcon = icon)}
      />
    </div>
    <div class="flex justify-end gap-2">
      <button
        onclick={closeEditPanel}
        class="rounded-lg border border-[var(--color-border)] px-3 py-1.5 text-xs text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]"
      >
        {$t('common.cancel')}
      </button>
      <button
        onclick={saveTabEdit}
        disabled={!editingTabName.trim()}
        class="rounded-lg bg-[var(--color-primary)] px-3 py-1.5 text-xs text-white hover:bg-[var(--color-primary-hover)] disabled:opacity-50"
      >
        {$t('common.save')}
      </button>
    </div>
  </div>
{/if}
