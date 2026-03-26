<script lang="ts">
  import { Loader2 } from 'lucide-svelte'
  import { t } from '../../i18n'
  import { loadConfigComponent } from '../../widgetRegistry'

  interface Props {
    type: string
    value?: string
    onchange?: (value: string) => void
  }

  const { type, value = '{}', onchange }: Props = $props()

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let ConfigComponent = $state<any>(null)
  let loading = $state(true)

  $effect(() => {
    ConfigComponent = null
    loading = true
    loadConfigComponent(type).then(comp => {
      ConfigComponent = comp ?? null
      loading = false
    })
  })
</script>

{#if loading}
  <div class="flex items-center justify-center py-4">
    <Loader2 size={16} class="animate-spin text-[var(--color-text-muted)]" />
  </div>
{:else if ConfigComponent}
  <ConfigComponent {value} {onchange} />
{:else}
  <p class="text-sm text-[var(--color-text-muted)]">{$t('config.noConfig')}</p>
{/if}
