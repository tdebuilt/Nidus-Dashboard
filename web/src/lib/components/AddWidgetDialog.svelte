<script lang="ts">
  import { X, Loader2 } from 'lucide-svelte'
  import { api } from '../api/client'
  import { toasts } from '../stores/toast'
  import { t, translate } from '../i18n'
  import WidgetConfigForm from './config/WidgetConfigForm.svelte'
  import { getAllWidgetTypes, getServiceToWidgetMap } from '../widgetRegistry'

  interface ServiceResponse {
    type: string
    enabled: boolean
  }

  interface Props {
    categoryId: number
    nextFreeY?: number
    open?: boolean
    onClose?: () => void
    onCreated?: () => void
  }

  const { categoryId, nextFreeY = 0, open = false, onClose, onCreated }: Props = $props()

  let selectedType = $state('')
  let title = $state('')
  let config = $state('{}')
  let loading = $state(false)
  let configuredServices = $state<string[]>([])
  let loadingServices = $state(false)

  const serviceToWidget = getServiceToWidgetMap()
  const allWidgetTypes = getAllWidgetTypes()

  const availableWidgetTypes = $derived(
    allWidgetTypes.filter((wt) => !wt.serviceType || configuredServices.includes(wt.type))
  )

  async function fetchConfiguredServices() {
    loadingServices = true
    try {
      const services = await api.get<ServiceResponse[]>('/api/services')
      configuredServices = (services ?? [])
        .filter((s) => s.enabled)
        .map((s) => serviceToWidget[s.type] ?? s.type)
    } catch {
      configuredServices = []
    } finally {
      loadingServices = false
    }
  }

  $effect(() => {
    if (open) {
      fetchConfiguredServices()
    }
  })

  function handleSelectType(type: string) {
    selectedType = type
    title = allWidgetTypes.find((w) => w.type === type)?.label || type
    config = '{}'
  }

  async function handleCreate() {
    if (!selectedType || !title.trim()) return
    loading = true
    try {
      await api.post(`/api/categories/${categoryId}/widgets`, {
        type: selectedType,
        title: title.trim(),
        config,
        pos_x: 0,
        pos_y: nextFreeY,
        width: 12,
        height: 0,
      })
      toasts.success(translate('widget.added'))
      selectedType = ''
      title = ''
      config = '{}'
      onCreated?.()
      onClose?.()
    } catch {
      toasts.error(translate('widget.addError'))
    } finally {
      loading = false
    }
  }

  function handleClose() {
    selectedType = ''
    title = ''
    config = '{}'
    onClose?.()
  }
</script>

{#if open}
  <button class="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm" onclick={handleClose} aria-label={$t('common.close')}></button>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4" data-testid="add-widget-dialog">
    <div class="w-full max-w-2xl max-h-[90vh] flex flex-col rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-6 shadow-2xl animate-[dialogIn_0.2s_ease-out]">
      <div class="mb-4 flex items-center justify-between">
        <h3 class="text-lg font-semibold text-[var(--color-text)]">{$t('widget.addTitle')}</h3>
        <button onclick={handleClose} class="touch-action-btn rounded p-2 text-[var(--color-text-muted)] hover:text-[var(--color-text)]" data-testid="add-widget-close">
          <X size={20} />
        </button>
      </div>

      {#if !selectedType}
        <p class="mb-3 text-sm text-[var(--color-text-secondary)]">{$t('widget.chooseType')}</p>
        {#if loadingServices}
          <div class="flex items-center justify-center gap-2 py-8 text-sm text-[var(--color-text-muted)]">
            <Loader2 size={16} class="animate-spin" />
            {$t('common.loading')}
          </div>
        {:else if availableWidgetTypes.length === 0}
          <div class="py-8 text-center text-sm text-[var(--color-text-muted)]">
            <p>{$t('widget.noServicesConfigured')}</p>
          </div>
        {:else}
          <div class="grid grid-cols-2 gap-2" data-testid="widget-type-grid">
            {#each availableWidgetTypes as wt (wt.type)}
              <button onclick={() => handleSelectType(wt.type)}
                class="flex items-center gap-3 rounded-lg border border-[var(--color-border)] px-3 py-3 text-start text-sm transition-colors hover:border-[var(--color-primary)] hover:bg-[var(--color-bg-tertiary)]"
                data-testid="widget-type-{wt.type}">
                <wt.icon size={20} class="text-[var(--color-primary)]" />
                <span class="text-[var(--color-text)]">{wt.label}</span>
              </button>
            {/each}
          </div>
        {/if}
      {:else}
        <div class="flex flex-1 flex-col gap-4 overflow-hidden">
          <div class="shrink-0">
            <label for="widget-title" class="mb-1 block text-sm text-[var(--color-text-secondary)]">{$t('widget.title')}</label>
            <input id="widget-title" type="text" bind:value={title}
              class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
              data-testid="widget-title-input" />
          </div>

          <div class="flex-1 overflow-y-auto">
            <span class="mb-1 block text-sm text-[var(--color-text-secondary)]">{$t('widget.configuration')}</span>
            <WidgetConfigForm type={selectedType} value={config} onchange={(v) => config = v} />
          </div>

          <div class="flex shrink-0 justify-end gap-2">
            <button onclick={() => selectedType = ''}
              class="rounded-lg border border-[var(--color-border)] px-4 py-2 text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]"
              data-testid="widget-back">{$t('common.back')}</button>
            <button onclick={handleCreate} disabled={loading || !title.trim()}
              class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white hover:bg-[var(--color-primary-hover)] disabled:opacity-50"
              data-testid="widget-create">{$t('common.add')}</button>
          </div>
        </div>
      {/if}
    </div>
  </div>
{/if}
