<script lang="ts">
  import { Play, Square, ArrowDown, ArrowUp } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { confirm } from '../../stores/confirm'
  import { t, translate } from '../../i18n'
  import { isViewer } from '../../stores/auth'

  interface TorrentInfo {
    id: number
    name: string
    status: string
    progress: number
    size: number
    downloaded: number
    speed_down: number
    speed_up: number
    eta: number
    ratio: number
    peers: number
    error?: string
  }

  interface Props {
    torrent: TorrentInfo
    onAction?: () => void
  }

  const { torrent, onAction }: Props = $props()

  function statusColor(status: string): string {
    if (status === 'downloading') return 'var(--color-primary)'
    if (status === 'seeding') return 'var(--color-success)'
    if (status === 'stopped') return 'var(--color-text-muted)'
    if (status === 'checking') return 'var(--color-warning)'
    return 'var(--color-text-muted)'
  }

  function formatSize(bytes: number): string {
    if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(1) + ' GB'
    if (bytes >= 1048576) return (bytes / 1048576).toFixed(0) + ' MB'
    if (bytes >= 1024) return (bytes / 1024).toFixed(0) + ' KB'
    return bytes + ' B'
  }

  function formatSpeed(bps: number): string {
    if (bps >= 1048576) return (bps / 1048576).toFixed(1) + ' MB/s'
    if (bps >= 1024) return (bps / 1024).toFixed(0) + ' KB/s'
    return ''
  }

  function formatETA(seconds: number): string {
    if (seconds <= 0) return ''
    const h = Math.floor(seconds / 3600)
    const m = Math.floor((seconds % 3600) / 60)
    if (h > 0) return `${h}h${m}m`
    if (m > 0) return `${m}m`
    return `${seconds}s`
  }

  async function startTorrent() {
    try {
      await api.post(`/api/transmission/torrents/${torrent.id}/start`)
      toasts.success(translate('transmission.actionSuccess'))
      onAction?.()
    } catch {
      toasts.error(translate('transmission.actionError'))
    }
  }

  async function stopTorrent() {
    const ok = await confirm({
      title: translate('transmission.stopConfirm'),
      message: translate('transmission.stopMessage', { name: torrent.name }),
      confirmLabel: translate('transmission.stop'),
      destructive: true,
    })
    if (!ok) return

    try {
      await api.post(`/api/transmission/torrents/${torrent.id}/stop`)
      toasts.success(translate('transmission.actionSuccess'))
      onAction?.()
    } catch {
      toasts.error(translate('transmission.actionError'))
    }
  }

  const color = $derived(statusColor(torrent.status))
  const isStopped = $derived(torrent.status === 'stopped')
  const isActive = $derived(torrent.status === 'downloading' || torrent.status === 'seeding')
  const progressWidth = $derived(Math.min(100, Math.max(0, torrent.progress)))
</script>

<div class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2" data-testid="torrent-card">
  <div class="flex items-center gap-2">
    <div class="h-2 w-2 flex-shrink-0 rounded-full" style="background-color: {color}"></div>
    <span class="min-w-0 flex-1 truncate text-sm font-medium text-[var(--color-text)]">{torrent.name}</span>
    {#if !$isViewer}
      <div class="flex gap-1">
        {#if isStopped}
          <button onclick={startTorrent} class="touch-action-btn rounded p-2 sm:p-1 text-[var(--color-text-muted)] hover:text-[var(--color-success)]" title={$t('transmission.start')}>
            <Play size={14} />
          </button>
        {/if}
        {#if isActive}
          <button onclick={stopTorrent} class="touch-action-btn rounded p-2 sm:p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]" title={$t('transmission.stop')}>
            <Square size={14} />
          </button>
        {/if}
      </div>
    {/if}
  </div>

  <!-- Progress bar -->
  <div class="mt-1.5 h-1.5 w-full overflow-hidden rounded-full bg-[var(--color-border)]">
    <div class="h-full rounded-full transition-all" style="width: {progressWidth}%; background-color: {color}"></div>
  </div>

  <!-- Details -->
  <div class="mt-1 flex items-center justify-between text-xs text-[var(--color-text-muted)]">
    <div class="flex items-center gap-2">
      <span>{torrent.progress.toFixed(1)}%</span>
      <span>{formatSize(torrent.downloaded)} / {formatSize(torrent.size)}</span>
    </div>
    <div class="flex items-center gap-2">
      {#if torrent.speed_down > 0}
        <span class="flex items-center gap-0.5">
          <ArrowDown size={10} /> {formatSpeed(torrent.speed_down)}
        </span>
      {/if}
      {#if torrent.speed_up > 0}
        <span class="flex items-center gap-0.5">
          <ArrowUp size={10} /> {formatSpeed(torrent.speed_up)}
        </span>
      {/if}
      {#if torrent.eta > 0}
        <span>{formatETA(torrent.eta)}</span>
      {/if}
      {#if torrent.ratio > 0}
        <span>R: {torrent.ratio.toFixed(1)}</span>
      {/if}
    </div>
  </div>

  {#if torrent.error}
    <div class="mt-1 truncate text-xs text-[var(--color-danger)]">{torrent.error}</div>
  {/if}
</div>
