<script lang="ts">
  import { Lock, Unlock } from 'lucide-svelte'
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

  const isLocked = $derived(entity.state === 'locked')

  async function toggle() {
    const service = isLocked ? 'unlock' : 'lock'
    try {
      await api.post(`/api/homeassistant/services/lock/${service}`, { entity_id: entity.entity_id })
      onAction?.()
    } catch {
      toasts.error(translate('homeassistant.actionError'))
    }
  }
</script>

<div class="flex items-center gap-3 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2" data-testid="lock-card">
  {#if !$isViewer}
    <button onclick={toggle} class="flex-shrink-0 rounded p-1" title={$t(isLocked ? 'homeassistant.unlock' : 'homeassistant.lock')}>
      {#if isLocked}
        <Lock size={18} style="color: var(--color-success)" />
      {:else}
        <Unlock size={18} style="color: var(--color-warning)" />
      {/if}
    </button>
  {:else}
    {#if isLocked}
      <Lock size={18} style="color: var(--color-success)" class="flex-shrink-0" />
    {:else}
      <Unlock size={18} style="color: var(--color-warning)" class="flex-shrink-0" />
    {/if}
  {/if}
  <span class="min-w-0 flex-1 truncate text-sm font-medium text-[var(--color-text)]">{entity.name}</span>
  <span class="text-xs font-medium" style="color: {isLocked ? 'var(--color-success)' : 'var(--color-warning)'}">
    {isLocked ? $t('homeassistant.locked') : $t('homeassistant.unlocked')}
  </span>
</div>
