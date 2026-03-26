<script lang="ts">
  import { onMount } from 'svelte'
  import { Plus, Trash2, Loader2, GripVertical } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { t } from '../../i18n'
  import ResponsiveColumnsConfig from './ResponsiveColumnsConfig.svelte'

  interface DashboardItem {
    uid: string
    title: string
  }

  interface PanelItem {
    id: number
    title: string
    type: string
  }

  interface DashboardDetail {
    uid: string
    title: string
    slug: string
    panels: PanelItem[]
  }

  interface PanelEntry {
    dashboardUid: string
    dashboardTitle: string
    panelId: number
    panelTitle: string
  }

  interface Props {
    value?: string
    onchange?: (value: string) => void
  }

  const { value = '{}', onchange }: Props = $props()

  let dashboards = $state<DashboardItem[]>([])
  let availablePanels = $state<PanelItem[]>([])
  let loading = $state(true)
  let loadingPanels = $state(false)

  let panels = $state<PanelEntry[]>([])
  let columns = $state(2)
  let columnsTablet = $state(0)
  let columnsMobile = $state(0)
  let theme = $state('dark')

  let selectedDashboardUid = $state('')
  let selectedPanelId = $state<number | null>(null)

  // Drag & drop reorder
  let dragIndex = $state<number | null>(null)
  let dragOverIndex = $state<number | null>(null)

  $effect(() => {
    try {
      const parsed = JSON.parse(value)
      if (parsed.panels && Array.isArray(parsed.panels)) panels = parsed.panels
      if (parsed.columns) columns = parsed.columns
      if (typeof parsed.columnsTablet === 'number') columnsTablet = parsed.columnsTablet
      if (typeof parsed.columnsMobile === 'number') columnsMobile = parsed.columnsMobile
      if (parsed.theme) theme = parsed.theme
    } catch { /* ignore */ }
  })

  function emitChange() {
    const config: Record<string, unknown> = { panels, columns, theme }
    if (columnsTablet > 0) config.columnsTablet = columnsTablet
    if (columnsMobile > 0) config.columnsMobile = columnsMobile
    onchange?.(JSON.stringify(config))
  }

  async function loadPanelsForDashboard(uid: string) {
    if (!uid) { availablePanels = []; return }
    loadingPanels = true
    try {
      const detail = await api.get<DashboardDetail>(`/api/grafana/dashboards/${uid}/panels`)
      availablePanels = detail.panels ?? []
    } catch { availablePanels = [] }
    finally { loadingPanels = false }
  }

  function addPanel() {
    if (!selectedDashboardUid || selectedPanelId === null) return
    const dash = dashboards.find(d => d.uid === selectedDashboardUid)
    const panel = availablePanels.find(p => p.id === selectedPanelId)
    if (!dash || !panel) return

    panels = [...panels, {
      dashboardUid: selectedDashboardUid,
      dashboardTitle: dash.title,
      panelId: selectedPanelId,
      panelTitle: panel.title,
    }]
    selectedPanelId = null
    emitChange()
  }

  function removePanel(index: number) {
    panels = panels.filter((_, i) => i !== index)
    emitChange()
  }

  function handleDragStart(index: number) { dragIndex = index }

  function handleDragOver(e: DragEvent, index: number) {
    e.preventDefault()
    dragOverIndex = index
  }

  function handleDrop(index: number) {
    if (dragIndex !== null && dragIndex !== index) {
      const reordered = [...panels]
      const [moved] = reordered.splice(dragIndex, 1)
      reordered.splice(index, 0, moved)
      panels = reordered
      emitChange()
    }
    dragIndex = null
    dragOverIndex = null
  }

  function handleDragEnd() {
    dragIndex = null
    dragOverIndex = null
  }

  onMount(async () => {
    try {
      dashboards = await api.get<DashboardItem[]>('/api/grafana/dashboards') ?? []
    } catch { dashboards = [] }
    finally { loading = false }
  })
</script>

<div class="space-y-3">
  {#if loading}
    <div class="flex items-center gap-2 text-sm text-[var(--color-text-muted)]">
      <Loader2 size={14} class="animate-spin" />
      {$t('common.loading')}
    </div>
  {:else}
    <!-- Add panel -->
    <div class="space-y-2 rounded-lg border border-[var(--color-border)] p-3">
      <span class="text-xs font-medium text-[var(--color-text-secondary)]">{$t('grafana.addPanel')}</span>
      <div class="flex flex-col gap-2">
        <select
          value={selectedDashboardUid}
          onchange={(e) => { selectedDashboardUid = (e.target as HTMLSelectElement).value; selectedPanelId = null; loadPanelsForDashboard(selectedDashboardUid) }}
          class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]"
        >
          <option value="">{$t('grafana.selectDashboard')}</option>
          {#each dashboards as d (d.uid)}
            <option value={d.uid}>{d.title}</option>
          {/each}
        </select>

        {#if selectedDashboardUid}
          {#if loadingPanels}
            <div class="flex items-center gap-1 text-xs text-[var(--color-text-muted)]">
              <Loader2 size={12} class="animate-spin" /> {$t('common.loading')}
            </div>
          {:else if availablePanels.length === 0}
            <p class="text-xs text-[var(--color-text-muted)]">{$t('grafana.noPanels')}</p>
          {:else}
            <select
              value={selectedPanelId ?? ''}
              onchange={(e) => { selectedPanelId = Number((e.target as HTMLSelectElement).value) }}
              class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]"
            >
              <option value="">{$t('grafana.selectPanel')}</option>
              {#each availablePanels as p (p.id)}
                <option value={p.id}>{p.title} ({p.type})</option>
              {/each}
            </select>
          {/if}
        {/if}

        <button
          onclick={addPanel}
          disabled={!selectedDashboardUid || selectedPanelId === null}
          class="flex items-center gap-1 self-start rounded-lg bg-[var(--color-primary)] px-3 py-1.5 text-xs text-white hover:bg-[var(--color-primary-hover)] disabled:opacity-50"
        >
          <Plus size={12} /> {$t('grafana.addPanel')}
        </button>
      </div>
    </div>

    <!-- Panel list -->
    {#if panels.length === 0}
      <p class="text-center text-xs text-[var(--color-text-muted)]">{$t('grafana.noPanel')}</p>
    {:else}
      <div class="space-y-1">
        {#each panels as panel, i (i)}
          <div
            draggable="true"
            role="listitem"
            ondragstart={() => handleDragStart(i)}
            ondragover={(e: DragEvent) => handleDragOver(e, i)}
            ondrop={() => handleDrop(i)}
            ondragend={handleDragEnd}
            class="flex items-center gap-2 rounded-lg border px-3 py-2 transition-colors select-none"
            class:border-[var(--color-primary)]={dragOverIndex === i && dragIndex !== i}
            class:border-[var(--color-border)]={dragOverIndex !== i || dragIndex === i}
            class:opacity-50={dragIndex === i}
          >
            <span class="cursor-grab text-[var(--color-text-muted)]"><GripVertical size={14} /></span>
            <div class="flex-1">
              <span class="text-sm font-medium text-[var(--color-text)]">{panel.panelTitle}</span>
              <span class="ms-2 text-xs text-[var(--color-text-muted)]">{panel.dashboardTitle}</span>
            </div>
            <button
              onclick={() => removePanel(i)}
              class="text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"
            >
              <Trash2 size={14} />
            </button>
          </div>
        {/each}
      </div>
    {/if}

    <!-- Columns -->
    <ResponsiveColumnsConfig
      {columns}
      columnsTablet={columnsTablet || Math.min(columns, 2)}
      columnsMobile={columnsMobile || 1}
      onchange={(d, t, m) => { columns = d; columnsTablet = t; columnsMobile = m; emitChange() }}
    />

    <!-- Theme -->
    <div class="flex items-center gap-2">
      <span class="text-sm text-[var(--color-text-secondary)]">{$t('grafana.theme')}</span>
      {#each [['dark', $t('grafana.themeDark')], ['light', $t('grafana.themeLight')]] as [val, label] (val)}
        <button
          onclick={() => { theme = val; emitChange() }}
          class="rounded-lg px-3 py-1 text-sm transition-colors"
          class:bg-[var(--color-primary)]={theme === val}
          class:text-white={theme === val}
          class:bg-[var(--color-bg-tertiary)]={theme !== val}
          class:text-[var(--color-text-secondary)]={theme !== val}
        >{label}</button>
      {/each}
    </div>

    <!-- Hint -->
    <p class="text-xs text-[var(--color-text-muted)]">{$t('grafana.embedHint')}</p>
  {/if}
</div>
