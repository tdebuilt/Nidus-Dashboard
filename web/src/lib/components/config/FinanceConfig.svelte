<script lang="ts">
  import { onMount } from 'svelte'
  import { Plus, Trash2, Search } from 'lucide-svelte'
  import { t } from '../../i18n'
  import { api } from '../../api/client'

  interface SearchResult {
    symbol: string
    name: string
    type: string
    exchange: string
  }

  interface Props {
    value?: string
    onchange?: (value: string) => void
  }

  const { value = '{}', onchange }: Props = $props()

  let symbols = $state<string[]>([])
  let showVolume = $state(true)
  let showMarketCap = $state(true)
  let compact = $state(false)

  let searchQuery = $state('')
  let searchResults = $state<SearchResult[]>([])
  let _searchLoading = $state(false)
  let showDropdown = $state(false)
  let searchTimeout: ReturnType<typeof setTimeout> | null = null

  const MAX_SYMBOLS = 20
  let globalCount = $state(0)
  const usedByOthers = $derived(Math.max(0, globalCount - symbols.length))
  const remaining = $derived(MAX_SYMBOLS - usedByOthers)
  const atLimit = $derived(symbols.length >= remaining)
  const totalUsed = $derived(usedByOthers + symbols.length)

  $effect(() => {
    try {
      const parsed = JSON.parse(value)
      if (parsed.symbols && Array.isArray(parsed.symbols)) {
        symbols = parsed.symbols
      }
      if (typeof parsed.show_volume === 'boolean') showVolume = parsed.show_volume
      if (typeof parsed.show_market_cap === 'boolean') showMarketCap = parsed.show_market_cap
      if (typeof parsed.compact === 'boolean') compact = parsed.compact
    } catch {
      // ignore
    }
  })

  onMount(async () => {
    try {
      const resp = await api.get<{ count: number }>('/api/finance/symbol-count')
      globalCount = resp?.count ?? 0
    } catch {
      globalCount = 0
    }
  })

  function emitChange() {
    onchange?.(
      JSON.stringify({
        symbols: symbols.filter((s) => s.trim()),
        show_volume: showVolume,
        show_market_cap: showMarketCap,
        compact,
      }),
    )
  }

  function addSymbol(symbol: string) {
    const s = symbol.trim().toUpperCase()
    if (!s || symbols.includes(s) || atLimit) return
    symbols = [...symbols, s]
    emitChange()
    searchQuery = ''
    searchResults = []
    showDropdown = false
  }

  function removeSymbol(index: number) {
    symbols = symbols.filter((_, i) => i !== index)
    emitChange()
  }

  function handleSearchInput(e: Event) {
    const q = (e.target as HTMLInputElement).value
    searchQuery = q

    if (searchTimeout) clearTimeout(searchTimeout)

    if (q.length < 2) {
      searchResults = []
      showDropdown = false
      return
    }

    searchTimeout = setTimeout(async () => {
      _searchLoading = true
      try {
        const results = await api.get<SearchResult[]>(
          `/api/finance/search?q=${encodeURIComponent(q)}`,
        )
        searchResults = results ?? []
        showDropdown = searchResults.length > 0
      } catch {
        searchResults = []
        showDropdown = false
      } finally {
        _searchLoading = false
      }
    }, 300)
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault()
      if (searchQuery.trim()) {
        addSymbol(searchQuery)
      }
    }
    if (e.key === 'Escape') {
      showDropdown = false
    }
  }

  function handleBlur() {
    // Delay to allow click on dropdown items
    setTimeout(() => {
      showDropdown = false
    }, 200)
  }
</script>

<div class="space-y-3">
  <!-- Symbol list -->
  <div class="flex items-center justify-between">
    <span class="text-sm text-[var(--color-text-secondary)]">{$t('finance.symbols')}</span>
    <span class="text-xs text-[var(--color-text-muted)]" class:text-[var(--color-danger)]={atLimit}>
      {totalUsed} / {MAX_SYMBOLS}
    </span>
  </div>
  <p class="text-xs text-[var(--color-text-muted)]">{$t('finance.symbolsHint')}</p>

  {#if symbols.length > 0}
    <div class="flex flex-wrap gap-1.5">
      {#each symbols as symbol, i (symbol)}
        <span
          class="inline-flex items-center gap-1 rounded-full bg-[var(--color-bg-tertiary)] px-2.5 py-1 text-sm text-[var(--color-text)]"
        >
          {symbol}
          <button
            onclick={() => removeSymbol(i)}
            class="ms-0.5 rounded-full p-0.5 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"
            title={$t('finance.removeSymbol')}
          >
            <Trash2 size={12} />
          </button>
        </span>
      {/each}
    </div>
  {/if}

  <!-- Search input -->
  <div class="relative">
    <div class="flex items-center gap-2">
      <div class="relative flex-1">
        <Search
          size={14}
          class="pointer-events-none absolute start-2 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]"
        />
        <input
          type="text"
          value={searchQuery}
          oninput={handleSearchInput}
          onkeydown={handleKeydown}
          onfocus={() => {
            if (searchResults.length > 0) showDropdown = true
          }}
          onblur={handleBlur}
          disabled={atLimit}
          placeholder={atLimit ? $t('finance.maxReached') : $t('finance.searchSymbol')}
          class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] py-1.5 ps-7 pe-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)] disabled:opacity-50"
        />
      </div>
      <button
        onclick={() => {
          if (searchQuery.trim()) addSymbol(searchQuery)
        }}
        disabled={atLimit}
        class="flex shrink-0 items-center gap-1 rounded-lg border border-dashed border-[var(--color-border)] px-3 py-1.5 text-sm text-[var(--color-text-muted)] transition-colors hover:border-[var(--color-primary)] hover:text-[var(--color-primary)] disabled:opacity-50"
      >
        <Plus size={14} /> {$t('finance.addSymbol')}
      </button>
    </div>

    <!-- Autocomplete dropdown -->
    {#if showDropdown}
      <div
        class="absolute start-0 end-0 z-50 mt-1 max-h-48 overflow-y-auto rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] shadow-lg"
      >
        {#each searchResults as result (result.symbol)}
          <button
            onclick={() => addSymbol(result.symbol)}
            class="flex w-full items-center justify-between px-3 py-2 text-start text-sm transition-colors hover:bg-[var(--color-bg-secondary)]"
          >
            <div>
              <span class="font-medium text-[var(--color-text)]">{result.symbol}</span>
              <span class="ms-2 text-[var(--color-text-muted)]">{result.name}</span>
            </div>
            <span class="text-xs text-[var(--color-text-muted)]"
              >{result.type} · {result.exchange}</span
            >
          </button>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Display options -->
  <div class="space-y-2 border-t border-[var(--color-border)] pt-3">
    <label class="flex items-center gap-2 text-sm text-[var(--color-text)]">
      <input
        type="checkbox"
        checked={showVolume}
        onchange={() => {
          showVolume = !showVolume
          emitChange()
        }}
        class="rounded"
      />
      {$t('finance.showVolume')}
    </label>
    <label class="flex items-center gap-2 text-sm text-[var(--color-text)]">
      <input
        type="checkbox"
        checked={showMarketCap}
        onchange={() => {
          showMarketCap = !showMarketCap
          emitChange()
        }}
        class="rounded"
      />
      {$t('finance.showMarketCap')}
    </label>
    <label class="flex items-center gap-2 text-sm text-[var(--color-text)]">
      <input
        type="checkbox"
        checked={compact}
        onchange={() => {
          compact = !compact
          emitChange()
        }}
        class="rounded"
      />
      {$t('finance.compact')}
    </label>
  </div>
</div>
