<script lang="ts">
  import { Download, Clock } from 'lucide-svelte'
  import { t } from '../../i18n'

  interface QueueItem {
    id: number
    title: string
    status: string
    trackedDownloadState: string
    size: number
    sizeleft: number
    timeleft: string
  }

  interface Props {
    item: QueueItem
  }

  const { item }: Props = $props()

  const progress = $derived(
    item.size > 0 ? Math.round(((item.size - item.sizeleft) / item.size) * 100) : 0
  )

  function formatSize(bytes: number): string {
    if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(1) + ' GB'
    if (bytes >= 1048576) return (bytes / 1048576).toFixed(0) + ' MB'
    return (bytes / 1024).toFixed(0) + ' KB'
  }
</script>

<div class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] p-2.5">
  <div class="mb-1.5 flex items-start justify-between gap-2">
    <span class="min-w-0 flex-1 truncate text-sm font-medium text-[var(--color-text)]">{item.title}</span>
    <span class="shrink-0 text-xs text-[var(--color-text-muted)]">{progress}%</span>
  </div>

  <!-- Progress bar -->
  <div class="mb-1.5 h-1.5 w-full overflow-hidden rounded-full bg-[var(--color-border)]">
    <div
      class="h-full rounded-full transition-all duration-300"
      class:bg-[var(--color-primary)]={progress < 100}
      class:bg-[var(--color-success)]={progress >= 100}
      style="width: {progress}%"
    ></div>
  </div>

  <div class="flex items-center justify-between text-xs text-[var(--color-text-muted)]">
    <span class="flex items-center gap-1">
      <Download size={10} />
      {formatSize(item.sizeleft)} {$t('arr.sizeLeft')}
    </span>
    {#if item.timeleft}
      <span class="flex items-center gap-1">
        <Clock size={10} />
        {item.timeleft}
      </span>
    {/if}
  </div>
</div>
