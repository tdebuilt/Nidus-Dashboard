<script lang="ts">
  import { Loader2, AlertCircle, RefreshCw } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { t } from '../../i18n'
  import RadarrLibrary from './RadarrLibrary.svelte'
  import SonarrLibrary from './SonarrLibrary.svelte'

  interface RadarrMovie {
    id: number; title: string; year: number; hasFile: boolean
    monitored: boolean; sizeOnDisk: number; runtime: number; status: string
  }

  interface SonarrSeries {
    id: number; title: string; year: number; seasonCount: number
    monitored: boolean; status: string
    seasons?: { seasonNumber: number; monitored: boolean }[]
    statistics: { episodeFileCount: number; episodeCount: number; percentOfEpisodes: number; sizeOnDisk: number }
  }

  type LibraryItem = RadarrMovie | SonarrSeries

  interface Props {
    serviceType: string
  }

  const { serviceType }: Props = $props()

  let loading = $state(true)
  let error = $state(false)
  let movies = $state<RadarrMovie[]>([])
  let series = $state<SonarrSeries[]>([])

  async function fetchLibrary() {
    loading = true
    error = false
    try {
      const res = await api.get<{ items: LibraryItem[]; total: number }>(`/api/arr/${serviceType}/library`)
      if (serviceType === 'radarr') {
        movies = (res.items ?? []) as RadarrMovie[]
      } else {
        series = (res.items ?? []) as SonarrSeries[]
      }
    } catch {
      error = true
    } finally {
      loading = false
    }
  }

  $effect(() => {
    fetchLibrary()
  })
</script>

<div class="mt-2 border-t border-[var(--color-border)] pt-2">
  {#if loading}
    <div class="flex items-center justify-center gap-2 py-3 text-xs text-[var(--color-text-muted)]">
      <Loader2 size={12} class="animate-spin" />
      {$t('arr.loadingLibrary')}
    </div>
  {:else if error}
    <div class="flex flex-col items-center gap-2 py-3 text-center text-xs">
      <AlertCircle size={16} class="text-[var(--color-danger)]" />
      <span class="text-[var(--color-text-muted)]">{$t('arr.libraryError')}</span>
      <button onclick={fetchLibrary} class="text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else}
    <div class="mb-1 flex items-center justify-end">
      <button
        onclick={fetchLibrary}
        class="rounded p-0.5 text-[var(--color-text-muted)] hover:bg-[var(--color-bg-tertiary)] hover:text-[var(--color-text)]"
        title={$t('arr.refreshLibrary')}
      >
        <RefreshCw size={12} />
      </button>
    </div>
    {#if serviceType === 'radarr'}
      <RadarrLibrary {movies} />
    {:else}
      <SonarrLibrary {series} />
    {/if}
  {/if}
</div>
