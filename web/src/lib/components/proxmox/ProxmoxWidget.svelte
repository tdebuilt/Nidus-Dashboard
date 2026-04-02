<script lang="ts">
  import { Loader2, AlertCircle, Settings, KeyRound, Cpu, MemoryStick } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { pollingInterval } from '../../stores/polling'
  import { usePolling } from '../../utils/usePolling'
  import { t } from '../../i18n'
  import VMCard from './VMCard.svelte'

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
    config?: string
    active?: boolean
  }

  const { config = '{}', active = true }: Props = $props()

  let loading = $state(true)
  let refreshing = $state(false)
  let error = $state<string | null>(null)
  let vms = $state.raw<VMInfo[]>([])

  const parsedConfig = $derived((() => {
    try { return JSON.parse(config) } catch { return {} }
  })())

  const filterNode = $derived(parsedConfig.node ?? '')
  const filterType = $derived(parsedConfig.type ?? '')

  const filteredVMs = $derived(vms.filter(vm => {
    if (filterNode && vm.node !== filterNode) return false
    if (filterType && vm.type !== filterType) return false
    return true
  }))

  const runningVMs = $derived(filteredVMs.filter(vm => vm.status === 'running'))

  const globalStats = $derived((() => {
    const running = runningVMs
    if (running.length === 0) return null
    let totalCpu = 0
    let totalMem = 0
    let totalMemMax = 0
    for (const vm of running) {
      totalCpu += vm.cpu_usage * 100
      totalMem += vm.mem_used
      totalMemMax += vm.mem_total
    }
    return { cpu: totalCpu / running.length, mem: totalMem, memMax: totalMemMax }
  })())

  function formatBytes(bytes: number): string {
    if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(1) + ' GB'
    if (bytes >= 1048576) return (bytes / 1048576).toFixed(0) + ' MB'
    return (bytes / 1024).toFixed(0) + ' KB'
  }

  async function fetchData() {
    const hadData = vms.length > 0
    if (!hadData) loading = true
    error = null
    try {
      const data = await api.get<VMInfo[]>('/api/proxmox/vms')
      vms = Array.isArray(data) ? data : []
    } catch (err: unknown) {
      const { status, message } = err as { status?: number; message?: string }
      if (status === 404) error = 'not_configured'
      else if (message === 'authentication_failed') { error = 'auth_error'; polling.stop() }
      else error = 'fetch_error'
    } finally {
      loading = false
    }
  }

  function refresh() {
    fetchData()
  }

  async function fetchDataWrapped() {
    const hadData = vms.length > 0
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

<div class="relative h-full" data-testid="proxmox-widget">
  {#if refreshing}<div class="absolute end-1 top-1 z-10"><Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" /></div>{/if}
  {#if loading && vms.length === 0}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('proxmox.loading')}
    </div>
  {:else if error === 'not_configured'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Settings size={24} />
      <p>{$t('proxmox.notConfigured')}</p>
      <p class="text-xs">{$t('proxmox.configureHint')}</p>
    </div>
  {:else if error === 'auth_error'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <KeyRound size={24} class="text-[var(--color-warning)]" />
      <p>{$t('proxmox.authError')}</p>
      <p class="text-xs">{$t('proxmox.authErrorHint')}</p>
      <button onclick={() => polling.start()} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if error === 'fetch_error' && vms.length === 0}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-danger)]" />
      <p>{$t('proxmox.fetchError')}</p>
      <button onclick={refresh} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if filteredVMs.length === 0}
    <div class="flex h-full items-center justify-center text-sm text-[var(--color-text-muted)]">
      {$t('proxmox.noVMs')}
    </div>
  {:else}
    <div class="space-y-2 overflow-y-auto">
      <!-- Global resource summary -->
      <div class="flex items-center justify-between rounded-lg bg-[var(--color-bg-secondary)] px-3 py-2">
        <div class="text-xs text-[var(--color-text-muted)]">
          <span class="font-semibold text-[var(--color-success)]">{runningVMs.length}</span>
          <span>/{filteredVMs.length} VMs</span>
        </div>
        {#if globalStats}
          <div class="flex items-center gap-3 text-xs">
            <div class="flex items-center gap-1 text-[var(--color-text-muted)]" title="CPU">
              <Cpu size={12} />
              <span class="font-medium text-[var(--color-text)]">{globalStats.cpu.toFixed(1)}%</span>
            </div>
            <div class="flex items-center gap-1 text-[var(--color-text-muted)]" title="RAM">
              <MemoryStick size={12} />
              <span class="font-medium text-[var(--color-text)]">{formatBytes(globalStats.mem)} / {formatBytes(globalStats.memMax)}</span>
            </div>
          </div>
        {/if}
      </div>

      {#each filteredVMs as vm (vm.vmid + ':' + vm.node)}
        <VMCard {vm} onAction={refresh} />
      {/each}
    </div>
  {/if}
</div>
