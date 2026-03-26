<script lang="ts">
  import { Settings, Palette, SlidersHorizontal, Server, Users, Bell, Webhook, Database, UserCircle } from 'lucide-svelte'
  import { isAdmin } from '../lib/stores/auth'
  import { t } from '../lib/i18n'

  import SettingsPillNav from '../lib/components/settings/SettingsPillNav.svelte'
  import AppearanceTab from '../lib/components/settings/AppearanceTab.svelte'
  import PreferencesTab from '../lib/components/settings/PreferencesTab.svelte'
  import ServicesTab from '../lib/components/settings/ServicesTab.svelte'
  import SecurityTab from '../lib/components/settings/SecurityTab.svelte'
  import NotificationsTab from '../lib/components/settings/NotificationsTab.svelte'
  import WebhooksTab from '../lib/components/settings/WebhooksTab.svelte'
  import BackupTab from '../lib/components/settings/BackupTab.svelte'
  import AccountTab from '../lib/components/settings/AccountTab.svelte'

  const allTabs = [
    { id: 'preferences', labelKey: 'settings.tabs.preferences', icon: SlidersHorizontal, adminOnly: false },
    { id: 'appearance', labelKey: 'settings.tabs.appearance', icon: Palette, adminOnly: false },
    { id: 'account', labelKey: 'settings.tabs.account', icon: UserCircle, adminOnly: false },
    { id: 'services', labelKey: 'settings.tabs.services', icon: Server, adminOnly: true },
    { id: 'users', labelKey: 'settings.tabs.users', icon: Users, adminOnly: true },
    { id: 'notifications', labelKey: 'settings.tabs.notifications', icon: Bell, adminOnly: true },
    { id: 'webhooks', labelKey: 'settings.tabs.webhooks', icon: Webhook, adminOnly: true },
    { id: 'backup', labelKey: 'settings.tabs.backup', icon: Database, adminOnly: true },
  ]

  let activeTab = $state('preferences')
  let slideDirection = $state<'right' | 'left'>('right')

  const visibleTabs = $derived(
    allTabs.filter(tab => !tab.adminOnly || $isAdmin)
  )

  const tabOrder = allTabs.map(t => t.id)

  function handleTabChange(tabId: string) {
    const oldIndex = tabOrder.indexOf(activeTab)
    const newIndex = tabOrder.indexOf(tabId)
    slideDirection = newIndex > oldIndex ? 'right' : 'left'
    activeTab = tabId
  }

  function handleSettingsReload() {
    // Force re-mount of the active tab by briefly switching
    const current = activeTab
    activeTab = ''
    requestAnimationFrame(() => { activeTab = current })
  }
</script>

<div data-testid="settings-page">
  <div class="mb-6 flex items-center gap-2">
    <Settings size={22} class="text-[var(--color-text-secondary)]" />
    <h2 class="text-xl font-semibold text-[var(--color-text)]">{$t('settings.title')}</h2>
  </div>

  <SettingsPillNav tabs={visibleTabs} {activeTab} onTabChange={handleTabChange} />

  <div class="overflow-hidden">
    {#key activeTab}
      <div class="animate-slide-{slideDirection}">
        {#if activeTab === 'appearance'}
          <AppearanceTab />
        {:else if activeTab === 'preferences'}
          <PreferencesTab />
        {:else if activeTab === 'services'}
          <ServicesTab />
        {:else if activeTab === 'account'}
          <AccountTab />
        {:else if activeTab === 'users'}
          <SecurityTab />
        {:else if activeTab === 'notifications'}
          <NotificationsTab />
        {:else if activeTab === 'webhooks'}
          <WebhooksTab />
        {:else if activeTab === 'backup'}
          <BackupTab onSettingsReload={handleSettingsReload} />
        {/if}
      </div>
    {/key}
  </div>
</div>
