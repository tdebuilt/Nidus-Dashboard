<script lang="ts">
  import { X, CheckCircle, AlertCircle, Info } from 'lucide-svelte'
  import { toasts } from '../stores/toast'
  import { t } from '../i18n'

  const icons = {
    success: CheckCircle,
    error: AlertCircle,
    info: Info,
  }

  const colors = {
    success: 'toast-success',
    error: 'toast-error',
    info: 'toast-info',
  }
</script>

<div class="fixed end-4 bottom-4 z-50 flex flex-col gap-2" data-testid="toast-container">
  {#each $toasts as toast (toast.id)}
    {@const Icon = icons[toast.type]}
    <div
      class="flex items-center gap-3 rounded-lg border px-4 py-3 shadow-lg backdrop-blur-sm {colors[toast.type]}"
      role="alert"
      data-testid="toast"
      data-toast-type={toast.type}
    >
      <Icon size={18} />
      <span class="flex-1 text-sm">{toast.message}</span>
      <button
        onclick={() => toasts.remove(toast.id)}
        class="rounded p-1 opacity-60 transition-opacity hover:opacity-100"
        aria-label={$t('common.close')}
        data-testid="toast-close"
      >
        <X size={14} />
      </button>
    </div>
  {/each}
</div>
