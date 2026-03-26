<script lang="ts">
  import { ChevronDown, ChevronRight, Play, Square, RefreshCw } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { confirm } from '../../stores/confirm'
  import { t, translate } from '../../i18n'
  import { isViewer } from '../../stores/auth'
  import ContainerCard from './ContainerCard.svelte'

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

  interface StackInfo {
    id: number
    name: string
    env_id: number
    status: string
    containers: ContainerInfo[]
  }

  interface ContainerStatsInfo {
    container_id: string
    cpu_percent: number
    mem_usage: number
    mem_limit: number
    mem_percent: number
  }

  interface Props {
    stack: StackInfo
    updates?: Map<string, boolean>
    stats?: Map<string, ContainerStatsInfo>
    envHost?: string
    onAction?: () => void
  }

  const { stack, updates = new Map(), stats = new Map(), envHost = '', onAction }: Props = $props()

  function formatBytes(bytes: number): string {
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
  }

  const stackStats = $derived((() => {
    let totalCpu = 0
    let totalMem = 0
    let count = 0
    for (const c of stack.containers) {
      const s = stats.get(c.id)
      if (s) {
        totalCpu += s.cpu_percent
        totalMem += s.mem_usage
        count++
      }
    }
    return count > 0 ? { cpu: totalCpu, mem: totalMem } : null
  })())
  let expanded = $state(false)

  function stackStatusColor(status: string): string {
    if (status === 'running') return 'var(--color-success)'
    if (status === 'stopped') return 'var(--color-danger)'
    return 'var(--color-warning)'
  }

  function stackStatusLabel(status: string): string {
    if (status === 'running') return translate('docker.statusRunning')
    if (status === 'stopped') return translate('docker.statusStopped')
    return translate('docker.statusPartial')
  }

  async function doStackAction(action: string) {
    if (action === 'stop') {
      const ok = await confirm({
        title: translate('docker.stopConfirm'),
        message: translate('docker.stopMessage', { name: stack.name }),
        confirmLabel: translate('docker.stop'),
        destructive: true,
      })
      if (!ok) return
    }

    try {
      await api.post(`/api/docker/stacks/${stack.id}/${action}`, {
        env_id: stack.env_id,
        stack_name: stack.name,
        container_ids: stack.containers.map((c) => c.id),
      })
      toasts.success(translate('docker.actionSuccess'))
    } catch {
      toasts.error(translate('docker.actionError'))
    } finally {
      onAction?.()
    }
  }

  const hasStackUpdate = $derived(
    stack.containers.some(c => updates.get(c.id))
  )

  async function doStackUpdate() {
    const ok = await confirm({
      title: translate('docker.updateStackConfirm'),
      message: translate('docker.updateStackMessage', { name: stack.name }),
      confirmLabel: translate('docker.updateStack'),
    })
    if (!ok) return

    try {
      await api.post(`/api/docker/stacks/${stack.id}/update`, {
        env_id: stack.env_id,
        pull_image: true,
      })
      toasts.info(translate('docker.updateStackStarted', { name: stack.name }))
    } catch {
      toasts.error(translate('docker.actionError'))
    } finally {
      onAction?.()
    }
  }

  const color = $derived(stackStatusColor(stack.status))
  const label = $derived(stackStatusLabel(stack.status))
  const isStopped = $derived(stack.status === 'stopped')
  const isRunning = $derived(stack.status === 'running')
</script>

<div class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)]" data-testid="stack-card">
  <!-- Stack header -->
  <div class="px-4 py-3">
    <div class="flex w-full items-center gap-2">
      <button
        onclick={() => expanded = !expanded}
        class="flex flex-1 items-center gap-2 text-start min-w-0"
        aria-expanded={expanded}
        title={expanded ? translate('docker.collapseStack') : translate('docker.expandStack')}
      >
        {#if expanded}
          <ChevronDown size={16} class="flex-shrink-0 text-[var(--color-text-muted)]" />
        {:else}
          <ChevronRight size={16} class="flex-shrink-0 text-[var(--color-text-muted)]" />
        {/if}

        <div class="h-2.5 w-2.5 flex-shrink-0 rounded-full" style="background-color: {color}"></div>

        <span class="flex-1 truncate text-sm font-semibold text-[var(--color-text)]">{stack.name}</span>

        <span class="flex-shrink-0 text-xs font-medium" style="color: {color}">{label}</span>
      </button>

      <!-- Stack actions -->
      {#if !$isViewer}
        <div class="flex flex-shrink-0 gap-1">
          {#if hasStackUpdate && isRunning}
            <button onclick={doStackUpdate} class="rounded p-1 text-[var(--color-warning)] hover:text-[var(--color-primary)]" title={$t('docker.updateStack')}>
              <RefreshCw size={14} />
            </button>
          {/if}
          {#if isStopped}
            <button onclick={() => doStackAction('start')} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-success)]" title={$t('docker.start')}>
              <Play size={14} />
            </button>
          {/if}
          {#if isRunning}
            <button onclick={() => doStackAction('stop')} class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]" title={$t('docker.stop')}>
              <Square size={14} />
            </button>
          {/if}
        </div>
      {/if}
    </div>
    <!-- Stats row -->
    <div class="mt-1 flex items-center gap-2 ps-7 text-xs text-[var(--color-text-muted)]">
      <span>{stack.containers.length} {$t('docker.containers').toLowerCase()}</span>
      {#if stackStats}
        <span class="text-[var(--color-border)]">·</span>
        <span>{stackStats.cpu.toFixed(1)}% CPU</span>
        <span class="text-[var(--color-border)]">·</span>
        <span>{formatBytes(stackStats.mem)}</span>
      {/if}
    </div>
  </div>

  <!-- Expanded container list -->
  {#if expanded}
    <div class="space-y-1 px-4 pb-3">
      {#each stack.containers as container (container.id)}
        <ContainerCard {container} hasUpdate={updates.get(container.id) ?? false} inStack={true} stats={stats.get(container.id) ?? null} {envHost} {onAction} />
      {/each}
    </div>
  {/if}
</div>
