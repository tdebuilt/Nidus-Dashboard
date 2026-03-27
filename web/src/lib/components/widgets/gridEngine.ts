export interface GridWidget {
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

export const ROW_UNIT = 10
export const GRID_COLS = 24
export const DRAG_THRESHOLD = 5

const ROW_GAP = 16 // vertical spacing between widgets in px
const GAP_ROWS = Math.ceil(ROW_GAP / ROW_UNIT) // gap in row units (2 rows = 20px)

/** Get the effective height of a widget in row units.
 *  For auto-height widgets (height=0), measure from the DOM if available. */
export function widgetRowSpan(w: GridWidget, gridRef: HTMLElement | null = null): number {
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

export function getColPitch(gridRef: HTMLElement | null, effectiveCols: number): number {
  if (!gridRef) return 80
  return gridRef.clientWidth / (effectiveCols > 1 ? effectiveCols : GRID_COLS)
}

/** Check if two widgets overlap in the grid */
export function overlaps(a: GridWidget, b: GridWidget, gridRef: HTMLElement | null = null): boolean {
  if (a.id === b.id) return false
  const aRight = a.pos_x + a.width
  const bRight = b.pos_x + b.width
  const aBottom = a.pos_y + widgetRowSpan(a, gridRef)
  const bBottom = b.pos_y + widgetRowSpan(b, gridRef)
  return a.pos_x < bRight && aRight > b.pos_x && a.pos_y < bBottom && aBottom > b.pos_y
}

/** Resolve collisions: push overlapping widgets down using a stable top-down sweep. */
export function resolveCollisions(allWidgets: GridWidget[], _movedWidget: GridWidget, gridRef: HTMLElement | null = null): GridWidget[] {
  // Process widgets top-to-bottom; if any widget overlaps with one above it, push it down.
  // Repeat until stable (max 100 iterations for safety).
  for (let iter = 0; iter < 100; iter++) {
    let moved = false
    const sorted = [...allWidgets].sort((a, b) => a.pos_y - b.pos_y || a.pos_x - b.pos_x)
    for (let i = 0; i < sorted.length; i++) {
      for (let j = i + 1; j < sorted.length; j++) {
        if (overlaps(sorted[i], sorted[j], gridRef)) {
          sorted[j].pos_y = sorted[i].pos_y + widgetRowSpan(sorted[i], gridRef)
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
export function compactWidgets(allWidgets: GridWidget[], excludeId?: number, gridRef: HTMLElement | null = null): void {
  allWidgets.sort((a, b) => a.pos_y - b.pos_y || a.pos_x - b.pos_x)
  for (const w of allWidgets) {
    if (w.id === excludeId) continue
    for (let y = 0; y <= w.pos_y; y++) {
      const blocked = allWidgets.some((other) => {
        if (other.id === w.id) return false
        const oRight = other.pos_x + other.width
        const wRight = w.pos_x + w.width
        const oBottom = other.pos_y + widgetRowSpan(other, gridRef)
        const testBottom = y + widgetRowSpan(w, gridRef)
        return w.pos_x < oRight && wRight > other.pos_x && y < oBottom && testBottom > other.pos_y
      })
      if (!blocked) {
        w.pos_y = y
        break
      }
    }
  }
}
