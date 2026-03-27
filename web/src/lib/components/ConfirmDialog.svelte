<script lang="ts">
  import { AlertTriangle } from 'lucide-svelte'
  import { confirmState, resolveConfirm } from '../stores/confirm'
  import { t } from '../i18n'
  import { focusTrap } from '../actions/focusTrap'
</script>

{#if $confirmState.open}
  <!-- Backdrop -->
  <button
    class="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm"
    onclick={() => resolveConfirm(false)}
    aria-label={$t('common.cancel')}
    data-testid="confirm-overlay"
  ></button>

  <!-- Dialog -->
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4" data-testid="confirm-dialog">
    <div class="w-full max-w-md rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-6 shadow-2xl animate-[dialogIn_0.2s_ease-out]" role="dialog" aria-modal="true" use:focusTrap={{ onClose: () => resolveConfirm(false) }}>
      <div class="mb-4 flex items-center gap-3">
        {#if $confirmState.options.destructive}
          <div class="flex h-10 w-10 items-center justify-center rounded-full" style="background-color: var(--color-error-bg)">
            <AlertTriangle size={20} class="text-[var(--color-danger)]" />
          </div>
        {/if}
        <h3 class="text-lg font-semibold text-[var(--color-text)]">{$confirmState.options.title}</h3>
      </div>

      <p class="mb-6 text-sm text-[var(--color-text-secondary)]">{$confirmState.options.message}</p>

      <div class="flex justify-end gap-3">
        <button
          onclick={() => resolveConfirm(false)}
          class="rounded-lg border border-[var(--color-border)] px-4 py-2 text-sm text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
          data-testid="confirm-cancel"
        >
          {$confirmState.options.cancelLabel || $t('common.cancel')}
        </button>
        <button
          onclick={() => resolveConfirm(true)}
          class="rounded-lg px-4 py-2 text-sm text-white transition-colors
            {$confirmState.options.destructive
              ? 'bg-[var(--color-danger)] hover:bg-[var(--color-danger-hover)]'
              : 'bg-[var(--color-primary)] hover:bg-[var(--color-primary-hover)]'}"
          data-testid="confirm-ok"
        >
          {$confirmState.options.confirmLabel || $t('common.confirm')}
        </button>
      </div>
    </div>
  </div>
{/if}
