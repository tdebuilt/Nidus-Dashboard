<script lang="ts">
  import { Loader2, AlertCircle, CheckCircle, XCircle, MinusCircle, RefreshCw, Zap } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { t } from '../../i18n'
  import { isViewer } from '../../stores/auth'

  interface ProwlarrIndexer {
    id: number
    name: string
    enable: boolean
    protocol: string
    priority: number
    status: string // "ok" | "error" | "disabled"
  }

  let loading = $state(true)
  let error = $state<string | null>(null)
  let indexers = $state<ProwlarrIndexer[]>([])
  let testingId = $state<number | null>(null)

  async function fetchIndexers() {
    loading = indexers.length === 0
    error = null
    try {
      indexers = await api.get<ProwlarrIndexer[]>('/api/arr/prowlarr/indexers')
    } catch {
      error = 'fetch_error'
    } finally {
      loading = false
    }
  }

  async function testIndexer(id: number) {
    testingId = id
    try {
      const result = await api.post<{ success: boolean; error?: string }>(`/api/arr/prowlarr/indexer/${id}/test`, {})
      const idx = indexers.findIndex(i => i.id === id)
      if (idx !== -1) {
        indexers[idx] = { ...indexers[idx], status: result.success ? 'ok' : 'error' }
      }
    } catch {
      // Silently handle — the user sees the status didn't change
    } finally {
      testingId = null
    }
  }

  fetchIndexers()
</script>

{#if loading}
  <div class="flex items-center justify-center gap-2 py-4 text-xs text-[var(--color-text-muted)]">
    <Loader2 size={14} class="animate-spin" />
    {$t('arr.loadingIndexers')}
  </div>
{:else if error}
  <div class="flex flex-col items-center gap-2 py-4 text-xs text-[var(--color-text-muted)]">
    <AlertCircle size={16} class="text-[var(--color-danger)]" />
    <p>{$t('arr.indexerError')}</p>
    <button onclick={fetchIndexers} class="text-[var(--color-primary)] hover:underline">{$t('common.retry')}</button>
  </div>
{:else if indexers.length === 0}
  <div class="py-4 text-center text-xs text-[var(--color-text-muted)]">{$t('arr.noIndexers')}</div>
{:else}
  <div class="space-y-1">
    <div class="mb-1.5 flex items-center gap-1.5 text-xs font-medium text-[var(--color-text-secondary)]">
      <Zap size={12} />
      {$t('arr.indexers')} ({indexers.length})
    </div>
    {#each indexers as indexer (indexer.id)}
      <div class="flex items-center gap-2 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2.5 py-2">
        <!-- Status icon -->
        {#if indexer.status === 'ok'}
          <CheckCircle size={14} class="shrink-0 text-[var(--color-success)]" />
        {:else if indexer.status === 'error'}
          <XCircle size={14} class="shrink-0 text-[var(--color-danger)]" />
        {:else}
          <MinusCircle size={14} class="shrink-0 text-[var(--color-text-muted)]" />
        {/if}

        <!-- Name & protocol -->
        <div class="min-w-0 flex-1">
          <span class="block truncate text-sm font-medium text-[var(--color-text)]">{indexer.name}</span>
          <span class="text-[10px] text-[var(--color-text-muted)]">{indexer.protocol}</span>
        </div>

        <!-- Test button -->
        {#if !$isViewer && indexer.enable}
          <button
            class="shrink-0 rounded p-1 text-[var(--color-text-muted)] transition-colors hover:bg-[var(--color-bg-tertiary)] hover:text-[var(--color-primary)] disabled:opacity-40"
            onclick={() => testIndexer(indexer.id)}
            disabled={testingId !== null}
            title={$t('arr.testIndexer')}
            aria-label={$t('arr.testIndexer')}
          >
            {#if testingId === indexer.id}
              <Loader2 size={14} class="animate-spin" />
            {:else}
              <RefreshCw size={14} />
            {/if}
          </button>
        {/if}
      </div>
    {/each}
  </div>
{/if}
