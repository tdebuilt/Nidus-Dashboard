<script lang="ts">
  import { Play, Square, Power, RotateCw, Monitor, Box, Loader2 } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { confirm } from '../../stores/confirm'
  import { t, translate } from '../../i18n'
  import { isViewer } from '../../stores/auth'

  interface VMInfo {
    vmid: number
    name: string
    node: string
    type: string
    status: string
    cpu_usage: number
    cpu_cores: number
    mem_used: number
    mem_total: number
    uptime: number
  }

  interface Props {
    vm: VMInfo
    onAction?: () => void
  }

  const { vm, onAction }: Props = $props()

  let actionInProgress = $state<string | null>(null)

  function statusColor(status: string): string {
    if (status === 'running') return 'var(--color-success)'
    if (status === 'stopped') return 'var(--color-danger)'
    if (status === 'paused') return 'var(--color-warning)'
    return 'var(--color-text-muted)'
  }

  function statusLabel(status: string): string {
    if (status === 'running') return translate('proxmox.statusRunning')
    if (status === 'stopped') return translate('proxmox.statusStopped')
    if (status === 'paused') return translate('proxmox.statusPaused')
    return status
  }

  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B'
    const units = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(1024))
    return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
  }

  function formatUptime(seconds: number): string {
    if (seconds === 0) return '-'
    const d = Math.floor(seconds / 86400)
    const h = Math.floor((seconds % 86400) / 3600)
    const m = Math.floor((seconds % 3600) / 60)
    if (d > 0) return `${d}d ${h}h`
    if (h > 0) return `${h}h ${m}m`
    return `${m}m`
  }

  function formatCPU(usage: number): string {
    return (usage * 100).toFixed(0) + '%'
  }

  async function doAction(action: string) {
    if (action === 'stop') {
      const ok = await confirm({
        title: translate('proxmox.stopConfirm'),
        message: translate('proxmox.stopMessage', { name: vm.name }),
        confirmLabel: translate('proxmox.stop'),
        destructive: true,
      })
      if (!ok) return
    }

    if (action === 'shutdown') {
      const ok = await confirm({
        title: translate('proxmox.shutdownConfirm'),
        message: translate('proxmox.shutdownMessage', { name: vm.name }),
        confirmLabel: translate('proxmox.shutdown'),
        destructive: true,
      })
      if (!ok) return
    }

    actionInProgress = action
    try {
      await api.post(`/api/proxmox/vms/${vm.node}/${vm.type}/${vm.vmid}/${action}`)
      toasts.success(translate('proxmox.actionSuccess'))
      // Poll until state changes (Proxmox actions are async)
      let retries = 0
      const poll = setInterval(async () => {
        retries++
        onAction?.()
        if (retries >= 10) {
          clearInterval(poll)
          actionInProgress = null
        }
      }, 3000)
    } catch {
      toasts.error(translate('proxmox.actionError'))
      actionInProgress = null
    }
  }

  const color = $derived(statusColor(vm.status))
  const label = $derived(statusLabel(vm.status))
  const isRunning = $derived(vm.status === 'running')
  const isStopped = $derived(vm.status === 'stopped')
  const typeLabel = $derived(vm.type === 'lxc' ? translate('proxmox.typeLxc') : translate('proxmox.typeQemu'))
</script>

<div class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] p-3" data-testid="vm-card">
  <!-- Header: icon, name, VMID, type, status -->
  <div class="mb-2 flex items-center gap-2">
    <div class="h-2.5 w-2.5 flex-shrink-0 rounded-full" style="background-color: {color}"></div>
    {#if vm.type === 'lxc'}
      <Box size={14} class="text-[var(--color-text-muted)]" />
    {:else}
      <Monitor size={14} class="text-[var(--color-text-muted)]" />
    {/if}
    <span class="flex-1 truncate text-sm font-medium text-[var(--color-text)]">{vm.name}</span>
    <span class="text-xs text-[var(--color-text-muted)]">{typeLabel} {vm.vmid}</span>
    {#if actionInProgress}
      <Loader2 size={12} class="animate-spin text-[var(--color-primary)]" />
      <span class="text-xs font-medium text-[var(--color-primary)]">{label}</span>
    {:else}
      <span class="text-xs font-medium" style="color: {color}">{label}</span>
    {/if}
  </div>

  <!-- Metrics (only when running) -->
  {#if isRunning}
    <div class="mb-2 grid grid-cols-3 gap-2 text-xs">
      <div class="min-w-0">
        <span class="text-[var(--color-text-muted)]">{$t('proxmox.cpu')}</span>
        <div class="truncate font-medium text-[var(--color-text)]">{formatCPU(vm.cpu_usage)} / {vm.cpu_cores}c</div>
      </div>
      <div class="min-w-0">
        <span class="text-[var(--color-text-muted)]">{$t('proxmox.ram')}</span>
        <div class="truncate font-medium text-[var(--color-text)]">{formatBytes(vm.mem_used)} / {formatBytes(vm.mem_total)}</div>
      </div>
      <div class="min-w-0">
        <span class="text-[var(--color-text-muted)]">{$t('proxmox.uptime')}</span>
        <div class="truncate font-medium text-[var(--color-text)]">{formatUptime(vm.uptime)}</div>
      </div>
    </div>
  {/if}

  <!-- Actions -->
  {#if !$isViewer}
    <div class="flex gap-1">
      {#if isStopped}
        <button onclick={() => doAction('start')} disabled={!!actionInProgress} class="touch-action-btn rounded p-2 sm:p-1 text-[var(--color-text-muted)] hover:text-[var(--color-success)] disabled:opacity-30" title={$t('proxmox.start')}>
          <Play size={14} />
        </button>
      {/if}
      {#if isRunning}
        <button onclick={() => doAction('shutdown')} disabled={!!actionInProgress} class="touch-action-btn rounded p-2 sm:p-1 text-[var(--color-text-muted)] hover:text-[var(--color-warning)] disabled:opacity-30" title={$t('proxmox.shutdown')}>
          <Power size={14} />
        </button>
        <button onclick={() => doAction('stop')} disabled={!!actionInProgress} class="touch-action-btn rounded p-2 sm:p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)] disabled:opacity-30" title={$t('proxmox.stop')}>
          <Square size={14} />
        </button>
        <button onclick={() => doAction('reboot')} disabled={!!actionInProgress} class="touch-action-btn rounded p-2 sm:p-1 text-[var(--color-text-muted)] hover:text-[var(--color-primary)] disabled:opacity-30" title={$t('proxmox.reboot')}>
          <RotateCw size={14} />
        </button>
      {/if}
    </div>
  {/if}
</div>
