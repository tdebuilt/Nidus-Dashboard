<script lang="ts">
  import { Plus, Search, Loader2 } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { t, translate } from '../../i18n'
  import { isDocker } from '../../stores/version'

  interface DiscoveredCamera {
    ip: string
    name: string
    model: string
  }

  interface Props {
    onAddCamera: (cam: DiscoveredCamera) => void
  }

  const { onAddCamera }: Props = $props()

  let discovering = $state(false)
  let discovered = $state<DiscoveredCamera[]>([])

  async function discover() {
    discovering = true
    discovered = []
    try {
      discovered = await api.get<DiscoveredCamera[]>('/api/reolink/discover')
    } catch {
      toasts.error(translate('reolink.fetchError'))
    } finally {
      discovering = false
    }
  }
</script>

<div class="mb-2 flex items-center justify-between">
  <span class="text-sm font-medium text-[var(--color-text)]">{$t('reolink.cameras')}</span>
  {#if !$isDocker}
    <button
      onclick={discover}
      disabled={discovering}
      class="flex items-center gap-1 rounded-lg bg-[var(--color-bg-tertiary)] px-2 py-1 text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-primary)]"
    >
      {#if discovering}
        <Loader2 size={12} class="animate-spin" />
      {:else}
        <Search size={12} />
      {/if}
      {$t('reolink.discover')}
    </button>
  {/if}
</div>

{#if discovered.length > 0}
  <div class="mb-2 space-y-1">
    <span class="text-xs text-[var(--color-text-muted)]">{$t('reolink.discovered')}</span>
    {#each discovered as cam (cam.ip)}
      <div class="flex items-center justify-between rounded-lg border border-dashed border-[var(--color-border)] px-3 py-2">
        <div>
          <span class="text-sm text-[var(--color-text)]">{cam.name}</span>
          <span class="ms-2 text-xs text-[var(--color-text-muted)]">{cam.ip}</span>
          {#if cam.model}
            <span class="ms-1 text-xs text-[var(--color-text-muted)]">({cam.model})</span>
          {/if}
        </div>
        <button
          onclick={() => onAddCamera(cam)}
          class="rounded bg-[var(--color-primary)] px-2 py-1 text-xs text-white hover:bg-[var(--color-primary-hover)]"
        >
          <Plus size={12} />
        </button>
      </div>
    {/each}
  </div>
{/if}
