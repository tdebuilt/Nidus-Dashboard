<script lang="ts">
  import { Bell, Trash2, Plus, Send, ToggleLeft, ToggleRight } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { confirm } from '../../stores/confirm'
  import { t, translate } from '../../i18n'
  import { onMount } from 'svelte'

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
  interface NotifRule {
    id: number
    event_type: string
    provider_id: number
    enabled: boolean
    config: string
  }

  let notifProviders = $state<NotifProvider[]>([])
  let notifRules = $state<NotifRule[]>([])
  let showAddProvider = $state(false)
  let newProviderType = $state('gotify')
  let newProviderName = $state('')
  let newProviderUrl = $state('')
  let newProviderToken = $state('')
  let newProviderConfig = $state('')
  let showAddRule = $state(false)
  let newRuleEventType = $state('container_down')
  let newRuleProviderId = $state<number>(0)

  onMount(() => { loadNotifications() })

  async function loadNotifications() {
    try {
      const [providers, rules] = await Promise.all([
        api.get<NotifProvider[]>('/api/notifications/providers'),
        api.get<NotifRule[]>('/api/notifications/rules'),
      ])
      notifProviders = providers ?? []
      notifRules = rules ?? []
    } catch {
      // not admin or error
    }
  }

  async function createProvider() {
    if (!newProviderName || !newProviderUrl) return
    try {
      await api.post('/api/notifications/providers', {
        type: newProviderType,
        name: newProviderName,
        url: newProviderUrl,
        token: newProviderToken,
        config: newProviderConfig || '{}',
      })
      showAddProvider = false
      newProviderName = ''
      newProviderUrl = ''
      newProviderToken = ''
      newProviderConfig = ''
      toasts.success(translate('notifications.providerCreated'))
      await loadNotifications()
    } catch {
      toasts.error(translate('notifications.providerError'))
    }
  }

  async function deleteProvider(id: number) {
    const ok = await confirm({
      title: translate('notifications.deleteProviderTitle'),
      message: translate('notifications.deleteProviderMessage'),
      confirmLabel: translate('common.delete'),
      destructive: true,
    })
    if (!ok) return
    try {
      await api.delete(`/api/notifications/providers/${id}`)
      toasts.success(translate('notifications.providerDeleted'))
      await loadNotifications()
    } catch {
      toasts.error(translate('notifications.providerError'))
    }
  }

  async function toggleProvider(id: number, enabled: boolean) {
    try {
      await api.put(`/api/notifications/providers/${id}`, { enabled: !enabled })
      await loadNotifications()
    } catch {
      toasts.error(translate('notifications.providerError'))
    }
  }

  async function testProvider(id: number) {
    try {
      await api.post('/api/notifications/test', { provider_id: id })
      toasts.success(translate('notifications.testSent'))
    } catch {
      toasts.error(translate('notifications.testError'))
    }
  }

  async function createRule() {
    if (!newRuleProviderId) return
    try {
      await api.post('/api/notifications/rules', {
        event_type: newRuleEventType,
        provider_id: newRuleProviderId,
        config: '{}',
      })
      showAddRule = false
      toasts.success(translate('notifications.ruleCreated'))
      await loadNotifications()
    } catch {
      toasts.error(translate('notifications.ruleError'))
    }
  }

  async function deleteRule(id: number) {
    try {
      await api.delete(`/api/notifications/rules/${id}`)
      toasts.success(translate('notifications.ruleDeleted'))
      await loadNotifications()
    } catch {
      toasts.error(translate('notifications.ruleError'))
    }
  }

  async function toggleRule(id: number, enabled: boolean) {
    try {
      await api.put(`/api/notifications/rules/${id}`, { enabled: !enabled })
      await loadNotifications()
    } catch {
      toasts.error(translate('notifications.ruleError'))
    }
  }
</script>

<div class="space-y-6">
  <h3 class="text-lg font-semibold text-[var(--color-text)]">{$t('settings.tabs.notifications')}</h3>

  <section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-notifications">
    <div class="mb-3 flex items-center gap-2">
      <Bell size={18} class="text-[var(--color-text-secondary)]" />
      <h3 class="font-semibold text-[var(--color-text)]">{$t('notifications.title')}</h3>
    </div>

    <!-- Providers -->
    <div class="mb-4">
      <h4 class="mb-2 text-sm font-medium text-[var(--color-text)]">{$t('notifications.providers')}</h4>
      {#if notifProviders.length === 0}
        <p class="text-xs text-[var(--color-text-muted)]">{$t('notifications.noProviders')}</p>
      {:else}
        <div class="space-y-2">
          {#each notifProviders as provider (provider.id)}
            <div class="flex items-center justify-between rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] p-3">
              <div class="flex items-center gap-2">
                <span class="rounded bg-[var(--color-bg-tertiary)] px-2 py-0.5 text-xs font-mono text-[var(--color-text-secondary)]">{provider.type}</span>
                <span class="text-sm text-[var(--color-text)]">{provider.name}</span>
                <span class="text-xs text-[var(--color-text-muted)]">{provider.url}</span>
              </div>
              <div class="flex items-center gap-1">
                <button onclick={() => testProvider(provider.id)}
                  class="rounded p-1.5 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]"
                  title={translate('notifications.testSend')}>
                  <Send size={14} />
                </button>
                <button onclick={() => toggleProvider(provider.id, provider.enabled)}
                  class="rounded p-1.5"
                  class:text-[var(--color-primary)]={provider.enabled}
                  class:text-[var(--color-text-muted)]={!provider.enabled}
                  title={provider.enabled ? translate('common.disable') : translate('common.enable')}>
                  {#if provider.enabled}<ToggleRight size={16} />{:else}<ToggleLeft size={16} />{/if}
                </button>
                <button onclick={() => deleteProvider(provider.id)}
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
            class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]">
            <option value="gotify">Gotify</option>
            <option value="ntfy">Ntfy</option>
            <option value="apprise">Apprise</option>
          </select>
          <input type="text" bind:value={newProviderName} placeholder={translate('notifications.providerName')}
            class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]" />
          <input type="text" bind:value={newProviderUrl} placeholder={translate('notifications.providerUrl')}
            class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]" />
          <input type="text" bind:value={newProviderToken} placeholder={translate('notifications.providerToken')}
            class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]" />
          {#if newProviderType === 'ntfy'}
            <input type="text" bind:value={newProviderConfig} placeholder={translate('notifications.ntfyTopic')}
              class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]" />
          {/if}
          <div class="flex justify-end gap-2">
            <button onclick={() => showAddProvider = false}
              class="rounded-lg border border-[var(--color-border)] px-3 py-1.5 text-sm text-[var(--color-text-secondary)]">{$t('common.cancel')}</button>
            <button onclick={createProvider}
              class="rounded-lg bg-[var(--color-primary)] px-3 py-1.5 text-sm text-white hover:bg-[var(--color-primary-hover)]">{$t('common.save')}</button>
          </div>
        </div>
      {/if}
    </div>

    <!-- Rules -->
    <div>
      <h4 class="mb-2 text-sm font-medium text-[var(--color-text)]">{$t('notifications.rules')}</h4>
      {#if notifRules.length === 0}
        <p class="text-xs text-[var(--color-text-muted)]">{$t('notifications.noRules')}</p>
      {:else}
        <div class="space-y-2">
          {#each notifRules as rule (rule.id)}
            <div class="flex items-center justify-between rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] p-3">
              <div class="flex items-center gap-2">
                <span class="rounded bg-[var(--color-bg-tertiary)] px-2 py-0.5 text-xs font-mono text-[var(--color-text-secondary)]">{rule.event_type}</span>
                <span class="text-xs text-[var(--color-text-muted)]">→ {notifProviders.find(p => p.id === rule.provider_id)?.name ?? '?'}</span>
              </div>
              <div class="flex items-center gap-1">
                <button onclick={() => toggleRule(rule.id, rule.enabled)}
                  class="rounded p-1.5"
                  class:text-[var(--color-primary)]={rule.enabled}
                  class:text-[var(--color-text-muted)]={!rule.enabled}>
                  {#if rule.enabled}<ToggleRight size={16} />{:else}<ToggleLeft size={16} />{/if}
                </button>
                <button onclick={() => deleteRule(rule.id)}
                  class="rounded p-1.5 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]">
                  <Trash2 size={14} />
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}

      {#if notifProviders.length > 0 && !showAddRule}
        <button onclick={() => { showAddRule = true; newRuleProviderId = notifProviders[0]?.id ?? 0 }}
          class="mt-2 flex items-center gap-1 text-xs text-[var(--color-primary)] hover:underline">
          <Plus size={12} /> {$t('notifications.addRule')}
        </button>
      {/if}

      {#if showAddRule}
        <div class="mt-3 rounded-lg border border-[var(--color-border)] p-3 space-y-2">
          <select bind:value={newRuleEventType}
            class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]">
            <option value="container_down">{$t('notifications.eventContainerDown')}</option>
            <option value="service_unreachable">{$t('notifications.eventServiceUnreachable')}</option>
          </select>
          <select bind:value={newRuleProviderId}
            class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]">
            {#each notifProviders as p (p.id)}
              <option value={p.id}>{p.name} ({p.type})</option>
            {/each}
          </select>
          <div class="flex justify-end gap-2">
            <button onclick={() => showAddRule = false}
              class="rounded-lg border border-[var(--color-border)] px-3 py-1.5 text-sm text-[var(--color-text-secondary)]">{$t('common.cancel')}</button>
            <button onclick={createRule}
              class="rounded-lg bg-[var(--color-primary)] px-3 py-1.5 text-sm text-white hover:bg-[var(--color-primary-hover)]">{$t('common.save')}</button>
          </div>
        </div>
      {/if}
    </div>
  </section>
</div>
