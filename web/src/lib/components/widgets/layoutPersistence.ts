import { api } from '../../api/client'
import { toasts } from '../../stores/toast'
import { translate } from '../../i18n'
import type { GridWidget } from './gridEngine'

let _saveTimer: ReturnType<typeof setTimeout> | null = null

/** Debounced batch save of all widget positions/sizes. */
export function saveAllWidgetLayouts(
  widgets: GridWidget[],
  onUpdate?: () => void,
): void {
  if (_saveTimer) clearTimeout(_saveTimer)
  _saveTimer = setTimeout(() => {
    const layouts = widgets.map((w) => ({
      id: w.id, pos_x: w.pos_x, pos_y: w.pos_y, width: w.width, height: w.height,
    }))
    api.put('/api/widgets/layout', { widgets: layouts })
      .then(() => onUpdate?.())
      .catch(() => toasts.error(translate('widget.moveError')))
  }, 500)
}

/** Recalculate grid min-height from rendered widget cards. */
export function updateGridMinHeight(
  gridRef: HTMLElement | null,
  resizingId: number | null,
  active: boolean,
): void {
  if (!gridRef || resizingId !== null || !active) return
  let maxBottom = 0
  const gridTop = gridRef.getBoundingClientRect().top
  gridRef.querySelectorAll('.widget-card').forEach((c) => {
    maxBottom = Math.max(maxBottom, c.getBoundingClientRect().bottom - gridTop)
  })
  if (maxBottom > 0) gridRef.style.minHeight = maxBottom + 'px'
}
