<script lang="ts">
  import { onMount } from 'svelte'
  import { Loader2, X, GripVertical } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { t } from '../../i18n'
  import ResponsiveColumnsConfig from './ResponsiveColumnsConfig.svelte'

  interface EntityInfo {
    entity_id: string
    domain: string
    name: string
  }

  interface Props {
    value?: string
    onchange?: (value: string) => void
  }

  const { value = '{}', onchange }: Props = $props()

  let allEntities = $state<EntityInfo[]>([])
  let loading = $state(true)
  let selectedEntities = $state<string[]>([])
  let selectedDomains = $state<string[]>([])
  let filterMode = $state<'entities' | 'domains'>('entities')
  let entitySearch = $state('')
  let entitySize = $state<'sm' | 'md' | 'lg'>('md')
  let columns = $state(1)
  let columnsTablet = $state(0)
  let columnsMobile = $state(0)

  $effect(() => {
    try {
      const parsed = JSON.parse(value)
      if (parsed.entities && Array.isArray(parsed.entities)) {
        selectedEntities = parsed.entities
        filterMode = 'entities'
      } else if (parsed.domains && Array.isArray(parsed.domains)) {
        selectedDomains = parsed.domains
        filterMode = 'domains'
      }
      if (parsed.entitySize === 'sm' || parsed.entitySize === 'lg') {
        entitySize = parsed.entitySize
      }
      if (typeof parsed.columns === 'number') {
        columns = parsed.columns
      }
      if (typeof parsed.columnsTablet === 'number') columnsTablet = parsed.columnsTablet
      if (typeof parsed.columnsMobile === 'number') columnsMobile = parsed.columnsMobile
    } catch {
      // ignore
    }
  })

  const availableDomains = $derived(
    [...new Set(allEntities.map((e) => e.domain))].sort()
  )

  const filteredEntities = $derived(
    entitySearch
      ? allEntities.filter(
          (e) =>
            e.entity_id.toLowerCase().includes(entitySearch.toLowerCase()) ||
            e.name.toLowerCase().includes(entitySearch.toLowerCase()),
        )
      : allEntities
  )

  function emitChange() {
    // Preserve fields set by the widget itself (e.g. cameraSizes)
    let existing: Record<string, unknown> = {}
    try { existing = JSON.parse(value) } catch { /* ignore */ }

    const config: Record<string, unknown> = {}
    if (existing.cameraSizes) {
      config.cameraSizes = existing.cameraSizes
    }
    if (filterMode === 'entities' && selectedEntities.length > 0) {
      config.entities = selectedEntities
    } else if (filterMode === 'domains' && selectedDomains.length > 0) {
      config.domains = selectedDomains
    }
    if (entitySize !== 'md') {
      config.entitySize = entitySize
    }
    if (columns > 1) {
      config.columns = columns
    }
    if (columnsTablet > 0) config.columnsTablet = columnsTablet
    if (columnsMobile > 0) config.columnsMobile = columnsMobile
    onchange?.(Object.keys(config).length > 0 ? JSON.stringify(config) : '{}')
  }

  function setEntitySize(s: 'sm' | 'md' | 'lg') {
    entitySize = s
    emitChange()
  }

  function toggleEntity(entityId: string) {
    if (selectedEntities.includes(entityId)) {
      selectedEntities = selectedEntities.filter((e) => e !== entityId)
    } else {
      selectedEntities = [...selectedEntities, entityId]
    }
    emitChange()
  }

  function toggleDomain(domain: string) {
    if (selectedDomains.includes(domain)) {
      selectedDomains = selectedDomains.filter((d) => d !== domain)
    } else {
      selectedDomains = [...selectedDomains, domain]
    }
    emitChange()
  }

  let dragIndex = $state<number | null>(null)
  let dragOverIndex = $state<number | null>(null)

  function handleDragStart(index: number) {
    dragIndex = index
  }

  function handleDragOver(e: DragEvent, index: number) {
    e.preventDefault()
    dragOverIndex = index
  }

  function handleDrop(index: number) {
    if (dragIndex !== null && dragIndex !== index) {
      const reordered = [...selectedEntities]
      const [moved] = reordered.splice(dragIndex, 1)
      reordered.splice(index, 0, moved)
      selectedEntities = reordered
      emitChange()
    }
    dragIndex = null
    dragOverIndex = null
  }

  function handleDragEnd() {
    dragIndex = null
    dragOverIndex = null
  }

  function switchMode(mode: 'entities' | 'domains') {
    filterMode = mode
    selectedEntities = []
    selectedDomains = []
    emitChange()
  }

  onMount(async () => {
    try {
      const data = await api.get<EntityInfo[]>('/api/homeassistant/entities')
      allEntities = data ?? []
    } catch {
      allEntities = []
    } finally {
      loading = false
    }
  })
</script>

<div class="space-y-3">
  {#if loading}
    <div class="flex items-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={14} class="animate-spin" />
      {$t('common.loading')}
    </div>
  {:else}
    <div class="flex gap-2">
      <button
        onclick={() => switchMode('entities')}
        class="rounded-lg px-3 py-1 text-sm transition-colors"
        class:bg-[var(--color-primary)]={filterMode === 'entities'}
        class:text-white={filterMode === 'entities'}
        class:bg-[var(--color-bg-tertiary)]={filterMode !== 'entities'}
        class:text-[var(--color-text-secondary)]={filterMode !== 'entities'}
      >{$t('config.byEntity')}</button>
      <button
        onclick={() => switchMode('domains')}
        class="rounded-lg px-3 py-1 text-sm transition-colors"
        class:bg-[var(--color-primary)]={filterMode === 'domains'}
        class:text-white={filterMode === 'domains'}
        class:bg-[var(--color-bg-tertiary)]={filterMode !== 'domains'}
        class:text-[var(--color-text-secondary)]={filterMode !== 'domains'}
      >{$t('config.byDomain')}</button>
    </div>

    <ResponsiveColumnsConfig
      {columns}
      columnsTablet={columnsTablet || Math.min(columns, 2)}
      columnsMobile={columnsMobile || 1}
      onchange={(d, t, m) => { columns = d; columnsTablet = t; columnsMobile = m; emitChange() }}
    />

    <div class="flex items-center gap-2">
      <span class="text-sm text-[var(--color-text-secondary)]">{$t('config.entitySize')}</span>
      {#each [['sm', $t('config.entitySizeSm')], ['md', $t('config.entitySizeMd')], ['lg', $t('config.entitySizeLg')]] as [size, label] (size)}
        <button
          onclick={() => setEntitySize(size as 'sm' | 'md' | 'lg')}
          class="rounded-lg px-3 py-1 text-sm transition-colors"
          class:bg-[var(--color-primary)]={entitySize === size}
          class:text-white={entitySize === size}
          class:bg-[var(--color-bg-tertiary)]={entitySize !== size}
          class:text-[var(--color-text-secondary)]={entitySize !== size}
        >{label}</button>
      {/each}
    </div>

    {#if filterMode === 'entities'}
      <input
        type="text"
        bind:value={entitySearch}
        placeholder={$t('config.searchEntities')}
        class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
      />
      {#if selectedEntities.length > 0}
        <p class="text-xs text-[var(--color-text-muted)]">{$t('config.reorderHint')}</p>
        <div class="flex flex-col gap-1">
          {#each selectedEntities as entityId, i (entityId)}
            <div
              draggable="true"
              role="listitem"
              ondragstart={() => handleDragStart(i)}
              ondragover={(e: DragEvent) => handleDragOver(e, i)}
              ondrop={() => handleDrop(i)}
              ondragend={handleDragEnd}
              class="flex items-center gap-1 rounded-lg border px-2 py-1 text-xs transition-colors select-none"
              class:border-[var(--color-primary)]={dragOverIndex === i && dragIndex !== i}
              class:border-transparent={dragOverIndex !== i || dragIndex === i}
              class:opacity-50={dragIndex === i}
              class:bg-[var(--color-bg-tertiary)]={true}
            >
              <span class="cursor-grab text-[var(--color-text-muted)]"><GripVertical size={12} /></span>
              <span class="flex-1 truncate text-[var(--color-text)]">{entityId}</span>
              <button onclick={() => toggleEntity(entityId)} class="text-[var(--color-text-muted)] hover:text-[var(--color-danger)]">
                <X size={12} />
              </button>
            </div>
          {/each}
        </div>
      {/if}
      <div class="max-h-40 overflow-y-auto rounded-lg border border-[var(--color-border)]">
        {#each filteredEntities.slice(0, 50) as entity (entity.entity_id)}
          <button
            onclick={() => toggleEntity(entity.entity_id)}
            class="flex w-full items-center gap-2 px-3 py-1.5 text-start text-sm transition-colors hover:bg-[var(--color-bg-tertiary)]"
            class:bg-[var(--color-bg-tertiary)]={selectedEntities.includes(entity.entity_id)}
          >
            <span class="text-[var(--color-text)]">{entity.name || entity.entity_id}</span>
            <span class="text-xs text-[var(--color-text-muted)]">{entity.entity_id}</span>
          </button>
        {/each}
      </div>
    {:else}
      <div class="flex flex-wrap gap-2">
        {#each availableDomains as domain (domain)}
          <button
            onclick={() => toggleDomain(domain)}
            class="rounded-lg border px-3 py-1.5 text-sm transition-colors"
            class:border-[var(--color-primary)]={selectedDomains.includes(domain)}
            class:bg-[var(--color-primary)]={selectedDomains.includes(domain)}
            class:text-white={selectedDomains.includes(domain)}
            class:border-[var(--color-border)]={!selectedDomains.includes(domain)}
            class:text-[var(--color-text)]={!selectedDomains.includes(domain)}
          >
            {domain}
          </button>
        {/each}
      </div>
    {/if}
  {/if}
</div>
