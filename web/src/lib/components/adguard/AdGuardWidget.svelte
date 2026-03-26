<script lang="ts">
  import { Loader2, AlertCircle, Settings, Shield, ShieldOff } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { pollingInterval } from '../../stores/polling'
  import { usePolling } from '../../utils/usePolling'
  import { confirm } from '../../stores/confirm'
  import { t, translate } from '../../i18n'
  import { isViewer } from '../../stores/auth'

  interface StatsInfo {
    total_queries: number
    blocked_queries: number
    blocked_percent: number
    avg_response_time: number
    filtering_enabled: boolean
    active_filters: number
    total_rules: number
  }

  interface Props {
    config?: string
    active?: boolean
  }

  const { config: _config = '{}', active = true }: Props = $props()

  let loading = $state(true)
  let refreshing = $state(false)
  let error = $state<string | null>(null)
  let stats = $state<StatsInfo | null>(null)

  function formatNumber(n: number): string {
    if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M'
    if (n >= 1000) return (n / 1000).toFixed(1) + 'k'
    return n.toString()
  }

  async function fetchData() {
    const hadData = stats !== null
    if (!hadData) loading = true
    error = null
    try {
      stats = await api.get<StatsInfo>('/api/adguard/stats')
    } catch (err: unknown) {
      const e = err as { status?: number }
      if (e?.status === 404) {
        error = 'not_configured'
      } else if (e?.status === 502) {
        error = 'fetch_error'
      } else {
        error = 'fetch_error'
        toasts.error(translate('adguard.fetchError'))
      }
    } finally {
      loading = false
    }
  }

  async function toggleFiltering() {
    if (!stats) return

    if (stats.filtering_enabled) {
      const ok = await confirm({
        title: translate('adguard.disableConfirm'),
        message: translate('adguard.disableMessage'),
        confirmLabel: translate('common.disable'),
        destructive: true,
      })
      if (!ok) return
    }

    try {
      await api.post('/api/adguard/filtering/toggle', { enabled: !stats.filtering_enabled })
      toasts.success(translate('adguard.toggleSuccess'))
      fetchData()
    } catch {
      toasts.error(translate('adguard.toggleError'))
    }
  }

  async function fetchDataWrapped() {
    const hadData = stats !== null
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

<div class="relative h-full" data-testid="adguard-widget">
  {#if refreshing}<div class="absolute end-1 top-1 z-10"><Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" /></div>{/if}
  {#if loading && !stats}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('adguard.loading')}
    </div>
  {:else if error === 'not_configured'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Settings size={24} />
      <p>{$t('adguard.notConfigured')}</p>
      <p class="text-xs">{$t('adguard.configureHint')}</p>
    </div>
  {:else if error === 'fetch_error' && !stats}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-danger)]" />
      <p>{$t('adguard.fetchError')}</p>
      <button onclick={fetchData} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if stats}
    <div class="space-y-3">
      <!-- Main stat -->
      <div class="text-center">
        <div class="text-3xl font-bold text-[var(--color-text)]">{formatNumber(stats.total_queries)}</div>
        <div class="text-xs text-[var(--color-text-muted)]">{$t('adguard.totalQueries')}</div>
      </div>

      <!-- Blocked stats -->
      <div class="flex items-center justify-between rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2">
        <div>
          <div class="text-sm font-semibold text-[var(--color-danger)]">{formatNumber(stats.blocked_queries)}</div>
          <div class="text-xs text-[var(--color-text-muted)]">{$t('adguard.blocked')}</div>
        </div>
        <div class="text-end">
          <div class="text-sm font-semibold text-[var(--color-text)]">{stats.blocked_percent.toFixed(1)}%</div>
          <div class="text-xs text-[var(--color-text-muted)]">{$t('adguard.blockedPercent')}</div>
        </div>
      </div>

      <!-- Response time + filters -->
      <div class="flex flex-wrap items-center justify-between gap-y-1 text-xs text-[var(--color-text-muted)]">
        <span>{$t('adguard.avgResponse')}: {(stats.avg_response_time * 1000).toFixed(1)}ms</span>
        <span>{stats.active_filters} {$t('adguard.filters')} ({formatNumber(stats.total_rules)} {$t('adguard.rules')})</span>
      </div>

      <!-- Filtering toggle -->
      {#if !$isViewer}
        <button
          onclick={toggleFiltering}
          class="flex w-full items-center justify-center gap-2 rounded-lg border px-3 py-2 text-sm font-medium transition-colors"
          class:border-[var(--color-success)]={stats.filtering_enabled}
          class:text-[var(--color-success)]={stats.filtering_enabled}
          class:border-[var(--color-danger)]={!stats.filtering_enabled}
          class:text-[var(--color-danger)]={!stats.filtering_enabled}
        >
          {#if stats.filtering_enabled}
            <Shield size={14} />
            {$t('adguard.filteringOn')}
          {:else}
            <ShieldOff size={14} />
            {$t('adguard.filteringOff')}
          {/if}
        </button>
      {/if}
    </div>
  {/if}
</div>
