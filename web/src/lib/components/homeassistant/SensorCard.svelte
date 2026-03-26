<script lang="ts">
  import { Activity } from 'lucide-svelte'

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
  }

  const { entity }: Props = $props()

  const unit = $derived(
    (entity.attributes.unit_of_measurement as string) ?? entity.unit_of_measurement ?? '',
  )
</script>

<div class="flex items-center gap-3 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2" data-testid="sensor-card">
  <Activity size={16} class="flex-shrink-0 text-[var(--color-primary)]" />
  <span class="min-w-0 flex-1 truncate text-sm font-medium text-[var(--color-text)]">{entity.name}</span>
  <span class="text-sm font-semibold text-[var(--color-text)]">
    {entity.state}{#if unit}&nbsp;{unit}{/if}
  </span>
</div>
