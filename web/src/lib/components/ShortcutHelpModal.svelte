<script lang="ts">
  import { X, Keyboard } from 'lucide-svelte'
  import { shortcutHelpOpen, closeShortcutHelp } from '../stores/shortcutHelp'
  import { t } from '../i18n'
  import { focusTrap } from '../actions/focusTrap'

  const shortcuts = [
    { keys: ['E'], translationKey: 'shortcuts.editMode' },
    { keys: ['1', '-', '9'], translationKey: 'shortcuts.switchCategory' },
    { keys: ['/'], translationKey: 'shortcuts.focusSearch' },
    { keys: ['?'], translationKey: 'shortcuts.showHelp' },
    { keys: ['Esc'], translationKey: 'shortcuts.closeDialog' },
  ]
</script>

{#if $shortcutHelpOpen}
  <!-- Backdrop -->
  <button
    class="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm"
    onclick={closeShortcutHelp}
    aria-label={$t('common.close')}
  ></button>

  <!-- Dialog -->
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <div class="w-full max-w-sm rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-6 shadow-2xl animate-[dialogIn_0.2s_ease-out]" role="dialog" aria-modal="true" use:focusTrap={{ onClose: closeShortcutHelp }}>
      <!-- Header -->
      <div class="mb-5 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <Keyboard size={20} class="text-[var(--color-primary)]" />
          <h3 class="text-lg font-semibold text-[var(--color-text)]">{$t('shortcuts.title')}</h3>
        </div>
        <button
          onclick={closeShortcutHelp}
          class="rounded-lg p-1 text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
          aria-label={$t('common.close')}
        >
          <X size={18} />
        </button>
      </div>

      <!-- Shortcut list -->
      <div class="space-y-3">
        {#each shortcuts as shortcut (shortcut.translationKey)}
          <div class="flex items-center justify-between">
            <span class="text-sm text-[var(--color-text-secondary)]">{$t(shortcut.translationKey)}</span>
            <div class="flex items-center gap-1">
              {#each shortcut.keys as key (key)}
                {#if key === '-'}
                  <span class="text-xs text-[var(--color-text-secondary)]">-</span>
                {:else}
                  <kbd class="inline-flex min-w-[28px] items-center justify-center rounded-md border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-2 py-1 text-xs font-mono text-[var(--color-text)]">
                    {key}
                  </kbd>
                {/if}
              {/each}
            </div>
          </div>
        {/each}
      </div>

      <!-- Footer -->
      <div class="mt-5 flex justify-end">
        <button
          onclick={closeShortcutHelp}
          class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white transition-colors hover:bg-[var(--color-primary-hover)]"
        >
          {$t('common.close')}
        </button>
      </div>
    </div>
  </div>
{/if}
