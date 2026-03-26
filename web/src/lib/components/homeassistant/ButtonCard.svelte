<script lang="ts">
  import { Play } from 'lucide-svelte'
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

  const serviceMap: Record<string, { service: string; label: string }> = {
    button: { service: 'press', label: 'homeassistant.press' },
    input_button: { service: 'press', label: 'homeassistant.press' },
    script: { service: 'turn_on', label: 'homeassistant.activate' },
    scene: { service: 'turn_on', label: 'homeassistant.activate' },
  }

  const config = $derived(serviceMap[entity.domain] ?? serviceMap.button)

  async function execute() {
    try {
      await api.post(`/api/homeassistant/services/${entity.domain}/${config.service}`, { entity_id: entity.entity_id })
      onAction?.()
    } catch {
      toasts.error(translate('homeassistant.actionError'))
    }
  }
</script>

<div class="flex items-center gap-3 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2" data-testid="button-card">
  <span class="min-w-0 flex-1 truncate text-sm font-medium text-[var(--color-text)]">{entity.name}</span>
  {#if !$isViewer}
    <button onclick={execute} class="flex items-center gap-1 rounded bg-[var(--color-primary)] px-2 py-1 text-xs font-medium text-white hover:opacity-80" title={$t(config.label)}>
      <Play size={12} />
      {$t(config.label)}
    </button>
  {/if}
</div>
