<script lang="ts">
  import { Play, Pause, Film, Tv, Music } from 'lucide-svelte'

  interface SessionData {
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

  interface Props {
    session: SessionData
    serverType: string
  }

  const { session, serverType }: Props = $props()

  const progressPercent = $derived(Math.round(session.progress * 100))

  const timeRemaining = $derived(() => {
    const remaining = session.duration - session.position
    if (remaining <= 0) return ''
    const hours = Math.floor(remaining / 3600)
    const minutes = Math.floor((remaining % 3600) / 60)
    if (hours > 0) return `${hours}h${minutes.toString().padStart(2, '0')}m`
    return `${minutes}m`
  })

  const thumbUrl = $derived(
    session.thumb_path
      ? `/api/mediaserver/${serverType}/proxy?path=${encodeURIComponent(session.thumb_path)}`
      : null
  )
</script>

<div
  class="flex gap-3 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] p-2.5"
  data-testid="session-card-{session.id}"
>
  <!-- Poster thumbnail -->
  {#if thumbUrl}
    <div class="h-16 w-11 flex-shrink-0 overflow-hidden rounded">
      <img
        src={thumbUrl}
        alt={session.title}
        class="h-full w-full object-cover"
        loading="lazy"
      />
    </div>
  {:else}
    <div class="flex h-16 w-11 flex-shrink-0 items-center justify-center rounded bg-[var(--color-bg-secondary)]">
      {#if session.media_type === 'episode'}
        <Tv size={16} class="text-[var(--color-text-muted)]" />
      {:else if session.media_type === 'track'}
        <Music size={16} class="text-[var(--color-text-muted)]" />
      {:else}
        <Film size={16} class="text-[var(--color-text-muted)]" />
      {/if}
    </div>
  {/if}

  <!-- Info -->
  <div class="min-w-0 flex-1">
    <div class="flex items-center gap-1.5">
      <div class="truncate text-sm font-medium text-[var(--color-text)]">{session.title}</div>
      {#if session.state === 'playing'}
        <Play size={12} class="flex-shrink-0 text-[var(--color-success)]" />
      {:else}
        <Pause size={12} class="flex-shrink-0 text-[var(--color-warning)]" />
      {/if}
    </div>

    {#if session.subtitle}
      <div class="truncate text-xs text-[var(--color-text-muted)]">{session.subtitle}</div>
    {/if}

    <div class="mt-0.5 truncate text-xs text-[var(--color-text-muted)]">
      {session.user_name} · {session.player}
    </div>

    <!-- Progress bar -->
    <div class="mt-1.5 flex items-center gap-2">
      <div class="h-1 flex-1 overflow-hidden rounded-full bg-[var(--color-bg-secondary)]">
        <div
          class="h-full rounded-full transition-all"
          style="width: {progressPercent}%; background-color: {session.state === 'playing' ? 'var(--color-primary)' : 'var(--color-warning)'}"
        ></div>
      </div>
      <div class="flex-shrink-0 text-[10px] text-[var(--color-text-muted)]">
        {timeRemaining()}
      </div>
    </div>
  </div>
</div>
