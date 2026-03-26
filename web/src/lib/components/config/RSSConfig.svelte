<script lang="ts">
  import { t } from '../../i18n'
  import { Trash2, Plus } from 'lucide-svelte'

  interface Props {
    value?: string
    onchange?: (value: string) => void
  }

  const { value = '{}', onchange }: Props = $props()

  let urls = $state<string[]>([''])
  let max = $state(20)

  $effect(() => {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed.urls) && parsed.urls.length > 0) {
        urls = parsed.urls
      }
      max = parsed.max ?? 20
    } catch {
      // ignore
    }
  })

  function emit() {
    const validUrls = urls.filter((u) => u.trim() !== '')
    const config: Record<string, unknown> = { max }
    if (validUrls.length > 0) {
      config.urls = validUrls
    }
    onchange?.(JSON.stringify(config))
  }

  function addUrl() {
    urls = [...urls, '']
  }

  function removeUrl(index: number) {
    urls = urls.filter((_, i) => i !== index)
    emit()
  }

  function updateUrl(index: number, val: string) {
    urls[index] = val
    emit()
  }
</script>

<div class="space-y-3">
  <div>
    <span class="block text-sm text-[var(--color-text-secondary)]">
      {$t('rss.feedUrls')}
    </span>
    <p class="mb-2 text-xs text-[var(--color-text-muted)]">{$t('rss.feedUrlsHint')}</p>
    {#each urls as url, i (i)}
      <div class="mb-2 flex gap-2">
        <input
          type="url"
          value={url}
          oninput={(e) => updateUrl(i, (e.target as HTMLInputElement).value)}
          placeholder="https://blog.example.com/feed.xml"
          class="flex-1 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
        />
        {#if urls.length > 1}
          <button
            type="button"
            onclick={() => removeUrl(i)}
            class="rounded-lg border border-[var(--color-border)] p-2 text-[var(--color-text-muted)] hover:border-[var(--color-danger)] hover:text-[var(--color-danger)]"
          >
            <Trash2 size={14} />
          </button>
        {/if}
      </div>
    {/each}
    <button
      type="button"
      onclick={addUrl}
      class="flex items-center gap-1 text-xs text-[var(--color-primary)] hover:underline"
    >
      <Plus size={12} />
      {$t('rss.addUrl')}
    </button>
  </div>

  <div>
    <label for="rss-max" class="block text-sm text-[var(--color-text-secondary)]">
      {$t('rss.maxArticles')}
    </label>
    <select
      id="rss-max"
      bind:value={max}
      onchange={emit}
      class="mt-1 w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
    >
      <option value={10}>10</option>
      <option value={20}>20</option>
      <option value={50}>50</option>
      <option value={100}>100</option>
    </select>
  </div>
</div>
