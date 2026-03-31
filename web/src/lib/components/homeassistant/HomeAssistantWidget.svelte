<script lang="ts">
  import { Loader2, AlertCircle, Settings } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { pollingInterval } from '../../stores/polling'
  import { breakpoint } from '../../stores/breakpoint'
  import { ws } from '../../stores/websocket'
  import { t, translate } from '../../i18n'
  import { getResponsiveColumns } from '../../utils/responsiveColumns'
  import { usePolling } from '../../utils/usePolling'
  import EntityCard from './EntityCard.svelte'

  interface EntityInfo {
    entity_id: string
    domain: string
    name: string
    state: string
    attributes: Record<string, unknown>
    icon?: string
    unit_of_measurement?: string
    last_changed: string
  }

  interface Props {
    config?: string
    active?: boolean
    widgetId?: number
    widgetType?: string
    widgetTitle?: string
  }

  const { config = '{}', active = true, widgetId, widgetType = '', widgetTitle = '' }: Props = $props()

  let loading = $state(true)
  let refreshing = $state(false)
  let error = $state<string | null>(null)
  let entities = $state.raw<EntityInfo[]>([])

  const parsedConfig = $derived((() => {
    try { return JSON.parse(config) } catch { return {} }
  })())

  const entitySize = $derived<'sm' | 'md' | 'lg'>(
    parsedConfig.entitySize === 'sm' ? 'sm' : parsedConfig.entitySize === 'lg' ? 'lg' : 'md'
  )
  const columns = $derived(getResponsiveColumns(parsedConfig, $breakpoint, 1))
  let cameraSizes = $state<Record<string, number>>({})
  const configCameraSizes = $derived<Record<string, number>>(parsedConfig.cameraSizes ?? {})
  // cameraSizes is also mutated locally in handleCameraResize, so writable $derived won't work
  // eslint-disable-next-line svelte/prefer-writable-derived
  $effect(() => { cameraSizes = configCameraSizes })

  async function handleCameraResize(entityId: string, width: number) {
    cameraSizes = { ...cameraSizes, [entityId]: width }
    if (!widgetId || !widgetType || !widgetTitle) return
    const newConfig = { ...parsedConfig, cameraSizes }
    try {
      await api.put(`/api/widgets/${widgetId}`, {
        type: widgetType,
        title: widgetTitle,
        config: JSON.stringify(newConfig),
      })
    } catch { /* silent */ }
  }

  const filteredEntities = $derived((() => {
    if (parsedConfig.entities && Array.isArray(parsedConfig.entities)) {
      const entityMap = new Map(entities.map((e) => [e.entity_id, e]))
      return parsedConfig.entities
        .map((id: string) => entityMap.get(id))
        .filter(Boolean) as EntityInfo[]
    }
    if (parsedConfig.domains && Array.isArray(parsedConfig.domains)) {
      return entities.filter((e) => parsedConfig.domains.includes(e.domain))
    }
    return entities.filter((e) =>
      ['light', 'switch', 'sensor', 'climate', 'camera', 'input_boolean', 'cover', 'button', 'input_button', 'script', 'lock', 'scene'].includes(e.domain),
    )
  })())

  async function fetchData() {
    const hadData = entities.length > 0
    if (!hadData) loading = true
    error = null
    try {
      const data = await api.get<EntityInfo[]>('/api/homeassistant/entities')
      entities = data ?? []
    } catch (err: unknown) {
      const status = (err as { status?: number })?.status
      if (status === 404) {
        error = 'not_configured'
      } else if (status === 502) {
        error = 'fetch_error'
      } else {
        error = 'fetch_error'
        toasts.error(translate('homeassistant.fetchError'))
      }
    } finally {
      loading = false
    }
  }

  function handleStateChanged(payload: unknown) {
    const data = payload as { new_state?: { entity_id: string; state: string; attributes?: Record<string, unknown>; last_changed?: string } }
    if (data?.new_state) {
      const entityId = data.new_state.entity_id
      entities = entities.map((e) => {
        if (e.entity_id === entityId) {
          return {
            ...e,
            state: data.new_state!.state,
            attributes: data.new_state!.attributes ?? e.attributes,
            last_changed: data.new_state!.last_changed ?? e.last_changed,
          }
        }
        return e
      })
    }
  }

  // WebSocket: always active (even on hidden tabs)
  $effect(() => {
    const unsub = ws.on('ha:state_changed', handleStateChanged)
    return () => unsub?.()
  })

  // Polling: controlled by active prop
  async function fetchDataWrapped() {
    const hadData = entities.length > 0
    if (hadData) refreshing = true
    await fetchData()
    refreshing = false
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

<div class="relative h-full" data-testid="homeassistant-widget">
  {#if refreshing}
    <div class="absolute end-1 top-1 z-10">
      <Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" />
    </div>
  {/if}
  {#if loading && entities.length === 0}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('homeassistant.loading')}
    </div>
  {:else if error === 'not_configured'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Settings size={24} />
      <p>{$t('homeassistant.notConfigured')}</p>
      <p class="text-xs">{$t('homeassistant.configureHint')}</p>
    </div>
  {:else if error === 'fetch_error' && entities.length === 0}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-danger)]" />
      <p>{$t('homeassistant.fetchError')}</p>
      <button onclick={fetchData} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if filteredEntities.length === 0}
    <div class="flex h-full items-center justify-center text-sm text-[var(--color-text-muted)]">
      {$t('homeassistant.noEntities')}
    </div>
  {:else}
    <div class="ha-size-{entitySize} grid gap-2 overflow-y-auto" style="grid-template-columns: repeat({columns}, 1fr);">
      {#each filteredEntities as entity (entity.entity_id)}
        <EntityCard {entity} onAction={fetchData}
          cameraWidth={cameraSizes[entity.entity_id]}
          onCameraResize={(w) => handleCameraResize(entity.entity_id, w)} />
      {/each}
    </div>
  {/if}
</div>
