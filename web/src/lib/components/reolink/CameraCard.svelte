<script lang="ts">
  import { onMount } from 'svelte'
  import { Video, Camera, Maximize2, X } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { t } from '../../i18n'
  import { isViewer } from '../../stores/auth'

  interface Props {
    id: string
    name: string
    active?: boolean
  }

  const { id, name, active = true }: Props = $props()

  let live = $state(false)
  let mseActive = $state(false)
  let fullscreen = $state(false)

  function toggleFullscreen() {
    fullscreen = !fullscreen
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && fullscreen) {
      fullscreen = false
    }
  }
  let snapshotTs = $state(Date.now())
  const imgSrc = $derived(`/api/reolink/cameras/${id}/snapshot?t=${snapshotTs}`)
  let timer: ReturnType<typeof setTimeout> | null = null
  let destroyed = false

  // MSE streaming state
  let go2rtcWsUrl = $state<string | null>(null)
  let videoEl: HTMLVideoElement | undefined = $state()
  let ws: WebSocket | null = null
  let mediaSource: MediaSource | null = null
  let sourceBuffer: SourceBuffer | null = null
  let bufferQueue: ArrayBuffer[] = []
  let mseObjectUrl: string | null = null

  // Fetch go2rtc stream URL on mount
  onMount(() => {
    api.get<{ go2rtc?: string }>(`/api/reolink/cameras/${id}/stream`)
      .then(info => { go2rtcWsUrl = info.go2rtc || null })
      .catch(() => { /* ignored */ })

    scheduleRefresh()
    return () => {
      destroyed = true
      stopRefresh()
      stopMSE()
    }
  })

  // Pause/resume based on active prop
  let wasActive = true
  $effect(() => {
    if (!active && wasActive) {
      stopRefresh()
      stopMSE()
      live = false
      wasActive = false
    } else if (active && !wasActive) {
      wasActive = true
      if (!destroyed) scheduleRefresh()
    }
  })

  // Snapshot polling
  let refreshGeneration = 0

  function scheduleRefresh() {
    if (destroyed || mseActive) return
    const gen = ++refreshGeneration
    const delay = live ? 500 : 5000
    timer = setTimeout(() => {
      if (gen !== refreshGeneration || destroyed || mseActive) return
      const ts = Date.now()
      const next = `/api/reolink/cameras/${id}/snapshot?t=${ts}`
      const img = new Image()
      img.onload = () => {
        if (gen === refreshGeneration && !destroyed && !mseActive) {
          snapshotTs = ts
          scheduleRefresh()
        }
      }
      img.onerror = () => {
        if (gen === refreshGeneration && !destroyed && !mseActive) scheduleRefresh()
      }
      img.src = next
    }, delay)
  }

  function stopRefresh() {
    refreshGeneration++
    if (timer) {
      clearTimeout(timer)
      timer = null
    }
  }

  // Live toggle
  function toggleLive() {
    live = !live
    if (live && go2rtcWsUrl) {
      stopRefresh()
      startMSE()
    } else if (live) {
      // No go2rtc — fast snapshot fallback
      stopRefresh()
      scheduleRefresh()
    } else {
      stopMSE()
      stopRefresh()
      scheduleRefresh()
    }
  }

  // Pending WebSocket URL — set by startMSE, consumed by $effect below
  let pendingWsUrl = $state<string | null>(null)

  // Reactive: attach MediaSource to <video> when element becomes available after startMSE
  let mseAttached = false
  $effect(() => {
    if (videoEl && mseObjectUrl && mseActive && !mseAttached) {
      mseAttached = true
      videoEl.src = mseObjectUrl
    }
    if (!mseActive) {
      mseAttached = false
    }
  })

  // MSE streaming via go2rtc WebSocket
  function startMSE() {
    if (!go2rtcWsUrl || destroyed) return

    // Build absolute WebSocket URL from relative or absolute path
    let wsUrl: string
    if (go2rtcWsUrl.startsWith('/')) {
      const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      wsUrl = `${proto}//${window.location.host}${go2rtcWsUrl}`
    } else {
      wsUrl = go2rtcWsUrl.replace(/^http/, 'ws')
    }

    mediaSource = new MediaSource()
    mseObjectUrl = URL.createObjectURL(mediaSource)
    pendingWsUrl = wsUrl
    mseActive = true

    mediaSource.addEventListener('sourceopen', () => {
      if (pendingWsUrl) connectWebSocket(pendingWsUrl)
    }, { once: true })
  }

  function connectWebSocket(wsUrl: string) {
    if (destroyed || !mseActive) return

    ws = new WebSocket(wsUrl)
    ws.binaryType = 'arraybuffer'

    let codecReceived = false

    ws.onopen = () => {
      // Request MSE stream from go2rtc
      ws!.send(JSON.stringify({ type: 'mse', value: '' }))
    }

    ws.onmessage = (event) => {
      // First message is text with codec info: {"type":"mse","value":"video/mp4; codecs=\"...\""}
      if (!codecReceived && typeof event.data === 'string') {
        codecReceived = true
        try {
          const msg = JSON.parse(event.data)
          const mimeCodec = msg.value || 'video/mp4; codecs="avc1.640029"'

          if (!MediaSource.isTypeSupported(mimeCodec)) {
            console.warn('[MSE] Codec not supported:', mimeCodec, '— falling back to snapshots')
            stopMSE()
            live = true
            scheduleRefresh()
            return
          }

          sourceBuffer = mediaSource!.addSourceBuffer(mimeCodec)
          sourceBuffer.mode = 'segments'
          sourceBuffer.addEventListener('updateend', flushQueue)
        } catch (e) {
          console.error('[MSE] Failed to create SourceBuffer:', e)
          stopMSE()
          scheduleRefresh()
        }
        return
      }

      // Binary messages = media segments
      if (event.data instanceof ArrayBuffer) {
        appendToBuffer(event.data)
      }
    }

    ws.onclose = () => {
      if (mseActive && !destroyed) {
        setTimeout(() => {
          if (mseActive && !destroyed) connectWebSocket(wsUrl)
        }, 2000)
      }
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  function appendToBuffer(data: ArrayBuffer) {
    if (!sourceBuffer) return

    if (sourceBuffer.updating) {
      bufferQueue.push(data)
      return
    }

    try {
      sourceBuffer.appendBuffer(data)
    } catch {
      // Buffer full or error — reset
      bufferQueue = []
    }
  }

  function flushQueue() {
    if (!sourceBuffer || sourceBuffer.updating || bufferQueue.length === 0) return

    // Keep video near live edge
    if (videoEl && videoEl.buffered.length > 0) {
      const bufferedEnd = videoEl.buffered.end(videoEl.buffered.length - 1)
      // Remove old buffer to avoid memory growth
      if (bufferedEnd - (videoEl.currentTime || 0) > 4) {
        videoEl.currentTime = bufferedEnd - 0.5
      }
      if (videoEl.buffered.start(0) < bufferedEnd - 10) {
        try {
          sourceBuffer.remove(0, bufferedEnd - 5)
          return // wait for updateend after remove
        } catch { /* ignored */ }
      }
    }

    const chunk = bufferQueue.shift()
    if (chunk) {
      try {
        sourceBuffer.appendBuffer(chunk)
      } catch {
        bufferQueue = []
      }
    }
  }

  function stopMSE() {
    mseActive = false
    pendingWsUrl = null
    bufferQueue = []

    if (ws) {
      ws.onclose = null
      ws.onerror = null
      ws.close()
      ws = null
    }

    if (sourceBuffer && mediaSource && mediaSource.readyState === 'open') {
      try {
        mediaSource.removeSourceBuffer(sourceBuffer)
      } catch { /* ignored */ }
    }
    sourceBuffer = null

    if (mediaSource && mediaSource.readyState === 'open') {
      try { mediaSource.endOfStream() } catch { /* ignored */ }
    }
    mediaSource = null

    if (mseObjectUrl) {
      URL.revokeObjectURL(mseObjectUrl)
      mseObjectUrl = null
    }

    if (videoEl) {
      videoEl.src = ''
    }
  }
</script>

<svelte:window onkeydown={onKeydown} />

<!-- Fullscreen backdrop -->
{#if fullscreen}
  <div
    class="fixed inset-0 z-50 bg-black/95"
    role="presentation"
    ondblclick={toggleFullscreen}
  ></div>
{/if}

<!-- Single camera container — repositioned via CSS for fullscreen -->
<div
  class="group overflow-hidden {fullscreen ? 'fixed inset-0 z-50 flex items-center justify-center bg-black' : 'relative rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)]'}"
  role="button"
  tabindex="0"
  ondblclick={toggleFullscreen}
>
  {#if mseActive}
    <video
      bind:this={videoEl}
      autoplay
      muted
      playsinline
      class={fullscreen ? 'h-full w-full object-contain' : 'w-full object-cover'}
    ></video>
  {:else}
    <img
      src={imgSrc}
      alt={name}
      class={fullscreen ? 'h-full w-full object-contain' : 'w-full object-cover'}
      loading="lazy"
    />
  {/if}

  <!-- Live badge -->
  {#if live}
    <div class="absolute start-2 top-2 flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] font-bold text-white {mseActive ? 'bg-red-600' : 'bg-amber-600'}">
      <span class="h-1.5 w-1.5 animate-pulse rounded-full bg-white"></span>
      {mseActive ? 'LIVE' : 'SNAPSHOT'}
    </div>
  {/if}

  <!-- Camera name -->
  <div class="absolute start-0 end-0 top-0 bg-gradient-to-b from-black/70 to-transparent px-2 pb-4 pt-1.5 text-center">
    <span class="{fullscreen ? 'text-sm' : 'text-xs'} font-medium text-white">{name}</span>
  </div>

  <!-- Live toggle (editor+) -->
  {#if !$isViewer}
    <button
      onclick={toggleLive}
      class="absolute end-2 top-2 rounded-full p-1.5 text-white transition-colors {live ? 'bg-red-600 hover:bg-red-700' : 'bg-black/50 hover:bg-black/70'}"
      title={live ? $t('reolink.snapshot') : (go2rtcWsUrl ? $t('reolink.liveStream') : $t('reolink.liveFast'))}
    >
      {#if live}
        <Camera size={fullscreen ? 18 : 14} />
      {:else}
        <Video size={fullscreen ? 18 : 14} />
      {/if}
    </button>
  {/if}

  <!-- Expand button (hover, inline only) -->
  {#if !fullscreen}
    <button
      onclick={toggleFullscreen}
      class="absolute bottom-2 end-2 rounded-full bg-black/50 p-1.5 text-white opacity-0 transition-opacity hover:bg-black/70 group-hover:opacity-100"
      title={$t('reolink.fullscreen')}
    >
      <Maximize2 size={14} />
    </button>
  {/if}

  <!-- Close button (fullscreen only) -->
  {#if fullscreen}
    <button
      onclick={() => fullscreen = false}
      class="absolute start-4 top-4 rounded-full bg-black/60 p-2 text-white hover:bg-black/80"
    >
      <X size={20} />
    </button>
  {/if}
</div>
