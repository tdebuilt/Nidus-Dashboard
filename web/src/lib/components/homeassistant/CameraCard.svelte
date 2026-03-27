<script lang="ts">
  import { Camera, ExternalLink } from 'lucide-svelte'
  import { t } from '../../i18n'

  interface EntityInfo {
    entity_id: string
    domain: string
    name: string
    state: string
    attributes: Record<string, unknown>
    icon?: string
    unit_of_measurement?: string
    last_changed: string
  }

  interface Props {
    entity: EntityInfo
    width?: number
    height?: number
    onResize?: (width: number) => void
  }

  const { entity, width, height, onResize }: Props = $props()

  let snapshotUrl = $state('')
  let containerRef = $state<HTMLElement | null>(null)
  let dragging = $state(false)
  let dragStartX = 0
  let dragStartWidth = 0

  function refreshSnapshot() {
    snapshotUrl = `/api/homeassistant/camera/${entity.entity_id}/snapshot?t=${Date.now()}`
  }

  function handlePointerDown(e: PointerEvent) {
    e.preventDefault()
    if (!containerRef) return
    dragging = true
    dragStartX = e.clientX
    dragStartWidth = containerRef.offsetWidth
    ;(e.target as HTMLElement).setPointerCapture(e.pointerId)
  }

  function clampWidth(raw: number): number {
    const maxW = containerRef?.parentElement?.clientWidth ?? Infinity
    return Math.max(120, Math.min(raw, maxW))
  }

  function handlePointerMove(e: PointerEvent) {
    if (!dragging) return
    const newWidth = clampWidth(dragStartWidth + (e.clientX - dragStartX))
    if (containerRef) {
      containerRef.style.width = `${newWidth}px`
    }
  }

  function handlePointerUp(e: PointerEvent) {
    if (!dragging) return
    dragging = false
    const finalWidth = clampWidth(dragStartWidth + (e.clientX - dragStartX))
    onResize?.(finalWidth)
  }

  $effect(() => {
    refreshSnapshot()
    const interval = setInterval(refreshSnapshot, 10000)
    return () => clearInterval(interval)
  })
</script>

<div
  bind:this={containerRef}
  class="relative rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] p-2"
  style={width ? `width: ${width}px; max-width: 100%` : ''}
  data-testid="camera-card"
>
  <div class="mb-1 flex items-center gap-2">
    <Camera size={14} class="text-[var(--color-primary)]" />
    <span class="flex-1 truncate text-sm font-medium text-[var(--color-text)]">{entity.name}</span>
    {#if typeof entity.attributes.stream_url === 'string'}
      <a href={entity.attributes.stream_url as string} target="_blank" rel="noopener" class="text-[var(--color-primary)]" title={$t('homeassistant.stream')}>
        <ExternalLink size={12} />
      </a>
    {/if}
  </div>
  {#if snapshotUrl}
    <img src={snapshotUrl} alt={entity.name}
      class="w-full rounded"
      style={height ? `height: ${height}px; object-fit: cover` : ''}
      loading="lazy" />
  {/if}
  <!-- Resize handle -->
  {#if onResize}
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="absolute end-0 bottom-0 cursor-se-resize p-1"
      onpointerdown={handlePointerDown}
      onpointermove={handlePointerMove}
      onpointerup={handlePointerUp}
    >
      <svg width="10" height="10" viewBox="0 0 10 10" class="text-[var(--color-text-muted)]">
        <path d="M9 1L1 9M9 5L5 9M9 9L9 9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
      </svg>
    </div>
  {/if}
</div>
