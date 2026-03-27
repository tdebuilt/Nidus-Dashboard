const FOCUSABLE = 'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'

export function focusTrap(node: HTMLElement, options?: { onClose?: () => void }) {
  const previouslyFocused = document.activeElement as HTMLElement | null

  // Focus first focusable element
  const first = node.querySelector<HTMLElement>(FOCUSABLE)
  first?.focus()

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      options?.onClose?.()
      return
    }
    if (e.key !== 'Tab') return

    const focusable = [...node.querySelectorAll<HTMLElement>(FOCUSABLE)]
    if (!focusable.length) return

    const firstEl = focusable[0]
    const lastEl = focusable[focusable.length - 1]

    if (e.shiftKey && document.activeElement === firstEl) {
      e.preventDefault()
      lastEl.focus()
    } else if (!e.shiftKey && document.activeElement === lastEl) {
      e.preventDefault()
      firstEl.focus()
    }
  }

  node.addEventListener('keydown', handleKeydown)

  return {
    destroy() {
      node.removeEventListener('keydown', handleKeydown)
      previouslyFocused?.focus()
    },
  }
}
