<script lang="ts">
  import { Circle, CircleCheck } from 'lucide-svelte'

  interface SonarrEpisode {
    id: number
    title: string
    episodeNumber: number
    seasonNumber: number
    hasFile: boolean
    monitored: boolean
    airDateUtc: string
    episodeFile?: { quality: { quality: { name: string } }; mediaInfo?: { audioCodec: string; videoCodec: string } }
  }

  interface Props {
    episode: SonarrEpisode
  }

  const { episode }: Props = $props()

  const code = $derived(
    `S${String(episode.seasonNumber).padStart(2, '0')}E${String(episode.episodeNumber).padStart(2, '0')}`
  )

  const dateStr = $derived(() => {
    if (!episode.airDateUtc) return ''
    const d = new Date(episode.airDateUtc)
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
  })
</script>

<div class="flex items-center gap-2 py-0.5 text-xs">
  {#if episode.hasFile}
    <CircleCheck size={12} class="shrink-0 text-[var(--color-success)]" />
  {:else}
    <Circle size={12} class="shrink-0 text-[var(--color-text-muted)]" />
  {/if}
  <span class="shrink-0 font-mono text-[var(--color-text-muted)]">{code}</span>
  <span class="min-w-0 flex-1 truncate text-[var(--color-text)]">{episode.title}</span>
  {#if episode.hasFile && episode.episodeFile}
    <span class="shrink-0 rounded bg-[var(--color-bg-tertiary)] px-1 py-0.5 text-[10px] text-[var(--color-text-muted)]">{episode.episodeFile.quality.quality.name}</span>
    {#if episode.episodeFile.mediaInfo?.audioCodec}
      <span class="shrink-0 rounded bg-[var(--color-bg-tertiary)] px-1 py-0.5 text-[10px] text-[var(--color-text-muted)]">{episode.episodeFile.mediaInfo.audioCodec}</span>
    {/if}
  {/if}
  {#if dateStr()}
    <span class="shrink-0 text-[var(--color-text-muted)]">{dateStr()}</span>
  {/if}
</div>
