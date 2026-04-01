<script lang="ts">
  import { t } from '../../i18n'

  interface Props {
    value?: string
    onchange?: (value: string) => void
  }

  const { value = '{}', onchange }: Props = $props()

  let slug = $state('')

  $effect(() => {
    try {
      const parsed = JSON.parse(value)
      slug = parsed.slug ?? ''
    } catch {
      // ignore
    }
  })

  function emit() {
    const config: Record<string, string> = {}
    if (slug.trim()) config.slug = slug.trim()
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
</div>
