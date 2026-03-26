<script lang="ts">
  import { Plus, Trash2, ChevronDown, ChevronUp, Pencil, Maximize2 } from 'lucide-svelte'
  import { api } from '../api/client'
  import { toasts } from '../stores/toast'
  import { confirm } from '../stores/confirm'
  import { editMode } from '../stores/editMode'
  import { breakpoint } from '../stores/breakpoint'
  import { t, translate } from '../i18n'
  import AddWidgetDialog from './AddWidgetDialog.svelte'
  import EditWidgetDialog from './EditWidgetDialog.svelte'
  import { getWidget, type WidgetDefinition } from '../widgetRegistry'

  interface Widget {
    id: number
    category_id: number
    type: string
    title: string
    config: string
    collapsed: boolean
    pos_x: number
    pos_y: number
    width: number
    height: number
  }

  interface Props {
    categoryId: number
    widgets?: Widget[]
    active?: boolean
    onUpdate?: () => void
    showAddDialog?: boolean
  }

  let { categoryId, widgets = [], active = true, onUpdate, showAddDialog = $bindable(false) }: Props = $props() // eslint-disable-line prefer-const
  let gridRef = $state<HTMLElement | null>(null)
  let draggingId = $state<number | null>(null)
  const dragOverId = $state<number | null>(null)
  let resizingId = $state<number | null>(null)

  function autoFocus(node: HTMLElement) {
    node.focus()
  }

  // Inline title editing
  let editingTitleId = $state<number | null>(null)
  let editingTitleValue = $state('')

  function startTitleEdit(widget: Widget) {
    editingTitleId = widget.id
    editingTitleValue = widget.title
  }

  async function saveTitleEdit(widget: Widget) {
    const newTitle = editingTitleValue.trim()
    if (!newTitle || newTitle === widget.title) {
      editingTitleId = null
      return
    }
    try {
      await api.put(`/api/widgets/${widget.id}`, {
        type: widget.type,
        title: newTitle,
        config: widget.config,
      })
      onUpdate?.()
    } catch {
      toasts.error(translate('widget.renameError'))
    }
    editingTitleId = null
  }

  function cancelTitleEdit() {
    editingTitleId = null
  }

  function handleTitleKeydown(e: KeyboardEvent, widget: Widget) {
    if (e.key === 'Enter') {
      e.preventDefault()
      saveTitleEdit(widget)
    } else if (e.key === 'Escape') {
      cancelTitleEdit()
    }
  }

  // Collapse toggle
  async function toggleCollapse(widget: Widget) {
    try {
      await api.patch(`/api/widgets/${widget.id}/toggle-collapse`, {
        collapsed: !widget.collapsed,
      })
      onUpdate?.()
    } catch {
      toasts.error(translate('widget.collapseError'))
    }
  }

  // Edit dialog
  let editingWidget = $state<Widget | null>(null)
  let showEditDialog = $state(false)

  function openEditDialog(widget: Widget) {
    editingWidget = widget
    showEditDialog = true
  }

  function closeEditDialog() {
    showEditDialog = false
    editingWidget = null
  }

  async function handleDelete(widget: Widget) {
    const confirmed = await confirm({
      title: translate('widget.deleteTitle'),
      message: translate('widget.deleteMessage', { name: widget.title }),
      confirmLabel: translate('common.delete'),
      destructive: true,
    })
    if (!confirmed) return
    try {
      await api.delete(`/api/widgets/${widget.id}`)
      toasts.success(translate('widget.deleted'))
      onUpdate?.()
    } catch {
      toasts.error(translate('widget.deleteError'))
    }
  }

  const ROW_UNIT = 10
  const GRID_COLS = 24
  const DRAG_THRESHOLD = 5

  // Responsive column count
  const effectiveCols = $derived(
    $breakpoint === 'mobile' ? 1
    : $breakpoint === 'tablet' ? Math.floor(GRID_COLS / 2)
    : GRID_COLS
  )

  // Grid metrics
  function getColPitch() {
    if (!gridRef) return 80
    return gridRef.clientWidth / (effectiveCols > 1 ? effectiveCols : GRID_COLS)
  }

  // --- Free-placement grid helpers ---

  /** Get the effective height of a widget in row units.
   *  For auto-height widgets (height=0), measure from the DOM if available. */
  const ROW_GAP = 16 // vertical spacing between widgets in px
  const GAP_ROWS = Math.ceil(ROW_GAP / ROW_UNIT) // gap in row units (2 rows = 20px)

  function widgetRowSpan(w: Widget): number {
    if (w.height > 0) return w.height + GAP_ROWS
    // Measure content height from DOM (scrollHeight = actual content, not stretched by grid)
    if (gridRef) {
      const el = gridRef.querySelector(`[data-widget-id="${w.id}"]`) as HTMLElement | null
      if (el) {
        const contentHeight = el.scrollHeight + ROW_GAP
        return Math.max(1, Math.ceil(contentHeight / ROW_UNIT))
      }
    }
    return 20 // fallback: ~200px
  }

  /** Check if two widgets overlap in the grid */
  function overlaps(a: Widget, b: Widget): boolean {
    if (a.id === b.id) return false
    const aRight = a.pos_x + a.width
    const bRight = b.pos_x + b.width
    const aBottom = a.pos_y + widgetRowSpan(a)
    const bBottom = b.pos_y + widgetRowSpan(b)
    return a.pos_x < bRight && aRight > b.pos_x && a.pos_y < bBottom && aBottom > b.pos_y
  }

  /** Resolve collisions: push overlapping widgets down using a stable top-down sweep. */
  function resolveCollisions(allWidgets: Widget[], _movedWidget: Widget): Widget[] {
    // Process widgets top-to-bottom; if any widget overlaps with one above it, push it down.
    // Repeat until stable (max 100 iterations for safety).
    for (let iter = 0; iter < 100; iter++) {
      let moved = false
      const sorted = [...allWidgets].sort((a, b) => a.pos_y - b.pos_y || a.pos_x - b.pos_x)
      for (let i = 0; i < sorted.length; i++) {
        for (let j = i + 1; j < sorted.length; j++) {
          if (overlaps(sorted[i], sorted[j])) {
            sorted[j].pos_y = sorted[i].pos_y + widgetRowSpan(sorted[i])
            moved = true
          }
        }
      }
      if (!moved) break
    }
    return allWidgets
  }

  /** Compact all widgets vertically: move each widget up as far as possible without overlapping.
   *  If excludeId is provided, that widget keeps its position (user just placed it there). */
  function compactWidgets(allWidgets: Widget[], excludeId?: number): void {
    allWidgets.sort((a, b) => a.pos_y - b.pos_y || a.pos_x - b.pos_x)
    for (const w of allWidgets) {
      if (w.id === excludeId) continue
      for (let y = 0; y <= w.pos_y; y++) {
        const blocked = allWidgets.some((other) => {
          if (other.id === w.id) return false
          const oRight = other.pos_x + other.width
          const wRight = w.pos_x + w.width
          const oBottom = other.pos_y + widgetRowSpan(other)
          const testBottom = y + widgetRowSpan(w)
          return w.pos_x < oRight && wRight > other.pos_x && y < oBottom && testBottom > other.pos_y
        })
        if (!blocked) {
          w.pos_y = y
          break
        }
      }
    }
  }

  // Sorted widgets for rendering
  const sortedWidgets = $derived(
    [...widgets].sort((a, b) => a.pos_y - b.pos_y || a.pos_x - b.pos_x)
  )

  // Compute total grid rows for the preview overlay
  const _totalGridRows = $derived(
    widgets.reduce((max, w) => Math.max(max, w.pos_y + widgetRowSpan(w)), 1)
  )

  // --- Drag preview state ---
  let previewCol = $state<number | null>(null)
  let previewRow = $state<number | null>(null)
  let previewWidth = $state(1)
  let previewHeight = $state(3)

  // --- Pointer-based drag for free placement ---

  function handleDragPointerDown(e: PointerEvent, widget: Widget) {
    const startX = e.clientX
    const startY = e.clientY
    const origPosX = widget.pos_x
    const origPosY = widget.pos_y
    const threshold = e.pointerType === 'touch' ? 15 : DRAG_THRESHOLD
    let activated = false
    let ghost: HTMLElement | null = null
    const sourceEl = (e.target as HTMLElement).closest('.widget-card') as HTMLElement | null

    previewWidth = widget.width
    previewHeight = widgetRowSpan(widget)

    function createGhost() {
      if (!sourceEl) return
      const rect = sourceEl.getBoundingClientRect()
      ghost = sourceEl.cloneNode(true) as HTMLElement
      ghost.style.position = 'fixed'
      ghost.style.left = `${rect.left}px`
      ghost.style.top = `${rect.top}px`
      ghost.style.width = `${rect.width}px`
      ghost.style.height = `${rect.height}px`
      ghost.style.opacity = '0.7'
      ghost.style.pointerEvents = 'none'
      ghost.style.zIndex = '1000'
      ghost.style.transition = 'none'
      ghost.style.margin = '0'
      document.body.appendChild(ghost)
    }

    function calculateGridPos(clientX: number, clientY: number) {
      if (!gridRef) return { col: 0, row: 0 }
      const cols = effectiveCols > 1 ? effectiveCols : GRID_COLS
      const rect = gridRef.getBoundingClientRect()
      const colPitch = rect.width / cols
      const w = Math.min(widget.width, cols)
      const relX = clientX - rect.left
      const relY = clientY - rect.top
      const col = Math.max(0, Math.min(cols - w, Math.round(relX / colPitch - w / 2)))
      const row = Math.max(0, Math.round(relY / ROW_UNIT))
      return { col, row }
    }

    function onMove(ev: PointerEvent) {
      const dx = ev.clientX - startX
      const dy = ev.clientY - startY
      if (!activated) {
        if (Math.abs(dx) < threshold && Math.abs(dy) < threshold) return
        activated = true
        draggingId = widget.id
        createGhost()
      }
      ev.preventDefault()
      if (ghost) {
        ghost.style.transform = `translate(${dx}px, ${dy}px)`
      }
      // Update placement preview
      const { col, row } = calculateGridPos(ev.clientX, ev.clientY)
      previewCol = col
      previewRow = row
    }

    function onUp(ev: PointerEvent) {
      // Hide ghost so elementFromPoint can see through it, then remove
      if (ghost) ghost.style.display = 'none'
      let elUnder: Element | null = null
      if (activated && gridRef) {
        elUnder = document.elementFromPoint(ev.clientX, ev.clientY)
      }
      if (ghost) {
        ghost.remove()
        ghost = null
      }
      if (activated && gridRef) {
        const { col, row } = calculateGridPos(ev.clientX, ev.clientY)
        const prevX = origPosX
        const prevY = origPosY

        // Only swap if the cursor is visually over another widget's rendered element
        let target: Widget | undefined
        if (elUnder) {
          const cardUnder = elUnder.closest('.widget-card') as HTMLElement | null
          if (cardUnder && cardUnder !== sourceEl) {
            const targetId = Number(cardUnder.dataset.widgetId)
            if (targetId) {
              target = widgets.find((w) => w.id === targetId)
            }
          }
        }

        widget.pos_x = col
        widget.pos_y = row

        if (target) {
          // Swap: move target to the dragged widget's original position
          target.pos_x = prevX
          target.pos_y = prevY
        }

        // Resolve any remaining collisions, then compact others (not the placed widget)
        resolveCollisions(widgets, widget)
        compactWidgets(widgets, widget.id)

        // Save all widget positions
        saveAllWidgetLayouts()
      }
      draggingId = null
      previewCol = null
      previewRow = null
      document.removeEventListener('pointermove', onMove)
      document.removeEventListener('pointerup', onUp)
    }

    document.addEventListener('pointermove', onMove)
    document.addEventListener('pointerup', onUp)
  }

  // Pointer-based resize for width + height
  function handleResizeStart(e: PointerEvent, widget: Widget) {
    e.preventDefault()
    e.stopPropagation()
    resizingId = widget.id
    const colPitch = getColPitch()
    const startX = e.clientX
    const startY = e.clientY
    const startW = widget.width
    const widgetEl = (e.target as HTMLElement).closest('.widget-card') as HTMLElement | null
    const startHeightPx = widgetEl?.clientHeight ?? 200

    // If auto-height, lock to current measured height before resize starts
    if (widget.height === 0) {
      widget.height = Math.max(1, Math.ceil(startHeightPx / ROW_UNIT))
    }
    const startH = widget.height

    function onMove(ev: PointerEvent) {
      const dx = ev.clientX - startX
      const dy = ev.clientY - startY
      const cols = effectiveCols > 1 ? effectiveCols : GRID_COLS
      widget.width = Math.max(1, Math.min(cols - widget.pos_x, startW + Math.round(dx / colPitch)))
      const deltaRows = Math.round(dy / ROW_UNIT)
      widget.height = Math.max(1, startH + deltaRows)
      // Resolve collisions live during resize
      resolveCollisions(widgets, widget)
    }

    function onUp() {
      resizingId = null
      resolveCollisions(widgets, widget)
      compactWidgets(widgets)
      saveAllWidgetLayouts()
      document.removeEventListener('pointermove', onMove)
      document.removeEventListener('pointerup', onUp)
    }

    document.addEventListener('pointermove', onMove)
    document.addEventListener('pointerup', onUp)
  }

  // Double-click resize handle to reset height to auto
  function handleResizeReset(widget: Widget) {
    if (widget.height === 0) return
    widget.height = 0
    // Wait for DOM to re-render in auto mode, then resolve layout
    requestAnimationFrame(() => {
      resolveCollisions(widgets, widget)
      compactWidgets(widgets)
      updateGridMinHeight()
      saveAllWidgetLayouts()
    })
  }

  // Toggle auto-height (fit content)
  function toggleAutoHeight(widget: Widget) {
    if (widget.height === 0) {
      const el = gridRef?.querySelector(`[data-widget-id="${widget.id}"]`) as HTMLElement | null
      const currentH = el?.clientHeight ?? ROW_UNIT * 3
      widget.height = Math.max(1, Math.round(currentH / ROW_UNIT))
      compactWidgets(widgets)
      saveAllWidgetLayouts()
    } else {
      widget.height = 0
      requestAnimationFrame(() => {
        resolveCollisions(widgets, widget)
        compactWidgets(widgets)
        updateGridMinHeight()
        saveAllWidgetLayouts()
      })
    }
  }

  /** Save positions of ALL widgets (needed after collision resolution) */
  function saveAllWidgetLayouts() {
    const layouts = widgets.map((w) => ({
      id: w.id,
      pos_x: w.pos_x,
      pos_y: w.pos_y,
      width: w.width,
      height: w.height,
    }))
    api.put('/api/widgets/layout', { widgets: layouts })
      .then(() => onUpdate?.())
      .catch(() => toasts.error(translate('widget.moveError')))
  }

  function _saveWidgetLayout(widget: Widget) {
    api.put('/api/widgets/layout', {
      widgets: [{ id: widget.id, pos_x: widget.pos_x, pos_y: widget.pos_y, width: widget.width, height: widget.height }],
    })
      .then(() => onUpdate?.())
      .catch(() => toasts.error(translate('widget.moveError')))
  }

  // Next free position for new widgets
  const nextFreeY = $derived(
    widgets.reduce((max, w) => Math.max(max, w.pos_y + widgetRowSpan(w)), 0)
  )

  function getWidgetProps(widget: Widget, def: WidgetDefinition): Record<string, unknown> {
    const props: Record<string, unknown> = { config: widget.config, active }
    for (const key of def.extraProps ?? []) {
      if (key === 'widgetId') props[key] = widget.id
      else if (key === 'widgetType') props[key] = widget.type
      else if (key === 'widgetTitle') props[key] = widget.title
    }
    return props
  }

  // Ensure grid min-height covers all widgets (needed because align-self:start
  // prevents widgets from stretching, so the grid height may be too small)
  function updateGridMinHeight() {
    if (!gridRef || resizingId !== null || !active) return
    const cards = gridRef.querySelectorAll('.widget-card')
    let maxBottom = 0
    const gridTop = gridRef.getBoundingClientRect().top
    cards.forEach((c) => {
      maxBottom = Math.max(maxBottom, c.getBoundingClientRect().bottom - gridTop)
    })
    if (maxBottom > 0) {
      gridRef.style.minHeight = maxBottom + 'px'
    }
  }

  $effect(() => {
    if (!gridRef || widgets.length === 0) return
    if (typeof ResizeObserver === 'undefined') return
    const observer = new ResizeObserver(updateGridMinHeight)
    const cards = gridRef.querySelectorAll('.widget-card')
    cards.forEach((c) => observer.observe(c))
    // Also update after a delay for async content loading
    const timer = setTimeout(updateGridMinHeight, 500)
    return () => {
      observer.disconnect()
      clearTimeout(timer)
    }
  })

  function widgetStyle(widget: Widget): string {
    if (effectiveCols <= 1) return '' // mobile: auto-flow handled by CSS
    const cols = effectiveCols
    const w = Math.min(widget.width, cols)
    const x = Math.min(widget.pos_x, cols - w)
    const col = `grid-column: ${x + 1} / span ${w}`
    const span = widgetRowSpan(widget)
    const row = `grid-row: ${widget.pos_y + 1} / span ${span}`
    if (widget.height > 0) {
      return `${col}; ${row}; height: ${widget.height * ROW_UNIT}px`
    }
    // Auto-height: span + align-self:start (widget doesn't stretch, grid reserves space)
    return `${col}; ${row}; align-self: start`
  }
</script>

<div data-testid="widget-grid">
  {#if widgets.length === 0}
    <div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-[var(--color-border)] py-16 text-center" data-testid="widget-grid-empty">
      <p class="mb-4 text-[var(--color-text-muted)]">{$t('widget.emptyState')}</p>
      {#if $editMode}
        <button onclick={() => showAddDialog = true}
          class="flex items-center gap-2 rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white hover:bg-[var(--color-primary-hover)]"
          data-testid="widget-add-empty">
          <Plus size={16} /> {$t('widget.addWidget')}
        </button>
      {/if}
    </div>
  {:else}
    <div bind:this={gridRef} class="widget-grid grid items-start gap-x-4 2xl:gap-x-6" style="grid-template-columns: repeat({GRID_COLS}, 1fr); grid-auto-rows: {ROW_UNIT}px;">
      {#each sortedWidgets as widget (widget.id)}
        <div
          class="widget-card group relative rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] shadow-sm transition-shadow hover:shadow-md overflow-hidden"
          class:opacity-50={draggingId === widget.id}
          class:ring-2={resizingId === widget.id || dragOverId === widget.id}
          class:ring-[var(--color-primary)]={resizingId === widget.id || dragOverId === widget.id}
          style={widgetStyle(widget)}
          data-widget-id={widget.id}
          data-testid="widget-card"
        >
          <div class="flex h-full flex-col p-4">
            <!-- Header (draggable) -->
            <div
              class="flex items-center justify-between select-none"
              class:mb-2={!widget.collapsed}
              class:touch-none={$editMode}
              class:cursor-grab={$editMode && draggingId !== widget.id}
              class:cursor-grabbing={$editMode && draggingId === widget.id}
              role="toolbar"
              tabindex="-1"
              onpointerdown={(e) => {
                if (!$editMode) return
                if ((e.target as HTMLElement).closest('button, input')) return
                handleDragPointerDown(e, widget)
              }}
            >
              <div class="flex items-center gap-2 min-w-0 flex-1">
                {#if $editMode}
                  <button
                    onclick={() => toggleCollapse(widget)}
                    class="shrink-0 rounded p-0.5 text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
                    title={widget.collapsed ? translate('widget.expand') : translate('widget.collapse')}
                    data-testid="widget-collapse"
                  >
                    {#if widget.collapsed}
                      <ChevronDown size={14} />
                    {:else}
                      <ChevronUp size={14} />
                    {/if}
                  </button>
                {/if}
                {#if editingTitleId === widget.id && $editMode}
                  <input
                    type="text"
                    bind:value={editingTitleValue}
                    onblur={() => saveTitleEdit(widget)}
                    onkeydown={(e) => handleTitleKeydown(e, widget)}
                    use:autoFocus
                    class="min-w-0 flex-1 rounded border border-[var(--color-primary)] bg-[var(--color-bg)] px-1 py-0.5 text-sm font-semibold text-[var(--color-text)] outline-none"
                    data-testid="widget-title-edit"
                  />
                {:else}
                  <h3
                    class="min-w-0 flex-1 truncate text-sm font-semibold text-[var(--color-text)]"
                    ondblclick={() => { if ($editMode) startTitleEdit(widget) }}
                    title={$editMode ? translate('widget.renameHint') : ''}
                    data-testid="widget-title"
                  >{widget.title}</h3>
                {/if}
              </div>
              {#if $editMode}
                <div class="flex shrink-0 items-center gap-1">
                  <button onclick={() => toggleAutoHeight(widget)}
                    class="touch-action-btn rounded p-2 sm:p-1 transition-colors"
                    class:text-[var(--color-primary)]={widget.height === 0}
                    class:text-[var(--color-text-muted)]={widget.height !== 0}
                    title={translate(widget.height === 0 ? 'widget.autoHeightOn' : 'widget.autoHeightOff')}
                    data-testid="widget-auto-height">
                    <Maximize2 size={12} />
                  </button>
                  <button onclick={() => openEditDialog(widget)} class="touch-action-btn rounded p-2 sm:p-1 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]" title={translate('widget.edit')} data-testid="widget-edit">
                    <Pencil size={12} />
                  </button>
                  <button onclick={() => handleDelete(widget)} class="touch-action-btn rounded p-2 sm:p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]" data-testid="widget-delete">
                    <Trash2 size={12} />
                  </button>
                </div>
              {/if}
            </div>
            <!-- Content -->
            {#if !widget.collapsed}
              {#if getWidget(widget.type)}
                {@const widgetDef = getWidget(widget.type)!}
                <div class="mt-2 flex-1 overflow-y-auto rounded-lg bg-[var(--color-bg)] p-2 text-xs text-[var(--color-text-muted)]">
                  <widgetDef.component {...getWidgetProps(widget, widgetDef)} />
                </div>
              {:else}
                <div class="mt-2 flex-1 overflow-y-auto rounded-lg bg-[var(--color-bg)] p-2 text-xs text-[var(--color-text-muted)]">
                  {$t('widget.placeholder')}
                </div>
              {/if}
            {/if}
          </div>
          <!-- Resize handle (double-click to reset height to auto) -->
          {#if !widget.collapsed && $editMode}
            <div
              class="resize-handle absolute bottom-0 end-0 flex h-5 w-5 cursor-se-resize items-center justify-center rounded-tl opacity-0 transition-opacity group-hover:opacity-100 touch-none"
              onpointerdown={(e) => handleResizeStart(e, widget)}
              ondblclick={() => handleResizeReset(widget)}
              role="separator"
              tabindex="-1"
              data-testid="widget-resize"
            >
              <svg width="10" height="10" viewBox="0 0 10 10" class="text-[var(--color-text-muted)]">
                <line x1="9" y1="1" x2="1" y2="9" stroke="currentColor" stroke-width="1.5" />
                <line x1="9" y1="5" x2="5" y2="9" stroke="currentColor" stroke-width="1.5" />
              </svg>
            </div>
          {/if}
        </div>
      {/each}

      <!-- Drag placement preview -->
      {#if previewCol !== null && previewRow !== null && draggingId !== null}
        <div
          class="pointer-events-none rounded-xl border-2 border-dashed border-[var(--color-primary)] bg-[var(--color-primary)]/10"
          style="grid-column: {previewCol + 1} / span {previewWidth}; grid-row: {previewRow + 1} / span {previewHeight};"
          data-testid="widget-drag-preview"
        ></div>
      {/if}
    </div>

    <!-- Add button -->
    {#if $editMode}
      <button onclick={() => showAddDialog = true}
        class="mt-4 flex w-full items-center justify-center gap-2 rounded-lg border border-dashed border-[var(--color-border)] py-3 text-sm text-[var(--color-text-muted)] transition-colors hover:border-[var(--color-primary)] hover:text-[var(--color-primary)]"
        data-testid="widget-add-btn">
        <Plus size={16} /> {$t('widget.addWidget')}
      </button>
    {/if}
  {/if}
</div>

<AddWidgetDialog {categoryId} {nextFreeY} open={showAddDialog} onClose={() => showAddDialog = false} onCreated={onUpdate} />
{#if editingWidget}
  <EditWidgetDialog widget={editingWidget} open={showEditDialog} onClose={closeEditDialog} onUpdated={onUpdate} />
{/if}

<style>
  @media (max-width: 767px) {
    .widget-grid {
      grid-template-columns: 1fr !important;
      grid-auto-rows: auto !important;
    }
    .widget-card {
      grid-column: auto !important;
      grid-row: auto !important;
      height: auto !important;
    }
  }
  @media (min-width: 768px) and (max-width: 1023px) {
    .widget-grid {
      grid-template-columns: repeat(12, 1fr) !important;
    }
  }
</style>
