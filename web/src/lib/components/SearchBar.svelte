<script lang="ts">
  import { Search, X, LayoutGrid, Layers } from 'lucide-svelte'
  import { api } from '../api/client'
  import { t } from '../i18n'
  import { navigate } from '../stores/router'
  import { categories } from '../stores/categories'

  interface SearchResult {
    type: string
    id: number
    name: string
    category_id?: number
    category_name?: string
  }

  let query = $state('')
  let results = $state.raw<SearchResult[]>([])
  let showResults = $state(false)
  let timer: ReturnType<typeof setTimeout> | null = null
  let containerEl: HTMLDivElement | undefined = $state()

  function debounceSearch() {
    if (timer) clearTimeout(timer)
    if (query.length < 2) {
      results = []
      showResults = false
      return
    }
    timer = setTimeout(async () => {
      try {
        const data = await api.get<{ results: SearchResult[] }>(`/api/search?q=${encodeURIComponent(query)}`)
        results = data.results ?? []
        showResults = true
      } catch {
        results = []
      }
    }, 300)
  }

  function selectResult(result: SearchResult) {
    const targetId = result.type === 'category' ? result.id : result.category_id
    if (targetId) {
      const cat = $categories.find((c) => c.id === targetId)
      if (cat) navigate(`/dashboard/${cat.slug}`)
    }
    query = ''
    results = []
    showResults = false
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      showResults = false
      query = ''
    }
  }

  function clear() {
    query = ''
    results = []
    showResults = false
  }

  function handleClickOutside(e: MouseEvent) {
    if (containerEl && !containerEl.contains(e.target as Node)) {
      showResults = false
    }
  }
</script>

<svelte:window onclick={handleClickOutside} />

<div class="relative" bind:this={containerEl} data-testid="search-bar">
  <div class="relative flex items-center">
    <Search size={16} class="pointer-events-none absolute start-3 text-[var(--color-text-muted)]" />
    <input
      type="text"
      bind:value={query}
      oninput={debounceSearch}
      onkeydown={handleKeydown}
      placeholder={$t('search.placeholder')}
      role="combobox"
      aria-expanded={showResults}
      aria-controls="search-results-list"
      aria-autocomplete="list"
      class="h-10 w-36 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] ps-9 pe-8 text-sm text-[var(--color-text)] placeholder-[var(--color-text-muted)] outline-none transition-all focus:w-56 sm:h-9 sm:w-48 sm:focus:w-64 focus:border-[var(--color-primary)]"
      data-testid="search-input"
    />
    {#if query.length > 0}
      <button
        onclick={clear}
        class="touch-action-btn absolute end-1 rounded p-1.5 text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
        aria-label={$t('common.close')}
        data-testid="search-clear"
      >
        <X size={14} />
      </button>
    {/if}
  </div>

  {#if showResults}
    <div
      id="search-results-list"
      role="listbox"
      aria-label={$t('search.placeholder')}
      class="absolute top-full end-0 z-50 mt-1 w-[calc(100vw-2rem)] max-w-72 overflow-hidden rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-secondary)] shadow-lg sm:start-0 sm:end-auto sm:w-72"
      data-testid="search-results"
    >
      {#if results.length === 0}
        <div class="px-4 py-3 text-sm text-[var(--color-text-muted)]">
          {$t('search.noResults')}
        </div>
      {:else}
        {#each results as result (result.type + '-' + result.id)}
          <button
            role="option"
            aria-selected="false"
            onclick={() => selectResult(result)}
            class="flex w-full items-center gap-3 px-4 py-2.5 text-start text-sm transition-colors hover:bg-[var(--color-bg-tertiary)]"
            data-testid="search-result-item"
          >
            {#if result.type === 'category'}
              <Layers size={16} class="shrink-0 text-[var(--color-primary)]" />
            {:else}
              <LayoutGrid size={16} class="shrink-0 text-[var(--color-text-muted)]" />
            {/if}
            <div class="min-w-0 flex-1">
              <div class="truncate text-[var(--color-text)]">{result.name}</div>
              {#if result.type === 'widget' && result.category_name}
                <div class="truncate text-xs text-[var(--color-text-muted)]">
                  {$t('search.category')}: {result.category_name}
                </div>
              {/if}
            </div>
            <span class="shrink-0 text-xs text-[var(--color-text-muted)]">
              {result.type === 'category' ? $t('search.category') : $t('search.widget')}
            </span>
          </button>
        {/each}
      {/if}
    </div>
  {/if}
</div>
