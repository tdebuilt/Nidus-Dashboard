<script lang="ts">
  import { Loader2, AlertCircle, Settings, KeyRound, Play, Square, Plus, ArrowDown, ArrowUp, Trash2, Search } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { pollingInterval } from '../../stores/polling'
  import { usePolling } from '../../utils/usePolling'
  import { t, translate } from '../../i18n'
  import { isViewer } from '../../stores/auth'
  import { loadPref, savePref } from '../../utils/widgetPrefs'
  import TorrentCard from './TorrentCard.svelte'
  import SortHeader from '../shared/SortHeader.svelte'
  import Pagination from '../shared/Pagination.svelte'
  import AddTorrentDialog from './AddTorrentDialog.svelte'

  type SortField = 'name' | 'progress' | 'size' | 'speed_down' | 'speed_up' | 'eta' | 'ratio' | 'status'

  interface TorrentInfo {
    id: number
    name: string
    status: string
    progress: number
    size: number
    downloaded: number
    speed_down: number
    speed_up: number
    eta: number
    ratio: number
    peers: number
    error?: string
  }

  interface TorrentsInfo {
    torrents: TorrentInfo[]
    download_speed: number
    upload_speed: number
    total_count: number
    active_count: number
  }

  interface Props {
    config?: string
    active?: boolean
  }

  const { config: _config = '{}', active = true }: Props = $props()

  const PREFIX = 'nidus-tx-'
  const SORT_FIELDS: SortField[] = ['name', 'progress', 'size', 'speed_down', 'speed_up', 'eta', 'ratio', 'status']
  const FILTERS = ['all', 'downloading', 'seeding', 'stopped', 'checking', 'completed']
  const PAGE_SIZES = [10, 20, 50, 100]

  let loading = $state(true)
  let refreshing = $state(false)
  let error = $state<string | null>(null)
  let data = $state<TorrentsInfo | null>(null)
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
    if (bps > 0) return bps + ' B/s'
    return '0'
  }

  const searchedTorrents = $derived(
    data?.torrents.filter((t) => {
      if (!search) return true
      return t.name.toLowerCase().includes(search.toLowerCase())
    }) ?? []
  )

  const filteredTorrents = $derived(
    searchedTorrents.filter((t) => {
      if (filter === 'all') return true
      if (filter === 'completed') return t.progress >= 100
      return t.status === filter
    })
  )

  const sortedTorrents = $derived(
    [...filteredTorrents].sort((a, b) => {
      const va = a[sortField as keyof TorrentInfo]
      const vb = b[sortField as keyof TorrentInfo]
      const cmp = typeof va === 'string' && typeof vb === 'string'
        ? va.localeCompare(vb)
        : (va as number) - (vb as number)
      return sortDirection === 'asc' ? cmp : -cmp
    })
  )

  const totalPages = $derived(Math.max(1, Math.ceil(sortedTorrents.length / pageSize)))
  const safePage = $derived(Math.min(page, totalPages))
  const paginatedTorrents = $derived(
    sortedTorrents.slice((safePage - 1) * pageSize, safePage * pageSize)
  )

  $effect(() => {
    void search
    void filter
    void sortField
    page = 1
  })

  async function fetchData() {
    if (data === null) loading = true
    error = null
    try {
      data = await api.get<TorrentsInfo>('/api/transmission/torrents')
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
    if (data !== null) refreshing = true
    try {
      await fetchData()
    } finally {
      refreshing = false
    }
  }

  async function startAll() {
    try {
      await api.post('/api/transmission/torrents/start-all')
      toasts.success(translate('transmission.allStarted'))
      fetchData()
    } catch {
      toasts.error(translate('transmission.actionError'))
    }
  }

  async function stopAll() {
    try {
      await api.post('/api/transmission/torrents/stop-all')
      toasts.success(translate('transmission.allStopped'))
      fetchData()
    } catch {
      toasts.error(translate('transmission.actionError'))
    }
  }

  async function cleanupCompleted() {
    try {
      const result = await api.post<{ removed: number }>('/api/transmission/torrents/cleanup')
      toasts.success(translate('transmission.cleanupDone', { count: String(result.removed) }))
      fetchData()
    } catch {
      toasts.error(translate('transmission.actionError'))
    }
  }

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

<div class="relative h-full" data-testid="transmission-widget">
  {#if refreshing}
    <div class="absolute top-0 end-0 z-10 p-1">
      <Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" />
    </div>
  {/if}
  {#if loading && !data}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('transmission.loading')}
    </div>
  {:else if error === 'not_configured'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Settings size={24} />
      <p>{$t('transmission.notConfigured')}</p>
      <p class="text-xs">{$t('transmission.configureHint')}</p>
    </div>
  {:else if error === 'auth_error'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <KeyRound size={24} class="text-[var(--color-danger)]" />
      <p>{$t('transmission.authError')}</p>
      <p class="text-xs">{$t('transmission.authErrorHint')}</p>
      <button onclick={() => { polling.start() }} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if error === 'fetch_error' && !data}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-danger)]" />
      <p>{$t('transmission.fetchError')}</p>
      <button onclick={fetchData} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if data}
    <div class="space-y-2">
      <!-- Header with global speeds -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3 text-xs text-[var(--color-text-muted)]">
          <span class="flex items-center gap-1">
            <ArrowDown size={12} class="text-[var(--color-success)]" />
            {formatSpeed(data.download_speed)}
          </span>
          <span class="flex items-center gap-1">
            <ArrowUp size={12} class="text-[var(--color-primary)]" />
            {formatSpeed(data.upload_speed)}
          </span>
          <span>{data.active_count}/{data.total_count}</span>
        </div>
        {#if !$isViewer}
          <div class="flex gap-1">
            <button onclick={startAll} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-success)]" title={$t('transmission.startAll')} aria-label={$t('transmission.startAll')}>
              <Play size={14} />
            </button>
            <button onclick={stopAll} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]" title={$t('transmission.stopAll')} aria-label={$t('transmission.stopAll')}>
              <Square size={14} />
            </button>
            <button onclick={cleanupCompleted} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-warning)]" title={$t('transmission.cleanup')} aria-label={$t('transmission.cleanup')}>
              <Trash2 size={14} />
            </button>
            <button onclick={() => showAddDialog = true} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]" title={$t('transmission.addTorrent')} aria-label={$t('transmission.addTorrent')}>
              <Plus size={14} />
            </button>
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
            placeholder={$t('transmission.searchPlaceholder')}
            aria-label={$t('transmission.searchPlaceholder')}
            class="w-full rounded border border-[var(--color-border)] bg-[var(--color-bg)] py-1 pl-6 pr-2 text-xs text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)] focus:outline-none"
          />
        </div>
        <select
          value={filter}
          onchange={(e) => { filter = (e.target as HTMLSelectElement).value; savePref(PREFIX, 'filter', filter) }}
          class="rounded border border-[var(--color-border)] bg-[var(--color-bg)] px-1.5 py-1 text-xs text-[var(--color-text)]"
          aria-label="Filter"
        >
          <option value="all">{$t('transmission.filterAll')}</option>
          <option value="downloading">{$t('transmission.filterDownloading')}</option>
          <option value="seeding">{$t('transmission.filterSeeding')}</option>
          <option value="stopped">{$t('transmission.filterStopped')}</option>
          <option value="checking">{$t('transmission.filterChecking')}</option>
          <option value="completed">{$t('transmission.filterCompleted')}</option>
        </select>
      </div>

      <!-- Sort header -->
      <SortHeader
        columns={[
          { field: 'name', labelKey: 'transmission.sortName', class: 'flex-1 text-left' },
          { field: 'progress', labelKey: 'transmission.sortProgress' },
          { field: 'size', labelKey: 'transmission.sortSize' },
          { field: 'speed_down', labelKey: 'transmission.sortSpeedDown' },
          { field: 'speed_up', labelKey: 'transmission.sortSpeedUp' },
          { field: 'ratio', labelKey: 'transmission.sortRatio' },
        ]}
        {sortField} {sortDirection} onToggleSort={toggleSort}
      />

      <!-- Torrents -->
      {#if sortedTorrents.length === 0}
        <div class="flex items-center justify-center py-4 text-sm text-[var(--color-text-muted)]">
          {$t('transmission.noTorrents')}
        </div>
      {:else}
        <div class="space-y-1 overflow-y-auto">
          {#each paginatedTorrents as torrent (torrent.id)}
            <TorrentCard {torrent} onAction={fetchData} />
          {/each}
        </div>
      {/if}

      <!-- Pagination -->
      <Pagination
        page={safePage}
        {totalPages}
        totalItems={sortedTorrents.length}
        {pageSize}
        onPageChange={(p) => page = p}
        onPageSizeChange={(s) => { pageSize = s; page = 1; savePref(PREFIX, 'pageSize', String(s)) }}
      />
    </div>
  {/if}
</div>

<AddTorrentDialog open={showAddDialog} onClose={() => showAddDialog = false} onAdded={fetchData} />
