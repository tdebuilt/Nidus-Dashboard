import {
  ROW_UNIT, GRID_COLS,
  getColPitch, resolveCollisions, compactWidgets,
  type GridWidget,
} from './gridEngine'

export interface ResizeHandlerOptions {
  gridRef: HTMLElement | null
  effectiveCols: number
  widgets: GridWidget[]
  widget: GridWidget
  onResizeStart: (widgetId: number) => void
  onResizeEnd: () => void
  onSaveLayouts: () => void
  onUpdateGridHeight: () => void
}

/**
 * Pointer-based resize handler for width + height.
 * Locks auto-height to current measured height before resizing starts.
 */
export function handleResizeStart(e: PointerEvent, opts: ResizeHandlerOptions): void {
  e.preventDefault()
  e.stopPropagation()
  const { gridRef, effectiveCols, widgets, widget, onResizeStart, onResizeEnd, onSaveLayouts } = opts
  onResizeStart(widget.id)

  const colPitch = getColPitch(gridRef, effectiveCols)
  const startX = e.clientX
  const startY = e.clientY
  const startW = widget.width
  const widgetEl = (e.target as HTMLElement).closest('.widget-card') as HTMLElement | null
  const startHeightPx = widgetEl?.clientHeight ?? 200

  if (widget.height === 0) {
    widget.height = Math.max(1, Math.ceil(startHeightPx / ROW_UNIT))
  }
  const startH = widget.height

  function onMove(ev: PointerEvent) {
    const dx = ev.clientX - startX
    const dy = ev.clientY - startY
    const cols = effectiveCols > 1 ? effectiveCols : GRID_COLS
    widget.width = Math.max(1, Math.min(cols - widget.pos_x, startW + Math.round(dx / colPitch)))
    widget.height = Math.max(1, startH + Math.round(dy / ROW_UNIT))
    resolveCollisions(widgets, widget, gridRef)
  }

  function onUp() {
    onResizeEnd()
    resolveCollisions(widgets, widget, gridRef)
    compactWidgets(widgets, undefined, gridRef)
    onSaveLayouts()
    document.removeEventListener('pointermove', onMove)
    document.removeEventListener('pointerup', onUp)
  }

  document.addEventListener('pointermove', onMove)
  document.addEventListener('pointerup', onUp)
}

/** Double-click resize handle: reset height to auto (0) */
export function handleResizeReset(
  widget: GridWidget, widgets: GridWidget[],
  gridRef: HTMLElement | null,
  onSaveLayouts: () => void, onUpdateGridHeight: () => void,
): void {
  if (widget.height === 0) return
  widget.height = 0
  requestAnimationFrame(() => {
    resolveCollisions(widgets, widget, gridRef)
    compactWidgets(widgets, undefined, gridRef)
    onUpdateGridHeight()
    onSaveLayouts()
  })
}

/** Toggle auto-height: if auto (0) lock to measured; if fixed, set to 0 */
export function toggleAutoHeight(
  widget: GridWidget, widgets: GridWidget[],
  gridRef: HTMLElement | null,
  onSaveLayouts: () => void, onUpdateGridHeight: () => void,
): void {
  if (widget.height === 0) {
    const el = gridRef?.querySelector(`[data-widget-id="${widget.id}"]`) as HTMLElement | null
    const currentH = el?.clientHeight ?? ROW_UNIT * 3
    widget.height = Math.max(1, Math.round(currentH / ROW_UNIT))
    compactWidgets(widgets, undefined, gridRef)
    onSaveLayouts()
  } else {
    widget.height = 0
    requestAnimationFrame(() => {
      resolveCollisions(widgets, widget, gridRef)
      compactWidgets(widgets, undefined, gridRef)
      onUpdateGridHeight()
      onSaveLayouts()
    })
  }
}
