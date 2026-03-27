<script lang="ts">
  import { icons } from 'lucide-svelte'
  import type { ComponentType } from 'svelte'
  import { Search } from 'lucide-svelte'
  import { t } from '../i18n'

  interface Props {
    selected?: string
    onSelect?: (icon: string) => void
  }

  let { selected = 'folder', onSelect }: Props = $props()  

  let search = $state('')

  const iconsMap = icons as Record<string, ComponentType>

  // Convert PascalCase to kebab-case for display
  function toKebab(s: string): string {
    return s.replace(/([a-z0-9])([A-Z])/g, '$1-$2').toLowerCase()
  }

  // All icon entries: [kebabName, PascalKey, Component]
  const allIcons: [string, string, ComponentType][] = Object.entries(iconsMap).map(
    ([key, comp]) => [toKebab(key), key, comp]
  )

  // Favorites shown at the top
  const favoriteNames = new Set([
    'server', 'monitor', 'hard-drive', 'cpu', 'network', 'shield', 'cloud',
    'database', 'folder', 'house', 'settings', 'globe', 'lock', 'wifi',
    'activity', 'chart-bar', 'box', 'container', 'layers', 'terminal',
    'code', 'zap', 'eye', 'radio',
  ])

  const favorites = allIcons.filter(([kebab]) => favoriteNames.has(kebab))

  const MAX_RESULTS = 60

  const filteredIcons = $derived.by(() => {
    const q = search.trim().toLowerCase()
    if (!q) return []
    return allIcons
      .filter(([kebab]) => kebab.includes(q))
      .slice(0, MAX_RESULTS)
  })

  const showingIcons = $derived(search.trim() ? filteredIcons : favorites)

  function handleSelect(kebabName: string) {
    selected = kebabName
    onSelect?.(kebabName)
  }

  function autoFocus(node: HTMLElement) {
    node.focus()
  }
</script>

<div data-testid="icon-picker">
  <div class="relative mb-3">
    <Search size={14} class="absolute start-2.5 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" />
    <input
      type="text"
      bind:value={search}
      use:autoFocus
      placeholder={$t('iconPicker.search')}
      class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] py-1.5 ps-8 pe-3 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
      data-testid="icon-search"
    />
  </div>

  {#if !search.trim()}
    <p class="mb-2 text-xs text-[var(--color-text-muted)]">{$t('iconPicker.favorites')}</p>
  {:else if filteredIcons.length === 0}
    <p class="py-4 text-center text-sm text-[var(--color-text-muted)]">{$t('iconPicker.noResults')}</p>
  {:else if filteredIcons.length === MAX_RESULTS}
    <p class="mb-2 text-xs text-[var(--color-text-muted)]">{$t('iconPicker.tooMany')}</p>
  {/if}

  <div class="grid grid-cols-6 gap-2 max-h-48 overflow-y-auto">
    {#each showingIcons as [kebab, _key, IconComponent] (kebab)}
      <button
        type="button"
        onclick={() => handleSelect(kebab)}
        class="flex items-center justify-center rounded-lg p-2 transition-colors
          {selected === kebab
            ? 'bg-[var(--color-primary)] text-white'
            : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]'}"
        title={kebab}
        data-testid="icon-option-{kebab}"
      >
        <IconComponent size={20} />
      </button>
    {/each}
  </div>
</div>
