<script lang="ts">
  import { ChevronUp, ChevronDown, Square } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { t, translate } from '../../i18n'
  import { isViewer } from '../../stores/auth'

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
    onAction?: () => void
  }

  const { entity, onAction }: Props = $props()

  const position = $derived(
    typeof entity.attributes.current_position === 'number'
      ? (entity.attributes.current_position as number)
      : null,
  )

  const stateLabel = $derived((() => {
    switch (entity.state) {
      case 'open': return 'homeassistant.open'
      case 'closed': return 'homeassistant.closed'
      case 'opening': return 'homeassistant.opening'
      case 'closing': return 'homeassistant.closing'
      default: return 'homeassistant.unknown'
    }
  })())

  const stateColor = $derived(
    entity.state === 'open' || entity.state === 'opening'
      ? 'var(--color-success)'
      : 'var(--color-text-muted)',
  )

  async function sendCommand(service: string) {
    try {
      await api.post(`/api/homeassistant/services/cover/${service}`, { entity_id: entity.entity_id })
      onAction?.()
    } catch {
      toasts.error(translate('homeassistant.actionError'))
    }
  }
</script>

<div class="flex items-center gap-3 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2" data-testid="cover-card">
  <div class="min-w-0 flex-1">
    <span class="truncate text-sm font-medium text-[var(--color-text)]">{entity.name}</span>
    <div class="flex items-center gap-2 text-xs" style="color: {stateColor}">
      <span>{$t(stateLabel)}</span>
      {#if position !== null}
        <span class="text-[var(--color-text-muted)]">{position}%</span>
      {/if}
    </div>
  </div>
  {#if !$isViewer}
    <div class="flex gap-1">
      <button onclick={() => sendCommand('open_cover')} class="rounded p-1 hover:bg-[var(--color-border)]" title={$t('homeassistant.openCover')} aria-label={$t('homeassistant.openCover')}>
        <ChevronUp size={16} class="text-[var(--color-text)]" />
      </button>
      <button onclick={() => sendCommand('stop_cover')} class="rounded p-1 hover:bg-[var(--color-border)]" title={$t('homeassistant.stopCover')} aria-label={$t('homeassistant.stopCover')}>
        <Square size={14} class="text-[var(--color-text)]" />
      </button>
      <button onclick={() => sendCommand('close_cover')} class="rounded p-1 hover:bg-[var(--color-border)]" title={$t('homeassistant.closeCover')} aria-label={$t('homeassistant.closeCover')}>
        <ChevronDown size={16} class="text-[var(--color-text)]" />
      </button>
    </div>
  {/if}
</div>
