<script lang="ts">
  import { Loader2, AlertCircle, Cpu, HardDrive, Clock } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { pollingInterval } from '../../stores/polling'
  import { usePolling } from '../../utils/usePolling'
  import { t, translate } from '../../i18n'

  interface DiskInfo {
    mount: string
    total: number
    used: number
    percent: number
  }

  interface SystemData {
    hostname: string
    os: string
    arch: string
    cpu_percent: number
    cpu_cores: number
    mem_total: number
    mem_used: number
    mem_percent: number
    disks: DiskInfo[]
    uptime: number
    temperature?: number
  }

  interface Props {
    config?: string
    active?: boolean
  }

  const { config: _config = '{}', active = true }: Props = $props()

  let loading = $state(true)
  let refreshing = $state(false)
  let error = $state<string | null>(null)
  let data = $state<SystemData | null>(null)

  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
  }

  function formatUptime(seconds: number): string {
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    const mins = Math.floor((seconds % 3600) / 60)

    if (days > 0) return `${days}${translate('system.days')} ${hours}h`
    if (hours > 0) return `${hours}h ${mins}min`
    return `${mins}min`
  }

  function gaugeColor(percent: number): string {
    if (percent > 90) return 'var(--color-danger)'
    if (percent > 70) return 'var(--color-warning, #f59e0b)'
    return 'var(--color-primary)'
  }

  async function fetchData() {
    if (!data) loading = true
    error = null
    try {
      data = await api.get<SystemData>('/api/system')
    } catch {
      error = 'fetch_error'
      toasts.error(translate('system.fetchError'))
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

  $effect(() => {
    if (active) polling.start(); else polling.stop()
    return () => polling.stop()
  })
</script>

<div class="relative h-full" data-testid="system-widget">
  {#if refreshing}<div class="absolute end-1 top-1 z-10"><Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" /></div>{/if}
  {#if loading && !data}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('system.loading')}
    </div>
  {:else if error === 'fetch_error' && !data}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-danger)]" />
      <p>{$t('system.fetchError')}</p>
      <button onclick={fetchData} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if data}
    <div class="space-y-3">
      <!-- Hostname + uptime -->
      <div class="flex items-center justify-between text-xs text-[var(--color-text-muted)]">
        <span class="font-medium text-[var(--color-text)]">{data.hostname}</span>
        <span class="flex items-center gap-1">
          <Clock size={10} />
          {formatUptime(data.uptime)}
        </span>
      </div>

      <!-- CPU -->
      <div>
        <div class="mb-1 flex items-center justify-between text-xs">
          <span class="flex items-center gap-1 text-[var(--color-text-muted)]">
            <Cpu size={12} />
            CPU
          </span>
          <span class="font-medium text-[var(--color-text)]">
            {data.cpu_percent.toFixed(1)}%
            {#if data.temperature}
              <span class="ms-1 font-normal text-[var(--color-text-muted)]">{data.temperature.toFixed(0)}°C</span>
            {/if}
          </span>
        </div>
        <div class="h-2 overflow-hidden rounded-full bg-[var(--color-bg)]">
          <div
            class="h-full rounded-full transition-all duration-500"
            style="width: {Math.min(data.cpu_percent, 100)}%; background: {gaugeColor(data.cpu_percent)}"
          ></div>
        </div>
        <div class="mt-0.5 text-[10px] text-[var(--color-text-muted)]">{data.cpu_cores} {$t('system.cores')}</div>
      </div>

      <!-- RAM -->
      <div>
        <div class="mb-1 flex items-center justify-between text-xs">
          <span class="text-[var(--color-text-muted)]">RAM</span>
          <span class="font-medium text-[var(--color-text)]">{data.mem_percent.toFixed(1)}%</span>
        </div>
        <div class="h-2 overflow-hidden rounded-full bg-[var(--color-bg)]">
          <div
            class="h-full rounded-full transition-all duration-500"
            style="width: {Math.min(data.mem_percent, 100)}%; background: {gaugeColor(data.mem_percent)}"
          ></div>
        </div>
        <div class="mt-0.5 text-[10px] text-[var(--color-text-muted)]">{formatBytes(data.mem_used)} / {formatBytes(data.mem_total)}</div>
      </div>

      <!-- Disks -->
      {#if data.disks && data.disks.length > 0}
        {#each data.disks as disk (disk.mount)}
          <div>
            <div class="mb-1 flex items-center justify-between text-xs">
              <span class="flex items-center gap-1 text-[var(--color-text-muted)] truncate max-w-[60%]">
                <HardDrive size={12} />
                {disk.mount}
              </span>
              <span class="font-medium text-[var(--color-text)]">{disk.percent.toFixed(1)}%</span>
            </div>
            <div class="h-2 overflow-hidden rounded-full bg-[var(--color-bg)]">
              <div
                class="h-full rounded-full transition-all duration-500"
                style="width: {Math.min(disk.percent, 100)}%; background: {gaugeColor(disk.percent)}"
              ></div>
            </div>
            <div class="mt-0.5 text-[10px] text-[var(--color-text-muted)]">{formatBytes(disk.used)} / {formatBytes(disk.total)}</div>
          </div>
        {/each}
      {/if}
    </div>
  {/if}
</div>
