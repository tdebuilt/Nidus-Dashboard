<script lang="ts">
  import { Search, ChevronLeft, ChevronRight } from 'lucide-svelte'
  import { t } from '../../i18n'
  import MovieCard from './MovieCard.svelte'

  interface RadarrMovie {
    id: number
    title: string
    year: number
    hasFile: boolean
    monitored: boolean
    sizeOnDisk: number
    runtime: number
    status: string
  }

  interface Props {
    movies: RadarrMovie[]
  }

  const { movies }: Props = $props()

  let search = $state('')
  let filter = $state('all')
  let sortBy = $state('title')
  let page = $state(1)
  const perPage = 20

  const filtered = $derived(() => {
    let list = movies
    const q = search.toLowerCase().trim()
    if (q) list = list.filter(m => m.title.toLowerCase().includes(q))
    if (filter === 'downloaded') list = list.filter(m => m.hasFile)
    else if (filter === 'missing') list = list.filter(m => !m.hasFile)
    else if (filter === 'monitored') list = list.filter(m => m.monitored)
    else if (filter === 'unmonitored') list = list.filter(m => !m.monitored)
    return list
  })

  const sorted = $derived(() => {
    const list = [...filtered()]
    if (sortBy === 'title') list.sort((a, b) => a.title.localeCompare(b.title))
    else if (sortBy === 'year') list.sort((a, b) => b.year - a.year)
    else if (sortBy === 'size') list.sort((a, b) => b.sizeOnDisk - a.sizeOnDisk)
    else if (sortBy === 'status') list.sort((a, b) => Number(b.hasFile) - Number(a.hasFile))
    return list
  })

  const totalPages = $derived(Math.max(1, Math.ceil(sorted().length / perPage)))
  const safePage = $derived(Math.min(page, totalPages))
  const paginated = $derived(sorted().slice((safePage - 1) * perPage, safePage * perPage))

  const downloadedCount = $derived(movies.filter(m => m.hasFile).length)
  const missingCount = $derived(movies.length - downloadedCount)

  $effect(() => {
    void search
    void filter
    void sortBy
    page = 1
  })
</script>

<div class="space-y-2">
  <!-- Stats -->
  <div class="flex gap-3 text-xs text-[var(--color-text-muted)]">
    <span class="text-[var(--color-success)]">{downloadedCount} {$t('arr.downloaded')}</span>
    <span>{missingCount} {$t('arr.missing')}</span>
  </div>

  <!-- Search + Filter + Sort -->
  <div class="flex gap-2">
    <div class="relative flex-1">
      <Search size={12} class="absolute start-2 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" />
      <input
        type="text"
        bind:value={search}
        placeholder={$t('arr.searchPlaceholder')}
        class="w-full rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] py-1 pe-2 ps-7 text-xs text-[var(--color-text)] placeholder:text-[var(--color-text-muted)]"
      />
    </div>
    <select
      bind:value={filter}
      class="rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1 text-xs text-[var(--color-text)]"
    >
      <option value="all">{$t('arr.filterAll')}</option>
      <option value="downloaded">{$t('arr.filterDownloaded')}</option>
      <option value="missing">{$t('arr.filterMissing')}</option>
      <option value="monitored">{$t('arr.filterMonitored')}</option>
      <option value="unmonitored">{$t('arr.filterUnmonitored')}</option>
    </select>
    <select
      bind:value={sortBy}
      class="rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1 text-xs text-[var(--color-text)]"
    >
      <option value="title">{$t('arr.sortTitle')}</option>
      <option value="year">{$t('arr.sortYear')}</option>
      <option value="size">{$t('arr.sortSize')}</option>
      <option value="status">{$t('arr.sortStatus')}</option>
    </select>
  </div>

  <!-- List -->
  {#if paginated.length === 0}
    <div class="py-4 text-center text-xs text-[var(--color-text-muted)]">{$t('arr.noMovies')}</div>
  {:else}
    <div class="space-y-1">
      {#each paginated as movie (movie.id)}
        <MovieCard {movie} />
      {/each}
    </div>
  {/if}

  <!-- Pagination -->
  {#if totalPages > 1}
    <div class="flex items-center justify-between text-xs text-[var(--color-text-muted)]">
      <span>{sorted().length} {$t('arr.movies')}</span>
      <div class="flex items-center gap-2">
        <button
          onclick={() => page = Math.max(1, safePage - 1)}
          disabled={safePage <= 1}
          class="rounded p-0.5 hover:bg-[var(--color-bg-tertiary)] disabled:opacity-30"
        >
          <ChevronLeft size={14} />
        </button>
        <span>{safePage} / {totalPages}</span>
        <button
          onclick={() => page = Math.min(totalPages, safePage + 1)}
          disabled={safePage >= totalPages}
          class="rounded p-0.5 hover:bg-[var(--color-bg-tertiary)] disabled:opacity-30"
        >
          <ChevronRight size={14} />
        </button>
      </div>
    </div>
  {/if}
</div>
