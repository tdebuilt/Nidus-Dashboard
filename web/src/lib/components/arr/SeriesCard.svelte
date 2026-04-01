<script lang="ts">
  import { ChevronDown, ChevronRight, Eye, EyeOff, Loader2 } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { t } from '../../i18n'
  import EpisodeRow from './EpisodeRow.svelte'

  interface SonarrStatistics {
    episodeFileCount: number
    episodeCount: number
    percentOfEpisodes: number
    sizeOnDisk: number
  }

  interface SonarrSeries {
    id: number
    title: string
    year: number
    seasonCount: number
    monitored: boolean
    status: string
    statistics: SonarrStatistics
  }

  interface SonarrEpisode {
    id: number
    title: string
    episodeNumber: number
    seasonNumber: number
    hasFile: boolean
    monitored: boolean
    airDateUtc: string
  }

  interface Props {
    series: SonarrSeries
  }

  const { series }: Props = $props()

  let expanded = $state(false)
  let episodes = $state<SonarrEpisode[]>([])
  let loadingEpisodes = $state(false)
  let episodesError = $state(false)
  let loaded = $state(false)

  const pct = $derived(Math.round(series.statistics.percentOfEpisodes))
  const progressColor = $derived(
    pct >= 100 ? 'var(--color-success)' : pct >= 50 ? 'var(--color-warning)' : 'var(--color-danger)'
  )

  const seasonGroups = $derived(() => {
    const grouped: Record<number, SonarrEpisode[]> = {}
    for (const ep of episodes) {
      (grouped[ep.seasonNumber] ??= []).push(ep)
    }
    return Object.entries(grouped)
      .map(([num, eps]) => [Number(num), eps] as [number, SonarrEpisode[]])
      .sort(([a], [b]) => a - b)
  })

  function formatSize(bytes: number): string {
    if (bytes <= 0) return '—'
    if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(1) + ' GB'
    if (bytes >= 1048576) return (bytes / 1048576).toFixed(0) + ' MB'
    return (bytes / 1024).toFixed(0) + ' KB'
  }

  async function toggle() {
    expanded = !expanded
    if (expanded && !loaded) {
      await fetchEpisodes()
    }
  }

  async function fetchEpisodes() {
    loadingEpisodes = true
    episodesError = false
    try {
      const res = await api.get<{ items: SonarrEpisode[]; total: number }>(`/api/arr/sonarr/episodes/${series.id}`)
      episodes = res.items ?? []
      loaded = true
    } catch {
      episodesError = true
    } finally {
      loadingEpisodes = false
    }
  }
</script>

<div class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)]">
  <button
    class="flex w-full items-center gap-2 px-2.5 py-2 text-left"
    onclick={toggle}
    aria-expanded={expanded}
  >
    {#if expanded}
      <ChevronDown size={14} class="shrink-0 text-[var(--color-text-muted)]" />
    {:else}
      <ChevronRight size={14} class="shrink-0 text-[var(--color-text-muted)]" />
    {/if}
    <div class="min-w-0 flex-1">
      <span class="truncate text-sm text-[var(--color-text)]">{series.title}</span>
      <span class="ml-1 text-xs text-[var(--color-text-muted)]">({series.year})</span>
    </div>
    <div class="flex shrink-0 items-center gap-2 text-xs text-[var(--color-text-muted)]">
      <span>{series.statistics.episodeFileCount}/{series.statistics.episodeCount}</span>
      <div class="h-1.5 w-12 overflow-hidden rounded-full bg-[var(--color-border)]">
        <div class="h-full rounded-full" style="width: {pct}%; background: {progressColor}"></div>
      </div>
      <span>{formatSize(series.statistics.sizeOnDisk)}</span>
      {#if series.monitored}
        <Eye size={12} class="text-[var(--color-primary)]" />
      {:else}
        <EyeOff size={12} />
      {/if}
    </div>
  </button>

  {#if expanded}
    <div class="border-t border-[var(--color-border)] px-3 py-2">
      {#if loadingEpisodes}
        <div class="flex items-center justify-center gap-2 py-2 text-xs text-[var(--color-text-muted)]">
          <Loader2 size={12} class="animate-spin" />
          {$t('arr.loadingEpisodes')}
        </div>
      {:else if episodesError}
        <div class="py-2 text-center text-xs text-[var(--color-danger)]">{$t('arr.libraryError')}</div>
      {:else if episodes.length === 0}
        <div class="py-2 text-center text-xs text-[var(--color-text-muted)]">{$t('arr.noEpisodes')}</div>
      {:else}
        {#each seasonGroups() as [seasonNum, seasonEpisodes] (seasonNum)}
          <div class="mb-2 last:mb-0">
            <div class="mb-1 text-xs font-medium text-[var(--color-text-secondary)]">
              {$t('arr.season')} {seasonNum}
            </div>
            {#each seasonEpisodes as episode (episode.id)}
              <EpisodeRow {episode} />
            {/each}
          </div>
        {/each}
      {/if}
    </div>
  {/if}
</div>
