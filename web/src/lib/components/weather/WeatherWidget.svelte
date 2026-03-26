<script lang="ts">
  import { Loader2, AlertCircle, Settings, Droplets, Wind, Thermometer } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { pollingInterval } from '../../stores/polling'
  import { usePolling } from '../../utils/usePolling'
  import { t, translate } from '../../i18n'

  interface CurrentWeather {
    temp: number
    feels_like: number
    temp_min: number
    temp_max: number
    humidity: number
    pressure: number
    wind_speed: number
    wind_deg: number
    description: string
    icon: string
    city: string
    country: string
    sunrise: number
    sunset: number
  }

  interface ForecastDay {
    date: string
    temp_min: number
    temp_max: number
    description: string
    icon: string
    humidity: number
    wind_speed: number
  }

  interface WeatherData {
    current: CurrentWeather
    forecast: ForecastDay[]
  }

  interface Props {
    config?: string
    active?: boolean
  }

  const { config = '{}', active = true }: Props = $props()

  let loading = $state(true)
  let refreshing = $state(false)
  let error = $state<string | null>(null)
  let data = $state<WeatherData | null>(null)

  const parsedConfig = $derived((() => {
    try { return JSON.parse(config) } catch { return {} }
  })())

  const apiKey = $derived(parsedConfig.api_key ?? '')
  const city = $derived(parsedConfig.city ?? '')
  const lat = $derived(parsedConfig.lat ?? '')
  const lon = $derived(parsedConfig.lon ?? '')
  const units = $derived(parsedConfig.units ?? 'metric')
  const lang = $derived(parsedConfig.lang ?? 'fr')

  function unitSymbol(): string {
    return units === 'imperial' ? '°F' : '°C'
  }

  function windUnit(): string {
    return units === 'imperial' ? 'mph' : 'm/s'
  }

  function iconUrl(code: string): string {
    if (!code) return ''
    return `https://openweathermap.org/img/wn/${code}@2x.png`
  }

  function formatDay(dateStr: string): string {
    const date = new Date(dateStr)
    return date.toLocaleDateString(lang === 'fr' ? 'fr-FR' : 'en-US', { weekday: 'short' })
  }

  async function fetchData() {
    if (!apiKey) {
      error = 'not_configured'
      loading = false
      return
    }
    if (!city && (!lat || !lon)) {
      error = 'not_configured'
      loading = false
      return
    }

    if (data === null) loading = true
    error = null
    try {
      const params = new URLSearchParams({ apikey: apiKey, units, lang })
      if (city) {
        params.set('city', city)
      } else {
        params.set('lat', lat)
        params.set('lon', lon)
      }
      data = await api.get<WeatherData>(`/api/weather?${params.toString()}`)
    } catch (err: unknown) {
      const status = (err as { status?: number })?.status
      if (status === 502) {
        error = 'fetch_error'
      } else {
        error = 'fetch_error'
        toasts.error(translate('weather.fetchError'))
      }
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

<div class="relative h-full" data-testid="weather-widget">
  {#if refreshing}
    <div class="absolute top-0 end-0 z-10 p-1">
      <Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" />
    </div>
  {/if}
  {#if loading && !data}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('weather.loading')}
    </div>
  {:else if error === 'not_configured'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Settings size={24} />
      <p>{$t('weather.notConfigured')}</p>
      <p class="text-xs">{$t('weather.configureHint')}</p>
    </div>
  {:else if error === 'fetch_error' && !data}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-danger)]" />
      <p>{$t('weather.fetchError')}</p>
      <button onclick={fetchData} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if data}
    <div class="space-y-3">
      <!-- Current weather -->
      <div class="flex items-center gap-3">
        {#if data.current.icon}
          <img src={iconUrl(data.current.icon)} alt={data.current.description} class="h-14 w-14 -my-2" />
        {/if}
        <div class="flex-1 min-w-0">
          <div class="flex items-baseline gap-2">
            <span class="text-2xl font-bold text-[var(--color-text)]">{Math.round(data.current.temp)}{unitSymbol()}</span>
            <span class="text-xs text-[var(--color-text-muted)] capitalize">{data.current.description}</span>
          </div>
          <div class="text-xs text-[var(--color-text-muted)]">
            {data.current.city}{#if data.current.country}, {data.current.country}{/if}
          </div>
        </div>
      </div>

      <!-- Details -->
      <div class="flex items-center gap-4 text-xs text-[var(--color-text-muted)]">
        <div class="flex items-center gap-1" title={$t('weather.feelsLike')}>
          <Thermometer size={12} />
          <span>{Math.round(data.current.feels_like)}{unitSymbol()}</span>
        </div>
        <div class="flex items-center gap-1" title={$t('weather.humidity')}>
          <Droplets size={12} />
          <span>{data.current.humidity}%</span>
        </div>
        <div class="flex items-center gap-1" title={$t('weather.wind')}>
          <Wind size={12} />
          <span>{data.current.wind_speed.toFixed(1)} {windUnit()}</span>
        </div>
        <div class="text-[var(--color-text-muted)]">
          {Math.round(data.current.temp_min)}° / {Math.round(data.current.temp_max)}°
        </div>
      </div>

      <!-- Forecast -->
      {#if data.forecast && data.forecast.length > 0}
        <div class="flex flex-wrap gap-1">
          {#each data.forecast.slice(0, 5) as day (day.date)}
            <div class="flex flex-1 flex-col items-center gap-0.5 rounded-lg bg-[var(--color-bg-secondary)] px-2 py-1.5 text-center">
              <span class="text-[10px] font-medium text-[var(--color-text-muted)]">{formatDay(day.date)}</span>
              {#if day.icon}
                <img src={iconUrl(day.icon)} alt={day.description} class="h-8 w-8 -my-1" />
              {/if}
              <div class="text-[10px] text-[var(--color-text-muted)]">
                <span class="font-medium text-[var(--color-text)]">{Math.round(day.temp_max)}°</span>
                <span>{Math.round(day.temp_min)}°</span>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>
