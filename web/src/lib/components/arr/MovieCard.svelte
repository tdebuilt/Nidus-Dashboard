<script lang="ts">
  import { CircleCheck, Circle, Eye, EyeOff } from 'lucide-svelte'

  interface RadarrMovie {
    id: number
    title: string
    year: number
    hasFile: boolean
    monitored: boolean
    sizeOnDisk: number
    runtime: number
    status: string
  }

  interface Props {
    movie: RadarrMovie
  }

  const { movie }: Props = $props()

  function formatSize(bytes: number): string {
    if (bytes <= 0) return '—'
    if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(1) + ' GB'
    if (bytes >= 1048576) return (bytes / 1048576).toFixed(0) + ' MB'
    return (bytes / 1024).toFixed(0) + ' KB'
  }
</script>

<div class="flex items-center gap-2 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2.5 py-2">
  {#if movie.hasFile}
    <CircleCheck size={14} class="shrink-0 text-[var(--color-success)]" />
  {:else}
    <Circle size={14} class="shrink-0 text-[var(--color-text-muted)]" />
  {/if}
  <div class="min-w-0 flex-1">
    <span class="truncate text-sm text-[var(--color-text)]">{movie.title}</span>
    <span class="ml-1 text-xs text-[var(--color-text-muted)]">({movie.year})</span>
  </div>
  <div class="flex shrink-0 items-center gap-2 text-xs text-[var(--color-text-muted)]">
    <span>{formatSize(movie.sizeOnDisk)}</span>
    {#if movie.monitored}
      <Eye size={12} class="text-[var(--color-primary)]" />
    {:else}
      <EyeOff size={12} />
    {/if}
  </div>
</div>
