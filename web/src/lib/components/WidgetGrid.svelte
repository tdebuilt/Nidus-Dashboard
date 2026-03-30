<script lang="ts">
  import { Plus } from 'lucide-svelte'
  import { editMode } from '../stores/editMode'
  import { breakpoint } from '../stores/breakpoint'
  import { t } from '../i18n'
  import AddWidgetDialog from './AddWidgetDialog.svelte'
  import EditWidgetDialog from './EditWidgetDialog.svelte'
  import WidgetPlaceholder from './WidgetPlaceholder.svelte'
  import WidgetCardHeader from './WidgetCardHeader.svelte'
  import WidgetCardContent from './WidgetCardContent.svelte'
  import { loadWidgetComponent } from '../widgetRegistry'
  import {
    ROW_UNIT, GRID_COLS,
    widgetRowSpan,
    type GridWidget,
  } from './widgets/gridEngine'
  import { handleDragPointerDown } from './widgets/widgetDrag'
  import {
    handleResizeStart, handleResizeReset, toggleAutoHeight,
  } from './widgets/widgetResize'
  import { saveAllWidgetLayouts, updateGridMinHeight } from './widgets/layoutPersistence'
  import { saveTitleEdit, toggleCollapse, deleteWidget } from './widgets/widgetActions'

  type Widget = GridWidget

  interface Props {
    categoryId: number
    widgets?: Widget[]
    active?: boolean
    onUpdate?: () => void
    showAddDialog?: boolean
  }

  let { categoryId, widgets = [], active = true, onUpdate, showAddDialog = $bindable(false) }: Props = $props()
  let gridRef = $state<HTMLElement | null>(null)
  let draggingId = $state<number | null>(null)
  let resizingId = $state<number | null>(null)

  // Inline title editing state
  let editingTitleId = $state<number | null>(null)
  let editingTitleValue = $state('')

  // Edit dialog state
  let editingWidget = $state<Widget | null>(null)
  let showEditDialog = $state(false)

  // Drag preview state
  let previewCol = $state<number | null>(null)
  let previewRow = $state<number | null>(null)
  let previewWidth = $state(1)
  let previewHeight = $state(3)

  // Responsive column count
  const effectiveCols = $derived(
    $breakpoint === 'mobile' ? 1
    : $breakpoint === 'tablet' ? Math.floor(GRID_COLS / 2)
    : GRID_COLS
  )

  const sortedWidgets = $derived(
    [...widgets].sort((a, b) => a.pos_y - b.pos_y || a.pos_x - b.pos_x)
  )

  const nextFreeY = $derived(
    widgets.reduce((max, w) => Math.max(max, w.pos_y + widgetRowSpan(w, gridRef)), 0)
  )

  // Lazy-loaded widget components
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let loadedComponents = $state<Record<string, any>>({})

  async function ensureWidgetLoaded(type: string) {
    if (loadedComponents[type]) return
    const comp = await loadWidgetComponent(type)
    if (comp) loadedComponents = { ...loadedComponents, [type]: comp }
  }

  $effect(() => {
    for (const type of new Set(widgets.map(w => w.type))) {
      ensureWidgetLoaded(type)
    }
  })

  // --- Layout persistence wrappers ---

  function saveLayouts() {
    saveAllWidgetLayouts(widgets, onUpdate)
  }

  function refreshGridHeight() {
    updateGridMinHeight(gridRef, resizingId, active)
  }

  $effect(() => {
    if (!gridRef || widgets.length === 0) return
    if (typeof ResizeObserver === 'undefined') return
    const observer = new ResizeObserver(refreshGridHeight)
    gridRef.querySelectorAll('.widget-card').forEach((c) => observer.observe(c))
    const timer = setTimeout(refreshGridHeight, 500)
    return () => { observer.disconnect(); clearTimeout(timer) }
  })

  // --- Widget style computation ---

  function widgetStyle(widget: Widget): string {
    if (effectiveCols <= 1) return ''
    const cols = effectiveCols
    const w = Math.min(widget.width, cols)
    const x = Math.min(widget.pos_x, cols - w)
    const col = `grid-column: ${x + 1} / span ${w}`
    const span = widgetRowSpan(widget, gridRef)
    const row = `grid-row: ${widget.pos_y + 1} / span ${span}`
    if (widget.height > 0) return `${col}; ${row}; height: ${widget.height * ROW_UNIT}px`
    return `${col}; ${row}; align-self: start`
  }

  // --- Widget actions ---

  function startTitleEdit(widget: Widget) {
    editingTitleId = widget.id
    editingTitleValue = widget.title
  }

  async function onSaveTitleEdit(widget: Widget) {
    await saveTitleEdit(widget, editingTitleValue, onUpdate)
    editingTitleId = null
  }

  function handleTitleKeydown(e: KeyboardEvent, widget: Widget) {
    if (e.key === 'Enter') { e.preventDefault(); onSaveTitleEdit(widget) }
    else if (e.key === 'Escape') editingTitleId = null
  }

  // --- Drag handler wiring ---

  function onWidgetDragStart(e: PointerEvent, widget: Widget) {
    handleDragPointerDown(e, {
      gridRef, effectiveCols, widgets, widget,
      onDragStart: (id, w, h) => { draggingId = id; previewWidth = w; previewHeight = h },
      onPreviewUpdate: (col, row) => { previewCol = col; previewRow = row },
      onDragEnd: () => { draggingId = null; previewCol = null; previewRow = null },
      onSaveLayouts: saveLayouts,
    })
  }

  // --- Resize handler wiring ---

  function onWidgetResizeStart(e: PointerEvent, widget: Widget) {
    handleResizeStart(e, {
      gridRef, effectiveCols, widgets, widget,
      onResizeStart: (id) => { resizingId = id },
      onResizeEnd: () => { resizingId = null },
      onSaveLayouts: saveLayouts,
      onUpdateGridHeight: refreshGridHeight,
    })
  }

  function onWidgetResizeReset(widget: Widget) {
    handleResizeReset(widget, widgets, gridRef, saveLayouts, refreshGridHeight)
  }

  function onWidgetAutoHeight(widget: Widget) {
    toggleAutoHeight(widget, widgets, gridRef, saveLayouts, refreshGridHeight)
  }
</script>

<div data-testid="widget-grid">
  {#if widgets.length === 0}
    <WidgetPlaceholder onAddWidget={() => showAddDialog = true} />
  {:else}
    <div bind:this={gridRef} class="widget-grid grid items-start gap-x-4 2xl:gap-x-6" style="grid-template-columns: repeat({GRID_COLS}, 1fr); grid-auto-rows: {ROW_UNIT}px;">
      {#each sortedWidgets as widget (widget.id)}
        <div
          class="widget-card group relative rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] shadow-sm transition-shadow hover:shadow-md overflow-hidden"
          class:opacity-50={draggingId === widget.id}
          class:ring-2={resizingId === widget.id}
          class:ring-[var(--color-primary)]={resizingId === widget.id}
          style={widgetStyle(widget)}
          data-widget-id={widget.id}
          data-testid="widget-card"
        >
          <div class="flex h-full flex-col p-4">
            <WidgetCardHeader
              {widget}
              {editingTitleId}
              {editingTitleValue}
              {draggingId}
              collapsed={widget.collapsed}
              onCollapse={() => toggleCollapse(widget, onUpdate)}
              onTitleStartEdit={() => startTitleEdit(widget)}
              onTitleSave={() => onSaveTitleEdit(widget)}
              onTitleValueChange={(v) => { editingTitleValue = v }}
              onTitleKeydown={(e) => handleTitleKeydown(e, widget)}
              onEdit={() => { editingWidget = widget; showEditDialog = true }}
              onDelete={() => deleteWidget(widget, onUpdate)}
              onAutoHeight={() => onWidgetAutoHeight(widget)}
              onDragStart={(e) => onWidgetDragStart(e, widget)}
            />
            {#if !widget.collapsed}
              <WidgetCardContent {widget} loadedComponent={loadedComponents[widget.type] ?? null} {active} />
            {/if}
          </div>
          {#if !widget.collapsed && $editMode}
            <div
              class="resize-handle absolute bottom-0 end-0 flex h-5 w-5 cursor-se-resize items-center justify-center rounded-tl opacity-0 transition-opacity group-hover:opacity-100 touch-none"
              onpointerdown={(e) => onWidgetResizeStart(e, widget)}
              ondblclick={() => onWidgetResizeReset(widget)}
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

      {#if previewCol !== null && previewRow !== null && draggingId !== null}
        <div
          class="pointer-events-none rounded-xl border-2 border-dashed border-[var(--color-primary)] bg-[var(--color-primary)]/10"
          style="grid-column: {previewCol + 1} / span {previewWidth}; grid-row: {previewRow + 1} / span {previewHeight};"
          data-testid="widget-drag-preview"
        ></div>
      {/if}
    </div>

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
  <EditWidgetDialog widget={editingWidget} open={showEditDialog} onClose={() => { showEditDialog = false; editingWidget = null }} onUpdated={onUpdate} />
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
