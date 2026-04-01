<script lang="ts">
  import { Search, ChevronLeft, ChevronRight } from 'lucide-svelte'
  import { t } from '../../i18n'
  import SeriesCard from './SeriesCard.svelte'

  interface SonarrStatistics {
    episodeFileCount: number
    episodeCount: number
    percentOfEpisodes: number
    sizeOnDisk: number
  }

  interface SonarrSeries {
    id: number
    title: string
    year: number
    seasonCount: number
    monitored: boolean
    status: string
    seasons?: { seasonNumber: number; monitored: boolean }[]
    statistics: SonarrStatistics
  }

  interface Props {
    series: SonarrSeries[]
  }

  const { series }: Props = $props()

  let search = $state('')
  let filter = $state('all')
  let sortBy = $state('title')
  let page = $state(1)
  const perPage = 20

  const filtered = $derived(() => {
    let list = series
    const q = search.toLowerCase().trim()
    if (q) list = list.filter(s => s.title.toLowerCase().includes(q))
    if (filter === 'complete') list = list.filter(s => s.statistics.percentOfEpisodes >= 100)
    else if (filter === 'incomplete') list = list.filter(s => s.statistics.percentOfEpisodes < 100)
    else if (filter === 'monitored') list = list.filter(s => s.monitored)
    else if (filter === 'unmonitored') list = list.filter(s => !s.monitored)
    return list
  })

  const sorted = $derived(() => {
    const list = [...filtered()]
    if (sortBy === 'title') list.sort((a, b) => a.title.localeCompare(b.title))
    else if (sortBy === 'year') list.sort((a, b) => b.year - a.year)
    else if (sortBy === 'size') list.sort((a, b) => b.statistics.sizeOnDisk - a.statistics.sizeOnDisk)
    else if (sortBy === 'completion') list.sort((a, b) => b.statistics.percentOfEpisodes - a.statistics.percentOfEpisodes)
    return list
  })

  const totalPages = $derived(Math.max(1, Math.ceil(sorted().length / perPage)))
  const safePage = $derived(Math.min(page, totalPages))
  const paginated = $derived(sorted().slice((safePage - 1) * perPage, safePage * perPage))

  const completeCount = $derived(series.filter(s => s.statistics.percentOfEpisodes >= 100).length)
  const incompleteCount = $derived(series.length - completeCount)

  $effect(() => {
    void search
    void filter
    void sortBy
    page = 1
  })
</script>

<div class="space-y-2">
  <!-- Stats + Legend -->
  <div class="flex items-center justify-between text-xs text-[var(--color-text-muted)]">
    <div class="flex gap-3">
      <span class="text-[var(--color-success)]">{completeCount} {$t('arr.filterComplete')}</span>
      <span>{incompleteCount} {$t('arr.filterIncomplete')}</span>
    </div>
    <div class="flex flex-wrap items-center gap-x-2 gap-y-0.5">
      <span class="flex items-center gap-1"><span class="inline-block h-2 w-2 rounded-full" style="background: var(--color-primary)"></span> {$t('arr.legendContinuing')}</span>
      <span class="flex items-center gap-1"><span class="inline-block h-2 w-2 rounded-full" style="background: var(--color-success)"></span> {$t('arr.legendEnded')}</span>
      <span class="flex items-center gap-1"><span class="inline-block h-2 w-2 rounded-full" style="background: var(--color-warning)"></span> {$t('arr.legendMissing')}</span>
      <span class="flex items-center gap-1"><span class="inline-block h-2 w-2 rounded-full" style="background: var(--color-danger)"></span> {$t('arr.legendUnmonitored')}</span>
    </div>
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
      <option value="complete">{$t('arr.filterComplete')}</option>
      <option value="incomplete">{$t('arr.filterIncomplete')}</option>
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
      <option value="completion">{$t('arr.sortCompletion')}</option>
    </select>
  </div>

  <!-- List -->
  {#if paginated.length === 0}
    <div class="py-4 text-center text-xs text-[var(--color-text-muted)]">{$t('arr.noSeries')}</div>
  {:else}
    <div class="space-y-1">
      {#each paginated as s (s.id)}
        <SeriesCard series={s} />
      {/each}
    </div>
  {/if}

  <!-- Pagination -->
  {#if totalPages > 1}
    <div class="flex items-center justify-between text-xs text-[var(--color-text-muted)]">
      <span>{sorted().length} {$t('arr.series')}</span>
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
