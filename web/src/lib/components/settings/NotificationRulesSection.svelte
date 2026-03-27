<script lang="ts">
  import { Trash2, Plus, ToggleLeft, ToggleRight } from 'lucide-svelte'
  import { t } from '../../i18n'

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

  interface Props {
    rules: NotifRule[]
    providers: NotifProvider[]
    onCreateRule: (data: { event_type: string; provider_id: number }) => Promise<boolean>
    onDelete: (id: number) => void
    onToggle: (id: number, enabled: boolean) => void
  }

  const { rules, providers, onCreateRule, onDelete, onToggle }: Props = $props()

  let showAddRule = $state(false)
  let newRuleEventType = $state('container_down')
  let newRuleProviderId = $state<number>(0)

  async function handleCreate() {
    if (!newRuleProviderId) return
    const success = await onCreateRule({
      event_type: newRuleEventType,
      provider_id: newRuleProviderId,
    })
    if (success) {
      showAddRule = false
    }
  }
</script>

<div>
  <h4 class="mb-2 text-sm font-medium text-[var(--color-text)]">{$t('notifications.rules')}</h4>
  {#if rules.length === 0}
    <p class="text-xs text-[var(--color-text-muted)]">{$t('notifications.noRules')}</p>
  {:else}
    <div class="space-y-2">
      {#each rules as rule (rule.id)}
        <div class="flex items-center justify-between rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] p-3">
          <div class="flex items-center gap-2">
            <span class="rounded bg-[var(--color-bg-tertiary)] px-2 py-0.5 text-xs font-mono text-[var(--color-text-secondary)]">{rule.event_type}</span>
            <span class="text-xs text-[var(--color-text-muted)]">&rarr; {providers.find(p => p.id === rule.provider_id)?.name ?? '?'}</span>
          </div>
          <div class="flex items-center gap-1">
            <button onclick={() => onToggle(rule.id, rule.enabled)}
              class="rounded p-1.5"
              class:text-[var(--color-primary)]={rule.enabled}
              class:text-[var(--color-text-muted)]={!rule.enabled}>
              {#if rule.enabled}<ToggleRight size={16} />{:else}<ToggleLeft size={16} />{/if}
            </button>
            <button onclick={() => onDelete(rule.id)}
              class="rounded p-1.5 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]">
              <Trash2 size={14} />
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}

  {#if providers.length > 0 && !showAddRule}
    <button onclick={() => { showAddRule = true; newRuleProviderId = providers[0]?.id ?? 0 }}
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
        {#each providers as p (p.id)}
          <option value={p.id}>{p.name} ({p.type})</option>
        {/each}
      </select>
      <div class="flex justify-end gap-2">
        <button onclick={() => showAddRule = false}
          class="rounded-lg border border-[var(--color-border)] px-3 py-1.5 text-sm text-[var(--color-text-secondary)]">{$t('common.cancel')}</button>
        <button onclick={handleCreate}
          class="rounded-lg bg-[var(--color-primary)] px-3 py-1.5 text-sm text-white hover:bg-[var(--color-primary-hover)]">{$t('common.save')}</button>
      </div>
    </div>
  {/if}
</div>
