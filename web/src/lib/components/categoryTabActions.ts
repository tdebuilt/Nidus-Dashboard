import { api } from '../api/client'
import { translate } from '../i18n'
import { toasts } from '../stores/toast'
import { confirm } from '../stores/confirm'
import type { Category } from '../stores/categories'

export type PopoverPosition = { top: number; left: number }

const POPOVER_WIDTH = 288
const POPOVER_EDGE_MARGIN = 8
const POPOVER_OFFSET_TOP = 4

/**
 * Compute popover position anchored below a target element,
 * clamped to stay within the viewport.
 */
export function computePopoverPosition(
  rect: DOMRect,
  anchor: 'left' | 'right' = 'left',
): PopoverPosition {
  let left = anchor === 'right' ? rect.right - POPOVER_WIDTH : rect.left
  if (left + POPOVER_WIDTH > window.innerWidth - POPOVER_EDGE_MARGIN) {
    left = window.innerWidth - POPOVER_WIDTH - POPOVER_EDGE_MARGIN
  }
  if (left < POPOVER_EDGE_MARGIN) left = POPOVER_EDGE_MARGIN
  return { top: rect.bottom + POPOVER_OFFSET_TOP, left }
}

/** Update a category's name and icon via API. */
export async function saveCategory(
  catId: number,
  name: string,
  icon: string,
  onSuccess: () => void,
): Promise<void> {
  try {
    await api.put(`/api/categories/${catId}`, { name, icon })
    toasts.success(translate('category.updated'))
    onSuccess()
  } catch {
    toasts.error(translate('category.updateError'))
  }
}

/** Confirm and delete a category via API. */
export async function deleteCategory(
  cat: Category,
  onSuccess: () => void,
): Promise<void> {
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
    onSuccess()
  } catch {
    toasts.error(translate('category.deleteError'))
  }
}

/** Reorder categories by sending the new ID order to the API. */
export async function reorderCategories(
  currentCategories: Category[],
  fromIndex: number,
  toIndex: number,
  onSuccess: () => void,
): Promise<void> {
  const reordered = [...currentCategories]
  const [moved] = reordered.splice(fromIndex, 1)
  reordered.splice(toIndex, 0, moved)
  try {
    await api.put('/api/categories/reorder', { ids: reordered.map((c) => c.id) })
    onSuccess()
  } catch {
    toasts.error(translate('category.reorderError'))
  }
}

/** Create a new category via API. */
export async function createCategory(
  name: string,
  icon: string,
  onSuccess: () => void,
): Promise<void> {
  try {
    await api.post('/api/categories', { name, icon })
    toasts.success(translate('category.created'))
    onSuccess()
  } catch {
    toasts.error(translate('category.createError'))
  }
}
