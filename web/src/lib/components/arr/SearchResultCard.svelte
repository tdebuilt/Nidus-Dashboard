<script lang="ts">
  import { Loader2 } from 'lucide-svelte'
  import { t } from '../../i18n'

  interface LookupResult {
    title: string
    year: number
    tmdbId?: number
    tvdbId?: number
    overview: string
    runtime?: number
    seasonCount?: number
    id?: number
  }

  interface Props {
    result: LookupResult
    serviceType: 'radarr' | 'sonarr'
    submitting: boolean
    onAdd: () => void
  }

  const { result, serviceType, submitting, onAdd }: Props = $props()

  const inLibrary = $derived((result.id ?? 0) > 0)

  const subtitle = $derived(() => {
    if (serviceType === 'sonarr' && result.seasonCount) return `${result.seasonCount} seasons`
    if (serviceType === 'radarr' && result.runtime) return `${result.runtime} min`
    return ''
  })
</script>

<div class="flex items-start gap-2 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] p-2.5">
  <div class="min-w-0 flex-1">
    <div class="flex items-baseline gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text)]">{result.title}</span>
      <span class="text-xs text-[var(--color-text-muted)]">({result.year})</span>
      {#if subtitle()}
        <span class="text-xs text-[var(--color-text-muted)]">· {subtitle()}</span>
      {/if}
    </div>
    {#if result.overview}
      <p class="mt-0.5 line-clamp-2 text-xs text-[var(--color-text-muted)]">{result.overview}</p>
    {/if}
  </div>
  <div class="shrink-0">
    {#if inLibrary}
      <span class="rounded-full bg-[var(--color-bg-tertiary)] px-2 py-0.5 text-[10px] text-[var(--color-text-muted)]">
        {$t('arr.alreadyInLibrary')}
      </span>
    {:else}
      <button
        onclick={onAdd}
        disabled={submitting}
        class="rounded-lg bg-[var(--color-primary)] px-3 py-1 text-xs text-white hover:bg-[var(--color-primary-hover)] disabled:opacity-50"
      >
        {#if submitting}
          <Loader2 size={12} class="animate-spin" />
        {:else}
          {$t('common.add')}
        {/if}
      </button>
    {/if}
  </div>
</div>
