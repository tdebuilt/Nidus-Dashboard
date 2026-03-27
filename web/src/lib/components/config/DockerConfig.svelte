<script lang="ts">
  import { onMount } from 'svelte'
  import { Loader2 } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { t } from '../../i18n'

  interface Environment {
    id: number
    name: string
    status: string
  }

  interface Props {
    value?: string
    onchange?: (value: string) => void
  }

  const { value = '{}', onchange }: Props = $props()

  let environments = $state.raw<Environment[]>([])
  let loading = $state(true)
  let selectedEnvId = $state<number | null>(null)

  $effect(() => {
    try {
      const parsed = JSON.parse(value)
      selectedEnvId = parsed.env_id ?? null
    } catch {
      selectedEnvId = null
    }
  })

  function handleChange(e: Event) {
    const select = e.target as HTMLSelectElement
    const envId = select.value ? parseInt(select.value, 10) : null
    selectedEnvId = envId
    const config = envId !== null ? { env_id: envId } : {}
    onchange?.(JSON.stringify(config))
  }

  onMount(async () => {
    try {
      const data = await api.get<Environment[]>('/api/docker/environments')
      environments = data ?? []
    } catch {
      environments = []
    } finally {
      loading = false
    }
  })
</script>

<div class="space-y-2">
  <label for="docker-env-select" class="block text-sm text-[var(--color-text-secondary)]">{$t('config.environment')}</label>
  {#if loading}
    <div class="flex items-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={14} class="animate-spin" />
      {$t('common.loading')}
    </div>
  {:else if environments.length === 0}
    <p class="text-sm text-[var(--color-text-muted)]">{$t('config.noEnvironments')}</p>
  {:else}
    <select
      id="docker-env-select"
      onchange={handleChange}
      class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
      data-testid="docker-config-env"
    >
      <option value="">{$t('config.allEnvironments')}</option>
      {#each environments as env, i (env.id ?? i)}
        <option value={env.id} selected={selectedEnvId === env.id}>
          {env.name} ({env.status})
        </option>
      {/each}
    </select>
  {/if}
</div>
