<script lang="ts">
  import { Loader2, Settings, TrendingUp, TrendingDown, Minus } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { pollingInterval } from '../../stores/polling'
  import { usePolling } from '../../utils/usePolling'
  import { t, translate } from '../../i18n'

  interface QuoteData {
    symbol: string
    short_name: string
    quote_type: string
    currency: string
    price: number
    change: number
    change_percent: number
    volume: number
    open: number
    day_high: number
    day_low: number
    market_cap: number
    market_state: string
  }

  interface QuotesResponse {
    quotes: QuoteData[]
    fetched_at: number
  }

  interface Props {
    config?: string
    active?: boolean
  }

  const { config = '{}', active = true }: Props = $props()

  let loading = $state(true)
  let refreshing = $state(false)
  let error = $state<string | null>(null)
  let data = $state<QuotesResponse | null>(null)

  const parsedConfig = $derived(
    (() => {
      try {
        return JSON.parse(config)
      } catch {
        return {}
      }
    })(),
  )
  const symbols = $derived<string[]>(parsedConfig.symbols ?? [])
  const showVolume = $derived(parsedConfig.show_volume !== false)
  // market cap not available from Yahoo Finance v8 chart API
  // let showMarketCap = $derived(parsedConfig.show_market_cap !== false)
  const compact = $derived(parsedConfig.compact === true)

  function formatPrice(value: number, currency: string): string {
    if (!value && value !== 0) return '--'
    try {
      return new Intl.NumberFormat(undefined, {
        style: 'currency',
        currency: currency || 'USD',
        minimumFractionDigits: value < 1 ? 4 : 2,
        maximumFractionDigits: value < 1 ? 6 : 2,
      }).format(value)
    } catch {
      return value.toFixed(2)
    }
  }

  function formatChange(change: number, pct: number): string {
    const sign = change >= 0 ? '+' : ''
    return `${sign}${change.toFixed(2)} (${sign}${pct.toFixed(2)}%)`
  }

  function formatCompact(value: number): string {
    if (!value) return '--'
    if (value >= 1e12) return (value / 1e12).toFixed(1) + 'T'
    if (value >= 1e9) return (value / 1e9).toFixed(1) + 'B'
    if (value >= 1e6) return (value / 1e6).toFixed(1) + 'M'
    if (value >= 1e3) return (value / 1e3).toFixed(1) + 'K'
    return value.toString()
  }

  function timeAgo(timestamp: number): string {
    const diff = Math.floor(Date.now() / 1000 - timestamp)
    if (diff < 60) return '< 1 min'
    if (diff < 3600) return `${Math.floor(diff / 60)} min`
    return `${Math.floor(diff / 3600)}h`
  }

  async function fetchData() {
    if (!symbols.length) {
      error = 'not_configured'
      loading = false
      return
    }
    if (!data) loading = true
    error = null
    try {
      const params = new URLSearchParams({ symbols: symbols.join(',') })
      const resp = await api.get<QuotesResponse>(`/api/finance/quotes?${params}`)
      data = resp
    } catch {
      if (!data) {
        error = 'fetch_error'
      }
      toasts.error(translate('finance.fetchError'))
    } finally {
      loading = false
    }
  }

  async function fetchDataWrapped() {
    const hadData = data !== null
    if (hadData) refreshing = true
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
  })

  $effect(() => {
    if (active) polling.start(); else polling.stop()
    return () => polling.stop()
  })

  // Re-fetch when symbols list changes (e.g. user adds a symbol in config)
  const symbolsKey = $derived(symbols.join(','))
  $effect(() => {
    void symbolsKey
    fetchData()
  })
</script>

<div class="relative flex h-full flex-col">
  {#if refreshing}<div class="absolute end-1 top-1 z-10"><Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" /></div>{/if}
  {#if loading && !data}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('finance.loading')}
    </div>
  {:else if error === 'not_configured' && !data}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm">
      <Settings size={24} class="text-[var(--color-text-muted)]" />
      <p class="text-[var(--color-text-muted)]">{$t('finance.notConfigured')}</p>
      <p class="text-xs text-[var(--color-text-muted)]">{$t('finance.configureHint')}</p>
    </div>
  {:else if error && !data}
    <div class="flex h-full items-center justify-center text-sm text-[var(--color-danger)]">
      {$t('finance.fetchError')}
    </div>
  {:else if data}
    <div class="flex-1 overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-[var(--color-border)] text-xs text-[var(--color-text-muted)]">
            <th class="px-1.5 py-1 text-start">{$t('finance.symbol')}</th>
            <th class="px-1.5 py-1 text-end">{$t('finance.price')}</th>
            <th class="px-1.5 py-1 text-end">{$t('finance.change')}</th>
            {#if showVolume && !compact}
              <th class="px-1.5 py-1 text-end">{$t('finance.volume')}</th>
            {/if}
            {#if !compact}
              <th class="px-1.5 py-1 text-end">{$t('finance.open')}</th>
              <th class="px-1.5 py-1 text-end">{$t('finance.highLow')}</th>
            {/if}
          </tr>
        </thead>
        <tbody>
          {#each data.quotes as quote (quote.symbol)}
            <tr class="border-b border-[var(--color-border)] last:border-0 hover:bg-[var(--color-bg-secondary)]">
              <td class="px-1.5 py-1.5">
                <div class="flex items-center gap-1.5">
                  {#if quote.change > 0}
                    <TrendingUp size={14} class="shrink-0 text-[var(--color-success)]" />
                  {:else if quote.change < 0}
                    <TrendingDown size={14} class="shrink-0 text-[var(--color-danger)]" />
                  {:else}
                    <Minus size={14} class="shrink-0 text-[var(--color-text-muted)]" />
                  {/if}
                  <div>
                    <span class="font-medium text-[var(--color-text)]">{quote.symbol}</span>
                    {#if !compact}
                      <div class="truncate text-xs text-[var(--color-text-muted)]" title={quote.short_name}>
                        {quote.short_name}
                      </div>
                    {/if}
                  </div>
                </div>
              </td>
              <td class="px-1.5 py-1.5 text-end font-medium text-[var(--color-text)]">
                {formatPrice(quote.price, quote.currency)}
              </td>
              <td
                class="px-1.5 py-1.5 text-end text-xs font-medium"
                class:text-[var(--color-success)]={quote.change > 0}
                class:text-[var(--color-danger)]={quote.change < 0}
                class:text-[var(--color-text-muted)]={quote.change === 0}
              >
                {formatChange(quote.change, quote.change_percent)}
              </td>
              {#if showVolume && !compact}
                <td class="px-1.5 py-1.5 text-end text-[var(--color-text-muted)]">
                  {formatCompact(quote.volume)}
                </td>
              {/if}
              {#if !compact}
                <td class="px-1.5 py-1.5 text-end text-[var(--color-text-muted)]">
                  {formatPrice(quote.open, quote.currency)}
                </td>
                <td class="whitespace-nowrap px-1.5 py-1.5 text-end text-xs text-[var(--color-text-muted)]">
                  {formatPrice(quote.day_high, quote.currency)} / {formatPrice(quote.day_low, quote.currency)}
                </td>
              {/if}
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
    {#if data.fetched_at}
      <div class="mt-1 text-end text-[10px] text-[var(--color-text-muted)]">
        {timeAgo(data.fetched_at)}
      </div>
    {/if}
  {/if}
</div>
