<script lang="ts">
  import { X, ArrowDownAZ, ArrowUpZA } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { t, translate } from '../../i18n'
  import DynamicIcon from '../DynamicIcon.svelte'
  import { getServiceIcon } from './serviceIcons'

  interface ServiceTypeDef {
    type: string
    display_name: string
    auth_type: string
    needs_url: boolean
  }

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

  const sortedServices = $derived(
    [...availableToAdd].sort((a, b) => {
      const cmp = serviceDisplayName(a).localeCompare(serviceDisplayName(b))
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
    selectedType = type
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

  function buildCredentials(): string {
    if (hasNoAuth(selectedType)) return ''
    if (hasJDAuth(selectedType)) {
      return jdEmail ? JSON.stringify({ email: jdEmail, password: jdPassword }) : ''
    }
    if (hasTokenOnly(selectedType)) {
      return serviceToken ? JSON.stringify({ token: serviceToken }) : ''
    }
    if (hasUserPassOnly(selectedType)) {
      return serviceUsername
        ? JSON.stringify({ username: serviceUsername, password: servicePassword })
        : ''
    }
    if (hasApiKeyOnly(selectedType)) {
      return serviceToken ? JSON.stringify({ api_key: serviceToken }) : ''
    }
    if (hasPasswordOnly(selectedType)) {
      return servicePassword ? JSON.stringify({ password: servicePassword }) : ''
    }
    if (serviceAuthMode === 'token') {
      if (selectedType === 'proxmox') {
        if (!proxmoxTokenId || !proxmoxTokenSecret) return ''
        return JSON.stringify({ token: `${proxmoxTokenId}=${proxmoxTokenSecret}` })
      }
      return serviceToken ? JSON.stringify({ token: serviceToken }) : ''
    }
    return serviceUsername
      ? JSON.stringify({ username: serviceUsername, password: servicePassword })
      : ''
  }

  async function handleSave() {
    if (!selectedType) return
    loading = true
    try {
      const creds = buildCredentials()
      const url = hasJDAuth(selectedType) ? 'https://api.jdownloader.org' : serviceUrl
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
    <div class="w-full max-w-2xl max-h-[90vh] flex flex-col rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-6 shadow-2xl animate-[dialogIn_0.2s_ease-out]">
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
                <span class="text-[var(--color-text)]">{serviceDisplayName(type)}</span>
              </button>
            {/each}
          </div>
        {/if}
      {:else}
        <div class="flex flex-1 flex-col gap-4 overflow-hidden">
          <div class="flex items-center gap-2 shrink-0">
            <DynamicIcon name={getServiceIcon(selectedType)} size={20} class="text-[var(--color-primary)]" />
            <span class="text-sm font-medium text-[var(--color-text)]">{serviceDisplayName(selectedType)}</span>
          </div>

          <div class="flex-1 space-y-3 overflow-y-auto">
            {#if needsURL(selectedType)}
              <div>
                <label for="svc-url-add" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.url')}</label>
                <input id="svc-url-add" type="url" bind:value={serviceUrl} placeholder="https://..."
                  class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                  data-testid="service-url-input" />
              </div>
            {/if}

            {#if hasJDAuth(selectedType)}
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label for="svc-jd-email-add" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.jdEmail')}</label>
                  <input id="svc-jd-email-add" type="email" bind:value={jdEmail}
                    placeholder={$t('settings.jdEmailHint')}
                    class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                    data-testid="service-jd-email-input" />
                </div>
                <div>
                  <label for="svc-jd-pass-add" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.jdPassword')}</label>
                  <input id="svc-jd-pass-add" type="password" bind:value={jdPassword}
                    placeholder={$t('settings.jdPasswordHint')}
                    class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                    data-testid="service-jd-password-input" />
                </div>
              </div>
            {:else if hasNoAuth(selectedType)}
              <p class="text-xs text-[var(--color-text-muted)]">{$t('settings.noAuthNeeded')}</p>
            {:else if hasTokenOnly(selectedType)}
              <div>
                <label for="svc-token-add" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.token')}</label>
                <input id="svc-token-add" type="password" bind:value={serviceToken}
                  placeholder={$t('settings.haTokenHint')}
                  class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                  data-testid="service-token-input" />
              </div>
            {:else if hasUserPassOnly(selectedType)}
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label for="svc-user-add" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.username')}</label>
                  <input id="svc-user-add" type="text" bind:value={serviceUsername}
                    class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                    data-testid="service-username-input" />
                </div>
                <div>
                  <label for="svc-pass-add" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.password')}</label>
                  <input id="svc-pass-add" type="password" bind:value={servicePassword}
                    class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                    data-testid="service-password-input" />
                </div>
              </div>
            {:else if hasApiKeyOnly(selectedType)}
              <div>
                <label for="svc-apikey-add" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('services.apiKey')}</label>
                <input id="svc-apikey-add" type="password" bind:value={serviceToken}
                  placeholder={$t('services.apiKey')}
                  class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]" />
              </div>
            {:else if hasPasswordOnly(selectedType)}
              <div>
                <label for="svc-pass-add" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.password')}</label>
                <input id="svc-pass-add" type="password" bind:value={servicePassword}
                  placeholder={$t('settings.password')}
                  class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]" />
              </div>
            {:else if hasDualAuth(selectedType)}
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
                  {#if selectedType === 'proxmox'}
                    <div class="grid grid-cols-2 gap-3">
                      <div>
                        <label for="svc-pve-tokenid-add" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.proxmoxTokenId')}</label>
                        <input id="svc-pve-tokenid-add" type="text" bind:value={proxmoxTokenId}
                          placeholder={$t('settings.proxmoxTokenIdHint')}
                          class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                          data-testid="service-pve-tokenid" />
                      </div>
                      <div>
                        <label for="svc-pve-tokensecret-add" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.proxmoxTokenSecret')}</label>
                        <input id="svc-pve-tokensecret-add" type="password" bind:value={proxmoxTokenSecret}
                          placeholder={$t('settings.proxmoxTokenSecretHint')}
                          class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                          data-testid="service-pve-tokensecret" />
                      </div>
                    </div>
                  {:else}
                    <div>
                      <label for="svc-token-add" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.token')}</label>
                      <input id="svc-token-add" type="password" bind:value={serviceToken}
                        placeholder={$t('settings.portainerTokenHint')}
                        class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                        data-testid="service-token-input" />
                    </div>
                  {/if}
                {:else}
                  <div class="grid grid-cols-2 gap-3">
                    <div>
                      <label for="svc-user-add" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.username')}</label>
                      <input id="svc-user-add" type="text" bind:value={serviceUsername}
                        class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                        data-testid="service-username-input" />
                    </div>
                    <div>
                      <label for="svc-pass-add" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.password')}</label>
                      <input id="svc-pass-add" type="password" bind:value={servicePassword}
                        class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                        data-testid="service-password-input" />
                    </div>
                  </div>
                {/if}
              </div>
            {/if}
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
