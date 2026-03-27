<script lang="ts">
  import { Plus, Check, X, Trash2 } from 'lucide-svelte'
  import { t } from '../../i18n'

  interface Props {
    name: string
    ip: string
    username: string
    password: string
    isEdit: boolean
    onSave: () => void
    onCancel: () => void
    onDelete?: () => void
  }

  let {
    name = $bindable(),
    ip = $bindable(),
    username = $bindable(),
    password = $bindable(),
    isEdit,
    onSave,
    onCancel,
    onDelete,
  }: Props = $props()
</script>

<div
  class="space-y-2 rounded-lg border p-3"
  class:border-[var(--color-primary)]={isEdit}
  class:border-[var(--color-border)]={!isEdit}
>
  {#if !isEdit}
    <span class="text-xs font-medium text-[var(--color-text-secondary)]">{$t('reolink.addCamera')}</span>
  {/if}
  <div class="grid grid-cols-2 gap-2">
    <input type="text" bind:value={name} placeholder={$t('reolink.cameraName')} aria-label={$t('reolink.cameraName')}
      class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]" />
    <input type="text" bind:value={ip} placeholder={$t('reolink.cameraIP')} aria-label={$t('reolink.cameraIP')}
      class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]" />
    <input type="text" bind:value={username} placeholder={$t('reolink.username')} aria-label={$t('reolink.username')}
      class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]" />
    <input type="password" bind:value={password} placeholder={$t('reolink.password')} aria-label={$t('reolink.password')}
      class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]" />
  </div>
  <div class="flex gap-2">
    {#if isEdit}
      <button onclick={onSave}
        class="flex items-center gap-1 rounded-lg bg-[var(--color-primary)] px-2 py-1 text-xs text-white hover:bg-[var(--color-primary-hover)]">
        <Check size={12} /> {$t('common.save')}
      </button>
      <button onclick={onCancel}
        class="flex items-center gap-1 rounded-lg bg-[var(--color-bg-tertiary)] px-2 py-1 text-xs text-[var(--color-text-secondary)]">
        <X size={12} /> {$t('common.cancel')}
      </button>
      {#if onDelete}
        <button onclick={onDelete}
          class="flex items-center gap-1 rounded-lg px-2 py-1 text-xs text-[var(--color-danger)] hover:bg-[var(--color-error-bg)]">
          <Trash2 size={12} /> {$t('common.delete')}
        </button>
      {/if}
    {:else}
      <button onclick={onSave}
        class="flex items-center gap-1 rounded-lg bg-[var(--color-primary)] px-3 py-1.5 text-xs text-white hover:bg-[var(--color-primary-hover)]">
        <Plus size={12} /> {$t('reolink.addCamera')}
      </button>
    {/if}
  </div>
</div>
