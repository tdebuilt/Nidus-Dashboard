<script lang="ts">
  import { Loader2, AlertCircle, Settings, Database, Calendar, Download } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { pollingInterval } from '../../stores/polling'
  import { usePolling } from '../../utils/usePolling'
  import { t } from '../../i18n'
  import QueueCard from './QueueCard.svelte'
  import CalendarCard from './CalendarCard.svelte'

  interface SystemStatus {
    version: string
    appName: string
    instanceName: string
  }

  interface QueueItem {
    id: number
    title: string
    status: string
    trackedDownloadState: string
    size: number
    sizeleft: number
    timeleft: string
  }

  interface CalendarItem {
    id: number
    title: string
    overview: string
    airDateUtc: string
    inCinemas: string
    physicalRelease: string
    hasFile: boolean
    seriesTitle: string
  }

  interface ArrOverview {
    type: string
    label: string
    status?: SystemStatus
    queue_count: number
    queue_items: QueueItem[]
    calendar_items: CalendarItem[]
    library_count: number
    error?: string
  }

  interface Props {
    config?: string
    active?: boolean
  }

  const { config: _config = '{}', active = true }: Props = $props()

  let loading = $state(true)
  let refreshing = $state(false)
  let error = $state<string | null>(null)
  let services = $state.raw<ArrOverview[]>([])
  let activeTab = $state(0)

  async function fetchData() {
    if (services.length === 0) loading = true
    error = null
    try {
      services = await api.get<ArrOverview[]>('/api/arr/overview')
      if (services.length === 0) {
        error = 'no_services'
      }
    } catch {
      error = 'fetch_error'
    } finally {
      loading = false
    }
  }

  async function fetchDataWrapped() {
    const hadData = services.length > 0
    if (hadData) refreshing = true
    try {
      await fetchData()
    } finally {
      refreshing = false
    }
  }

  const polling = usePolling({
    fetchFn: fetchDataWrapped,
    active: () => active,
    pollingStore: pollingInterval,
  })

  $effect(() => {
    if (active) polling.start(); else polling.stop()
    return () => polling.stop()
  })
</script>

<div class="relative h-full" data-testid="arr-widget">
  {#if refreshing}<div class="absolute end-1 top-1 z-10"><Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" /></div>{/if}
  {#if loading && services.length === 0}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('arr.loading')}
    </div>
  {:else if error === 'no_services'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Settings size={24} />
      <p>{$t('arr.noServices')}</p>
      <p class="text-xs">{$t('arr.configureHint')}</p>
    </div>
  {:else if error === 'fetch_error' && services.length === 0}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-danger)]" />
      <p>{$t('arr.fetchError')}</p>
      <button onclick={fetchData} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if services.length > 0}
    <!-- Tabs -->
    {#if services.length > 1}
      <div class="mb-3 flex gap-1 overflow-x-auto border-b border-[var(--color-border)]">
        {#each services as svc, i (svc.label)}
          <button
            class="shrink-0 px-3 py-1.5 text-xs font-medium transition-colors
              {i === activeTab
                ? 'border-b-2 border-[var(--color-primary)] text-[var(--color-primary)]'
                : 'text-[var(--color-text-muted)] hover:text-[var(--color-text)]'}"
            onclick={() => activeTab = i}
          >
            {svc.label}
          </button>
        {/each}
      </div>
    {/if}

    <!-- Active service content -->
    {@const svc = services[activeTab]}
    {#if svc}
      {#if svc.error}
        <div class="flex items-center gap-2 rounded-lg border border-[var(--color-danger)] bg-[var(--color-error-bg)] p-2 text-xs text-[var(--color-danger)]">
          <AlertCircle size={14} />
          {svc.error}
        </div>
      {:else}
        <div class="space-y-3">
          <!-- Header: version + library count -->
          <div class="flex items-center justify-between text-xs text-[var(--color-text-muted)]">
            <div class="flex items-center gap-1.5">
              <Database size={12} />
              <span>{svc.library_count} {$t('arr.' + (svc.type === 'sonarr' ? 'series' : svc.type === 'radarr' ? 'movies' : svc.type === 'lidarr' ? 'artists' : 'indexers'))}</span>
            </div>
            {#if svc.status?.version}
              <span class="rounded bg-[var(--color-bg-tertiary)] px-1.5 py-0.5 text-[10px]">v{svc.status.version}</span>
            {/if}
          </div>

          <!-- Queue -->
          {#if svc.queue_items && svc.queue_items.length > 0}
            <div>
              <div class="mb-1.5 flex items-center gap-1.5 text-xs font-medium text-[var(--color-text-secondary)]">
                <Download size={12} />
                {$t('arr.queue')} ({svc.queue_count})
              </div>
              <div class="space-y-1.5">
                {#each svc.queue_items.slice(0, 5) as item (item.id)}
                  <QueueCard {item} />
                {/each}
              </div>
            </div>
          {:else}
            <div class="text-center text-xs text-[var(--color-text-muted)]">{$t('arr.noQueue')}</div>
          {/if}

          <!-- Calendar -->
          {#if svc.calendar_items && svc.calendar_items.length > 0}
            <div>
              <div class="mb-1.5 flex items-center gap-1.5 text-xs font-medium text-[var(--color-text-secondary)]">
                <Calendar size={12} />
                {$t('arr.upcoming')} ({svc.calendar_items.length})
              </div>
              <div class="space-y-1.5">
                {#each svc.calendar_items.slice(0, 5) as item (item.id)}
                  <CalendarCard {item} />
                {/each}
              </div>
            </div>
          {/if}
        </div>
      {/if}
    {/if}
  {/if}
</div>
