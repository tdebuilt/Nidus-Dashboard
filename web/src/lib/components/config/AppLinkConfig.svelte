<script lang="ts">
  import { Plus, Trash2, GripVertical } from 'lucide-svelte'
  import { t } from '../../i18n'
  import ResponsiveColumnsConfig from './ResponsiveColumnsConfig.svelte'

  interface LinkEntry {
    name: string
    url: string
    group: string
  }

  interface Props {
    value?: string
    onchange?: (value: string) => void
  }

  const { value = '{}', onchange }: Props = $props()

  let links = $state<LinkEntry[]>([])
  let columns = $state(1)
  let columnsTablet = $state(0)
  let columnsMobile = $state(0)
  let sortBy = $state('manual')
  let fillMode = $state<'row' | 'column'>('row')
  let healthCheckInterval = $state(5)
  let dragIndex = $state<number | null>(null)
  let dragOverIndex = $state<number | null>(null)

  // Collect unique group names for datalist suggestions
  const groupNames = $derived([...new Set(links.map((l) => l.group).filter(Boolean))])

  $effect(() => {
    try {
      const parsed = JSON.parse(value)
      if (parsed.columns) columns = parsed.columns
      if (typeof parsed.columnsTablet === 'number') columnsTablet = parsed.columnsTablet
      if (typeof parsed.columnsMobile === 'number') columnsMobile = parsed.columnsMobile
      if (parsed.sortBy) sortBy = parsed.sortBy
      if (parsed.fillMode) fillMode = parsed.fillMode
      if (typeof parsed.healthCheckInterval === 'number') healthCheckInterval = parsed.healthCheckInterval
      if (parsed.links && Array.isArray(parsed.links)) {
        links = parsed.links.map((l: Record<string, string>) => ({
          name: l.name ?? '',
          url: l.url ?? '',
          group: l.group ?? '',
        }))
      }
    } catch {
      // ignore
    }
  })

  function emitChange() {
    const validLinks = links
      .filter((l) => l.name.trim() && l.url.trim())
      .map((l) => {
        const entry: Record<string, string> = { name: l.name.trim(), url: l.url.trim() }
        if (l.group.trim()) entry.group = l.group.trim()
        return entry
      })
    const config: Record<string, unknown> = { links: validLinks, columns, sortBy, fillMode }
    if (columnsTablet > 0) config.columnsTablet = columnsTablet
    if (columnsMobile > 0) config.columnsMobile = columnsMobile
    if (healthCheckInterval !== 5) config.healthCheckInterval = healthCheckInterval
    onchange?.(JSON.stringify(config))
  }

  function addLink() {
    // Pre-fill group from last link if available
    const lastGroup = links.length > 0 ? links[links.length - 1].group : ''
    links = [...links, { name: '', url: '', group: lastGroup }]
  }

  function removeLink(index: number) {
    links = links.filter((_, i) => i !== index)
    emitChange()
  }

  function updateLink(index: number, field: keyof LinkEntry, val: string) {
    links[index] = { ...links[index], [field]: val }
    emitChange()
  }

  function handleDragStart(index: number) {
    dragIndex = index
  }

  function handleDragOver(e: DragEvent, index: number) {
    e.preventDefault()
    dragOverIndex = index
  }

  function handleDrop(e: DragEvent, targetIndex: number) {
    e.preventDefault()
    if (dragIndex === null || dragIndex === targetIndex) {
      dragIndex = null
      dragOverIndex = null
      return
    }
    const reordered = [...links]
    const [moved] = reordered.splice(dragIndex, 1)
    reordered.splice(targetIndex, 0, moved)
    links = reordered
    dragIndex = null
    dragOverIndex = null
    emitChange()
  }

  function handleDragEnd() {
    dragIndex = null
    dragOverIndex = null
  }
</script>

<div class="space-y-3">
  <!-- Display options -->
  <div class="space-y-3">
    <ResponsiveColumnsConfig
      {columns}
      columnsTablet={columnsTablet || Math.min(columns, 2)}
      columnsMobile={columnsMobile || 1}
      onchange={(d, t, m) => { columns = d; columnsTablet = t; columnsMobile = m; emitChange() }}
    />
    <div class="flex items-center gap-2">
      <label for="applink-sort-select" class="shrink-0 text-sm text-[var(--color-text-secondary)]">{$t('applink.sortBy')}</label>
      <select
        id="applink-sort-select"
        value={sortBy}
        onchange={(e) => { sortBy = (e.target as HTMLSelectElement).value; emitChange() }}
        class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1 text-sm text-[var(--color-text)]"
      >
        <option value="manual">{$t('applink.sortManual')}</option>
        <option value="name-asc">{$t('applink.sortNameAsc')}</option>
        <option value="name-desc">{$t('applink.sortNameDesc')}</option>
        <option value="status-up">{$t('applink.sortStatusUp')}</option>
        <option value="status-down">{$t('applink.sortStatusDown')}</option>
      </select>
    </div>
    <div class="flex items-center gap-2">
      <label for="applink-fill-select" class="shrink-0 text-sm text-[var(--color-text-secondary)]">{$t('applink.fillMode')}</label>
      <select
        id="applink-fill-select"
        value={fillMode}
        onchange={(e) => { fillMode = (e.target as HTMLSelectElement).value as 'row' | 'column'; emitChange() }}
        class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1 text-sm text-[var(--color-text)]"
      >
        <option value="row">{$t('applink.fillByRow')}</option>
        <option value="column">{$t('applink.fillByColumn')}</option>
      </select>
    </div>
    <div class="flex items-center gap-2">
      <label for="applink-health-interval-select" class="shrink-0 text-sm text-[var(--color-text-secondary)]">{$t('applink.healthCheckInterval')}</label>
      <select
        id="applink-health-interval-select"
        value={healthCheckInterval}
        onchange={(e) => { healthCheckInterval = Number((e.target as HTMLSelectElement).value); emitChange() }}
        class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1 text-sm text-[var(--color-text)]"
      >
        <option value={0}>{$t('applink.healthCheckDisabled')}</option>
        <option value={1}>{$t('applink.healthCheckMinutes', { count: 1 })}</option>
        <option value={5}>{$t('applink.healthCheckMinutes', { count: 5 })}</option>
        <option value={15}>{$t('applink.healthCheckMinutes', { count: 15 })}</option>
        <option value={30}>{$t('applink.healthCheckMinutes', { count: 30 })}</option>
      </select>
    </div>
  </div>

  <span class="block text-sm text-[var(--color-text-secondary)]">{$t('config.links')}</span>

  {#each links as link, i (i)}
    <div
      class="flex gap-2 rounded-lg border p-2 transition-colors"
      class:border-[var(--color-primary)]={dragOverIndex === i}
      class:border-[var(--color-border)]={dragOverIndex !== i}
      class:opacity-50={dragIndex === i}
      draggable="true"
      ondragstart={() => handleDragStart(i)}
      ondragover={(e) => handleDragOver(e, i)}
      ondrop={(e) => handleDrop(e, i)}
      ondragend={handleDragEnd}
      role="listitem"
    >
      <div class="flex shrink-0 cursor-grab items-center text-[var(--color-text-muted)]">
        <GripVertical size={14} />
      </div>
      <div class="flex flex-1 flex-col gap-1.5">
        <div class="flex gap-1.5">
          <input
            type="text"
            value={link.name}
            oninput={(e) => updateLink(i, 'name', (e.target as HTMLInputElement).value)}
            placeholder={$t('config.linkName')}
            class="flex-1 rounded border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
          />
          <input
            type="text"
            value={link.group}
            oninput={(e) => updateLink(i, 'group', (e.target as HTMLInputElement).value)}
            placeholder={$t('applink.groupPlaceholder')}
            list="applink-groups"
            class="w-28 rounded border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
          />
        </div>
        <input
          type="url"
          value={link.url}
          oninput={(e) => updateLink(i, 'url', (e.target as HTMLInputElement).value)}
          placeholder={$t('config.linkUrl')}
          class="w-full rounded border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
        />
      </div>
      <button
        onclick={() => removeLink(i)}
        class="shrink-0 self-start rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"
        title={$t('config.removeLink')}
      >
        <Trash2 size={14} />
      </button>
    </div>
  {/each}

  <!-- Datalist for group autocomplete -->
  <datalist id="applink-groups">
    {#each groupNames as group (group)}
      <option value={group}></option>
    {/each}
  </datalist>

  <button
    onclick={addLink}
    class="flex w-full items-center justify-center gap-1 rounded-lg border border-dashed border-[var(--color-border)] py-2 text-sm text-[var(--color-text-muted)] transition-colors hover:border-[var(--color-primary)] hover:text-[var(--color-primary)]"
  >
    <Plus size={14} /> {$t('config.addLink')}
  </button>
</div>
