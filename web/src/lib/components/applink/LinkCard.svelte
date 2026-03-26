<script lang="ts">
  import { ExternalLink, Globe } from 'lucide-svelte'
  import { t } from '../../i18n'

  interface Props {
    name: string
    url: string
    status: 'up' | 'down' | 'pending'
  }

  const { name, url, status }: Props = $props()

  const faviconUrl = $derived(
    url ? `/api/applinks/favicon?url=${encodeURIComponent(url)}` : ''
  )

  let faviconError = $state(false)

  function onFaviconError() {
    faviconError = true
  }

  function statusColor(s: string): string {
    if (s === 'up') return 'var(--color-success)'
    if (s === 'down') return 'var(--color-danger)'
    return 'var(--color-text-muted)'
  }

  function statusLabel(s: string): string {
    if (s === 'up') return $t('applink.statusUp')
    if (s === 'down') return $t('applink.statusDown')
    return $t('applink.statusPending')
  }

  const color = $derived(statusColor(status))
  const label = $derived(statusLabel(status))
</script>

<a
  href={url}
  target="_blank"
  rel="noopener"
  class="flex items-center gap-3 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 transition-colors hover:border-[var(--color-primary)]"
  data-testid="link-card"
>
  <!-- Favicon -->
  <div class="flex-shrink-0 text-[var(--color-primary)]">
    {#if faviconUrl && !faviconError}
      <img src={faviconUrl} alt="" width="18" height="18" class="rounded-sm" onerror={onFaviconError} />
    {:else}
      <Globe size={18} />
    {/if}
  </div>

  <!-- Name -->
  <div class="min-w-0 flex-1">
    <span class="truncate text-sm font-medium text-[var(--color-text)]">{name}</span>
  </div>

  <!-- Health dot -->
  <div class="h-2.5 w-2.5 flex-shrink-0 rounded-full" style="background-color: {color}" title={label}></div>

  <!-- External link icon -->
  <ExternalLink size={12} class="flex-shrink-0 text-[var(--color-text-muted)]" />
</a>
