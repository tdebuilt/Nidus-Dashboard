<script lang="ts">
  import { Loader2, AlertCircle, Settings, ArrowUp, ArrowDown } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { pollingInterval } from '../../stores/polling'
  import { breakpoint } from '../../stores/breakpoint'
  import { usePolling } from '../../utils/usePolling'
  import { getResponsiveColumns } from '../../utils/responsiveColumns'
  import { t } from '../../i18n'
  import MonitorCard from './MonitorCard.svelte'

  interface MonitorInfo {
    id: number
    name: string
    type: string
    status: number
    uptime_24h: number
    latency: number
    message: string
  }

  interface MonitorsOverview {
    monitors: MonitorInfo[]
    total_up: number
    total_down: number
    total_count: number
    status_page: string
  }

  interface Props {
    config?: string
    active?: boolean
  }

  const { config = '{}', active = true }: Props = $props()

  let loading = $state(true)
  let refreshing = $state(false)
  let error = $state<string | null>(null)
  let data = $state<MonitorsOverview | null>(null)

  const parsedConfig = $derived(() => {
    try { return JSON.parse(config) } catch { return {} }
  })

  const slug = $derived(parsedConfig().slug || 'default')
  const cols = $derived(getResponsiveColumns(parsedConfig(), $breakpoint, 1))

  async function fetchData() {
    const hadData = data !== null
    if (!hadData) loading = true
    error = null
    try {
      data = await api.get<MonitorsOverview>(`/api/uptimekuma/monitors/${slug}`)
    } catch (err: unknown) {
      const status = (err as { status?: number })?.status
      if (status === 404) {
        error = 'not_configured'
      } else if (status === 504) {
        error = 'timeout'
        polling.stop()
      } else {
        error = 'fetch_error'
        polling.stop()
      }
    } finally {
      loading = false
    }
  }

  async function fetchDataWrapped() {
    const hadData = data !== null
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

  function retry() {
    error = null
    polling.start()
  }

  $effect(() => {
    if (active) polling.start(); else polling.stop()
    return () => polling.stop()
  })
</script>

<div class="relative h-full" data-testid="uptimekuma-widget">
  {#if refreshing}<div class="absolute end-1 top-1 z-10"><Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" /></div>{/if}
  {#if loading && !data}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('uptimekuma.loading')}
    </div>
  {:else if error === 'not_configured'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Settings size={24} />
      <p>{$t('uptimekuma.notConfigured')}</p>
      <p class="text-xs">{$t('uptimekuma.configureHint')}</p>
    </div>
  {:else if error === 'timeout' && !data}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-warning)]" />
      <p>{$t('uptimekuma.timeoutError')}</p>
      <p class="text-xs">{$t('uptimekuma.timeoutHint')}</p>
      <button onclick={retry} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if error === 'fetch_error' && !data}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-danger)]" />
      <p>{$t('uptimekuma.fetchError')}</p>
      <button onclick={retry} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if data}
    <div class="space-y-3">
      <!-- Summary bar -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="flex items-center gap-1 text-sm font-semibold text-[var(--color-success)]">
            <ArrowUp size={14} />
            {data.total_up}
          </div>
          {#if data.total_down > 0}
            <div class="flex items-center gap-1 text-sm font-semibold text-[var(--color-danger)]">
              <ArrowDown size={14} />
              {data.total_down}
            </div>
          {/if}
        </div>
        <div class="text-xs text-[var(--color-text-muted)]">
          {data.total_count} {$t('uptimekuma.monitors')}
        </div>
      </div>

      <!-- Monitor list -->
      <div class="grid gap-1.5" style="grid-template-columns: repeat({cols}, 1fr)">
        {#each data.monitors as monitor (monitor.id)}
          <MonitorCard {monitor} />
        {/each}
      </div>

      {#if data.monitors.length === 0}
        <div class="py-4 text-center text-xs text-[var(--color-text-muted)]">
          {$t('uptimekuma.noMonitors')}
        </div>
      {/if}
    </div>
  {/if}
</div>
