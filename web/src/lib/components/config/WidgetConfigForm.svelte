<script lang="ts">
  import { t } from '../../i18n'
  import { getConfigComponent } from '../../widgetRegistry'

  interface Props {
    type: string
    value?: string
    onchange?: (value: string) => void
  }

  const { type, value = '{}', onchange }: Props = $props()

  const ConfigComponent = $derived(getConfigComponent(type))
</script>

{#if ConfigComponent}
  <ConfigComponent {value} {onchange} />
{:else}
  <p class="text-sm text-[var(--color-text-muted)]">{$t('config.noConfig')}</p>
{/if}
