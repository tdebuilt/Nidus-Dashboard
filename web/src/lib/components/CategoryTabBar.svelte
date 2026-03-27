<script lang="ts">
  import { api } from '../api/client'
  import { translate } from '../i18n'
  import { categories } from '../stores/categories'
  import { editMode } from '../stores/editMode'
  import { toasts } from '../stores/toast'
  import { confirm } from '../stores/confirm'
  import { navigate } from '../stores/router'
  import { Plus, SquarePlus } from 'lucide-svelte'
  import CategoryPopover from './CategoryPopover.svelte'
  import CategoryTab from './CategoryTab.svelte'

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
  const showEditPanel = $derived(editingTabId !== null)

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

  function handleTabDblClick(e: MouseEvent, catId: number, catName: string, catIcon: string) {
    const target = e.currentTarget as HTMLElement
    const rect = target.getBoundingClientRect()
    const panelWidth = 288
    let left = rect.left
    if (left + panelWidth > window.innerWidth - 8) left = window.innerWidth - panelWidth - 8
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
    if (!cat || (editingTabName.trim() === cat.name && editingTabIcon === cat.icon)) {
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

  async function handleTabDelete() {
    const catId = editingTabId
    editingTabId = null
    if (catId === null) return
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
    if (e.dataTransfer) e.dataTransfer.effectAllowed = 'move'
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
      await api.put('/api/categories/reorder', { ids: reordered.map((c) => c.id) })
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
      await api.post('/api/categories', { name: newCategoryName.trim(), icon: newCategoryIcon })
      toasts.success(translate('category.created'))
      newCategoryName = ''
      newCategoryIcon = 'folder'
      showCreateInput = false
      onCategoryUpdate()
    } catch {
      toasts.error(translate('category.createError'))
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
      <CategoryTab
        category={cat}
        selected={selectedCategoryId === cat.id}
        editMode={$editMode}
        {editingTabId}
        dragOverIndex={dragOverTabIndex === i && dragTabIndex !== i}
        draggedIndex={dragTabIndex === i}
        onClick={handleTabClick}
        onDblClick={handleTabDblClick}
        onDragStart={(e) => handleTabDragStart(e, i)}
        onDragOver={(e) => handleTabDragOver(e, i)}
        onDrop={(e) => handleTabDrop(e, i)}
        onDragEnd={handleTabDragEnd}
      />
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

<!-- Add category popover -->
<CategoryPopover
  bind:open={showCreateInput}
  position={createPanelPos}
  title={translate('category.addCategory')}
  bind:name={newCategoryName}
  bind:icon={newCategoryIcon}
  showDelete={false}
  onSave={handleCreateCategory}
  onClose={closeCreatePanel}
/>

<!-- Edit category popover -->
<CategoryPopover
  open={showEditPanel}
  position={editPanelPos}
  title={translate('category.editCategory')}
  bind:name={editingTabName}
  bind:icon={editingTabIcon}
  showDelete={true}
  onSave={saveTabEdit}
  onDelete={handleTabDelete}
  onClose={closeEditPanel}
/>
