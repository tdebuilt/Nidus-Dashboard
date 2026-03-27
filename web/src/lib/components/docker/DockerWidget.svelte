<script lang="ts">
  import { Loader2, AlertCircle, Settings, Cpu, MemoryStick, SearchCheck } from 'lucide-svelte'
  import { SvelteMap } from 'svelte/reactivity'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { pollingInterval } from '../../stores/polling'
  import { usePolling } from '../../utils/usePolling'
  import { t, translate } from '../../i18n'
  import StackCard from './StackCard.svelte'
  import ContainerCard from './ContainerCard.svelte'

  interface ContainerInfo {
    id: string
    name: string
    image: string
    state: string
    status: string
    health: string
    has_update: boolean
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

  interface Props {
    config?: string
    active?: boolean
  }

  const { config = '{}', active = true }: Props = $props()

  interface UpdateInfo {
    container_id: string
    has_update: boolean
  }

  interface ContainerStatsInfo {
    container_id: string
    cpu_percent: number
    mem_usage: number
    mem_limit: number
    mem_percent: number
  }

  let loading = $state(true)
  let refreshing = $state(false)
  let checkingUpdates = $state(false)
  let error = $state<string | null>(null)
  let stacks = $state.raw<StackInfo[]>([])
  let standalone = $state.raw<ContainerInfo[]>([])
  // eslint-disable-next-line svelte/no-unnecessary-state-wrap
  let updates: SvelteMap<string, boolean> = $state(new SvelteMap())
  // eslint-disable-next-line svelte/no-unnecessary-state-wrap
  let stats: SvelteMap<string, ContainerStatsInfo> = $state(new SvelteMap())
  let envHost = $state('')

  const parsedConfig = $derived((() => {
    try { return JSON.parse(config) } catch { return {} }
  })())

  function formatBytes(bytes: number): string {
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
  }

  const globalStats = $derived((() => {
    let totalCpu = 0
    let totalMem = 0
    let totalMemLimit = 0
    let count = 0
    for (const s of stats.values()) {
      totalCpu += s.cpu_percent
      totalMem += s.mem_usage
      totalMemLimit = Math.max(totalMemLimit, s.mem_limit)
      count++
    }
    return count > 0 ? { cpu: totalCpu, mem: totalMem, memLimit: totalMemLimit, count } : null
  })())

  const totalContainers = $derived(
    stacks.reduce((sum, s) => sum + s.containers.length, 0) + standalone.length
  )

  const runningContainers = $derived(
    stacks.reduce((sum, s) => sum + s.containers.filter(c => c.state === 'running').length, 0)
    + standalone.filter(c => c.state === 'running').length
  )

  const configEnvId = $derived(parsedConfig.env_id ?? null)
  const configHost = $derived(parsedConfig.host ?? null)
  let resolvedEnvId = $state<number | null>(null)

  async function resolveEnvId(): Promise<number | null> {
    if (resolvedEnvId !== null && envHost !== '') return resolvedEnvId

    const envs = await api.get<{ id: number; name: string; status: string; host: string }[]>('/api/docker/environments').catch(() => null)
    if (envs && envs.length > 0) {
      const targetId = configEnvId ?? envs[0].id
      const env = envs.find((e) => e.id === targetId) ?? envs[0]
      resolvedEnvId = env.id
      envHost = configHost || env.host || ''
      return resolvedEnvId
    }

    if (configEnvId !== null) {
      resolvedEnvId = configEnvId
      return configEnvId
    }
    return null
  }

  function handleFetchError(err: unknown) {
    const status = (err as { status?: number })?.status
    if (status === 404) { error = 'not_configured'; return }
    error = 'fetch_error'
    if (status !== 502) toasts.error(translate('docker.fetchError'))
  }

  function fetchUpdatesAndStats(envId: number) {
    if (updates.size === 0) checkingUpdates = true
    api.get<UpdateInfo[]>(`/api/docker/environments/${envId}/updates`).then((upd) => {
      const map = new SvelteMap<string, boolean>()
      for (const u of upd ?? []) map.set(u.container_id, u.has_update)
      updates = map
    }).catch(() => {}).finally(() => { checkingUpdates = false })

    api.get<ContainerStatsInfo[]>(`/api/docker/environments/${envId}/stats`).then((st) => {
      const map = new SvelteMap<string, ContainerStatsInfo>()
      for (const s of st ?? []) map.set(s.container_id, s)
      stats = map
    }).catch(() => {})
  }

  async function fetchData() {
    const hadData = stacks.length > 0 || standalone.length > 0
    if (!hadData) loading = true
    error = null
    try {
      const envId = await resolveEnvId()
      if (envId === null) { error = 'not_configured'; return }

      const data = await api.get<{ stacks: StackInfo[]; standalone: ContainerInfo[] }>(
        `/api/docker/environments/${envId}/containers`
      )
      stacks = data.stacks ?? []
      standalone = data.standalone ?? []
      fetchUpdatesAndStats(envId)
    } catch (err: unknown) {
      handleFetchError(err)
    } finally {
      loading = false
    }
  }

  function refresh() {
    fetchData()
  }

  async function fetchDataWrapped() {
    const hadData = stacks.length > 0 || standalone.length > 0
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

<div class="relative h-full" data-testid="docker-widget">
  {#if refreshing}<div class="absolute end-1 top-1 z-10"><Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" /></div>{/if}
  {#if loading && stacks.length === 0}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('docker.loading')}
    </div>
  {:else if error === 'not_configured'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Settings size={24} />
      <p>{$t('docker.notConfigured')}</p>
      <p class="text-xs">{$t('docker.configureHint')}</p>
    </div>
  {:else if error === 'fetch_error' && stacks.length === 0}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-danger)]" />
      <p>{$t('docker.fetchError')}</p>
      <button onclick={refresh} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if stacks.length === 0 && standalone.length === 0}
    <div class="flex h-full items-center justify-center text-sm text-[var(--color-text-muted)]">
      {$t('docker.noContainers')}
    </div>
  {:else}
    <div class="space-y-2 overflow-y-auto">
      <!-- Global resource summary -->
      <div class="flex items-center justify-between rounded-lg bg-[var(--color-bg-secondary)] px-3 py-2" data-testid="docker-global-stats">
        <div class="text-xs text-[var(--color-text-muted)]">
          <span class="font-semibold text-[var(--color-success)]">{runningContainers}</span>
          <span>/{totalContainers} {$t('docker.containers').toLowerCase()}</span>
        </div>
        <div class="flex items-center gap-3 text-xs">
          {#if checkingUpdates}
            <div class="flex items-center gap-1 text-[var(--color-text-muted)]" title={$t('docker.checkingUpdates')}>
              <SearchCheck size={12} class="animate-pulse" />
              <span class="hidden sm:inline">{$t('docker.checkingUpdates')}</span>
            </div>
          {/if}
          {#if globalStats}
            <div class="flex items-center gap-1 text-[var(--color-text-muted)]" title="CPU">
              <Cpu size={12} />
              <span class="font-medium text-[var(--color-text)]">{globalStats.cpu.toFixed(1)}%</span>
            </div>
            <div class="flex items-center gap-1 text-[var(--color-text-muted)]" title="RAM">
              <MemoryStick size={12} />
              <span class="font-medium text-[var(--color-text)]">{formatBytes(globalStats.mem)}</span>
            </div>
          {/if}
        </div>
      </div>

      {#if stacks.length > 0}
        {#each stacks as stack (stack.name + ':' + stack.env_id)}
          <StackCard {stack} {updates} {stats} {envHost} onAction={refresh} />
        {/each}
      {/if}

      {#if standalone.length > 0}
        {#if stacks.length > 0}
          <div class="pt-1 text-xs font-medium text-[var(--color-text-muted)]">
            {$t('docker.standalone')}
          </div>
        {/if}
        {#each standalone as container (container.id)}
          <ContainerCard {container} hasUpdate={updates.get(container.id) ?? false} stats={stats.get(container.id) ?? null} {envHost} onAction={refresh} />
        {/each}
      {/if}
    </div>
  {/if}
</div>
