<script lang="ts">
  import { X, Search, Loader2 } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { t, translate } from '../../i18n'
  import { focusTrap } from '../../actions/focusTrap'
  import SearchResultCard from './SearchResultCard.svelte'

  interface QualityProfile { id: number; name: string }
  interface RootFolder { id: number; path: string; freeSpace: number }
  interface LookupResult {
    title: string; year: number; tmdbId?: number; tvdbId?: number
    overview: string; runtime?: number; seasonCount?: number; id?: number
  }

  interface Props {
    open: boolean
    serviceType: 'radarr' | 'sonarr'
    onClose: () => void
    onAdded?: () => void
  }

  const { open, serviceType, onClose, onAdded }: Props = $props()

  let searchTerm = $state('')
  let searchResults = $state<LookupResult[]>([])
  let searching = $state(false)
  let profiles = $state<QualityProfile[]>([])
  let folders = $state<RootFolder[]>([])
  let selectedProfileId = $state(0)
  let selectedFolder = $state('')
  let searchOnAdd = $state(true)
  let loadingOptions = $state(true)
  let submittingId = $state<number | null>(null)
  let searchTimer: ReturnType<typeof setTimeout> | null = null

  const title = $derived(serviceType === 'radarr' ? $t('arr.addMovie') : $t('arr.addSeries'))
  const placeholder = $derived(
    serviceType === 'radarr' ? $t('arr.searchMoviePlaceholder') : $t('arr.searchSeriesPlaceholder')
  )

  function formatSize(bytes: number): string {
    if (bytes >= 1099511627776) return (bytes / 1099511627776).toFixed(1) + ' TB'
    if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(0) + ' GB'
    return (bytes / 1048576).toFixed(0) + ' MB'
  }

  async function loadOptions() {
    loadingOptions = true
    try {
      const [p, f] = await Promise.all([
        api.get<QualityProfile[]>(`/api/arr/${serviceType}/qualityprofiles`),
        api.get<RootFolder[]>(`/api/arr/${serviceType}/rootfolders`),
      ])
      profiles = p
      folders = f
      if (p.length > 0) selectedProfileId = p[0].id
      if (f.length > 0) selectedFolder = f[0].path
    } catch {
      toasts.error(translate('arr.addError'))
    } finally {
      loadingOptions = false
    }
  }

  function handleSearchInput() {
    if (searchTimer) clearTimeout(searchTimer)
    if (searchTerm.trim().length < 2) {
      searchResults = []
      return
    }
    searchTimer = setTimeout(doSearch, 400)
  }

  async function doSearch() {
    const term = searchTerm.trim()
    if (term.length < 2) return
    searching = true
    try {
      searchResults = await api.get<LookupResult[]>(
        `/api/arr/${serviceType}/lookup?term=${encodeURIComponent(term)}`
      )
    } catch {
      searchResults = []
    } finally {
      searching = false
    }
  }

  function buildAddBody(result: LookupResult) {
    const base = {
      title: result.title, year: result.year,
      qualityProfileId: selectedProfileId, rootFolderPath: selectedFolder, monitored: true,
    }
    if (serviceType === 'radarr') {
      return { ...base, tmdbId: result.tmdbId, addOptions: { searchForMovie: searchOnAdd } }
    }
    return { ...base, tvdbId: result.tvdbId, seasonFolder: true, addOptions: { searchForMissingEpisodes: searchOnAdd } }
  }

  function getAddErrorKey(err: unknown): string {
    const msg = (err as { message?: string })?.message ?? ''
    return msg.includes('already') || msg.includes('exists') ? 'arr.duplicateError' : 'arr.addError'
  }

  async function handleAdd(result: LookupResult) {
    if (submittingId !== null) return
    submittingId = (serviceType === 'radarr' ? result.tmdbId : result.tvdbId) ?? 0
    try {
      await api.post(`/api/arr/${serviceType}/add`, buildAddBody(result))
      toasts.success(translate(serviceType === 'radarr' ? 'arr.movieAdded' : 'arr.seriesAdded'))
      searchResults = searchResults.map(r => r === result ? { ...r, id: 1 } : r)
      onAdded?.()
    } catch (err: unknown) {
      toasts.error(translate(getAddErrorKey(err)))
    } finally {
      submittingId = null
    }
  }

  function reset() {
    searchTerm = ''
    searchResults = []
    searching = false
    submittingId = null
  }

  function handleClose() {
    reset()
    onClose()
  }

  $effect(() => {
    if (open) loadOptions()
  })
</script>

{#if open}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onclick={handleClose} onkeydown={(e) => { if (e.key === 'Escape') handleClose() }}>
    <div
      class="mx-4 flex w-full max-w-lg flex-col rounded-xl bg-[var(--color-bg-secondary)] p-6 shadow-xl animate-[dialogIn_0.2s_ease-out]"
      style="max-height: 80vh"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
      role="dialog" aria-modal="true" tabindex="-1"
      use:focusTrap={{ onClose: handleClose }}
    >
      <!-- Header -->
      <div class="mb-4 flex items-center justify-between">
        <h2 class="text-lg font-semibold text-[var(--color-text)]">{title}</h2>
        <button onclick={handleClose} class="rounded p-2 text-[var(--color-text-muted)] hover:text-[var(--color-text)]" aria-label={$t('common.close')}>
          <X size={16} />
        </button>
      </div>

      <!-- Search -->
      <div class="relative mb-3">
        <Search size={14} class="absolute start-3 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" />
        <input
          type="text"
          bind:value={searchTerm}
          oninput={handleSearchInput}
          {placeholder}
          class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] py-2 pe-3 ps-9 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)] focus:outline-none"
        />
        {#if searching}
          <Loader2 size={14} class="absolute end-3 top-1/2 -translate-y-1/2 animate-spin text-[var(--color-text-muted)]" />
        {/if}
      </div>

      <!-- Results -->
      <div class="mb-3 min-h-0 flex-1 overflow-y-auto">
        {#if searchResults.length > 0}
          <div class="space-y-1.5">
            {#each searchResults as result (result.tmdbId ?? result.tvdbId ?? result.title)}
              <SearchResultCard
                {result}
                {serviceType}
                submitting={submittingId === (result.tmdbId ?? result.tvdbId ?? 0)}
                onAdd={() => handleAdd(result)}
              />
            {/each}
          </div>
        {:else if searchTerm.trim().length >= 2 && !searching}
          <div class="py-4 text-center text-xs text-[var(--color-text-muted)]">{$t('arr.noSearchResults')}</div>
        {/if}
      </div>

      <!-- Options -->
      {#if loadingOptions}
        <div class="flex items-center justify-center gap-2 py-2 text-xs text-[var(--color-text-muted)]">
          <Loader2 size={12} class="animate-spin" />
          {$t('arr.loadingProfiles')}
        </div>
      {:else}
        <div class="space-y-2 border-t border-[var(--color-border)] pt-3">
          <div class="flex gap-2">
            <label class="flex-1">
              <span class="mb-1 block text-xs text-[var(--color-text-muted)]">{$t('arr.qualityProfile')}</span>
              <select bind:value={selectedProfileId} class="w-full rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-xs text-[var(--color-text)]">
                {#each profiles as p (p.id)}
                  <option value={p.id}>{p.name}</option>
                {/each}
              </select>
            </label>
            <label class="flex-1">
              <span class="mb-1 block text-xs text-[var(--color-text-muted)]">{$t('arr.rootFolder')}</span>
              <select bind:value={selectedFolder} class="w-full rounded-md border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-xs text-[var(--color-text)]">
                {#each folders as f (f.id)}
                  <option value={f.path}>{f.path} ({formatSize(f.freeSpace)} {$t('arr.freeSpace')})</option>
                {/each}
              </select>
            </label>
          </div>
          <label class="flex items-center gap-2 text-xs text-[var(--color-text)]">
            <input type="checkbox" bind:checked={searchOnAdd} class="rounded" />
            {$t('arr.searchOnAdd')}
          </label>
        </div>
      {/if}

      <!-- Footer -->
      <div class="mt-4 flex justify-end">
        <button onclick={handleClose} class="rounded-lg border border-[var(--color-border)] px-4 py-2 text-sm text-[var(--color-text)] hover:bg-[var(--color-bg)]">
          {$t('common.close')}
        </button>
      </div>
    </div>
  </div>
{/if}
