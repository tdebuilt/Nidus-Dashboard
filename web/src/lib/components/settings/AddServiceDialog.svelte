<script lang="ts">
  import { X, ArrowDownAZ, ArrowUpZA } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { t, translate } from '../../i18n'
  import DynamicIcon from '../DynamicIcon.svelte'
  import { getServiceIcon } from './serviceIcons'
  import { focusTrap } from '../../actions/focusTrap'
  import ServiceAuthFields from './ServiceAuthFields.svelte'
  import {
    type ServiceTypeDef,
    serviceDisplayName,
    needsURL,
    hasJDAuth,
    buildCredentials,
  } from './serviceAuth'

  interface Props {
    open?: boolean
    serviceTypeDefs: ServiceTypeDef[]
    availableToAdd: string[]
    onClose?: () => void
    onCreated?: () => void
  }

  const { open = false, serviceTypeDefs, availableToAdd, onClose, onCreated }: Props = $props()

  let sortAsc = $state(true)
  let selectedType = $state('')
  let serviceUrl = $state('')
  let serviceToken = $state('')
  let serviceUsername = $state('')
  let servicePassword = $state('')
  let serviceAuthMode = $state<'token' | 'userpass'>('token')
  let proxmoxTokenId = $state('')
  let proxmoxTokenSecret = $state('')
  let jdEmail = $state('')
  let jdPassword = $state('')
  let loading = $state(false)

  const sortedServices = $derived(
    [...availableToAdd].sort((a, b) => {
      const cmp = serviceDisplayName(serviceTypeDefs, a).localeCompare(serviceDisplayName(serviceTypeDefs, b))
      return sortAsc ? cmp : -cmp
    })
  )

  function resetForm() {
    selectedType = ''
    serviceUrl = ''
    serviceToken = ''
    serviceUsername = ''
    servicePassword = ''
    serviceAuthMode = 'token'
    proxmoxTokenId = ''
    proxmoxTokenSecret = ''
    jdEmail = ''
    jdPassword = ''
  }

  function handleSelectType(type: string) {
    resetForm()
    selectedType = type
  }

  async function handleSave() {
    if (!selectedType) return
    loading = true
    try {
      const creds = buildCredentials(serviceTypeDefs, {
        serviceType: selectedType,
        authMode: serviceAuthMode,
        token: serviceToken,
        username: serviceUsername,
        password: servicePassword,
        proxmoxTokenId,
        proxmoxTokenSecret,
        jdEmail,
        jdPassword,
      })
      const url = hasJDAuth(serviceTypeDefs, selectedType) ? 'https://api.jdownloader.org' : serviceUrl
      await api.put(`/api/services/${selectedType}`, {
        name: selectedType,
        url,
        credentials: creds || undefined,
      })
      toasts.success(translate('settings.serviceConfigured'))
      resetForm()
      onCreated?.()
      onClose?.()
    } catch {
      toasts.error(translate('settings.serviceConfigError'))
    } finally {
      loading = false
    }
  }

  function handleClose() {
    resetForm()
    onClose?.()
  }
</script>

{#if open}
  <button class="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm" onclick={handleClose} aria-label={$t('common.close')}></button>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4" data-testid="add-service-dialog">
    <div class="w-full max-w-2xl max-h-[90vh] flex flex-col rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-6 shadow-2xl animate-[dialogIn_0.2s_ease-out]" role="dialog" aria-modal="true" use:focusTrap={{ onClose: handleClose }}>
      <div class="mb-4 flex items-center justify-between">
        <h3 class="text-lg font-semibold text-[var(--color-text)]">{$t('settings.addService')}</h3>
        <button onclick={handleClose} class="touch-action-btn rounded p-2 text-[var(--color-text-muted)] hover:text-[var(--color-text)]" data-testid="add-service-close">
          <X size={20} />
        </button>
      </div>

      {#if !selectedType}
        <div class="mb-3 flex items-center justify-between">
          <p class="text-sm text-[var(--color-text-secondary)]">{$t('settings.chooseService')}</p>
          {#if availableToAdd.length > 1}
            <button onclick={() => sortAsc = !sortAsc}
              class="rounded p-1.5 text-[var(--color-text-muted)] transition-colors hover:text-[var(--color-text)]"
              title={sortAsc ? 'Z → A' : 'A → Z'}
              data-testid="service-sort-toggle">
              {#if sortAsc}
                <ArrowDownAZ size={16} />
              {:else}
                <ArrowUpZA size={16} />
              {/if}
            </button>
          {/if}
        </div>
        {#if availableToAdd.length === 0}
          <div class="py-8 text-center text-sm text-[var(--color-text-muted)]">
            <p>{$t('settings.allServicesConfigured')}</p>
          </div>
        {:else}
          <div class="grid grid-cols-2 gap-2" data-testid="service-type-grid">
            {#each sortedServices as type (type)}
              <button onclick={() => handleSelectType(type)}
                class="flex items-center gap-3 rounded-lg border border-[var(--color-border)] px-3 py-3 text-start text-sm transition-colors hover:border-[var(--color-primary)] hover:bg-[var(--color-bg-tertiary)]"
                data-testid="service-type-{type}">
                <DynamicIcon name={getServiceIcon(type)} size={20} class="text-[var(--color-primary)]" />
                <span class="text-[var(--color-text)]">{serviceDisplayName(serviceTypeDefs, type)}</span>
              </button>
            {/each}
          </div>
        {/if}
      {:else}
        <div class="flex flex-1 flex-col gap-4 overflow-hidden">
          <div class="flex items-center gap-2 shrink-0">
            <DynamicIcon name={getServiceIcon(selectedType)} size={20} class="text-[var(--color-primary)]" />
            <span class="text-sm font-medium text-[var(--color-text)]">{serviceDisplayName(serviceTypeDefs, selectedType)}</span>
          </div>

          <div class="flex-1 space-y-3 overflow-y-auto">
            {#if needsURL(serviceTypeDefs, selectedType)}
              <div>
                <label for="svc-url-add" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.url')}</label>
                <input id="svc-url-add" type="url" bind:value={serviceUrl} placeholder="https://..."
                  class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                  data-testid="service-url-input" />
              </div>
            {/if}

            <ServiceAuthFields
              serviceType={selectedType}
              {serviceTypeDefs}
              bind:authMode={serviceAuthMode}
              bind:token={serviceToken}
              bind:username={serviceUsername}
              bind:password={servicePassword}
              bind:proxmoxTokenId
              bind:proxmoxTokenSecret
              bind:jdEmail
              bind:jdPassword
              idSuffix="add"
            />
          </div>

          <div class="flex shrink-0 justify-end gap-2">
            <button onclick={() => selectedType = ''}
              class="rounded-lg border border-[var(--color-border)] px-4 py-2 text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]"
              data-testid="service-back">{$t('common.back')}</button>
            <button onclick={handleSave} disabled={loading}
              class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white hover:bg-[var(--color-primary-hover)] disabled:opacity-50"
              data-testid="service-save-btn">{$t('common.save')}</button>
          </div>
        </div>
      {/if}
    </div>
  </div>
{/if}
