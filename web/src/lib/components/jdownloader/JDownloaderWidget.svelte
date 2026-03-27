<script lang="ts">
  import { Loader2, AlertCircle, Settings, Play, Pause, Plus, Trash2 } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { pollingInterval } from '../../stores/polling'
  import { usePolling } from '../../utils/usePolling'
  import { t, translate } from '../../i18n'
  import { isViewer } from '../../stores/auth'
  import PackageCard from './PackageCard.svelte'
  import AddLinksDialog from './AddLinksDialog.svelte'

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

  let loading = $state(true)
  let refreshing = $state(false)
  let error = $state<string | null>(null)
  let queue = $state<QueueInfo | null>(null)
  let showAddDialog = $state(false)

  function formatSpeed(bps: number): string {
    if (bps >= 1048576) return (bps / 1048576).toFixed(1) + ' MB/s'
    if (bps >= 1024) return (bps / 1024).toFixed(0) + ' KB/s'
    return bps + ' B/s'
  }

  async function fetchData() {
    if (queue === null) loading = true
    error = null
    try {
      queue = await api.get<QueueInfo>('/api/jdownloader/queue')
    } catch (err: unknown) {
      const status = (err as { status?: number })?.status
      if (status === 404) {
        error = 'not_configured'
      } else if (status === 502) {
        error = 'fetch_error'
      } else {
        error = 'fetch_error'
        toasts.error(translate('jdownloader.fetchError'))
      }
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

      <!-- Packages -->
      {#if queue.packages.length === 0}
        <div class="flex items-center justify-center py-4 text-sm text-[var(--color-text-muted)]">
          {$t('jdownloader.emptyQueue')}
        </div>
      {:else}
        <div class="space-y-1 overflow-y-auto">
          {#each queue.packages as pkg (pkg.uuid)}
            <PackageCard {pkg} />
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>

<AddLinksDialog open={showAddDialog} onClose={() => showAddDialog = false} onAdded={fetchData} />
