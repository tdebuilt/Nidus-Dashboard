<script lang="ts">
  import { t } from '../../i18n'

  interface Props {
    value?: string
    onchange?: (value: string) => void
  }

  const { value = '{}', onchange }: Props = $props()

  let apiKey = $state('')
  let city = $state('')
  let lat = $state('')
  let lon = $state('')
  let units = $state('metric')

  $effect(() => {
    try {
      const parsed = JSON.parse(value)
      apiKey = parsed.api_key ?? ''
      city = parsed.city ?? ''
      lat = parsed.lat ?? ''
      lon = parsed.lon ?? ''
      units = parsed.units ?? 'metric'
    } catch {
      // ignore
    }
  })

  function emit() {
    const config: Record<string, string> = { api_key: apiKey, units }
    if (city) {
      config.city = city
    } else if (lat && lon) {
      config.lat = lat
      config.lon = lon
    }
    onchange?.(JSON.stringify(config))
  }
</script>

<div class="space-y-3">
  <div>
    <label for="weather-apikey" class="block text-sm text-[var(--color-text-secondary)]">
      {$t('weather.apiKey')}
    </label>
    <input
      id="weather-apikey"
      type="password"
      bind:value={apiKey}
      oninput={emit}
      placeholder="abc123..."
      class="mt-1 w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
    />
    <p class="mt-1 text-xs text-[var(--color-text-muted)]">{$t('weather.apiKeyHint')}</p>
  </div>

  <div>
    <label for="weather-city" class="block text-sm text-[var(--color-text-secondary)]">
      {$t('weather.city')}
    </label>
    <input
      id="weather-city"
      type="text"
      bind:value={city}
      oninput={emit}
      placeholder="Paris,FR"
      class="mt-1 w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
    />
    <p class="mt-1 text-xs text-[var(--color-text-muted)]">{$t('weather.cityHint')}</p>
  </div>

  <div class="flex gap-3">
    <div class="flex-1">
      <label for="weather-lat" class="block text-sm text-[var(--color-text-secondary)]">
        {$t('weather.latitude')}
      </label>
      <input
        id="weather-lat"
        type="text"
        bind:value={lat}
        oninput={emit}
        placeholder="48.8566"
        class="mt-1 w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
      />
    </div>
    <div class="flex-1">
      <label for="weather-lon" class="block text-sm text-[var(--color-text-secondary)]">
        {$t('weather.longitude')}
      </label>
      <input
        id="weather-lon"
        type="text"
        bind:value={lon}
        oninput={emit}
        placeholder="2.3522"
        class="mt-1 w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
      />
    </div>
  </div>

  <div>
    <label for="weather-units" class="block text-sm text-[var(--color-text-secondary)]">
      {$t('weather.units')}
    </label>
    <select
      id="weather-units"
      bind:value={units}
      onchange={emit}
      class="mt-1 w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
    >
      <option value="metric">°C (Celsius)</option>
      <option value="imperial">°F (Fahrenheit)</option>
    </select>
  </div>
</div>
