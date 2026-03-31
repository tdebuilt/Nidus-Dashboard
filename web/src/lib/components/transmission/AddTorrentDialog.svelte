<script lang="ts">
  import { X, Link, Upload } from 'lucide-svelte'
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

  let mode = $state<'url' | 'file'>('url')
  let url = $state('')
  let fileName = $state('')
  let fileBase64 = $state('')
  let submitting = $state(false)
  let fileInput: HTMLInputElement | undefined = $state()

  function reset() {
    url = ''
    fileName = ''
    fileBase64 = ''
    mode = 'url'
  }

  function handleFileChange(e: Event) {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) return

    fileName = file.name
    const reader = new FileReader()
    reader.onload = () => {
      const result = reader.result as string
      // Remove data:...;base64, prefix
      fileBase64 = result.split(',')[1] ?? ''
    }
    reader.readAsDataURL(file)
  }

  async function handleSubmit() {
    submitting = true
    try {
      if (mode === 'file' && fileBase64) {
        await api.post('/api/transmission/torrents', { metainfo: fileBase64 })
      } else if (mode === 'url' && url.trim()) {
        await api.post('/api/transmission/torrents', { url: url.trim() })
      } else {
        return
      }
      toasts.success(translate('transmission.torrentAdded'))
      reset()
      onClose()
      onAdded?.()
    } catch {
      toasts.error(translate('transmission.addError'))
    } finally {
      submitting = false
    }
  }

  const canSubmit = $derived(
    (mode === 'url' && url.trim().length > 0) ||
    (mode === 'file' && fileBase64.length > 0)
  )
</script>

{#if open}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onclick={onClose} onkeydown={(e) => { if (e.key === 'Escape') onClose() }}>
    <div class="mx-4 w-full max-w-md rounded-xl bg-[var(--color-bg-secondary)] p-6 shadow-xl animate-[dialogIn_0.2s_ease-out]" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()} role="dialog" aria-modal="true" tabindex="-1" use:focusTrap={{ onClose: onClose }}>
      <div class="mb-4 flex items-center justify-between">
        <h2 class="text-lg font-semibold text-[var(--color-text)]">{$t('transmission.addTorrent')}</h2>
        <button onclick={onClose} class="touch-action-btn rounded p-2 text-[var(--color-text-muted)] hover:text-[var(--color-text)]" aria-label={$t('common.close')}>
          <X size={16} />
        </button>
      </div>

      <!-- Mode tabs -->
      <div class="mb-3 flex gap-1 rounded-lg bg-[var(--color-bg)] p-1">
        <button
          onclick={() => mode = 'url'}
          class="flex flex-1 items-center justify-center gap-1.5 rounded-md px-3 py-1.5 text-sm transition-colors {mode === 'url' ? 'bg-[var(--color-bg-secondary)] text-[var(--color-text)] shadow-sm' : 'text-[var(--color-text-muted)] hover:text-[var(--color-text)]'}"
        >
          <Link size={14} />
          {$t('transmission.urlTab')}
        </button>
        <button
          onclick={() => mode = 'file'}
          class="flex flex-1 items-center justify-center gap-1.5 rounded-md px-3 py-1.5 text-sm transition-colors {mode === 'file' ? 'bg-[var(--color-bg-secondary)] text-[var(--color-text)] shadow-sm' : 'text-[var(--color-text-muted)] hover:text-[var(--color-text)]'}"
        >
          <Upload size={14} />
          {$t('transmission.fileTab')}
        </button>
      </div>

      {#if mode === 'url'}
        <input
          type="text"
          bind:value={url}
          placeholder={$t('transmission.urlPlaceholder')}
          aria-label={$t('transmission.urlPlaceholder')}
          class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)] focus:outline-none"
        />
      {:else}
        <div
          class="flex flex-col items-center gap-2 rounded-lg border-2 border-dashed border-[var(--color-border)] bg-[var(--color-bg)] p-4 text-center transition-colors hover:border-[var(--color-primary)]"
          role="button"
          tabindex="0"
          onclick={() => fileInput?.click()}
          onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') fileInput?.click() }}
        >
          <Upload size={24} class="text-[var(--color-text-muted)]" />
          {#if fileName}
            <span class="text-sm text-[var(--color-text)]">{fileName}</span>
          {:else}
            <span class="text-sm text-[var(--color-text-muted)]">{$t('transmission.fileDropHint')}</span>
          {/if}
          <input
            bind:this={fileInput}
            type="file"
            accept=".torrent"
            class="hidden"
            onchange={handleFileChange}
          />
        </div>
      {/if}

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
