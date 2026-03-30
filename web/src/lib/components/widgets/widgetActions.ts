import { api } from '../../api/client'
import { toasts } from '../../stores/toast'
import { confirm } from '../../stores/confirm'
import { translate } from '../../i18n'
import type { GridWidget } from './gridEngine'

/** Save an inline title edit via the API. Returns true if title changed. */
export async function saveTitleEdit(
  widget: GridWidget,
  newTitle: string,
  onUpdate?: () => void,
): Promise<boolean> {
  const trimmed = newTitle.trim()
  if (!trimmed || trimmed === widget.title) return false
  try {
    await api.put(`/api/widgets/${widget.id}`, {
      type: widget.type, title: trimmed, config: widget.config,
    })
    onUpdate?.()
    return true
  } catch {
    toasts.error(translate('widget.renameError'))
    return false
  }
}

/** Toggle the collapsed state of a widget. */
export async function toggleCollapse(
  widget: GridWidget,
  onUpdate?: () => void,
): Promise<void> {
  try {
    await api.patch(`/api/widgets/${widget.id}/toggle-collapse`, {
      collapsed: !widget.collapsed,
    })
    onUpdate?.()
  } catch {
    toasts.error(translate('widget.collapseError'))
  }
}

/** Show a confirm dialog and delete the widget if confirmed. */
export async function deleteWidget(
  widget: GridWidget,
  onUpdate?: () => void,
): Promise<void> {
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
