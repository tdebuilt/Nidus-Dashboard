<script lang="ts">
  import { Trash2, Plus, Send, ToggleLeft, ToggleRight } from 'lucide-svelte'
  import { t, translate } from '../../i18n'

  interface NotifProvider {
    id: number
    type: string
    name: string
    url: string
    has_token: boolean
    enabled: boolean
    config: string
    created_at: string
  }

  interface Props {
    providers: NotifProvider[]
    onCreateProvider: (data: { type: string; name: string; url: string; token: string; config: string }) => Promise<boolean>
    onDelete: (id: number) => void
    onToggle: (id: number, enabled: boolean) => void
    onTest: (id: number) => void
  }

  const { providers, onCreateProvider, onDelete, onToggle, onTest }: Props = $props()

  let showAddProvider = $state(false)
  let newProviderType = $state('gotify')
  let newProviderName = $state('')
  let newProviderUrl = $state('')
  let newProviderToken = $state('')
  let newProviderConfig = $state('')

  async function handleCreate() {
    if (!newProviderName || !newProviderUrl) return
    const success = await onCreateProvider({
      type: newProviderType,
      name: newProviderName,
      url: newProviderUrl,
      token: newProviderToken,
      config: newProviderConfig || '{}',
    })
    if (success) {
      showAddProvider = false
      newProviderName = ''
      newProviderUrl = ''
      newProviderToken = ''
      newProviderConfig = ''
    }
  }
</script>

<div class="mb-4">
  <h4 class="mb-2 text-sm font-medium text-[var(--color-text)]">{$t('notifications.providers')}</h4>
  {#if providers.length === 0}
    <p class="text-xs text-[var(--color-text-muted)]">{$t('notifications.noProviders')}</p>
  {:else}
    <div class="space-y-2">
      {#each providers as provider (provider.id)}
        <div class="flex items-center justify-between rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] p-3">
          <div class="flex items-center gap-2">
            <span class="rounded bg-[var(--color-bg-tertiary)] px-2 py-0.5 text-xs font-mono text-[var(--color-text-secondary)]">{provider.type}</span>
            <span class="text-sm text-[var(--color-text)]">{provider.name}</span>
            <span class="text-xs text-[var(--color-text-muted)]">{provider.url}</span>
          </div>
          <div class="flex items-center gap-1">
            <button onclick={() => onTest(provider.id)}
              class="rounded p-1.5 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]"
              title={translate('notifications.testSend')}>
              <Send size={14} />
            </button>
            <button onclick={() => onToggle(provider.id, provider.enabled)}
              class="rounded p-1.5"
              class:text-[var(--color-primary)]={provider.enabled}
              class:text-[var(--color-text-muted)]={!provider.enabled}
              title={provider.enabled ? translate('common.disable') : translate('common.enable')}>
              {#if provider.enabled}<ToggleRight size={16} />{:else}<ToggleLeft size={16} />{/if}
            </button>
            <button onclick={() => onDelete(provider.id)}
              class="rounded p-1.5 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]">
              <Trash2 size={14} />
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}

  {#if !showAddProvider}
    <button onclick={() => showAddProvider = true}
      class="mt-2 flex items-center gap-1 text-xs text-[var(--color-primary)] hover:underline">
      <Plus size={12} /> {$t('notifications.addProvider')}
    </button>
  {:else}
    <div class="mt-3 rounded-lg border border-[var(--color-border)] p-3 space-y-2">
      <select bind:value={newProviderType}
        aria-label={translate('notifications.addProvider')}
        class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]">
        <option value="gotify">Gotify</option>
        <option value="ntfy">Ntfy</option>
        <option value="apprise">Apprise</option>
      </select>
      <input type="text" bind:value={newProviderName} placeholder={translate('notifications.providerName')}
        aria-label={translate('notifications.providerName')}
        class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]" />
      <input type="text" bind:value={newProviderUrl} placeholder={translate('notifications.providerUrl')}
        aria-label={translate('notifications.providerUrl')}
        class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]" />
      <input type="text" bind:value={newProviderToken} placeholder={translate('notifications.providerToken')}
        aria-label={translate('notifications.providerToken')}
        class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]" />
      {#if newProviderType === 'ntfy'}
        <input type="text" bind:value={newProviderConfig} placeholder={translate('notifications.ntfyTopic')}
          aria-label={translate('notifications.ntfyTopic')}
          class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]" />
      {/if}
      <div class="flex justify-end gap-2">
        <button onclick={() => showAddProvider = false}
          class="rounded-lg border border-[var(--color-border)] px-3 py-1.5 text-sm text-[var(--color-text-secondary)]">{$t('common.cancel')}</button>
        <button onclick={handleCreate}
          class="rounded-lg bg-[var(--color-primary)] px-3 py-1.5 text-sm text-white hover:bg-[var(--color-primary-hover)]">{$t('common.save')}</button>
      </div>
    </div>
  {/if}
</div>
