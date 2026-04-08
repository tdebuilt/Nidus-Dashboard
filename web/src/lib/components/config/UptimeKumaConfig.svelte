<script lang="ts">
  import { t } from '../../i18n'
  import ResponsiveColumnsConfig from './ResponsiveColumnsConfig.svelte'

  interface Props {
    value?: string
    onchange?: (value: string) => void
  }

  const { value = '{}', onchange }: Props = $props()

  let slug = $state('')
  let columns = $state(1)
  let columnsTablet = $state(0)
  let columnsMobile = $state(0)

  $effect(() => {
    try {
      const parsed = JSON.parse(value)
      slug = parsed.slug ?? ''
      if (parsed.columns) columns = parsed.columns
      if (typeof parsed.columnsTablet === 'number') columnsTablet = parsed.columnsTablet
      if (typeof parsed.columnsMobile === 'number') columnsMobile = parsed.columnsMobile
    } catch {
      // ignore
    }
  })

  function emit() {
    const config: Record<string, unknown> = {}
    if (slug.trim()) config.slug = slug.trim()
    if (columns !== 1) config.columns = columns
    if (columnsTablet > 0) config.columnsTablet = columnsTablet
    if (columnsMobile > 0) config.columnsMobile = columnsMobile
    onchange?.(JSON.stringify(config))
  }
</script>

<div class="space-y-3">
  <div>
    <label for="kuma-slug" class="block text-sm text-[var(--color-text-secondary)]">
      {$t('uptimekuma.slug')}
    </label>
    <input
      id="kuma-slug"
      type="text"
      bind:value={slug}
      oninput={emit}
      placeholder="default"
      class="mt-1 w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
    />
    <p class="mt-1 text-xs text-[var(--color-text-muted)]">{$t('uptimekuma.slugHint')}</p>
  </div>

  <ResponsiveColumnsConfig
    {columns}
    {columnsTablet}
    {columnsMobile}
    onchange={(d, tab, m) => { columns = d; columnsTablet = tab; columnsMobile = m; emit() }}
  />
</div>
