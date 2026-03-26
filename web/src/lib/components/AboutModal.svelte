<script lang="ts">
  import { X, Info } from 'lucide-svelte'
  import { aboutModalOpen, closeAboutModal } from '../stores/aboutModal'
  import { appVersion } from '../stores/version'
  import { t } from '../i18n'
</script>

{#if $aboutModalOpen}
  <!-- Backdrop -->
  <button
    class="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm"
    onclick={closeAboutModal}
    aria-label={$t('common.close')}
  ></button>

  <!-- Dialog -->
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <div class="w-full max-w-sm rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-6 shadow-2xl animate-[dialogIn_0.2s_ease-out]">
      <!-- Header -->
      <div class="mb-5 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <Info size={20} class="text-[var(--color-primary)]" />
          <h3 class="text-lg font-semibold text-[var(--color-text)]">{$t('settings.about')}</h3>
        </div>
        <button
          onclick={closeAboutModal}
          class="rounded-lg p-1 text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
          aria-label={$t('common.close')}
        >
          <X size={18} />
        </button>
      </div>

      <!-- Content -->
      <div class="space-y-3 text-sm text-[var(--color-text-secondary)]">
        <div class="flex items-center justify-between">
          <span>{$t('settings.version')}</span>
          <span class="font-mono text-[var(--color-text)]" data-testid="about-version">{$appVersion || 'dev'}</span>
        </div>
        <div class="flex items-center justify-between">
          <span>{$t('settings.license')}</span>
          <span class="text-[var(--color-text)]">MIT</span>
        </div>
        <div class="flex items-center justify-between">
          <span>GitHub</span>
          <a href="https://github.com/tdebuilt/Nidus-Dashboard" target="_blank" rel="noopener"
            class="text-[var(--color-primary)] hover:underline">tdebuilt/Nidus-Dashboard</a>
        </div>
      </div>

      <!-- Footer -->
      <div class="mt-5 flex justify-end">
        <button
          onclick={closeAboutModal}
          class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white transition-colors hover:bg-[var(--color-primary-hover)]"
        >
          {$t('common.close')}
        </button>
      </div>
    </div>
  </div>
{/if}
