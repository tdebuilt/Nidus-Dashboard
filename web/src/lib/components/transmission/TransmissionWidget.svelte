<script lang="ts">
  import { Loader2, AlertCircle, Settings, Play, Square, Plus, ArrowDown, ArrowUp, Trash2 } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { pollingInterval } from '../../stores/polling'
  import { usePolling } from '../../utils/usePolling'
  import { t, translate } from '../../i18n'
  import { isViewer } from '../../stores/auth'
  import TorrentCard from './TorrentCard.svelte'
  import AddTorrentDialog from './AddTorrentDialog.svelte'

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

  let loading = $state(true)
  let refreshing = $state(false)
  let error = $state<string | null>(null)
  let data = $state<TorrentsInfo | null>(null)
  let showAddDialog = $state(false)

  function formatSpeed(bps: number): string {
    if (bps >= 1048576) return (bps / 1048576).toFixed(1) + ' MB/s'
    if (bps >= 1024) return (bps / 1024).toFixed(0) + ' KB/s'
    if (bps > 0) return bps + ' B/s'
    return '0'
  }

  async function fetchData() {
    if (data === null) loading = true
    error = null
    try {
      data = await api.get<TorrentsInfo>('/api/transmission/torrents')
    } catch (err: unknown) {
      const status = (err as { status?: number })?.status
      if (status === 404) {
        error = 'not_configured'
      } else if (status === 502) {
        error = 'fetch_error'
      } else {
        error = 'fetch_error'
        toasts.error(translate('transmission.fetchError'))
      }
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
            <button onclick={startAll} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-success)]" title={$t('transmission.startAll')}>
              <Play size={14} />
            </button>
            <button onclick={stopAll} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]" title={$t('transmission.stopAll')}>
              <Square size={14} />
            </button>
            <button onclick={cleanupCompleted} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-warning)]" title={$t('transmission.cleanup')}>
              <Trash2 size={14} />
            </button>
            <button onclick={() => showAddDialog = true} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]" title={$t('transmission.addTorrent')}>
              <Plus size={14} />
            </button>
          </div>
        {/if}
      </div>

      <!-- Torrents -->
      {#if data.torrents.length === 0}
        <div class="flex items-center justify-center py-4 text-sm text-[var(--color-text-muted)]">
          {$t('transmission.noTorrents')}
        </div>
      {:else}
        <div class="space-y-1 overflow-y-auto">
          {#each data.torrents as torrent (torrent.id)}
            <TorrentCard {torrent} onAction={fetchData} />
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>

<AddTorrentDialog open={showAddDialog} onClose={() => showAddDialog = false} onAdded={fetchData} />
