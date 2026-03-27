<script lang="ts">
  import { Webhook, Trash2, Plus, Copy, Link, ToggleLeft, ToggleRight } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { confirm } from '../../stores/confirm'
  import { t, translate } from '../../i18n'

  let webhooks = $state<Array<{id: number; name: string; has_secret: boolean; enabled: boolean; url: string}>>([])
  let _webhooksLoading = $state(true)
  let showAddWebhook = $state(false)
  let newWebhookName = $state('')
  let createdWebhookSecret = $state<string | null>(null)
  let expandedWebhookId = $state<number | null>(null)
  const webhookActions = $state<Record<number, Array<{id: number; webhook_id: number; action_type: string; config: string}>>>({})
  let newActionType = $state('notify')
  let newActionConfig = $state('{}')

  $effect(() => { loadWebhooks() })

  async function loadWebhooks() {
    try { webhooks = await api.get('/api/webhooks') }
    catch { webhooks = [] }
    finally { _webhooksLoading = false }
  }

  async function createWebhook() {
    if (!newWebhookName.trim()) return
    try {
      const result: { secret: string } = await api.post('/api/webhooks', { name: newWebhookName })
      createdWebhookSecret = result.secret
      newWebhookName = ''
      showAddWebhook = false
      loadWebhooks()
      toasts.success(translate('webhooks.created'))
    } catch { toasts.error(translate('webhooks.error')) }
  }

  async function deleteWebhook(id: number) {
    const ok = await confirm({ title: translate('webhooks.deleteTitle'), message: translate('webhooks.deleteMessage'), confirmLabel: translate('common.delete'), destructive: true })
    if (!ok) return
    try {
      await api.delete('/api/webhooks/' + id)
      loadWebhooks()
      toasts.success(translate('webhooks.deleted'))
    } catch { toasts.error(translate('webhooks.error')) }
  }

  async function toggleWebhook(id: number, enabled: boolean) {
    try {
      await api.put('/api/webhooks/' + id, { enabled: !enabled })
      loadWebhooks()
    } catch { toasts.error(translate('webhooks.error')) }
  }

  async function loadWebhookActions(webhookId: number) {
    try {
      webhookActions[webhookId] = await api.get('/api/webhooks/' + webhookId + '/actions')
    } catch { webhookActions[webhookId] = [] }
  }

  async function toggleExpandWebhook(id: number) {
    if (expandedWebhookId === id) { expandedWebhookId = null; return }
    expandedWebhookId = id
    await loadWebhookActions(id)
  }

  async function addAction(webhookId: number) {
    try {
      await api.post('/api/webhooks/' + webhookId + '/actions', { action_type: newActionType, config: newActionConfig })
      newActionType = 'notify'
      newActionConfig = '{}'
      loadWebhookActions(webhookId)
    } catch { toasts.error(translate('webhooks.error')) }
  }

  async function deleteAction(actionId: number, webhookId: number) {
    try {
      await api.delete('/api/webhooks/' + webhookId + '/actions/' + actionId)
      loadWebhookActions(webhookId)
    } catch { toasts.error(translate('webhooks.error')) }
  }

  function copyToClipboard(text: string, toastKey: string) {
    navigator.clipboard.writeText(text)
    toasts.success(translate(toastKey))
  }
</script>

<div class="space-y-6">
  <h3 class="text-lg font-semibold text-[var(--color-text)]">{$t('settings.tabs.webhooks')}</h3>

  <section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-webhooks">
    <div class="mb-3 flex items-center justify-between">
      <div class="flex items-center gap-2">
        <Webhook size={18} class="text-[var(--color-text-secondary)]" />
        <h3 class="font-semibold text-[var(--color-text)]">{$t('webhooks.title')}</h3>
      </div>
      <button onclick={() => showAddWebhook = !showAddWebhook}
        class="flex items-center gap-1.5 rounded-lg bg-[var(--color-primary)] px-3 py-1.5 text-xs text-white hover:bg-[var(--color-primary-hover)]">
        <Plus size={14} /> {$t('webhooks.addWebhook')}
      </button>
    </div>

    {#if createdWebhookSecret}
      <div class="mb-3 rounded-lg border border-[var(--color-success)] bg-[var(--color-success)]/10 p-3">
        <p class="mb-1 text-xs font-medium text-[var(--color-success)]">{$t('webhooks.secretHint')}</p>
        <div class="flex items-center gap-2">
          <code class="flex-1 rounded bg-[var(--color-bg)] px-2 py-1 text-xs text-[var(--color-text)]">{createdWebhookSecret}</code>
          <button onclick={() => { copyToClipboard(createdWebhookSecret!, 'webhooks.secretCopied'); createdWebhookSecret = null }}
            class="rounded border border-[var(--color-border)] p-1 text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]"
            aria-label={$t('common.copy')}>
            <Copy size={14} />
          </button>
        </div>
      </div>
    {/if}

    {#if showAddWebhook}
      <div class="mb-3 rounded-lg border border-[var(--color-border)] p-3">
        <input type="text" bind:value={newWebhookName} placeholder={$t('webhooks.name')}
          class="mb-2 w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]"
          onkeydown={(e) => { if (e.key === 'Enter') createWebhook() }} />
        <button onclick={createWebhook}
          class="rounded-lg bg-[var(--color-primary)] px-3 py-1.5 text-xs text-white hover:bg-[var(--color-primary-hover)]">
          {$t('common.save')}
        </button>
      </div>
    {/if}

    {#if webhooks.length === 0}
      <p class="text-center text-sm text-[var(--color-text-muted)]">{$t('webhooks.noWebhooks')}</p>
    {:else}
      <div class="space-y-2">
        {#each webhooks as wh (wh.id)}
          <div class="rounded-lg border border-[var(--color-border)]">
            <div class="flex items-center justify-between px-4 py-3">
              <div class="min-w-0 flex-1">
                <span class="text-sm font-medium text-[var(--color-text)]">{wh.name}</span>
                <div class="flex items-center gap-2 text-xs text-[var(--color-text-muted)]">
                  <Link size={10} />
                  <code>{window.location.origin}{wh.url}</code>
                  <button onclick={() => copyToClipboard(window.location.origin + wh.url, 'webhooks.urlCopied')}
                    class="text-[var(--color-text-secondary)] hover:text-[var(--color-primary)]"
                    aria-label={$t('common.copy')}>
                    <Copy size={10} />
                  </button>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <button onclick={() => toggleExpandWebhook(wh.id)}
                  class="rounded border border-[var(--color-border)] px-2 py-1 text-xs text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]">
                  {$t('webhooks.actions')}
                </button>
                <button onclick={() => toggleWebhook(wh.id, wh.enabled)}
                  class="text-[var(--color-text-secondary)] hover:text-[var(--color-primary)]"
                  aria-label={wh.enabled ? $t('common.disable') : $t('common.enable')}>
                  {#if wh.enabled}
                    <ToggleRight size={24} class="text-[var(--color-primary)]" />
                  {:else}
                    <ToggleLeft size={24} />
                  {/if}
                </button>
                <button onclick={() => deleteWebhook(wh.id)}
                  class="text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"
                  aria-label={$t('common.delete')}>
                  <Trash2 size={16} />
                </button>
              </div>
            </div>

            {#if expandedWebhookId === wh.id}
              <div class="border-t border-[var(--color-border)] px-4 py-3">
                <div class="mb-2 text-xs font-medium text-[var(--color-text-secondary)]">{$t('webhooks.actions')}</div>
                {#if webhookActions[wh.id]?.length > 0}
                  <div class="mb-2 space-y-1">
                    {#each webhookActions[wh.id] as action (action.id)}
                      <div class="flex items-center justify-between rounded bg-[var(--color-bg)] px-2 py-1.5 text-xs">
                        <span class="font-medium text-[var(--color-text)]">{action.action_type}</span>
                        <div class="flex items-center gap-2">
                          <code class="text-[var(--color-text-muted)]">{action.config}</code>
                          <button onclick={() => deleteAction(action.id, wh.id)}
                            class="text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"
                            aria-label={$t('common.delete')}>
                            <Trash2 size={12} />
                          </button>
                        </div>
                      </div>
                    {/each}
                  </div>
                {/if}
                <div class="flex gap-2">
                  <select bind:value={newActionType}
                    class="rounded border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1 text-xs text-[var(--color-text)]">
                    <option value="notify">{$t('webhooks.actionNotify')}</option>
                    <option value="refresh_widget">{$t('webhooks.actionRefreshWidget')}</option>
                    <option value="invalidate_cache">{$t('webhooks.actionInvalidateCache')}</option>
                  </select>
                  <input type="text" bind:value={newActionConfig} placeholder={'{"provider_id": 1}'}
                    class="flex-1 rounded border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1 text-xs text-[var(--color-text)]" />
                  <button onclick={() => addAction(wh.id)}
                    class="rounded bg-[var(--color-primary)] px-2 py-1 text-xs text-white hover:bg-[var(--color-primary-hover)]"
                    aria-label={$t('common.add')}>
                    <Plus size={12} />
                  </button>
                </div>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </section>
</div>
