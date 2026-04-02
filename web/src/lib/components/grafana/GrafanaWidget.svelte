<script lang="ts">
  import { untrack } from 'svelte'
  import { Settings, AlertCircle, KeyRound } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { t } from '../../i18n'
  import { breakpoint } from '../../stores/breakpoint'
  import { getResponsiveColumns } from '../../utils/responsiveColumns'
  import GrafanaPanelCard from './GrafanaPanelCard.svelte'

  interface PanelConfig {
    dashboardUid: string
    dashboardTitle: string
    panelId: number
    panelTitle: string
  }

  interface Props {
    config?: string
    active?: boolean
  }

  const { config = '{}', active = true }: Props = $props()

  let loading = $state(true)
  let error = $state<string | null>(null)
  let embedUrls = $state.raw<{ title: string; url: string }[]>([])

  const parsedConfig = $derived((() => {
    try { return JSON.parse(config) } catch { return {} }
  })())

  const panels = $derived<PanelConfig[]>(parsedConfig.panels ?? [])
  const columns = $derived(getResponsiveColumns(parsedConfig, $breakpoint, 2))
  const theme = $derived<string>(parsedConfig.theme ?? 'dark')

  async function fetchEmbedUrls() {
    if (panels.length === 0) {
      error = 'not_configured'
      loading = false
      return
    }
    error = null
    try {
      const urls = await Promise.all(panels.map(async (p) => {
        const params = new URLSearchParams({
          dashboardUid: p.dashboardUid,
          panelId: String(p.panelId),
          theme,
        })
        const result = await api.get<{ url: string }>(`/api/grafana/embed-url?${params}`)
        return { title: p.panelTitle, url: result.url }
      }))
      embedUrls = urls
    } catch (err: unknown) {
      const { status, message } = err as { status?: number; message?: string }
      if (status === 404) error = 'not_configured'
      else if (message === 'authentication_failed') error = 'auth_error'
      else error = 'fetch_error'
    } finally {
      loading = false
    }
  }

  $effect(() => {
    if (active && panels.length > 0) {
      untrack(() => { fetchEmbedUrls() })
    } else if (panels.length === 0) {
      untrack(() => {
        error = 'not_configured'
        loading = false
      })
    }
  })
</script>

<div class="h-full" data-testid="grafana-widget">
  {#if loading && embedUrls.length === 0}
    <div class="flex h-full items-center justify-center text-sm text-[var(--color-text-muted)]">
      {$t('grafana.loading')}
    </div>
  {:else if error === 'not_configured'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Settings size={24} />
      <p>{$t('grafana.notConfigured')}</p>
      <p class="text-xs">{$t('grafana.configureHint')}</p>
    </div>
  {:else if error === 'auth_error'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <KeyRound size={24} class="text-[var(--color-danger)]" />
      <p>{$t('grafana.authError')}</p>
      <p class="text-xs">{$t('grafana.authErrorHint')}</p>
      <button onclick={fetchEmbedUrls} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if error === 'fetch_error'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-danger)]" />
      <p>{$t('grafana.fetchError')}</p>
      <button onclick={fetchEmbedUrls} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else}
    <div class="grid gap-2" style="grid-template-columns: repeat({columns}, 1fr);">
      {#each embedUrls as panel (panel.url)}
        <GrafanaPanelCard url={panel.url} title={panel.title} />
      {/each}
    </div>
  {/if}
</div>
