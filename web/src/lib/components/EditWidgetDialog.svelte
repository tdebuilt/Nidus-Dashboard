<script lang="ts">
  import { X } from 'lucide-svelte'
  import { api } from '../api/client'
  import { toasts } from '../stores/toast'
  import { t, translate } from '../i18n'
  import WidgetConfigForm from './config/WidgetConfigForm.svelte'
  import { focusTrap } from '../actions/focusTrap'

  interface Widget {
    id: number
    type: string
    title: string
    config: string
  }

  interface Props {
    widget: Widget
    open?: boolean
    onClose?: () => void
    onUpdated?: () => void
  }

  const { widget, open = false, onClose, onUpdated }: Props = $props()

  let title = $state('')
  let config = $state('{}')
  let loading = $state(false)

  $effect(() => {
    if (open && widget) {
      title = widget.title
      config = widget.config || '{}'
    }
  })

  async function handleSave() {
    if (!title.trim()) return
    loading = true
    try {
      await api.put(`/api/widgets/${widget.id}`, {
        type: widget.type,
        title: title.trim(),
        config,
      })
      toasts.success(translate('widget.updated'))
      onUpdated?.()
      onClose?.()
    } catch {
      toasts.error(translate('widget.editError'))
    } finally {
      loading = false
    }
  }

  function handleClose() {
    onClose?.()
  }
</script>

{#if open}
  <button class="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm" onclick={handleClose} aria-label={$t('common.close')}></button>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4" data-testid="edit-widget-dialog">
    <div class="w-full max-w-2xl max-h-[90vh] flex flex-col rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-6 shadow-2xl animate-[dialogIn_0.2s_ease-out]" role="dialog" aria-modal="true" use:focusTrap={{ onClose: handleClose }}>
      <div class="mb-4 flex shrink-0 items-center justify-between">
        <h3 class="text-lg font-semibold text-[var(--color-text)]">{$t('widget.editTitle')}</h3>
        <button onclick={handleClose} class="touch-action-btn rounded p-2 text-[var(--color-text-muted)] hover:text-[var(--color-text)]" data-testid="edit-widget-close">
          <X size={20} />
        </button>
      </div>

      <div class="flex flex-1 flex-col gap-4 overflow-hidden">
        <div class="shrink-0">
          <label for="edit-widget-title" class="mb-1 block text-sm text-[var(--color-text-secondary)]">{$t('widget.title')}</label>
          <input id="edit-widget-title" type="text" bind:value={title}
            class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
            data-testid="edit-widget-title-input" />
        </div>

        <div class="flex-1 overflow-y-auto">
          <span class="mb-1 block text-sm text-[var(--color-text-secondary)]">{$t('widget.configuration')}</span>
          <WidgetConfigForm type={widget.type} value={config} onchange={(v) => config = v} />
        </div>

        <div class="flex shrink-0 justify-end gap-2">
          <button onclick={handleClose}
            class="rounded-lg border border-[var(--color-border)] px-4 py-2 text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]"
            data-testid="edit-widget-cancel">{$t('common.cancel')}</button>
          <button onclick={handleSave} disabled={loading || !title.trim()}
            class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white hover:bg-[var(--color-primary-hover)] disabled:opacity-50"
            data-testid="edit-widget-save">{$t('common.save')}</button>
        </div>
      </div>
    </div>
  </div>
{/if}
