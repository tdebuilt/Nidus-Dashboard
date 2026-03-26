<script lang="ts">
  import { CalendarDays, Check, Clock } from 'lucide-svelte'

  interface CalendarItem {
    id: number
    title: string
    overview: string
    airDateUtc: string
    inCinemas: string
    physicalRelease: string
    hasFile: boolean
    seriesTitle: string
  }

  interface Props {
    item: CalendarItem
  }

  const { item }: Props = $props()

  const dateStr = $derived(() => {
    const raw = item.airDateUtc || item.inCinemas || item.physicalRelease
    if (!raw) return ''
    const d = new Date(raw)
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
  })

  const displayTitle = $derived(
    item.seriesTitle ? `${item.seriesTitle} — ${item.title}` : item.title
  )
</script>

<div class="flex items-center gap-2 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] p-2.5">
  <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-[var(--color-bg-secondary)]">
    {#if item.hasFile}
      <Check size={14} class="text-[var(--color-success)]" />
    {:else}
      <Clock size={14} class="text-[var(--color-text-muted)]" />
    {/if}
  </div>
  <div class="min-w-0 flex-1">
    <div class="truncate text-sm font-medium text-[var(--color-text)]">{displayTitle}</div>
    {#if dateStr()}
      <div class="flex items-center gap-1 text-xs text-[var(--color-text-muted)]">
        <CalendarDays size={10} />
        {dateStr()}
      </div>
    {/if}
  </div>
</div>
