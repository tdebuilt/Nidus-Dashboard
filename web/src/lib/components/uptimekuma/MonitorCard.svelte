<script lang="ts">
  import { ArrowUp, ArrowDown, Clock, Minus } from 'lucide-svelte'
  import { t } from '../../i18n'

  interface Monitor {
    id: number
    name: string
    type: string
    status: number
    uptime_24h: number
    latency: number
    message: string
  }

  interface Props {
    monitor: Monitor
  }

  const { monitor }: Props = $props()

  const statusColor = $derived(
    monitor.status === 1 ? 'var(--color-success)' :
    monitor.status === 0 ? 'var(--color-danger)' :
    monitor.status === 3 ? 'var(--color-warning)' :
    'var(--color-text-muted)'
  )

  const statusLabel = $derived(
    monitor.status === 1 ? $t('uptimekuma.statusUp') :
    monitor.status === 0 ? $t('uptimekuma.statusDown') :
    monitor.status === 3 ? $t('uptimekuma.statusMaintenance') :
    $t('uptimekuma.statusPending')
  )
</script>

<div
  class="flex items-center gap-3 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2"
  data-testid="monitor-card-{monitor.id}"
>
  <!-- Status dot -->
  <div
    class="h-2.5 w-2.5 flex-shrink-0 rounded-full"
    style="background-color: {statusColor}"
    title={statusLabel}
  ></div>

  <!-- Name -->
  <div class="min-w-0 flex-1">
    <div class="truncate text-sm font-medium text-[var(--color-text)]">{monitor.name}</div>
  </div>

  <!-- Latency -->
  {#if monitor.latency > 0}
    <div class="flex items-center gap-1 text-xs text-[var(--color-text-muted)]" title={$t('uptimekuma.latency')}>
      <Clock size={12} />
      <span>{monitor.latency.toFixed(0)}ms</span>
    </div>
  {/if}

  <!-- Uptime -->
  <div
    class="text-xs font-semibold"
    style="color: {monitor.uptime_24h >= 0.99 ? 'var(--color-success)' : monitor.uptime_24h >= 0.95 ? 'var(--color-warning)' : 'var(--color-danger)'}"
    title={$t('uptimekuma.uptime24h')}
  >
    {(monitor.uptime_24h * 100).toFixed(1)}%
  </div>

  <!-- Status icon -->
  {#if monitor.status === 1}
    <ArrowUp size={14} style="color: var(--color-success)" />
  {:else if monitor.status === 0}
    <ArrowDown size={14} style="color: var(--color-danger)" />
  {:else}
    <Minus size={14} style="color: var(--color-text-muted)" />
  {/if}
</div>
