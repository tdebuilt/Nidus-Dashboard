<script lang="ts">
  import { Loader2, AlertCircle, Settings, KeyRound, Play, Pause, Plus, Trash2, Search } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { pollingInterval } from '../../stores/polling'
  import { usePolling } from '../../utils/usePolling'
  import { t, translate } from '../../i18n'
  import { isViewer } from '../../stores/auth'
  import { loadPref, savePref } from '../../utils/widgetPrefs'
  import PackageCard from './PackageCard.svelte'
  import SortHeader from '../shared/SortHeader.svelte'
  import Pagination from '../shared/Pagination.svelte'
  import AddLinksDialog from './AddLinksDialog.svelte'

  type SortField = 'name' | 'progress' | 'size' | 'speed' | 'eta' | 'status'

  interface PackageInfo {
    uuid: number
    name: string
    status: string
    progress: number
    size: number
    downloaded: number
    speed: number
    eta: number
    finished: boolean
    link_count: number
  }

  interface QueueInfo {
    packages: PackageInfo[]
    total_speed: number
    running: boolean
  }

  interface Props {
    config?: string
    active?: boolean
  }

  const { config: _config = '{}', active = true }: Props = $props()

  const PREFIX = 'nidus-jd-'
  const SORT_FIELDS: SortField[] = ['name', 'progress', 'size', 'speed', 'eta', 'status']
  const FILTERS = ['all', 'downloading', 'finished', 'queued']
  const PAGE_SIZES = [10, 20, 50, 100]

  let loading = $state(true)
  let refreshing = $state(false)
  let error = $state<string | null>(null)
  let queue = $state<QueueInfo | null>(null)
  let showAddDialog = $state(false)

  let search = $state('')
  let filter = $state(loadPref(PREFIX, 'filter', 'all', FILTERS))
  let sortField = $state<SortField>(loadPref<SortField>(PREFIX, 'sortField', 'name', SORT_FIELDS))
  let sortDirection = $state<'asc' | 'desc'>(loadPref<'asc' | 'desc'>(PREFIX, 'sortDir', 'asc', ['asc', 'desc']))
  let page = $state(1)
  let pageSize = $state(Number(loadPref(PREFIX, 'pageSize', '20', PAGE_SIZES.map(String))) || 20)

  function formatSpeed(bps: number): string {
    if (bps >= 1048576) return (bps / 1048576).toFixed(1) + ' MB/s'
    if (bps >= 1024) return (bps / 1024).toFixed(0) + ' KB/s'
    return bps + ' B/s'
  }

  const searchedPackages = $derived(
    queue?.packages.filter((p) => {
      if (!search) return true
      return p.name.toLowerCase().includes(search.toLowerCase())
    }) ?? []
  )

  const filteredPackages = $derived(
    searchedPackages.filter((p) => {
      if (filter === 'all') return true
      if (filter === 'finished') return p.finished
      return p.status === filter
    })
  )

  const sortedPackages = $derived(
    [...filteredPackages].sort((a, b) => {
      const va = a[sortField as keyof PackageInfo]
      const vb = b[sortField as keyof PackageInfo]
      const cmp = typeof va === 'string' && typeof vb === 'string'
        ? va.localeCompare(vb)
        : Number(va) - Number(vb)
      return sortDirection === 'asc' ? cmp : -cmp
    })
  )

  const totalPages = $derived(Math.max(1, Math.ceil(sortedPackages.length / pageSize)))
  const safePage = $derived(Math.min(page, totalPages))
  const paginatedPackages = $derived(
    sortedPackages.slice((safePage - 1) * pageSize, safePage * pageSize)
  )

  $effect(() => {
    void search
    void filter
    void sortField
    page = 1
  })

  async function fetchData() {
    if (queue === null) loading = true
    error = null
    try {
      queue = await api.get<QueueInfo>('/api/jdownloader/queue')
    } catch (err: unknown) {
      const { status, message } = err as { status?: number; message?: string }
      if (status === 404) error = 'not_configured'
      else if (message === 'authentication_failed') { error = 'auth_error'; polling.stop() }
      else error = 'fetch_error'
    } finally {
      loading = false
    }
  }

  async function fetchDataWrapped() {
    if (queue !== null) refreshing = true
    try {
      await fetchData()
    } finally {
      refreshing = false
    }
  }

  async function startQueue() {
    try {
      await api.post('/api/jdownloader/queue/start')
      toasts.success(translate('jdownloader.started'))
      fetchData()
    } catch {
      toasts.error(translate('jdownloader.actionError'))
    }
  }

  async function pauseQueue() {
    try {
      await api.post('/api/jdownloader/queue/pause')
      toasts.success(translate('jdownloader.paused'))
      fetchData()
    } catch {
      toasts.error(translate('jdownloader.actionError'))
    }
  }

  async function cleanupFinished() {
    try {
      const result = await api.post<{ removed: number }>('/api/jdownloader/queue/cleanup')
      if (result.removed > 0) {
        toasts.success(translate('jdownloader.cleanupSuccess', { count: result.removed }))
      } else {
        toasts.info(translate('jdownloader.cleanupNone'))
      }
      fetchData()
    } catch {
      toasts.error(translate('jdownloader.actionError'))
    }
  }

  const hasFinished = $derived(queue?.packages.some(p => p.finished) ?? false)

  function toggleSort(field: string) {
    const f = field as SortField
    if (sortField === f) {
      sortDirection = sortDirection === 'asc' ? 'desc' : 'asc'
    } else {
      sortField = f
      sortDirection = 'asc'
    }
    savePref(PREFIX, 'sortField', sortField)
    savePref(PREFIX, 'sortDir', sortDirection)
  }

  const polling = usePolling({
    fetchFn: fetchDataWrapped,
    active: () => active,
    pollingStore: pollingInterval,
    intervalTransform: (ms) => Math.min(ms, 10000),
  })

  $effect(() => {
    if (active) polling.start()
    else polling.stop()
    return () => polling.stop()
  })
</script>

<div class="relative h-full" data-testid="jdownloader-widget">
  {#if refreshing}
    <div class="absolute top-0 end-0 z-10 p-1">
      <Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" />
    </div>
  {/if}
  {#if loading && !queue}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('jdownloader.loading')}
    </div>
  {:else if error === 'not_configured'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Settings size={24} />
      <p>{$t('jdownloader.notConfigured')}</p>
      <p class="text-xs">{$t('jdownloader.configureHint')}</p>
    </div>
  {:else if error === 'auth_error'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <KeyRound size={24} class="text-[var(--color-danger)]" />
      <p>{$t('jdownloader.authError')}</p>
      <p class="text-xs">{$t('jdownloader.authErrorHint')}</p>
      <button onclick={() => { polling.start() }} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if error === 'fetch_error' && !queue}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-danger)]" />
      <p>{$t('jdownloader.fetchError')}</p>
      <button onclick={fetchData} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if queue}
    <div class="space-y-2">
      <!-- Header with speed and controls -->
      <div class="flex items-center justify-between">
        <div class="text-sm font-medium text-[var(--color-text)]">
          {#if queue.total_speed > 0}
            {formatSpeed(queue.total_speed)}
          {:else}
            {$t('jdownloader.idle')}
          {/if}
        </div>
        {#if !$isViewer}
          <div class="flex gap-1">
            {#if queue.running}
              <button onclick={pauseQueue} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-warning)]" title={$t('jdownloader.pause')} aria-label={$t('jdownloader.pause')}>
                <Pause size={14} />
              </button>
            {:else}
              <button onclick={startQueue} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-success)]" title={$t('jdownloader.start')} aria-label={$t('jdownloader.start')}>
                <Play size={14} />
              </button>
            {/if}
            <button onclick={() => showAddDialog = true} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]" title={$t('jdownloader.addLinks')} aria-label={$t('jdownloader.addLinks')}>
              <Plus size={14} />
            </button>
            {#if hasFinished}
              <button onclick={cleanupFinished} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]" title={$t('jdownloader.cleanup')} aria-label={$t('jdownloader.cleanup')}>
                <Trash2 size={14} />
              </button>
            {/if}
          </div>
        {/if}
      </div>

      <!-- Search + filter -->
      <div class="flex gap-2">
        <div class="relative flex-1">
          <Search size={12} class="absolute left-2 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" />
          <input
            type="text"
            bind:value={search}
            placeholder={$t('jdownloader.searchPlaceholder')}
            aria-label={$t('jdownloader.searchPlaceholder')}
            class="w-full rounded border border-[var(--color-border)] bg-[var(--color-bg)] py-1 pl-6 pr-2 text-xs text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)] focus:outline-none"
          />
        </div>
        <select
          value={filter}
          onchange={(e) => { filter = (e.target as HTMLSelectElement).value; savePref(PREFIX, 'filter', filter) }}
          class="rounded border border-[var(--color-border)] bg-[var(--color-bg)] px-1.5 py-1 text-xs text-[var(--color-text)]"
          aria-label="Filter"
        >
          <option value="all">{$t('jdownloader.filterAll')}</option>
          <option value="downloading">{$t('jdownloader.filterDownloading')}</option>
          <option value="finished">{$t('jdownloader.filterFinished')}</option>
          <option value="queued">{$t('jdownloader.filterQueued')}</option>
        </select>
      </div>

      <!-- Sort header -->
      <SortHeader
        columns={[
          { field: 'name', labelKey: 'jdownloader.sortName', class: 'flex-1 text-left' },
          { field: 'progress', labelKey: 'jdownloader.sortProgress' },
          { field: 'size', labelKey: 'jdownloader.sortSize' },
          { field: 'speed', labelKey: 'jdownloader.sortSpeed' },
          { field: 'eta', labelKey: 'jdownloader.sortETA' },
        ]}
        {sortField} {sortDirection} onToggleSort={toggleSort}
      />

      <!-- Packages -->
      {#if sortedPackages.length === 0}
        <div class="flex items-center justify-center py-4 text-sm text-[var(--color-text-muted)]">
          {$t('jdownloader.emptyQueue')}
        </div>
      {:else}
        <div class="space-y-1 overflow-y-auto">
          {#each paginatedPackages as pkg (pkg.uuid)}
            <PackageCard {pkg} />
          {/each}
        </div>
      {/if}

      <!-- Pagination -->
      <Pagination
        page={safePage}
        {totalPages}
        totalItems={sortedPackages.length}
        {pageSize}
        onPageChange={(p) => page = p}
        onPageSizeChange={(s) => { pageSize = s; page = 1; savePref(PREFIX, 'pageSize', String(s)) }}
      />
    </div>
  {/if}
</div>

<AddLinksDialog open={showAddDialog} onClose={() => showAddDialog = false} onAdded={fetchData} />
