<script lang="ts">
  import { Server, Plus, Video, Play, Square, RotateCw, Loader2 } from 'lucide-svelte'
  import { api, ApiError } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { confirm } from '../../stores/confirm'
  import { t, translate } from '../../i18n'
  import { onMount, onDestroy } from 'svelte'
  import ServiceCard from './ServiceCard.svelte'
  import { getServiceGroup, groupOrder } from './serviceIcons'
  import { serviceStatuses, startServiceStatusPolling, stopServiceStatusPolling } from '../../stores/serviceStatus'
  import AddServiceDialog from './AddServiceDialog.svelte'

  // Service type definitions
  interface ServiceTypeDef {
    type: string
    display_name: string
    auth_type: string
    needs_url: boolean
  }

  let services = $state<Array<{ type: string; name: string; url: string; enabled: boolean; has_credentials: boolean }>>([])
  let _servicesLoading = $state(true)
  let serviceTypeDefs = $state<ServiceTypeDef[]>([])

  let showAddDialog = $state(false)
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

  // go2rtc
  interface Go2RTCStatus {
    available: boolean
    running: boolean
    uptime?: string
    cameras: number
  }

  let go2rtcStatus = $state<Go2RTCStatus>({ available: false, running: false, cameras: 0 })
  let go2rtcLoading = $state(false)
  let go2rtcPollTimer: ReturnType<typeof setInterval> | null = null

  const serviceTypes = $derived(serviceTypeDefs.map(d => d.type))
  const configuredServices = $derived(serviceTypes.filter(type => services.some(s => s.type === type)))
  const availableToAdd = $derived(serviceTypes.filter(type => !services.some(s => s.type === type)))

  const groupedServices = $derived((() => {
    const groups: Record<string, string[]> = {}
    for (const type of configuredServices) {
      const group = getServiceGroup(type)
      if (!groups[group]) groups[group] = []
      groups[group].push(type)
    }
    const distinctGroups = Object.keys(groups).length
    return groupOrder
      .filter(g => groups[g])
      .map(g => ({ name: g, services: groups[g], showHeader: distinctGroups >= 2 }))
  })())

  onMount(() => {
    loadServiceTypes()
    loadServices()
    loadGo2rtcStatus()
    go2rtcPollTimer = setInterval(loadGo2rtcStatus, 10000)
  })

  onDestroy(() => {
    stopServiceStatusPolling()
    if (go2rtcPollTimer) clearInterval(go2rtcPollTimer)
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
      if (withUrl.length > 0) {
        startServiceStatusPolling(withUrl)
      }
    }
    catch { services = [] }
    finally { _servicesLoading = false }
  }

  async function loadGo2rtcStatus() {
    try { go2rtcStatus = await api.get<Go2RTCStatus>('/api/go2rtc/status') }
    catch { go2rtcStatus = { available: false, running: false, cameras: 0 } }
  }

  async function go2rtcAction(action: 'start' | 'stop' | 'restart') {
    go2rtcLoading = true
    try { go2rtcStatus = await api.post<Go2RTCStatus>(`/api/go2rtc/${action}`) }
    catch { toasts.error(translate('go2rtc.actionError')) }
    finally { go2rtcLoading = false }
  }

  function getServiceDef(type: string): ServiceTypeDef | undefined {
    return serviceTypeDefs.find(d => d.type === type)
  }

  function serviceDisplayName(type: string): string {
    return getServiceDef(type)?.display_name || type.charAt(0).toUpperCase() + type.slice(1)
  }

  function getAuthType(type: string): string {
    return getServiceDef(type)?.auth_type || 'none'
  }

  function needsURL(type: string): boolean {
    return getServiceDef(type)?.needs_url ?? true
  }

  function hasDualAuth(type: string): boolean { return getAuthType(type) === 'dual' }
  function hasTokenOnly(type: string): boolean { return getAuthType(type) === 'token' }
  function hasUserPassOnly(type: string): boolean { return getAuthType(type) === 'userpass' }
  function hasApiKeyOnly(type: string): boolean { return getAuthType(type) === 'apikey' }
  function hasPasswordOnly(type: string): boolean { return getAuthType(type) === 'password' }
  function hasJDAuth(type: string): boolean { return getAuthType(type) === 'jdownloader' }
  function hasNoAuth(type: string): boolean { return getAuthType(type) === 'none' }

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

  function buildCredentials(): string {
    if (!editingService) return ''
    if (hasNoAuth(editingService)) return ''
    if (hasJDAuth(editingService)) {
      return jdEmail ? JSON.stringify({ email: jdEmail, password: jdPassword }) : ''
    }
    if (hasTokenOnly(editingService)) {
      return serviceToken ? JSON.stringify({ token: serviceToken }) : ''
    }
    if (hasUserPassOnly(editingService)) {
      return serviceUsername
        ? JSON.stringify({ username: serviceUsername, password: servicePassword })
        : ''
    }
    if (hasApiKeyOnly(editingService)) {
      return serviceToken ? JSON.stringify({ api_key: serviceToken }) : ''
    }
    if (hasPasswordOnly(editingService)) {
      return servicePassword ? JSON.stringify({ password: servicePassword }) : ''
    }
    if (serviceAuthMode === 'token') {
      if (editingService === 'proxmox') {
        if (!proxmoxTokenId || !proxmoxTokenSecret) return ''
        return JSON.stringify({ token: `${proxmoxTokenId}=${proxmoxTokenSecret}` })
      }
      return serviceToken ? JSON.stringify({ token: serviceToken }) : ''
    }
    return serviceUsername
      ? JSON.stringify({ username: serviceUsername, password: servicePassword })
      : ''
  }

  async function saveService() {
    if (!editingService) return
    try {
      const creds = buildCredentials()
      const url = hasJDAuth(editingService) ? 'https://api.jdownloader.org' : serviceUrl
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
      message: translate('settings.deleteServiceConfirm', { name: serviceDisplayName(type) }),
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

{#snippet authFields(type: string)}
  {#if hasJDAuth(type)}
    <div class="grid grid-cols-2 gap-3">
      <div>
        <label for="svc-jd-email-{type}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.jdEmail')}</label>
        <input id="svc-jd-email-{type}" type="email" bind:value={jdEmail}
          placeholder={$t('settings.jdEmailHint')}
          class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
          data-testid="service-jd-email-input" />
      </div>
      <div>
        <label for="svc-jd-pass-{type}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.jdPassword')}</label>
        <input id="svc-jd-pass-{type}" type="password" bind:value={jdPassword}
          placeholder={$t('settings.jdPasswordHint')}
          class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
          data-testid="service-jd-password-input" />
      </div>
    </div>
  {:else if hasNoAuth(type)}
    <p class="text-xs text-[var(--color-text-muted)]">{$t('settings.noAuthNeeded')}</p>
  {:else if hasTokenOnly(type)}
    <div>
      <label for="svc-token-{type}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.token')}</label>
      <input id="svc-token-{type}" type="password" bind:value={serviceToken}
        placeholder={$t('settings.haTokenHint')}
        class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
        data-testid="service-token-input" />
    </div>
  {:else if hasUserPassOnly(type)}
    <div class="grid grid-cols-2 gap-3">
      <div>
        <label for="svc-user-{type}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.username')}</label>
        <input id="svc-user-{type}" type="text" bind:value={serviceUsername}
          class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
          data-testid="service-username-input" />
      </div>
      <div>
        <label for="svc-pass-{type}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.password')}</label>
        <input id="svc-pass-{type}" type="password" bind:value={servicePassword}
          class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
          data-testid="service-password-input" />
      </div>
    </div>
  {:else if hasApiKeyOnly(type)}
    <div>
      <label for="svc-apikey-{type}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('services.apiKey')}</label>
      <input id="svc-apikey-{type}" type="password" bind:value={serviceToken}
        placeholder={$t('services.apiKey')}
        class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]" />
    </div>
  {:else if hasPasswordOnly(type)}
    <div>
      <label for="svc-pass-{type}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.password')}</label>
      <input id="svc-pass-{type}" type="password" bind:value={servicePassword}
        placeholder={$t('settings.password')}
        class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]" />
    </div>
  {:else if hasDualAuth(type)}
    <div>
      <span class="mb-2 block text-xs text-[var(--color-text-secondary)]">{$t('settings.authMode')}</span>
      <div class="mb-3 flex gap-2">
        <button
          onclick={() => serviceAuthMode = 'token'}
          class="rounded-lg px-3 py-1 text-sm transition-colors"
          class:bg-[var(--color-primary)]={serviceAuthMode === 'token'}
          class:text-white={serviceAuthMode === 'token'}
          class:bg-[var(--color-bg-tertiary)]={serviceAuthMode !== 'token'}
          class:text-[var(--color-text-secondary)]={serviceAuthMode !== 'token'}
          data-testid="service-auth-token"
        >{$t('settings.authToken')}</button>
        <button
          onclick={() => serviceAuthMode = 'userpass'}
          class="rounded-lg px-3 py-1 text-sm transition-colors"
          class:bg-[var(--color-primary)]={serviceAuthMode === 'userpass'}
          class:text-white={serviceAuthMode === 'userpass'}
          class:bg-[var(--color-bg-tertiary)]={serviceAuthMode !== 'userpass'}
          class:text-[var(--color-text-secondary)]={serviceAuthMode !== 'userpass'}
          data-testid="service-auth-userpass"
        >{$t('settings.authUserPass')}</button>
      </div>

      {#if serviceAuthMode === 'token'}
        {#if type === 'proxmox'}
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label for="svc-pve-tokenid" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.proxmoxTokenId')}</label>
              <input id="svc-pve-tokenid" type="text" bind:value={proxmoxTokenId}
                placeholder={$t('settings.proxmoxTokenIdHint')}
                class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                data-testid="service-pve-tokenid" />
            </div>
            <div>
              <label for="svc-pve-tokensecret" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.proxmoxTokenSecret')}</label>
              <input id="svc-pve-tokensecret" type="password" bind:value={proxmoxTokenSecret}
                placeholder={$t('settings.proxmoxTokenSecretHint')}
                class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                data-testid="service-pve-tokensecret" />
            </div>
          </div>
        {:else}
          <div>
            <label for="svc-token-{type}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.token')}</label>
            <input id="svc-token-{type}" type="password" bind:value={serviceToken}
              placeholder={$t('settings.portainerTokenHint')}
              class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
              data-testid="service-token-input" />
          </div>
        {/if}
      {:else}
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label for="svc-user-{type}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.username')}</label>
            <input id="svc-user-{type}" type="text" bind:value={serviceUsername}
              class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
              data-testid="service-username-input" />
          </div>
          <div>
            <label for="svc-pass-{type}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.password')}</label>
            <input id="svc-pass-{type}" type="password" bind:value={servicePassword}
              class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
              data-testid="service-password-input" />
          </div>
        </div>
      {/if}
    </div>
  {/if}
{/snippet}

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
      {#if availableToAdd.length > 0}
        <button onclick={() => showAddDialog = true}
          class="ml-auto flex items-center justify-center rounded-lg border border-[var(--color-border)] p-1.5 text-[var(--color-text-secondary)] transition-colors hover:border-[var(--color-primary)] hover:text-[var(--color-primary)]"
          title={$t('settings.addService')}
          data-testid="service-add-btn">
          <Plus size={16} />
        </button>
      {/if}
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
      <div class="space-y-4">
        {#each groupedServices as group (group.name)}
          <div>
            {#if group.showHeader}
              <div class="mb-2 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">
                {$t(`settings.serviceGroup.${group.name}`)}
              </div>
            {/if}
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {#each group.services as type (type)}
                {@const svc = services.find((s) => s.type === type)}
                {#if editingService === type}
                  <div class="col-span-full">
                    <ServiceCard
                      {type}
                      displayName={serviceDisplayName(type)}
                      url={svc?.url || ''}
                      status={$serviceStatuses[type]}
                      isEditing={true}
                      onEdit={() => startEditService(type)}
                      onTest={() => testService(type)}
                      onDelete={() => deleteService(type)}
                    />
                    <div class="space-y-3 rounded-b-lg border border-t-0 border-[var(--color-border)] px-4 py-4" data-testid="service-form-{type}">
                      {#if needsURL(type)}
                        <div>
                          <label for="svc-url-{type}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.url')}</label>
                          <input id="svc-url-{type}" type="url" bind:value={serviceUrl} placeholder="https://..."
                            class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                            data-testid="service-url-input" />
                        </div>
                      {/if}
                      {@render authFields(type)}
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
                    displayName={serviceDisplayName(type)}
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
          </div>
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
  {#if go2rtcStatus.available}
  <section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-go2rtc">
    <div class="mb-3 flex items-center gap-2">
      <Video size={18} class="text-[var(--color-text-secondary)]" />
      <h3 class="font-semibold text-[var(--color-text)]">{$t('go2rtc.title')}</h3>
    </div>

    <div class="space-y-3">
      <div class="flex items-center gap-3">
        <span class="inline-block h-2.5 w-2.5 rounded-full {go2rtcStatus.running ? 'bg-green-500' : 'bg-red-500'}"></span>
        <span class="text-sm text-[var(--color-text)]">
          {go2rtcStatus.running ? $t('go2rtc.running') : $t('go2rtc.stopped')}
        </span>
        {#if go2rtcStatus.running && go2rtcStatus.uptime}
          <span class="text-xs text-[var(--color-text-muted)]">({go2rtcStatus.uptime})</span>
        {/if}
        {#if go2rtcStatus.cameras > 0}
          <span class="text-xs text-[var(--color-text-muted)]">— {go2rtcStatus.cameras} {$t('go2rtc.cameras')}</span>
        {/if}
      </div>

      <div class="flex gap-2">
        {#if !go2rtcStatus.running}
          <button
            onclick={() => go2rtcAction('start')}
            disabled={go2rtcLoading}
            class="flex items-center gap-1.5 rounded-lg bg-green-600 px-3 py-1.5 text-xs text-white hover:bg-green-700 disabled:opacity-50"
          >
            {#if go2rtcLoading}<Loader2 size={12} class="animate-spin" />{:else}<Play size={12} />{/if}
            {$t('go2rtc.start')}
          </button>
        {:else}
          <button
            onclick={() => go2rtcAction('stop')}
            disabled={go2rtcLoading}
            class="flex items-center gap-1.5 rounded-lg bg-red-600 px-3 py-1.5 text-xs text-white hover:bg-red-700 disabled:opacity-50"
          >
            {#if go2rtcLoading}<Loader2 size={12} class="animate-spin" />{:else}<Square size={12} />{/if}
            {$t('go2rtc.stop')}
          </button>
          <button
            onclick={() => go2rtcAction('restart')}
            disabled={go2rtcLoading}
            class="flex items-center gap-1.5 rounded-lg bg-[var(--color-bg-tertiary)] px-3 py-1.5 text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-primary)] disabled:opacity-50"
          >
            {#if go2rtcLoading}<Loader2 size={12} class="animate-spin" />{:else}<RotateCw size={12} />{/if}
            {$t('go2rtc.restart')}
          </button>
        {/if}
      </div>

      <p class="text-xs text-[var(--color-text-muted)]">{$t('go2rtc.hint')}</p>
    </div>
  </section>
  {/if}
</div>
