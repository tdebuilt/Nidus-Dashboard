<script lang="ts">
  import { Play, Square, RotateCw, RefreshCw, ExternalLink, ArrowUpCircle } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { confirm } from '../../stores/confirm'
  import { t, translate } from '../../i18n'
  import { isViewer } from '../../stores/auth'

  interface ContainerInfo {
    id: string
    name: string
    image: string
    state: string
    status: string
    health: string
    ports: { IP: string; PrivatePort: number; PublicPort: number; Type: string }[]
    env_id: number
  }

  interface ContainerStatsInfo {
    container_id: string
    cpu_percent: number
    mem_usage: number
    mem_limit: number
    mem_percent: number
  }

  interface Props {
    container: ContainerInfo
    hasUpdate?: boolean
    inStack?: boolean
    envHost?: string
    stats?: ContainerStatsInfo | null
    onAction?: () => void
  }

  const { container, hasUpdate = false, inStack = false, envHost = '', stats = null, onAction }: Props = $props()

  function formatBytes(bytes: number): string {
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
  }

  function statusColor(state: string, health: string): string {
    if (state === 'running') {
      if (health === 'unhealthy') return 'var(--color-danger)'
      if (health === 'starting') return 'var(--color-warning)'
      return 'var(--color-success)'
    }
    if (state === 'exited' || state === 'dead') return 'var(--color-danger)'
    if (state === 'paused') return 'var(--color-warning)'
    return 'var(--color-text-muted)'
  }

  function statusLabel(state: string, health: string): string {
    if (health === 'healthy') return translate('docker.healthHealthy')
    if (health === 'unhealthy') return translate('docker.healthUnhealthy')
    if (health === 'starting') return translate('docker.healthStarting')
    if (state === 'running') return translate('docker.statusRunning')
    if (state === 'exited' || state === 'dead') return translate('docker.statusStopped')
    if (state === 'paused') return translate('docker.statusPaused')
    return state
  }

  function quickLink(ports: ContainerInfo['ports'], envHost: string): string | null {
    const mapped = ports.find(p => p.PublicPort > 0)
    if (!mapped) return null
    let host = mapped.IP
    if (host === '0.0.0.0' || host === '::' || host === '') {
      host = envHost || window.location.hostname
    }
    return `http://${host}:${mapped.PublicPort}`
  }

  async function doAction(action: string) {
    if (action === 'stop') {
      const ok = await confirm({
        title: translate('docker.stopConfirm'),
        message: translate('docker.stopMessage', { name: container.name }),
        confirmLabel: translate('docker.stop'),
        destructive: true,
      })
      if (!ok) return
    }

    try {
      await api.post(`/api/docker/environments/${container.env_id}/containers/${container.id}/${action}`)
      toasts.success(translate('docker.actionSuccess'))
    } catch {
      toasts.error(translate('docker.actionError'))
    } finally {
      onAction?.()
    }
  }

  async function doRecreate() {
    const ok = await confirm({
      title: translate('docker.recreateConfirm'),
      message: translate('docker.recreateWarning', { name: container.name }),
      confirmLabel: translate('docker.recreate'),
      destructive: true,
    })
    if (!ok) return

    try {
      await api.post(`/api/docker/environments/${container.env_id}/containers/${container.id}/recreate`, { pull_image: true })
      toasts.info(translate('docker.recreateStarted', { name: container.name }))
    } catch {
      toasts.error(translate('docker.actionError'))
    } finally {
      setTimeout(() => onAction?.(), 5000)
    }
  }

  const link = $derived(quickLink(container.ports, envHost))
  const color = $derived(statusColor(container.state, container.health))
  const label = $derived(statusLabel(container.state, container.health))
  const isRunning = $derived(container.state === 'running')
  const isStopped = $derived(container.state === 'exited' || container.state === 'dead')
</script>

<div class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2" data-testid="container-card">
  <!-- Row 1: name + status + actions -->
  <div class="flex items-center gap-2">
    <div class="h-2.5 w-2.5 flex-shrink-0 rounded-full" style="background-color: {color}" title={label}></div>

    <span class="min-w-0 flex-1 truncate text-sm font-medium text-[var(--color-text)]">{container.name}</span>

    {#if link}
      <a href={link} target="_blank" rel="noopener" class="touch-action-btn flex-shrink-0 inline-flex items-center justify-center rounded p-1 text-[var(--color-primary)] hover:text-[var(--color-primary-hover)]" title={$t('docker.openLink')}>
        <ExternalLink size={14} />
      </a>
    {/if}

    <span class="hidden flex-shrink-0 whitespace-nowrap text-xs font-medium sm:inline" style="color: {color}">{label}</span>

    <!-- Actions -->
    {#if !$isViewer}
      <div class="flex flex-shrink-0 gap-1">
        {#if isStopped}
          <button onclick={() => doAction('start')} class="touch-action-btn rounded p-2 sm:p-1 text-[var(--color-text-muted)] hover:text-[var(--color-success)]" title={$t('docker.start')}>
            <Play size={14} />
          </button>
        {/if}
        {#if isRunning}
          <button onclick={() => doAction('stop')} class="touch-action-btn rounded p-2 sm:p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]" title={$t('docker.stop')}>
            <Square size={14} />
          </button>
          <button onclick={() => doAction('restart')} class="touch-action-btn rounded p-2 sm:p-1 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]" title={$t('docker.restart')}>
            <RotateCw size={14} />
          </button>
        {/if}
        {#if hasUpdate && isRunning && !inStack}
          <button onclick={doRecreate} class="touch-action-btn rounded p-2 sm:p-1 text-[var(--color-warning)] hover:text-[var(--color-primary)]" title={$t('docker.recreate')}>
            <RefreshCw size={14} />
          </button>
        {/if}
      </div>
    {/if}
  </div>

  <!-- Row 2: image + stats -->
  <div class="mt-1 flex items-center gap-2 ps-4 text-xs text-[var(--color-text-muted)]">
    <span class="truncate">{container.image}</span>
    {#if hasUpdate}
      <span class="flex-shrink-0 text-[var(--color-warning)]" title={$t('docker.updateAvailable')}>
        <ArrowUpCircle size={12} />
      </span>
    {/if}
  </div>
  {#if stats && container.state === 'running'}
    <div class="flex items-center gap-2 ps-4 text-xs text-[var(--color-text-muted)]">
      <span title="CPU">{stats.cpu_percent.toFixed(1)}% CPU</span>
      <span class="text-[var(--color-border)]">·</span>
      <span title="RAM">{formatBytes(stats.mem_usage)} / {formatBytes(stats.mem_limit)}</span>
    </div>
  {/if}
</div>
