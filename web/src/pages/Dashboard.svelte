<script lang="ts">
  import { api } from '../lib/api/client'
  import { get } from 'svelte/store'
  import WidgetGrid from '../lib/components/WidgetGrid.svelte'
  import CategoryTabBar from '../lib/components/CategoryTabBar.svelte'
  import { t } from '../lib/i18n'
  import { currentPath, navigate } from '../lib/stores/router'
  import { categories } from '../lib/stores/categories'

  interface Widget {
    id: number
    category_id: number
    type: string
    title: string
    config: string
    pos_x: number
    pos_y: number
    width: number
    height: number
    collapsed: boolean
  }

  let widgets = $state<Record<number, Widget[]>>({})
  let selectedCategoryId = $state<number | null>(null)
  let loading = $state(true)

  // Add widget (per-category)
  const showAddWidgetMap = $state<Record<number, boolean>>({})

  // Ref to CategoryTabBar for triggering create panel
  let categoryTabBar = $state<ReturnType<typeof CategoryTabBar> | undefined>(undefined)

  function getCategoryFromPath(path: string, cats: { id: number; slug: string }[]): number | null {
    // Backward compat: /dashboard/category/3
    const oldMatch = path.match(/\/dashboard\/category\/(\d+)/)
    if (oldMatch) {
      return parseInt(oldMatch[1], 10)
    }
    // New format: /dashboard/{slug}
    const slugMatch = path.match(/\/dashboard\/([^/]+)/)
    if (slugMatch) {
      const cat = cats.find((c) => c.slug === slugMatch[1])
      return cat ? cat.id : null
    }
    return null
  }

  function selectDefaultCategory() {
    const cats = get(categories)
    if (cats.length > 0 && selectedCategoryId === null) {
      const fromUrl = getCategoryFromPath($currentPath, cats)
      const validId = fromUrl && cats.some((c) => c.id === fromUrl) ? fromUrl : null
      selectedCategoryId = validId ?? cats[0].id
      if (!validId) {
        const cat = cats.find((c) => c.id === selectedCategoryId)
        if (cat) navigate(`/dashboard/${cat.slug}`)
      }
    }
  }

  $effect(() => {
    const cats = get(categories)
    const fromUrl = getCategoryFromPath($currentPath, cats)
    if (fromUrl && cats.some((c) => c.id === fromUrl)) {
      selectedCategoryId = fromUrl
    }
  })

  async function loadWidgets(categoryId: number) {
    try {
      widgets = { ...widgets, [categoryId]: await api.get<Widget[]>(`/api/categories/${categoryId}/widgets`) }
    } catch (e) {
      console.error('dashboard: failed to load widgets', categoryId, e)
      widgets = { ...widgets, [categoryId]: [] }
    }
  }

  async function loadAll() {
    await categories.reload()
    selectDefaultCategory()
    const cats = get(categories)
    for (const cat of cats) {
      if (!(cat.id in showAddWidgetMap)) showAddWidgetMap[cat.id] = false
      await loadWidgets(cat.id)
    }
    loading = false
  }

  $effect(() => {
    loadAll()
  })

  async function handleCategoryUpdate() {
    await categories.reload()
    const cats = get(categories)
    if (selectedCategoryId !== null && !cats.some((c) => c.id === selectedCategoryId)) {
      selectedCategoryId = cats.length > 0 ? cats[0].id : null
    }
    selectDefaultCategory()
    for (const cat of cats) {
      if (!(cat.id in showAddWidgetMap)) showAddWidgetMap[cat.id] = false
      await loadWidgets(cat.id)
    }
  }

  function handleWidgetUpdate() {
    if (selectedCategoryId !== null) {
      loadWidgets(selectedCategoryId)
    }
  }

  $effect(() => {
    if (selectedCategoryId !== null && !widgets[selectedCategoryId]) {
      loadWidgets(selectedCategoryId)
    }
  })

  function handleTabSelect(catId: number) {
    selectedCategoryId = catId
  }

  function handleShowAddWidget() {
    if (selectedCategoryId !== null) {
      showAddWidgetMap[selectedCategoryId] = true
    }
  }
</script>

<div data-testid="dashboard-page">
  {#if loading}
    <p class="text-[var(--color-text-secondary)]">{$t('common.loading')}</p>
  {:else}
    <CategoryTabBar
      bind:this={categoryTabBar}
      {selectedCategoryId}
      onSelect={handleTabSelect}
      onCategoryUpdate={handleCategoryUpdate}
      onShowAddWidget={handleShowAddWidget}
    />

    <!-- Widgets -->
    {#if $categories.length > 0}
      {#each $categories as cat (cat.id)}
        <div style={selectedCategoryId === cat.id ? '' : 'display: none'}>
          <WidgetGrid
            categoryId={cat.id}
            widgets={widgets[cat.id] || []}
            active={selectedCategoryId === cat.id}
            onUpdate={handleWidgetUpdate}
            bind:showAddDialog={showAddWidgetMap[cat.id]}
          />
        </div>
      {/each}
    {:else}
      <div class="flex flex-col items-center justify-center py-16 text-center">
        <p class="mb-2 text-[var(--color-text-muted)]">{$t('category.noCategories')}</p>
        <p class="text-sm text-[var(--color-text-muted)]">{$t('category.noCategoriesHint')}</p>
        <button onclick={() => categoryTabBar?.openCreatePanel()}
          class="mt-4 rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white hover:bg-[var(--color-primary-hover)]"
          data-testid="create-first-category">
          {$t('category.createFirst')}
        </button>
      </div>
    {/if}
  {/if}
</div>
