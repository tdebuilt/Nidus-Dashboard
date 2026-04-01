<script lang="ts">
  import { X } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { t, translate } from '../../i18n'
  import { focusTrap } from '../../actions/focusTrap'

  interface Props {
    open: boolean
    onClose: () => void
    onAdded?: () => void
  }

  const { open, onClose, onAdded }: Props = $props()

  let url = $state('')
  let submitting = $state(false)

  function reset() {
    url = ''
  }

  async function handleSubmit() {
    if (!url.trim()) return
    submitting = true
    try {
      await api.post('/api/qbittorrent/torrents', { url: url.trim() })
      toasts.success(translate('qbittorrent.torrentAdded'))
      reset()
      onClose()
      onAdded?.()
    } catch {
      toasts.error(translate('qbittorrent.addError'))
    } finally {
      submitting = false
    }
  }

  const canSubmit = $derived(url.trim().length > 0)
</script>

{#if open}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onclick={onClose} onkeydown={(e) => { if (e.key === 'Escape') onClose() }}>
    <div class="mx-4 w-full max-w-md rounded-xl bg-[var(--color-bg-secondary)] p-6 shadow-xl animate-[dialogIn_0.2s_ease-out]" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()} role="dialog" aria-modal="true" tabindex="-1" use:focusTrap={{ onClose: onClose }}>
      <div class="mb-4 flex items-center justify-between">
        <h2 class="text-lg font-semibold text-[var(--color-text)]">{$t('qbittorrent.addTorrent')}</h2>
        <button onclick={onClose} class="touch-action-btn rounded p-2 text-[var(--color-text-muted)] hover:text-[var(--color-text)]" aria-label={$t('common.close')}>
          <X size={16} />
        </button>
      </div>

      <input
        type="text"
        bind:value={url}
        placeholder={$t('qbittorrent.urlPlaceholder')}
        aria-label={$t('qbittorrent.urlPlaceholder')}
        class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)] focus:outline-none"
      />

      <div class="mt-4 flex justify-end gap-2">
        <button onclick={() => { reset(); onClose() }}
          class="rounded-lg border border-[var(--color-border)] px-4 py-2 text-sm text-[var(--color-text)] hover:bg-[var(--color-bg)]">
          {$t('common.cancel')}
        </button>
        <button onclick={handleSubmit} disabled={submitting || !canSubmit}
          class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white hover:bg-[var(--color-primary-hover)] disabled:opacity-50">
          {submitting ? $t('common.loading') : $t('common.add')}
        </button>
      </div>
    </div>
  </div>
{/if}
