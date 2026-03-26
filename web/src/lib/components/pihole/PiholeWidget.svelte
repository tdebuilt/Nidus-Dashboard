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
    unique_domains: number
    cached_queries: number
    forwarded_queries: number
    blocking_enabled: boolean
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
    if (!stats) loading = true
    error = null
    try {
      stats = await api.get<StatsInfo>('/api/pihole/stats')
    } catch (err: unknown) {
      const status = (err as { status?: number })?.status
      if (status === 404) {
        error = 'not_configured'
      } else {
        error = 'fetch_error'
      }
    } finally {
      loading = false
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

  async function toggleBlocking() {
    if (!stats) return

    if (stats.blocking_enabled) {
      const ok = await confirm({
        title: translate('pihole.disableConfirm'),
        message: translate('pihole.disableMessage'),
        confirmLabel: translate('common.disable'),
        destructive: true,
      })
      if (!ok) return
    }

    try {
      await api.post('/api/pihole/blocking', { blocking: !stats.blocking_enabled })
      toasts.success(translate('pihole.toggleSuccess'))
      fetchData()
    } catch {
      toasts.error(translate('pihole.toggleError'))
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

<div class="relative h-full" data-testid="pihole-widget">
  {#if refreshing}<div class="absolute end-1 top-1 z-10"><Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" /></div>{/if}
  {#if loading && !stats}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('pihole.loading')}
    </div>
  {:else if error === 'not_configured'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Settings size={24} />
      <p>{$t('pihole.notConfigured')}</p>
      <p class="text-xs">{$t('pihole.configureHint')}</p>
    </div>
  {:else if error === 'fetch_error' && !stats}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-danger)]" />
      <p>{$t('pihole.fetchError')}</p>
      <button onclick={fetchData} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if stats}
    <div class="space-y-3">
      <!-- Main stat -->
      <div class="text-center">
        <div class="text-3xl font-bold text-[var(--color-text)]">{formatNumber(stats.total_queries)}</div>
        <div class="text-xs text-[var(--color-text-muted)]">{$t('pihole.totalQueries')}</div>
      </div>

      <!-- Blocked stats -->
      <div class="flex items-center justify-between rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2">
        <div>
          <div class="text-sm font-semibold text-[var(--color-danger)]">{formatNumber(stats.blocked_queries)}</div>
          <div class="text-xs text-[var(--color-text-muted)]">{$t('pihole.blockedQueries')}</div>
        </div>
        <div class="text-end">
          <div class="text-sm font-semibold text-[var(--color-text)]">{stats.blocked_percent.toFixed(1)}%</div>
          <div class="text-xs text-[var(--color-text-muted)]">{$t('pihole.blockedPercent')}</div>
        </div>
      </div>

      <!-- Additional stats -->
      <div class="grid grid-cols-3 gap-2 text-center text-xs">
        <div>
          <div class="font-medium text-[var(--color-text)]">{formatNumber(stats.unique_domains)}</div>
          <div class="text-[var(--color-text-muted)]">{$t('pihole.uniqueDomains')}</div>
        </div>
        <div>
          <div class="font-medium text-[var(--color-text)]">{formatNumber(stats.cached_queries)}</div>
          <div class="text-[var(--color-text-muted)]">{$t('pihole.cached')}</div>
        </div>
        <div>
          <div class="font-medium text-[var(--color-text)]">{formatNumber(stats.forwarded_queries)}</div>
          <div class="text-[var(--color-text-muted)]">{$t('pihole.forwarded')}</div>
        </div>
      </div>

      <!-- Blocking toggle -->
      {#if !$isViewer}
        <button
          onclick={toggleBlocking}
          class="flex w-full items-center justify-center gap-2 rounded-lg border px-3 py-2 text-sm font-medium transition-colors"
          class:border-[var(--color-success)]={stats.blocking_enabled}
          class:text-[var(--color-success)]={stats.blocking_enabled}
          class:border-[var(--color-danger)]={!stats.blocking_enabled}
          class:text-[var(--color-danger)]={!stats.blocking_enabled}
        >
          {#if stats.blocking_enabled}
            <Shield size={14} />
            {$t('pihole.blockingEnabled')}
          {:else}
            <ShieldOff size={14} />
            {$t('pihole.blockingDisabled')}
          {/if}
        </button>
      {/if}
    </div>
  {/if}
</div>
