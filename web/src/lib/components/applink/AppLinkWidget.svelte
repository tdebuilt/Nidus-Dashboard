<script lang="ts">
  import { Loader2, Link } from 'lucide-svelte'
  import { SvelteMap } from 'svelte/reactivity'
  import { api } from '../../api/client'
  import { t } from '../../i18n'
  import { breakpoint } from '../../stores/breakpoint'
  import { getResponsiveColumns } from '../../utils/responsiveColumns'
  import { usePolling } from '../../utils/usePolling'
  import LinkCard from './LinkCard.svelte'

  interface LinkConfig {
    name: string
    url: string
    group?: string
  }

  interface Props {
    config?: string
    active?: boolean
  }

  const { config = '{}', active = true }: Props = $props()

  let loading = $state(true)
  let refreshing = $state(false)
  const healthMap = $state<Record<string, 'up' | 'down' | 'pending'>>({})

  const parsedConfig = $derived((() => {
    try { return JSON.parse(config) } catch { return {} }
  })())

  const links = $derived<LinkConfig[]>(parsedConfig.links ?? [])
  const columns = $derived(getResponsiveColumns(parsedConfig, $breakpoint, 1))
  const sortBy = $derived<string>(parsedConfig.sortBy ?? 'manual')

  function sortLinks(items: LinkConfig[]): LinkConfig[] {
    if (sortBy === 'manual') return items
    const sorted = [...items]
    switch (sortBy) {
      case 'name-asc':
        return sorted.sort((a, b) => a.name.localeCompare(b.name))
      case 'name-desc':
        return sorted.sort((a, b) => b.name.localeCompare(a.name))
      case 'status-up':
        return sorted.sort((a, b) => {
          const sa = healthMap[a.url] ?? 'pending'
          const sb = healthMap[b.url] ?? 'pending'
          const order = { up: 0, pending: 1, down: 2 }
          return (order[sa] ?? 1) - (order[sb] ?? 1)
        })
      case 'status-down':
        return sorted.sort((a, b) => {
          const sa = healthMap[a.url] ?? 'pending'
          const sb = healthMap[b.url] ?? 'pending'
          const order = { down: 0, pending: 1, up: 2 }
          return (order[sa] ?? 1) - (order[sb] ?? 1)
        })
      default:
        return items
    }
  }

  interface GroupedLinks {
    name: string
    links: LinkConfig[]
  }

  const groupedLinks = $derived<GroupedLinks[]>((() => {
    if (links.length === 0) return []

    const groups = new SvelteMap<string, LinkConfig[]>()
    const order: string[] = []

    for (const link of links) {
      const group = link.group || ''
      if (!groups.has(group)) {
        groups.set(group, [])
        order.push(group)
      }
      groups.get(group)!.push(link)
    }

    return order.map((name) => ({ name, links: sortLinks(groups.get(name)!) }))
  })())

  const hasGroups = $derived(groupedLinks.some((g) => g.name !== ''))

  async function checkHealth(link: LinkConfig) {
    try {
      const data = await api.get<{ status: string }>(`/api/applinks/health?url=${encodeURIComponent(link.url)}`)
      healthMap[link.url] = data.status === 'up' ? 'up' : 'down'
    } catch {
      healthMap[link.url] = 'down'
    }
  }

  async function fetchAllHealth() {
    if (links.length === 0) {
      loading = false
      return
    }
    await Promise.all(links.map(checkHealth))
    loading = false
  }

  async function fetchAllHealthWrapped() {
    const hadData = Object.values(healthMap).some((v) => v !== 'pending')
    if (hadData) refreshing = true
    for (const link of links) {
      if (!healthMap[link.url]) healthMap[link.url] = 'pending'
    }
    await fetchAllHealth()
    refreshing = false
  }

  const polling = usePolling({
    fetchFn: fetchAllHealthWrapped,
    active: () => active,
    fixedIntervalMs: 60000,
  })

  $effect(() => {
    if (active) polling.start(); else polling.stop()
    return () => polling.stop()
  })
</script>

<div class="relative h-full" data-testid="applink-widget">
  {#if refreshing}
    <div class="absolute end-1 top-1 z-10">
      <Loader2 size={12} class="animate-spin text-[var(--color-text-muted)]" />
    </div>
  {/if}
  {#if loading && links.length > 0}
    <div class="flex h-full items-center justify-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={16} class="animate-spin" />
      {$t('applink.loading')}
    </div>
  {:else if links.length === 0}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Link size={24} />
      <p>{$t('applink.noLinks')}</p>
      <p class="text-xs">{$t('applink.configureHint')}</p>
    </div>
  {:else}
    <div class="space-y-3 overflow-y-auto">
      {#each groupedLinks as group (group.name)}
        {#if hasGroups && group.name}
          <div class="text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">
            {group.name}
          </div>
        {/if}
        <div class="grid gap-1.5" style="grid-template-columns: repeat({columns}, 1fr);">
          {#each group.links as link (link.url)}
            <LinkCard
              name={link.name}
              url={link.url}
              status={healthMap[link.url] ?? 'pending'}
            />
          {/each}
        </div>
      {/each}
    </div>
  {/if}
</div>
