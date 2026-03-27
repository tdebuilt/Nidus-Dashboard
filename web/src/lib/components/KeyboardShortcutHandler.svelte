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

  function handleNumberKey(key: string, e: KeyboardEvent) {
    if (get(currentRoute) !== 'dashboard') return
    const cats = get(categories)
    const sorted = [...cats].sort((a, b) => a.sort_order - b.sort_order)
    const index = parseInt(key) - 1
    if (index < sorted.length) {
      navigate('/dashboard/' + sorted[index].slug)
      e.preventDefault()
    }
  }

  const shortcuts: Record<string, (e: KeyboardEvent) => void> = {
    'e': (e) => { if (get(isEditor)) { toggleEditMode(); e.preventDefault() } },
    'E': (e) => { if (get(isEditor)) { toggleEditMode(); e.preventDefault() } },
    '?': (e) => { toggleShortcutHelp(); e.preventDefault() },
    '/': (e) => {
      e.preventDefault()
      const searchInput = document.querySelector<HTMLInputElement>('[data-testid="search-input"]')
      searchInput?.focus()
    },
  }

  function handleEscape(e: KeyboardEvent): boolean {
    if (e.key !== 'Escape') return false
    if (get(shortcutHelpOpen)) { closeShortcutHelp(); e.preventDefault() }
    return true
  }

  function hasModifier(e: KeyboardEvent): boolean {
    return e.ctrlKey || e.altKey || e.metaKey
  }

  function handleKeydown(e: KeyboardEvent) {
    if (!get(keyboardShortcutsEnabled)) return
    if (hasModifier(e)) return
    if (!allowedRoutes.includes(get(currentRoute))) return
    if (handleEscape(e)) return
    if (isInputFocused()) return

    const handler = shortcuts[e.key]
    if (handler) { handler(e); return }
    if (e.key >= '1' && e.key <= '9') handleNumberKey(e.key, e)
  }
</script>

<svelte:window onkeydown={handleKeydown} />
