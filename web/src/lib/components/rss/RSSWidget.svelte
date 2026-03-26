<script lang="ts">
  import { Loader2, AlertCircle, Settings, Rss, ExternalLink } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { pollingInterval } from '../../stores/polling'
  import { usePolling } from '../../utils/usePolling'
  import { t, translate } from '../../i18n'

  interface FeedItem {
    title: string
    link: string
    description?: string
    published?: string
    author?: string
    source?: string
  }

  interface FeedData {
    items: FeedItem[]
  }

  interface Props {
    config?: string
    active?: boolean
  }

  const { config = '{}', active = true }: Props = $props()

  let loading = $state(true)
  let refreshing = $state(false)
  let error = $state<string | null>(null)
  let data = $state<FeedData | null>(null)

  const parsedConfig = $derived((() => {
    try { return JSON.parse(config) } catch { return {} }
  })())

  const urls = $derived<string[]>(parsedConfig.urls ?? [])
  const maxItems = $derived<number>(parsedConfig.max ?? 20)

  function formatDate(dateStr: string): string {
    if (!dateStr) return ''
    const d = new Date(dateStr)
    const now = new Date()
    const diff = now.getTime() - d.getTime()
    const hours = Math.floor(diff / 3600000)
    const minutes = Math.floor(diff / 60000)

    if (minutes < 1) return translate('rss.justNow')
    if (minutes < 60) return `${minutes}min`
    if (hours < 24) return `${hours}h`
    if (hours < 48) return translate('rss.yesterday')
    return d.toLocaleDateString([], { day: 'numeric', month: 'short' })
  }

  async function fetchData() {
    if (!urls.length) {
      error = 'not_configured'
      loading = false
      return
    }

    if (data === null) loading = true
    error = null
    try {
      const params = new URLSearchParams({
        urls: urls.join(','),
        max: String(maxItems),
      })
      data = await api.get<FeedData>(`/api/rss?${params.toString()}`)
    } catch {
      error = 'fetch_error'
      toasts.error(translate('rss.fetchError'))
    } finally {
      loading = false
    }
  }

  async function fetchDataWrapped() {
    if (data !== null) refreshing = true
    try {
      await fetchData()
    } finally {
      refreshing = false
    }
  }

  const polling = usePolling({
    fetchFn: fetchDataWrapped,
    active: () => active,
    pollingStore: pollingInterval,
    intervalTransform: (ms) => Math.max(ms * 10, 300000),
  })

  $effect(() => {
    if (active) polling.start()
    else polling.stop()
    return () => polling.stop()
  })
</script>

<div class="relative h-full" data-testid="rss-widget">
  {#if refreshing}
    <div class="absolute top-0 end-0 z-10 p-1">
      <Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" />
    </div>
  {/if}
  {#if loading && !data}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('rss.loading')}
    </div>
  {:else if error === 'not_configured'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Settings size={24} />
      <p>{$t('rss.notConfigured')}</p>
      <p class="text-xs">{$t('rss.configureHint')}</p>
    </div>
  {:else if error === 'fetch_error' && !data}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-danger)]" />
      <p>{$t('rss.fetchError')}</p>
      <button onclick={fetchData} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if data}
    {#if !data.items || data.items.length === 0}
      <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
        <Rss size={24} />
        <p>{$t('rss.noArticles')}</p>
      </div>
    {:else}
      <div class="space-y-1 overflow-y-auto">
        {#each data.items as item (item.link + item.title)}
          <a
            href={item.link}
            target="_blank"
            rel="noopener noreferrer"
            class="group flex items-start gap-2 rounded-lg px-2 py-1.5 transition-colors hover:bg-[var(--color-bg-secondary)]"
          >
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-1.5">
                <span class="text-sm font-medium text-[var(--color-text)] line-clamp-2 group-hover:text-[var(--color-primary)]">{item.title}</span>
                <ExternalLink size={10} class="flex-shrink-0 text-[var(--color-text-muted)] opacity-0 group-hover:opacity-100" />
              </div>
              <div class="flex flex-wrap items-center gap-x-2 text-xs text-[var(--color-text-muted)]">
                {#if item.source}
                  <span class="font-medium">{item.source}</span>
                {/if}
                {#if item.published}
                  <span>{formatDate(item.published)}</span>
                {/if}
                {#if item.author}
                  <span>{item.author}</span>
                {/if}
              </div>
            </div>
          </a>
        {/each}
      </div>
    {/if}
  {/if}
</div>
