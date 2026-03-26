<script lang="ts">
  import { get } from 'svelte/store'
  import { keyboardShortcutsEnabled } from '../stores/keyboardShortcuts'
  import { shortcutHelpOpen, toggleShortcutHelp, closeShortcutHelp } from '../stores/shortcutHelp'
  import { toggleEditMode } from '../stores/editMode'
  import { isEditor } from '../stores/auth'
  import { categories } from '../stores/categories'
  import { currentRoute, navigate } from '../stores/router'

  const allowedRoutes = ['dashboard', 'settings', 'help']

  function isInputFocused(): boolean {
    const el = document.activeElement
    if (!el) return false
    const tag = el.tagName.toLowerCase()
    if (tag === 'input' || tag === 'textarea' || tag === 'select') return true
    if ((el as HTMLElement).isContentEditable) return true
    return false
  }

  function handleKeydown(e: KeyboardEvent) {
    if (!get(keyboardShortcutsEnabled)) return
    if (e.ctrlKey || e.altKey || e.metaKey) return
    if (!allowedRoutes.includes(get(currentRoute))) return

    // Escape always works, even in inputs (to close help modal)
    if (e.key === 'Escape') {
      if (get(shortcutHelpOpen)) {
        closeShortcutHelp()
        e.preventDefault()
      }
      return
    }

    if (isInputFocused()) return

    switch (e.key) {
      case 'e':
      case 'E':
        if (get(isEditor)) {
          toggleEditMode()
          e.preventDefault()
        }
        break

      case '?':
        toggleShortcutHelp()
        e.preventDefault()
        break

      case '/': {
        e.preventDefault()
        const searchInput = document.querySelector<HTMLInputElement>('[data-testid="search-input"]')
        searchInput?.focus()
        break
      }

      case '1': case '2': case '3': case '4': case '5':
      case '6': case '7': case '8': case '9': {
        if (get(currentRoute) !== 'dashboard') break
        const cats = get(categories)
        const sorted = [...cats].sort((a, b) => a.sort_order - b.sort_order)
        const index = parseInt(e.key) - 1
        if (index < sorted.length) {
          navigate('/dashboard/' + sorted[index].slug)
          e.preventDefault()
        }
        break
      }
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />
