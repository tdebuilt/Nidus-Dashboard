<script lang="ts">
  import { ServerOff } from 'lucide-svelte'
  import { t } from '../lib/i18n'

  interface Props {
    message?: string
    onRetry?: () => void
  }

  const { message, onRetry }: Props = $props()
</script>

<div class="flex min-h-[60vh] flex-col items-center justify-center text-center" data-testid="service-unavailable-page">
  <ServerOff size={64} class="mb-4 text-[var(--color-text-muted)]" />
  <h1 class="mb-2 text-2xl font-bold text-[var(--color-text)]">{$t('errors.serviceUnavailable')}</h1>
  <p class="mb-6 text-[var(--color-text-secondary)]">{message || $t('errors.serviceMessage')}</p>
  {#if onRetry}
    <button
      onclick={onRetry}
      class="rounded-lg bg-[var(--color-primary)] px-6 py-2 text-white transition-colors hover:bg-[var(--color-primary-hover)]"
      data-testid="retry-button"
    >
      {$t('common.retry')}
    </button>
  {/if}
</div>
