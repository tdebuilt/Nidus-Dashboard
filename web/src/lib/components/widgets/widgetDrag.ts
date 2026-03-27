import {
  ROW_UNIT, GRID_COLS, DRAG_THRESHOLD,
  widgetRowSpan, resolveCollisions, compactWidgets,
  type GridWidget,
} from './gridEngine'

export interface DragHandlerOptions {
  gridRef: HTMLElement | null
  effectiveCols: number
  widgets: GridWidget[]
  widget: GridWidget
  onDragStart: (widgetId: number, width: number, height: number) => void
  onPreviewUpdate: (col: number, row: number) => void
  onDragEnd: () => void
  onSaveLayouts: () => void
}

/** Calculate the grid column/row from pointer coordinates */
function calculateGridPos(
  clientX: number, clientY: number,
  gridRef: HTMLElement, cols: number, widgetWidth: number,
): { col: number; row: number } {
  const rect = gridRef.getBoundingClientRect()
  const colPitch = rect.width / cols
  const w = Math.min(widgetWidth, cols)
  const relX = clientX - rect.left
  const relY = clientY - rect.top
  const col = Math.max(0, Math.min(cols - w, Math.round(relX / colPitch - w / 2)))
  const row = Math.max(0, Math.round(relY / ROW_UNIT))
  return { col, row }
}

/** Clone a widget card element as a fixed-position drag ghost */
function createGhost(sourceEl: HTMLElement): HTMLElement {
  const rect = sourceEl.getBoundingClientRect()
  const ghost = sourceEl.cloneNode(true) as HTMLElement
  Object.assign(ghost.style, {
    position: 'fixed', left: `${rect.left}px`, top: `${rect.top}px`,
    width: `${rect.width}px`, height: `${rect.height}px`,
    opacity: '0.7', pointerEvents: 'none', zIndex: '1000',
    transition: 'none', margin: '0',
  })
  document.body.appendChild(ghost)
  return ghost
}

/** Mutable state shared between move and up handlers */
interface DragState {
  activated: boolean
  ghost: HTMLElement | null
}

/** Build the pointermove handler for an active drag session */
function buildMoveHandler(
  state: DragState, startX: number, startY: number, threshold: number,
  sourceEl: HTMLElement | null, cols: number, opts: DragHandlerOptions,
) {
  const { gridRef, widget, onDragStart, onPreviewUpdate } = opts
  return (ev: PointerEvent) => {
    const dx = ev.clientX - startX
    const dy = ev.clientY - startY
    if (!state.activated) {
      if (Math.abs(dx) < threshold && Math.abs(dy) < threshold) return
      state.activated = true
      onDragStart(widget.id, widget.width, widgetRowSpan(widget, gridRef))
      if (sourceEl) state.ghost = createGhost(sourceEl)
    }
    ev.preventDefault()
    if (state.ghost) state.ghost.style.transform = `translate(${dx}px, ${dy}px)`
    if (gridRef) {
      const pos = calculateGridPos(ev.clientX, ev.clientY, gridRef, cols, widget.width)
      onPreviewUpdate(pos.col, pos.row)
    }
  }
}

/** Build the pointerup handler; uses cleanup callback to remove both listeners */
function buildUpHandler(
  state: DragState, origPosX: number, origPosY: number,
  sourceEl: HTMLElement | null, cols: number,
  opts: DragHandlerOptions, cleanup: () => void,
) {
  const { gridRef, widgets, widget, onDragEnd, onSaveLayouts } = opts
  return (ev: PointerEvent) => {
    if (state.ghost) state.ghost.style.display = 'none'
    const elUnder = (state.activated && gridRef)
      ? document.elementFromPoint(ev.clientX, ev.clientY) : null
    if (state.ghost) { state.ghost.remove(); state.ghost = null }
    if (state.activated && gridRef) {
      const pos = calculateGridPos(ev.clientX, ev.clientY, gridRef, cols, widget.width)
      applyDrop(widgets, widget, pos.col, pos.row, origPosX, origPosY, elUnder, sourceEl, gridRef)
      onSaveLayouts()
    }
    onDragEnd()
    cleanup()
  }
}

/**
 * Pointer-down handler for dragging a widget on the grid.
 * Activates after a threshold, creates a ghost, shows a preview, and
 * applies the drop (swap + collision resolution) on pointer up.
 */
export function handleDragPointerDown(e: PointerEvent, opts: DragHandlerOptions): void {
  const state: DragState = { activated: false, ghost: null }
  const sourceEl = (e.target as HTMLElement).closest('.widget-card') as HTMLElement | null
  const cols = opts.effectiveCols > 1 ? opts.effectiveCols : GRID_COLS
  const threshold = e.pointerType === 'touch' ? 15 : DRAG_THRESHOLD

  const ac = new AbortController()
  const onMove = buildMoveHandler(state, e.clientX, e.clientY, threshold, sourceEl, cols, opts)
  const onUp = buildUpHandler(
    state, opts.widget.pos_x, opts.widget.pos_y, sourceEl, cols, opts, () => ac.abort(),
  )
  document.addEventListener('pointermove', onMove, { signal: ac.signal })
  document.addEventListener('pointerup', onUp, { signal: ac.signal })
}

/** Apply the drop: update positions, handle swap, resolve collisions and compact */
function applyDrop(
  widgets: GridWidget[], widget: GridWidget,
  col: number, row: number,
  origPosX: number, origPosY: number,
  elUnder: Element | null, sourceEl: HTMLElement | null,
  gridRef: HTMLElement,
): void {
  let target: GridWidget | undefined
  if (elUnder) {
    const cardUnder = elUnder.closest('.widget-card') as HTMLElement | null
    if (cardUnder && cardUnder !== sourceEl) {
      const targetId = Number(cardUnder.dataset.widgetId)
      if (targetId) target = widgets.find((w) => w.id === targetId)
    }
  }

  widget.pos_x = col
  widget.pos_y = row
  if (target) { target.pos_x = origPosX; target.pos_y = origPosY }

  resolveCollisions(widgets, widget, gridRef)
  compactWidgets(widgets, widget.id, gridRef)
}
