<script lang="ts">
  import { Thermometer, ChevronUp, ChevronDown } from 'lucide-svelte'
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

  const currentTemp = $derived(
    typeof entity.attributes.current_temperature === 'number'
      ? (entity.attributes.current_temperature as number)
      : null,
  )
  const targetTemp = $derived(
    typeof entity.attributes.temperature === 'number'
      ? (entity.attributes.temperature as number)
      : null,
  )
  const hvacMode = $derived(entity.state)
  const unit = $derived((entity.attributes.temperature_unit as string) ?? '°C')

  function modeColor(mode: string): string {
    if (mode === 'heat') return 'var(--color-danger)'
    if (mode === 'cool') return 'var(--color-primary)'
    if (mode === 'heat_cool' || mode === 'auto') return 'var(--color-warning)'
    return 'var(--color-text-muted)'
  }

  async function adjustTemp(delta: number) {
    if (targetTemp === null) return
    const newTemp = targetTemp + delta
    try {
      await api.post('/api/homeassistant/services/climate/set_temperature', {
        entity_id: entity.entity_id,
        data: { temperature: newTemp },
      })
      onAction?.()
    } catch {
      toasts.error(translate('homeassistant.actionError'))
    }
  }
</script>

<div class="flex items-center gap-3 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2" data-testid="climate-card">
  <Thermometer size={16} class="flex-shrink-0" style="color: {modeColor(hvacMode)}" />
  <div class="min-w-0 flex-1">
    <span class="truncate text-sm font-medium text-[var(--color-text)]">{entity.name}</span>
    <div class="flex items-center gap-2 text-xs text-[var(--color-text-muted)]">
      {#if currentTemp !== null}
        <span>{$t('homeassistant.current')}: {currentTemp}{unit}</span>
      {/if}
      <span class="capitalize" style="color: {modeColor(hvacMode)}">{hvacMode}</span>
    </div>
  </div>
  {#if targetTemp !== null}
    <div class="flex items-center gap-1">
      {#if !$isViewer}
        <button onclick={() => adjustTemp(-0.5)} class="touch-action-btn rounded p-2 sm:p-0.5 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]" title={$t('homeassistant.decrease')}>
          <ChevronDown size={14} />
        </button>
      {/if}
      <span class="min-w-[3rem] text-center text-sm font-semibold text-[var(--color-text)]">{targetTemp}{unit}</span>
      {#if !$isViewer}
        <button onclick={() => adjustTemp(0.5)} class="touch-action-btn rounded p-2 sm:p-0.5 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]" title={$t('homeassistant.increase')}>
          <ChevronUp size={14} />
        </button>
      {/if}
    </div>
  {/if}
</div>
