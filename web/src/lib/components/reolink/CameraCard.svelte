<script lang="ts">
  import { Video, Camera, Maximize2, X } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { t } from '../../i18n'
  import { isViewer } from '../../stores/auth'
  import { MsePlayer, buildWsUrl } from './msePlayer'

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
    if (e.key === 'Escape' && fullscreen) fullscreen = false
  }

  let snapshotTs = $state(Date.now())
  const imgSrc = $derived(`/api/reolink/cameras/${id}/snapshot?t=${snapshotTs}`)
  let imgEl: HTMLImageElement | undefined = $state()
  let timer: ReturnType<typeof setTimeout> | null = null
  let destroyed = false

  // MSE streaming
  let go2rtcWsUrl = $state<string | null>(null)
  let videoEl: HTMLVideoElement | undefined = $state()
  let player: MsePlayer | null = null

  $effect(() => {
    api
      .get<{ go2rtc?: string }>(`/api/reolink/cameras/${id}/stream`)
      .then((info) => { go2rtcWsUrl = info.go2rtc || null })
      .catch(() => { /* ignored */ })

    scheduleRefresh()
    return () => {
      destroyed = true
      stopRefresh()
      player?.destroy()
      player = null
      imgEl = undefined
    }
  })

  // Pause/resume based on active prop
  let wasActive = true
  $effect(() => {
    if (!active && wasActive) {
      stopRefresh()
      player?.stop()
      player = null
      live = false
      mseActive = false
      wasActive = false
    } else if (active && !wasActive) {
      wasActive = true
      if (!destroyed) scheduleRefresh()
    }
  })

  // Reactive: attach video src when element becomes available
  let mseAttached = false
  $effect(() => {
    if (videoEl && mseActive && player && !mseAttached) {
      mseAttached = true
    }
    if (!mseActive) mseAttached = false
  })

  // Snapshot polling
  let refreshGeneration = 0

  function scheduleRefresh() {
    if (destroyed || mseActive) return
    const gen = ++refreshGeneration
    const delay = live ? 500 : 5000
    timer = setTimeout(() => {
      if (gen !== refreshGeneration || destroyed || mseActive) return
      preloadSnapshot(gen)
    }, delay)
  }

  function preloadSnapshot(gen: number) {
    const ts = Date.now()
    const next = `/api/reolink/cameras/${id}/snapshot?t=${ts}`
    const img = new Image()
    img.onload = () => {
      if (gen === refreshGeneration && !destroyed && !mseActive) {
        snapshotTs = ts
        // Force DOM update in case Svelte reactivity doesn't trigger re-render
        if (imgEl) imgEl.src = next
        scheduleRefresh()
      }
    }
    img.onerror = () => {
      if (gen === refreshGeneration && !destroyed && !mseActive) scheduleRefresh()
    }
    img.src = next
  }

  function stopRefresh() {
    refreshGeneration++
    if (timer) {
      clearTimeout(timer)
      timer = null
    }
  }

  function toggleLive() {
    live = !live
    if (live && go2rtcWsUrl) {
      stopRefresh()
      startMSE()
    } else if (live) {
      stopRefresh()
      scheduleRefresh()
    } else {
      player?.stop()
      player = null
      mseActive = false
      stopRefresh()
      scheduleRefresh()
    }
  }

  function startMSE() {
    if (!go2rtcWsUrl || destroyed) return
    const wsUrl = buildWsUrl(go2rtcWsUrl)
    player = new MsePlayer(wsUrl, {
      onActive: (active) => { mseActive = active },
      onFallback: () => {
        live = true
        scheduleRefresh()
      },
    })
    mseActive = true
    // Wait for videoEl to be rendered via mseActive=true, then start in next tick
    requestAnimationFrame(() => {
      if (videoEl && player) player.start(videoEl)
    })
  }
</script>

<svelte:window onkeydown={onKeydown} />

<!-- Fullscreen backdrop -->
{#if fullscreen}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-50 bg-black/95"
    ondblclick={toggleFullscreen}
  ></div>
{/if}

<!-- Single camera container -->
<div
  class="group overflow-hidden {fullscreen ? 'fixed inset-0 z-50 flex items-center justify-center bg-black' : 'relative rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)]'}"
  role="button"
  tabindex="0"
  aria-label={fullscreen ? $t('reolink.exitFullscreen') : $t('reolink.fullscreen')}
  ondblclick={toggleFullscreen}
  onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); toggleFullscreen() } }}
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
      bind:this={imgEl}
      src={imgSrc}
      alt={name}
      class={fullscreen ? 'h-full w-full object-contain' : 'w-full object-cover'}
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
      aria-label={live ? $t('reolink.snapshot') : (go2rtcWsUrl ? $t('reolink.liveStream') : $t('reolink.liveFast'))}
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
      class="absolute bottom-2 end-2 rounded-full bg-black/50 p-1.5 text-white opacity-100 transition-opacity hover:bg-black/70 sm:opacity-0 sm:group-hover:opacity-100"
      title={$t('reolink.fullscreen')}
      aria-label={$t('reolink.fullscreen')}
    >
      <Maximize2 size={14} />
    </button>
  {/if}

  <!-- Close button (fullscreen only) -->
  {#if fullscreen}
    <button
      onclick={() => fullscreen = false}
      class="absolute start-4 top-4 rounded-full bg-black/60 p-2 text-white hover:bg-black/80"
      aria-label={$t('common.close')}
    >
      <X size={20} />
    </button>
  {/if}
</div>
