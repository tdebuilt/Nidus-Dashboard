<script lang="ts">
  import { Bell } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { confirm } from '../../stores/confirm'
  import { t, translate } from '../../i18n'
  import NotificationProvidersSection from './NotificationProvidersSection.svelte'
  import NotificationRulesSection from './NotificationRulesSection.svelte'

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

  let notifProviders = $state.raw<NotifProvider[]>([])
  let notifRules = $state.raw<NotifRule[]>([])

  $effect(() => { loadNotifications() })

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

  async function createProvider(data: { type: string; name: string; url: string; token: string; config: string }): Promise<boolean> {
    try {
      await api.post('/api/notifications/providers', data)
      toasts.success(translate('notifications.providerCreated'))
      await loadNotifications()
      return true
    } catch {
      toasts.error(translate('notifications.providerError'))
      return false
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

  async function createRule(data: { event_type: string; provider_id: number }): Promise<boolean> {
    try {
      await api.post('/api/notifications/rules', { ...data, config: '{}' })
      toasts.success(translate('notifications.ruleCreated'))
      await loadNotifications()
      return true
    } catch {
      toasts.error(translate('notifications.ruleError'))
      return false
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

    <NotificationProvidersSection
      providers={notifProviders}
      onCreateProvider={createProvider}
      onDelete={deleteProvider}
      onToggle={toggleProvider}
      onTest={testProvider}
    />

    <NotificationRulesSection
      rules={notifRules}
      providers={notifProviders}
      onCreateRule={createRule}
      onDelete={deleteRule}
      onToggle={toggleRule}
    />
  </section>
</div>
