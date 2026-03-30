<script lang="ts">
  import { Lightbulb } from 'lucide-svelte'
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

  const isOn = $derived(entity.state === 'on')
  const brightness = $derived(
    typeof entity.attributes.brightness === 'number'
      ? Math.round((entity.attributes.brightness as number) / 255 * 100)
      : null,
  )

  async function toggle() {
    const service = isOn ? 'turn_off' : 'turn_on'
    try {
      await api.post(`/api/homeassistant/services/light/${service}`, { entity_id: entity.entity_id })
      onAction?.()
    } catch {
      toasts.error(translate('homeassistant.actionError'))
    }
  }

  async function setBrightness(e: Event) {
    const value = parseInt((e.target as HTMLInputElement).value)
    const haValue = Math.round(value / 100 * 255)
    try {
      await api.post('/api/homeassistant/services/light/turn_on', {
        entity_id: entity.entity_id,
        data: { brightness: haValue },
      })
      onAction?.()
    } catch {
      toasts.error(translate('homeassistant.actionError'))
    }
  }
</script>

<div class="flex items-center gap-3 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2" data-testid="light-card">
  {#if !$isViewer}
    <button onclick={toggle} class="touch-action-btn flex-shrink-0 rounded p-2 sm:p-1" title={$t(isOn ? 'homeassistant.turnOff' : 'homeassistant.turnOn')} aria-label={$t(isOn ? 'homeassistant.turnOff' : 'homeassistant.turnOn')}>
      <Lightbulb size={18} style="color: {isOn ? '#f59e0b' : 'var(--color-text-muted)'}" />
    </button>
  {:else}
    <Lightbulb size={18} style="color: {isOn ? '#f59e0b' : 'var(--color-text-muted)'}" class="flex-shrink-0" />
  {/if}
  <div class="min-w-0 flex-1">
    <span class="truncate text-sm font-medium text-[var(--color-text)]">{entity.name}</span>
    {#if isOn && brightness !== null}
      {#if !$isViewer}
        <div class="mt-1 flex items-center gap-2">
          <input type="range" min="1" max="100" value={brightness} onchange={setBrightness}
            class="h-1 w-full cursor-pointer appearance-none rounded-lg bg-[var(--color-border)]" />
          <span class="text-xs text-[var(--color-text-muted)]">{brightness}%</span>
        </div>
      {:else}
        <div class="mt-1 text-xs text-[var(--color-text-muted)]">{brightness}%</div>
      {/if}
    {/if}
  </div>
  <span class="text-xs font-medium" style="color: {isOn ? '#f59e0b' : 'var(--color-text-muted)'}">
    {isOn ? $t('homeassistant.on') : $t('homeassistant.off')}
  </span>
</div>
