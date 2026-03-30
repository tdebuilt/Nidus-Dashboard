<script lang="ts">
  import { Settings, Zap } from 'lucide-svelte'
  import { t } from '../../i18n'
  import DynamicIcon from '../DynamicIcon.svelte'
  import { getServiceIcon } from './serviceIcons'
  import ServiceOverflowMenu from './ServiceOverflowMenu.svelte'
  import type { ServiceStatus } from '../../stores/serviceStatus'

  interface Props {
    type: string
    displayName: string
    url: string
    status?: ServiceStatus
    isEditing: boolean
    onEdit: () => void
    onTest: () => void
    onDelete: () => void
  }

  const { type, displayName, url, status, isEditing, onEdit, onTest, onDelete }: Props = $props()

  function statusColor(s: ServiceStatus): string {
    if (s === 'up') return 'var(--color-success)'
    if (s === 'down') return 'var(--color-danger)'
    return 'var(--color-text-muted)'
  }

  function statusLabel(s: ServiceStatus): string {
    if (s === 'up') return $t('settings.serviceOnline')
    if (s === 'down') return $t('settings.serviceOffline')
    return $t('settings.serviceChecking')
  }
</script>

<div class="flex flex-col rounded-lg border border-[var(--color-border)] p-4" data-testid="service-row-{type}">
  <!-- Row 1: Icon + Name + Status + Overflow -->
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-2">
      <DynamicIcon name={getServiceIcon(type)} size={18} class="text-[var(--color-primary)]" />
      <span class="text-sm font-medium text-[var(--color-text)]">{displayName}</span>
      {#if status}
        <span
          class="inline-block h-2 w-2 rounded-full"
          style="background-color: {statusColor(status)}"
          title={statusLabel(status)}
        ></span>
      {/if}
    </div>
    {#if !isEditing}
      <ServiceOverflowMenu {onDelete} />
    {/if}
  </div>

  <!-- Row 2: URL -->
  {#if url}
    <a href={url} target="_blank" rel="noopener"
      class="mt-1.5 truncate text-xs text-[var(--color-text-muted)] hover:text-[var(--color-primary)] hover:underline"
      onclick={(e) => e.stopPropagation()}>{url}</a>
  {:else}
    <span class="mt-1.5 text-xs text-[var(--color-text-muted)]">—</span>
  {/if}

  <!-- Row 3: Action buttons -->
  {#if !isEditing}
    <div class="mt-3 flex items-center gap-1.5">
      <button onclick={onEdit}
        class="rounded border border-[var(--color-border)] p-1.5 text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]"
        title={$t('common.configure')}
        aria-label={$t('common.configure')}
        data-testid="service-config-btn">
        <Settings size={14} />
      </button>
      <button onclick={onTest}
        class="rounded border border-[var(--color-border)] p-1.5 text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)] disabled:cursor-not-allowed disabled:opacity-40"
        title={$t('common.test')}
        aria-label={$t('common.test')}
        disabled={!url}
        data-testid="service-test-btn">
        <Zap size={14} />
      </button>
    </div>
  {/if}
</div>
