<script lang="ts">
  import { X } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { t, translate } from '../../i18n'

  interface Props {
    open: boolean
    onClose: () => void
    onAdded?: () => void
  }

  const { open, onClose, onAdded }: Props = $props()

  let links = $state('')
  let submitting = $state(false)

  async function handleSubmit() {
    const urls = links
      .split('\n')
      .map((l) => l.trim())
      .filter((l) => l.length > 0)

    if (urls.length === 0) return

    submitting = true
    try {
      await api.post('/api/jdownloader/links', { links: urls })
      toasts.success(translate('jdownloader.linksAdded'))
      links = ''
      onClose()
      onAdded?.()
    } catch {
      toasts.error(translate('jdownloader.addError'))
    } finally {
      submitting = false
    }
  }
</script>

{#if open}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onclick={onClose} onkeydown={() => {}}>
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="mx-4 w-full max-w-md rounded-xl bg-[var(--color-bg-secondary)] p-6 shadow-xl animate-[dialogIn_0.2s_ease-out]" onclick={(e) => e.stopPropagation()} onkeydown={() => {}}>
      <div class="mb-4 flex items-center justify-between">
        <h2 class="text-lg font-semibold text-[var(--color-text)]">{$t('jdownloader.addLinks')}</h2>
        <button onclick={onClose} class="touch-action-btn rounded p-2 text-[var(--color-text-muted)] hover:text-[var(--color-text)]">
          <X size={16} />
        </button>
      </div>

      <textarea
        bind:value={links}
        placeholder={$t('jdownloader.linksPlaceholder')}
        class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] p-3 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)] focus:outline-none"
        rows="6"
      ></textarea>

      <div class="mt-4 flex justify-end gap-2">
        <button onclick={onClose}
          class="rounded-lg border border-[var(--color-border)] px-4 py-2 text-sm text-[var(--color-text)] hover:bg-[var(--color-bg)]">
          {$t('common.cancel')}
        </button>
        <button onclick={handleSubmit} disabled={submitting || links.trim().length === 0}
          class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white hover:bg-[var(--color-primary-hover)] disabled:opacity-50">
          {submitting ? $t('common.loading') : $t('common.add')}
        </button>
      </div>
    </div>
  </div>
{/if}
