<script lang="ts">
  import { Loader2, AlertCircle, Settings, MapPin, Calendar, Clock } from 'lucide-svelte'
  import { SvelteDate, SvelteMap } from 'svelte/reactivity'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { pollingInterval } from '../../stores/polling'
  import { usePolling } from '../../utils/usePolling'
  import { t, translate } from '../../i18n'

  interface CalendarEvent {
    uid: string
    summary: string
    description?: string
    location?: string
    start: string
    end: string
    all_day: boolean
    color?: string
  }

  interface CalendarData {
    events: CalendarEvent[]
  }

  interface Props {
    config?: string
    active?: boolean
  }

  const { config = '{}', active = true }: Props = $props()

  let loading = $state(true)
  let refreshing = $state(false)
  let error = $state<string | null>(null)
  let data = $state<CalendarData | null>(null)

  const parsedConfig = $derived((() => {
    try { return JSON.parse(config) } catch { return {} }
  })())

  const urls = $derived<string[]>(parsedConfig.urls ?? [])
  const days = $derived<number>(parsedConfig.days ?? 14)

  function isToday(dateStr: string): boolean {
    const d = new Date(dateStr)
    const now = new Date()
    return d.getFullYear() === now.getFullYear() &&
      d.getMonth() === now.getMonth() &&
      d.getDate() === now.getDate()
  }

  function isTomorrow(dateStr: string): boolean {
    const d = new Date(dateStr)
    const tomorrow = new SvelteDate()
    tomorrow.setDate(tomorrow.getDate() + 1)
    return d.getFullYear() === tomorrow.getFullYear() &&
      d.getMonth() === tomorrow.getMonth() &&
      d.getDate() === tomorrow.getDate()
  }

  function formatDate(dateStr: string, allDay: boolean): string {
    const d = new Date(dateStr)
    if (isToday(dateStr)) {
      return allDay ? translate('calendar.today') : d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    }
    if (isTomorrow(dateStr)) {
      return allDay ? translate('calendar.tomorrow') : translate('calendar.tomorrow') + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    }
    if (allDay) {
      return d.toLocaleDateString([], { weekday: 'short', day: 'numeric', month: 'short' })
    }
    return d.toLocaleDateString([], { weekday: 'short', day: 'numeric', month: 'short' }) +
      ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  function groupLabel(dateStr: string): string {
    if (isToday(dateStr)) return translate('calendar.today')
    if (isTomorrow(dateStr)) return translate('calendar.tomorrow')
    const d = new Date(dateStr)
    return d.toLocaleDateString([], { weekday: 'long', day: 'numeric', month: 'long' })
  }

  function getDateKey(dateStr: string): string {
    const d = new Date(dateStr)
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
  }

  interface GroupedDay {
    key: string
    label: string
    events: CalendarEvent[]
  }

  const groupedEvents = $derived<GroupedDay[]>((() => {
    if (!data?.events) return []
    const groups = new SvelteMap<string, { label: string; events: CalendarEvent[] }>()
    for (const event of data.events) {
      const key = getDateKey(event.start)
      if (!groups.has(key)) {
        groups.set(key, { label: groupLabel(event.start), events: [] })
      }
      groups.get(key)!.events.push(event)
    }
    return Array.from(groups.entries())
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([key, val]) => ({ key, ...val }))
  })())

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
        days: String(days),
      })
      data = await api.get<CalendarData>(`/api/calendar?${params.toString()}`)
    } catch {
      error = 'fetch_error'
      toasts.error(translate('calendar.fetchError'))
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

<div class="relative h-full" data-testid="calendar-widget">
  {#if refreshing}
    <div class="absolute top-0 end-0 z-10 p-1">
      <Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" />
    </div>
  {/if}
  {#if loading && !data}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('calendar.loading')}
    </div>
  {:else if error === 'not_configured'}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Settings size={24} />
      <p>{$t('calendar.notConfigured')}</p>
      <p class="text-xs">{$t('calendar.configureHint')}</p>
    </div>
  {:else if error === 'fetch_error' && !data}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <AlertCircle size={24} class="text-[var(--color-danger)]" />
      <p>{$t('calendar.fetchError')}</p>
      <button onclick={fetchData} class="text-xs text-[var(--color-primary)] hover:underline">
        {$t('common.retry')}
      </button>
    </div>
  {:else if data}
    {#if !data.events || data.events.length === 0}
      <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
        <Calendar size={24} />
        <p>{$t('calendar.noEvents')}</p>
      </div>
    {:else}
      <div class="space-y-3 overflow-y-auto">
        {#each groupedEvents as group (group.key)}
          <div>
            <div class="mb-1 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">
              {group.label}
            </div>
            <div class="space-y-1">
              {#each group.events as event (event.uid)}
                <div class="flex items-start gap-2 rounded-lg bg-[var(--color-bg-secondary)] px-2.5 py-1.5">
                  <div class="mt-0.5 h-2 w-2 flex-shrink-0 rounded-full {isToday(event.start) ? 'bg-[var(--color-primary)]' : 'bg-[var(--color-text-muted)]'}"></div>
                  <div class="min-w-0 flex-1">
                    <div class="text-sm font-medium text-[var(--color-text)] truncate">{event.summary}</div>
                    <div class="flex flex-wrap items-center gap-x-3 gap-y-0.5 text-xs text-[var(--color-text-muted)]">
                      <span class="flex items-center gap-1">
                        <Clock size={10} />
                        {#if event.all_day}
                          {$t('calendar.allDay')}
                        {:else}
                          {formatDate(event.start, false)}
                        {/if}
                      </span>
                      {#if event.location}
                        <span class="flex items-center gap-1 truncate">
                          <MapPin size={10} />
                          {event.location}
                        </span>
                      {/if}
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  {/if}
</div>
