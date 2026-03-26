<script lang="ts">
  import { Loader2, AlertCircle, Settings, Play, MonitorPlay } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { pollingInterval } from '../../stores/polling'
  import { usePolling } from '../../utils/usePolling'
  import { t, translate } from '../../i18n'
  import SessionCard from './SessionCard.svelte'

  interface Session {
    id: string
    user_name: string
    title: string
    subtitle?: string
    media_type: string
    year?: number
    progress: number
    state: string
    player: string
    platform?: string
    thumb_path?: string
    duration: number
    position: number
  }

  interface MediaOverview {
    sessions: Session[]
    session_count: number
    server_name: string
    server_type: string
  }

  interface Props {
    config?: string
    active?: boolean
  }

  const { config = '{}', active = true }: Props = $props()

  let loading = $state(true)
  let refreshing = $state(false)
  let error = $state<string | null>(null)
  let data = $state<MediaOverview | null>(null)

  const parsedConfig = $derived(() => {
    try { return JSON.parse(config) } catch { return {} }
  })

  const serverType = $derived(parsedConfig().server_type || 'jellyfin')

  async function fetchData() {
    const hadData = data !== null
    if (!hadData) loading = true
    error = null
    try {
      data = await api.get<MediaOverview>(`/api/mediaserver/${serverType}/sessions`)
    } catch (err: unknown) {
      const status = (err as { status?: number })?.status
      if (status === 404) {
        error = 'not_configured'
      } else if (status === 502) {
        error = 'fetch_error'
      } else {
        error = 'fetch_error'
        toasts.error(translate('mediaserver.fetchError'))
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

  $effect(() => {
    if (active) polling.start(); else polling.stop()
    return () => polling.stop()
  })
</script>

<div class="relative h-full" data-testid="mediaserver-widget">
  {#if refreshing}<div class="absolute end-1 top-1 z-10"><Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" /></div>{/if}
  {#if loading && !data}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('mediaserver.loading')}
    </div>
  {:else if error === 'not_configured'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Settings size={24} />
      <p>{$t('mediaserver.notConfigured')}</p>
      <p class="text-xs">{$t('mediaserver.configureHint')}</p>
    </div>
  {:else if error === 'fetch_error' && !data}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-danger)]" />
      <p>{$t('mediaserver.fetchError')}</p>
      <button onclick={fetchData} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if data}
    <div class="space-y-3">
      <!-- Header -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          {#if data.session_count > 0}
            <div class="flex items-center gap-1 text-sm font-semibold text-[var(--color-success)]">
              <Play size={14} />
              {data.session_count}
            </div>
          {/if}
          {#if data.server_name}
            <div class="text-xs text-[var(--color-text-muted)]">{data.server_name}</div>
          {/if}
        </div>
        <div class="text-xs text-[var(--color-text-muted)]">
          {data.server_type === 'plex' ? 'Plex' : 'Jellyfin'}
        </div>
      </div>

      <!-- Sessions -->
      {#if data.session_count > 0}
        <div class="space-y-1.5">
          {#each data.sessions as session (session.id)}
            <SessionCard {session} serverType={data.server_type} />
          {/each}
        </div>
      {:else}
        <div class="flex flex-col items-center justify-center gap-2 py-4 text-center">
          <MonitorPlay size={24} class="text-[var(--color-text-muted)]" />
          <p class="text-xs text-[var(--color-text-muted)]">{$t('mediaserver.noSessions')}</p>
        </div>
      {/if}
    </div>
  {/if}
</div>
