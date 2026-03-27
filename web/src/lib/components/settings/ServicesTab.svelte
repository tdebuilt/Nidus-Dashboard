<script lang="ts">
  import { Server, Plus, ArrowDownAZ, ArrowUpZA } from 'lucide-svelte'
  import { api, ApiError } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { confirm } from '../../stores/confirm'
  import { t, translate } from '../../i18n'
  import ServiceCard from './ServiceCard.svelte'
  import ServiceAuthFields from './ServiceAuthFields.svelte'
  import { serviceStatuses, startServiceStatusPolling, stopServiceStatusPolling } from '../../stores/serviceStatus'
  import AddServiceDialog from './AddServiceDialog.svelte'
  import Go2rtcSection from './Go2rtcSection.svelte'
  import {
    type ServiceTypeDef,
    serviceDisplayName,
    needsURL,
    hasJDAuth,
    buildCredentials,
  } from './serviceAuth'

  let services = $state<Array<{ type: string; name: string; url: string; enabled: boolean; has_credentials: boolean }>>([])
  let _servicesLoading = $state(true)
  let serviceTypeDefs = $state.raw<ServiceTypeDef[]>([])

  let showAddDialog = $state(false)
  let sortAsc = $state(true)
  let editingService = $state<string | null>(null)
  let serviceUrl = $state('')
  let serviceToken = $state('')
  let serviceUsername = $state('')
  let servicePassword = $state('')
  let serviceAuthMode = $state<'token' | 'userpass'>('token')
  let proxmoxTokenId = $state('')
  let proxmoxTokenSecret = $state('')
  let jdEmail = $state('')
  let jdPassword = $state('')

  const serviceTypes = $derived(serviceTypeDefs.map(d => d.type))
  const configuredServices = $derived(serviceTypes.filter(type => services.some(s => s.type === type)))
  const availableToAdd = $derived(serviceTypes.filter(type => !services.some(s => s.type === type)))

  const sortedConfigured = $derived(
    [...configuredServices].sort((a, b) => {
      const cmp = serviceDisplayName(serviceTypeDefs, a).localeCompare(serviceDisplayName(serviceTypeDefs, b))
      return sortAsc ? cmp : -cmp
    })
  )

  $effect(() => {
    loadServiceTypes()
    loadServices()
    return () => stopServiceStatusPolling()
  })

  async function loadServiceTypes() {
    try {
      const defs = await api.get<ServiceTypeDef[]>('/api/services?types=true')
      serviceTypeDefs = defs.sort((a, b) => a.display_name.localeCompare(b.display_name))
    } catch { serviceTypeDefs = [] }
  }

  async function loadServices() {
    try {
      services = await api.get('/api/services')
      const withUrl = services.filter(s => s.url).map(s => s.type)
      if (withUrl.length > 0) startServiceStatusPolling(withUrl)
    }
    catch { services = [] }
    finally { _servicesLoading = false }
  }

  function startEditService(type: string) {
    const svc = services.find((s) => s.type === type)
    editingService = type
    serviceUrl = svc?.url || ''
    serviceToken = ''
    serviceUsername = ''
    servicePassword = ''
    serviceAuthMode = 'token'
    proxmoxTokenId = ''
    proxmoxTokenSecret = ''
    jdEmail = ''
    jdPassword = ''
  }

  async function saveService() {
    if (!editingService) return
    try {
      const creds = buildCredentials(serviceTypeDefs, {
        serviceType: editingService,
        authMode: serviceAuthMode,
        token: serviceToken,
        username: serviceUsername,
        password: servicePassword,
        proxmoxTokenId,
        proxmoxTokenSecret,
        jdEmail,
        jdPassword,
      })
      const url = hasJDAuth(serviceTypeDefs, editingService) ? 'https://api.jdownloader.org' : serviceUrl
      await api.put(`/api/services/${editingService}`, {
        name: editingService,
        url,
        credentials: creds || undefined,
      })
      toasts.success(translate('settings.serviceConfigured'))
      editingService = null
      loadServices()
    } catch {
      toasts.error(translate('settings.serviceConfigError'))
    }
  }

  async function testService(type: string) {
    try {
      const result = await api.post<{ success: boolean; message: string }>(`/api/services/${type}/test`)
      if (result.success) {
        toasts.success(`${type}: ${result.message}`)
      } else {
        toasts.error(`${type}: ${result.message}`)
      }
    } catch (e) {
      toasts.error(translate('settings.serviceTestError', { message: e instanceof ApiError ? e.message : 'unknown' }))
    }
  }

  async function deleteService(type: string) {
    const ok = await confirm({
      title: translate('settings.deleteService'),
      message: translate('settings.deleteServiceConfirm', { name: serviceDisplayName(serviceTypeDefs, type) }),
      destructive: true,
      confirmLabel: translate('common.delete'),
    })
    if (!ok) return
    try {
      await api.delete(`/api/services/${type}`)
      toasts.success(translate('settings.serviceDeleted'))
      loadServices()
    } catch {
      toasts.error(translate('settings.serviceDeleteError'))
    }
  }
</script>

<div class="space-y-6">
  <h3 class="text-lg font-semibold text-[var(--color-text)]">{$t('settings.tabs.services')}</h3>

  <!-- Services -->
  <section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-services">
    <div class="mb-3 flex items-center gap-2">
      <Server size={18} class="text-[var(--color-text-secondary)]" />
      <h3 class="font-semibold text-[var(--color-text)]">
        {#if configuredServices.length > 0}
          {$t('settings.servicesCount', { count: configuredServices.length })}
        {:else}
          {$t('settings.servicesSection')}
        {/if}
      </h3>
      <div class="ml-auto flex items-center gap-1">
        {#if configuredServices.length > 1}
          <button onclick={() => sortAsc = !sortAsc}
            class="flex items-center justify-center rounded-lg p-1.5 text-[var(--color-text-muted)] transition-colors hover:text-[var(--color-text)]"
            title={sortAsc ? 'Z → A' : 'A → Z'}
            data-testid="service-sort-toggle">
            {#if sortAsc}
              <ArrowDownAZ size={16} />
            {:else}
              <ArrowUpZA size={16} />
            {/if}
          </button>
        {/if}
        {#if availableToAdd.length > 0}
          <button onclick={() => showAddDialog = true}
            class="flex items-center justify-center rounded-lg border border-[var(--color-border)] p-1.5 text-[var(--color-text-secondary)] transition-colors hover:border-[var(--color-primary)] hover:text-[var(--color-primary)]"
            title={$t('settings.addService')}
            data-testid="service-add-btn">
            <Plus size={16} />
          </button>
        {/if}
      </div>
    </div>

    {#if configuredServices.length === 0}
      <div class="rounded-lg border border-dashed border-[var(--color-border)] p-6 text-center" data-testid="services-empty">
        <p class="mb-3 text-sm text-[var(--color-text-secondary)]">{$t('settings.noServices')}</p>
        <button onclick={() => showAddDialog = true}
          class="inline-flex items-center gap-1.5 rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white hover:bg-[var(--color-primary-hover)]"
          data-testid="service-add-empty-btn">
          <Plus size={16} />
          {$t('settings.addService')}
        </button>
      </div>
    {:else}
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {#each sortedConfigured as type (type)}
          {@const svc = services.find((s) => s.type === type)}
          {#if editingService === type}
            <div class="col-span-full">
              <ServiceCard
                {type}
                displayName={serviceDisplayName(serviceTypeDefs, type)}
                url={svc?.url || ''}
                status={$serviceStatuses[type]}
                isEditing={true}
                onEdit={() => startEditService(type)}
                onTest={() => testService(type)}
                onDelete={() => deleteService(type)}
              />
              <div class="space-y-3 rounded-b-lg border border-t-0 border-[var(--color-border)] px-4 py-4" data-testid="service-form-{type}">
                {#if needsURL(serviceTypeDefs, type)}
                  <div>
                    <label for="svc-url-{type}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.url')}</label>
                    <input id="svc-url-{type}" type="url" bind:value={serviceUrl} placeholder="https://..."
                      class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                      data-testid="service-url-input" />
                  </div>
                {/if}
                <ServiceAuthFields
                  serviceType={type}
                  {serviceTypeDefs}
                  bind:authMode={serviceAuthMode}
                  bind:token={serviceToken}
                  bind:username={serviceUsername}
                  bind:password={servicePassword}
                  bind:proxmoxTokenId
                  bind:proxmoxTokenSecret
                  bind:jdEmail
                  bind:jdPassword
                  idSuffix={type}
                />
                <div class="flex justify-end gap-2 pt-1">
                  <button onclick={() => editingService = null}
                    class="rounded-lg border border-[var(--color-border)] px-4 py-1.5 text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]"
                    data-testid="service-cancel">{$t('common.cancel')}</button>
                  <button onclick={saveService}
                    class="rounded-lg bg-[var(--color-primary)] px-4 py-1.5 text-sm text-white hover:bg-[var(--color-primary-hover)]"
                    data-testid="service-save">{$t('common.save')}</button>
                </div>
              </div>
            </div>
          {:else}
            <ServiceCard
              {type}
              displayName={serviceDisplayName(serviceTypeDefs, type)}
              url={svc?.url || ''}
              status={$serviceStatuses[type]}
              isEditing={false}
              onEdit={() => startEditService(type)}
              onTest={() => testService(type)}
              onDelete={() => deleteService(type)}
            />
          {/if}
        {/each}
      </div>
    {/if}

    <AddServiceDialog
      open={showAddDialog}
      {serviceTypeDefs}
      {availableToAdd}
      onClose={() => showAddDialog = false}
      onCreated={() => { loadServices(); showAddDialog = false }}
    />
  </section>

  <!-- go2rtc Streaming -->
  <Go2rtcSection />
</div>
