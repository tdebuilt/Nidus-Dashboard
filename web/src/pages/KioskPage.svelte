<script lang="ts">
  import { api } from '../lib/api/client'
  import { get } from 'svelte/store'
  import WidgetGrid from '../lib/components/WidgetGrid.svelte'
  import DynamicIcon from '../lib/components/DynamicIcon.svelte'
  import { t } from '../lib/i18n'
  import { categories } from '../lib/stores/categories'
  import { editMode } from '../lib/stores/editMode'
  import { navigate } from '../lib/stores/router'

  interface Widget {
    id: number
    category_id: number
    type: string
    title: string
    config: string
    pos_x: number
    pos_y: number
    width: number
    height: number
    collapsed: boolean
  }

  const widgets = $state<Record<number, Widget[]>>({})
  let selectedCategoryIndex = $state(0)
  let loading = $state(true)
  let rotationInterval = $state(30)
  let rotationTimer: ReturnType<typeof setInterval> | null = null
  let rotationEnabled = $state(true)
  let kioskRef = $state<HTMLElement | null>(null)

  // Read rotation interval from URL params
  function getRotationFromURL(): number {
    if (typeof window === 'undefined') return 30
    const params = new URLSearchParams(window.location.search)
    const val = params.get('rotate')
    if (val) {
      const n = parseInt(val, 10)
      if (n > 0) return n
    }
    return 30
  }

  const selectedCategoryId = $derived(
    $categories.length > 0 ? $categories[selectedCategoryIndex % $categories.length]?.id ?? null : null
  )

  async function loadWidgets(categoryId: number) {
    try {
      widgets[categoryId] = await api.get<Widget[]>(`/api/categories/${categoryId}/widgets`)
    } catch (e) {
      console.error('kiosk: failed to load widgets', categoryId, e)
      widgets[categoryId] = []
    }
  }

  async function loadAll() {
    await categories.reload()
    const cats = get(categories)
    for (const cat of cats) {
      await loadWidgets(cat.id)
    }
    loading = false
  }

  function startRotation() {
    stopRotation()
    if ($categories.length <= 1) return
    rotationTimer = setInterval(() => {
      selectedCategoryIndex = (selectedCategoryIndex + 1) % $categories.length
    }, rotationInterval * 1000)
  }

  function stopRotation() {
    if (rotationTimer) {
      clearInterval(rotationTimer)
      rotationTimer = null
    }
  }

  function toggleRotation() {
    rotationEnabled = !rotationEnabled
    if (rotationEnabled) {
      startRotation()
    } else {
      stopRotation()
    }
  }

  async function enterFullscreen() {
    try {
      if (kioskRef && !document.fullscreenElement) {
        await kioskRef.requestFullscreen()
      }
    } catch {
      // Fullscreen not supported or denied
    }
  }

  function exitKiosk() {
    stopRotation()
    if (document.fullscreenElement) {
      document.exitFullscreen().catch(() => {})
    }
    navigate('/')
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      exitKiosk()
    } else if (e.key === 'ArrowRight' || e.key === ' ') {
      e.preventDefault()
      selectedCategoryIndex = (selectedCategoryIndex + 1) % $categories.length
    } else if (e.key === 'ArrowLeft') {
      e.preventDefault()
      selectedCategoryIndex = (selectedCategoryIndex - 1 + $categories.length) % $categories.length
    } else if (e.key === 'r' || e.key === 'R') {
      toggleRotation()
    } else if (e.key === 'f' || e.key === 'F') {
      enterFullscreen()
    }
  }

  $effect(() => {
    editMode.set(false)
    rotationInterval = getRotationFromURL()
    loadAll().then(() => {
      if ($categories.length > 1) {
        startRotation()
      }
      enterFullscreen()
    })

    if (typeof window !== 'undefined') {
      window.addEventListener('keydown', handleKeydown)
    }

    return () => {
      stopRotation()
      if (typeof window !== 'undefined') {
        window.removeEventListener('keydown', handleKeydown)
      }
    }
  })
</script>

<div bind:this={kioskRef} class="kiosk-container flex min-h-screen flex-col bg-[var(--color-bg)]" data-testid="kiosk-page">
  <!-- Top bar: minimal, auto-hides -->
  <div class="kiosk-topbar flex items-center justify-between px-4 py-2 opacity-0 transition-opacity duration-300 hover:opacity-100">
    <div class="flex items-center gap-2">
      {#each $categories as cat, i (cat.id)}
        <button
          onclick={() => { selectedCategoryIndex = i; if (rotationEnabled) { startRotation() } }}
          class="flex items-center gap-1 rounded-lg px-3 py-1.5 text-xs transition-colors
            {selectedCategoryIndex === i
              ? 'bg-[var(--color-primary)] text-white'
              : 'text-[var(--color-text-muted)] hover:text-[var(--color-text)]'}"
          data-testid="kiosk-tab"
        >
          <DynamicIcon name={cat.icon} size={14} />
          {cat.name}
        </button>
      {/each}
    </div>
    <div class="flex items-center gap-2">
      <button onclick={toggleRotation}
        class="rounded-lg px-2 py-1 text-xs {rotationEnabled ? 'text-[var(--color-primary)]' : 'text-[var(--color-text-muted)]'}"
        title={rotationEnabled ? $t('kiosk.pauseRotation') : $t('kiosk.resumeRotation')}
        data-testid="kiosk-rotation-toggle">
        {rotationEnabled ? '⏸' : '▶'}
      </button>
      <button onclick={exitKiosk}
        class="rounded-lg px-2 py-1 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
        data-testid="kiosk-exit">
        {$t('kiosk.exit')}
      </button>
    </div>
  </div>

  <!-- Content -->
  <main class="flex-1 overflow-y-auto p-4 sm:p-6 2xl:p-10">
    {#if loading}
      <div class="flex items-center justify-center py-16">
        <p class="text-[var(--color-text-secondary)]">{$t('common.loading')}</p>
      </div>
    {:else if selectedCategoryId !== null}
      <WidgetGrid categoryId={selectedCategoryId} widgets={widgets[selectedCategoryId] || []} />
    {/if}
  </main>

  <!-- Rotation indicator -->
  {#if rotationEnabled && $categories.length > 1}
    <div class="kiosk-indicator flex items-center justify-center gap-1 pb-2 2xl:gap-2 2xl:pb-4">
      {#each $categories as _, i (i)}
        <div
          class="h-1.5 rounded-full transition-all duration-300 2xl:h-2.5
            {i === selectedCategoryIndex ? 'w-6 bg-[var(--color-primary)] 2xl:w-10' : 'w-1.5 bg-[var(--color-border)] 2xl:w-2.5'}"
        ></div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .kiosk-container {
    cursor: none;
  }
  .kiosk-container:hover {
    cursor: default;
  }
  .kiosk-topbar:hover {
    opacity: 1 !important;
  }
</style>
