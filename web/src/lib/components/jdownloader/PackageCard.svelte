<script lang="ts">
  import { t } from '../../i18n'

  interface PackageInfo {
    uuid: number
    name: string
    status: string
    progress: number
    size: number
    downloaded: number
    speed: number
    eta: number
    finished: boolean
    link_count: number
  }

  interface Props {
    pkg: PackageInfo
  }

  const { pkg }: Props = $props()

  function statusColor(status: string): string {
    if (status === 'downloading') return 'var(--color-primary)'
    if (status === 'finished') return 'var(--color-success)'
    if (status === 'queued') return 'var(--color-text-muted)'
    return 'var(--color-warning)'
  }

  function formatSize(bytes: number): string {
    if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(1) + ' GB'
    if (bytes >= 1048576) return (bytes / 1048576).toFixed(0) + ' MB'
    if (bytes >= 1024) return (bytes / 1024).toFixed(0) + ' KB'
    return bytes + ' B'
  }

  function formatETA(seconds: number): string {
    if (seconds <= 0) return ''
    const h = Math.floor(seconds / 3600)
    const m = Math.floor((seconds % 3600) / 60)
    const s = seconds % 60
    if (h > 0) return `${h}h${m}m`
    if (m > 0) return `${m}m${s}s`
    return `${s}s`
  }

  function formatSpeed(bps: number): string {
    if (bps >= 1048576) return (bps / 1048576).toFixed(1) + ' MB/s'
    if (bps >= 1024) return (bps / 1024).toFixed(0) + ' KB/s'
    return ''
  }

  const color = $derived(statusColor(pkg.status))
  const progressWidth = $derived(Math.min(100, Math.max(0, pkg.progress)))
</script>

<div class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2" data-testid="package-card">
  <div class="flex items-center gap-2">
    <div class="h-2 w-2 flex-shrink-0 rounded-full" style="background-color: {color}"></div>
    <span class="min-w-0 flex-1 truncate text-sm font-medium text-[var(--color-text)]">{pkg.name}</span>
    <span class="text-xs text-[var(--color-text-muted)]">{pkg.link_count} {$t('jdownloader.files')}</span>
  </div>

  <!-- Progress bar -->
  <div class="mt-1.5 h-1.5 w-full overflow-hidden rounded-full bg-[var(--color-border)]">
    <div class="h-full rounded-full transition-all" style="width: {progressWidth}%; background-color: {color}"></div>
  </div>

  <!-- Details -->
  <div class="mt-1 flex items-center justify-between text-xs text-[var(--color-text-muted)]">
    <span>{formatSize(pkg.downloaded)} / {formatSize(pkg.size)}</span>
    <div class="flex items-center gap-2">
      {#if pkg.speed > 0}
        <span>{formatSpeed(pkg.speed)}</span>
      {/if}
      {#if pkg.eta > 0}
        <span>{formatETA(pkg.eta)}</span>
      {/if}
      <span>{pkg.progress.toFixed(0)}%</span>
    </div>
  </div>
</div>
