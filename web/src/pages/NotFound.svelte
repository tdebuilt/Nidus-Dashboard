<script lang="ts">
  import { FileQuestion } from 'lucide-svelte'
  import { navigate } from '../lib/stores/router'
  import { t } from '../lib/i18n'
</script>

<div class="flex min-h-[60vh] flex-col items-center justify-center text-center" data-testid="not-found-page">
  <FileQuestion size={64} class="mb-4 text-[var(--color-text-muted)]" />
  <h1 class="mb-2 text-4xl font-bold text-[var(--color-text)]">{$t('errors.notFoundCode')}</h1>
  <p class="mb-6 text-[var(--color-text-secondary)]">{$t('errors.notFound')}</p>
  <button
    onclick={() => navigate('/')}
    class="rounded-lg bg-[var(--color-primary)] px-6 py-2 text-white transition-colors hover:bg-[var(--color-primary-hover)]"
  >
    {$t('errors.backToDashboard')}
  </button>
</div>
