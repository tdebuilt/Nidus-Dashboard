<script lang="ts">
  import { t } from '../../i18n'
  import {
    type ServiceTypeDef,
    hasJDAuth,
    hasNoAuth,
    hasTokenOnly,
    hasUserPassOnly,
    hasApiKeyOnly,
    hasPasswordOnly,
    hasDualAuth,
  } from './serviceAuth'

  const inputClass = 'w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]'

  interface Props {
    serviceType: string
    serviceTypeDefs: ServiceTypeDef[]
    authMode: 'token' | 'userpass'
    token: string
    username: string
    password: string
    proxmoxTokenId: string
    proxmoxTokenSecret: string
    jdEmail: string
    jdPassword: string
    idSuffix: string
  }

  let {
    serviceType,
    serviceTypeDefs,
    authMode = $bindable(),
    token = $bindable(),
    username = $bindable(),
    password = $bindable(),
    proxmoxTokenId = $bindable(),
    proxmoxTokenSecret = $bindable(),
    jdEmail = $bindable(),
    jdPassword = $bindable(),
    idSuffix,
  }: Props = $props()
</script>

{#if hasJDAuth(serviceTypeDefs, serviceType)}
  <div class="grid grid-cols-2 gap-3">
    <div>
      <label for="svc-jd-email-{idSuffix}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.jdEmail')}</label>
      <input id="svc-jd-email-{idSuffix}" type="email" bind:value={jdEmail}
        placeholder={$t('settings.jdEmailHint')}
        class={inputClass}
        data-testid="service-jd-email-input" />
    </div>
    <div>
      <label for="svc-jd-pass-{idSuffix}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.jdPassword')}</label>
      <input id="svc-jd-pass-{idSuffix}" type="password" bind:value={jdPassword}
        placeholder={$t('settings.jdPasswordHint')}
        class={inputClass}
        data-testid="service-jd-password-input" />
    </div>
  </div>
{:else if hasNoAuth(serviceTypeDefs, serviceType)}
  <p class="text-xs text-[var(--color-text-muted)]">{$t('settings.noAuthNeeded')}</p>
{:else if hasTokenOnly(serviceTypeDefs, serviceType)}
  <div>
    <label for="svc-token-{idSuffix}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.token')}</label>
    <input id="svc-token-{idSuffix}" type="password" bind:value={token}
      placeholder={$t('settings.haTokenHint')}
      class={inputClass}
      data-testid="service-token-input" />
  </div>
{:else if hasUserPassOnly(serviceTypeDefs, serviceType)}
  <div class="grid grid-cols-2 gap-3">
    <div>
      <label for="svc-user-{idSuffix}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.username')}</label>
      <input id="svc-user-{idSuffix}" type="text" bind:value={username}
        class={inputClass}
        data-testid="service-username-input" />
    </div>
    <div>
      <label for="svc-pass-{idSuffix}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.password')}</label>
      <input id="svc-pass-{idSuffix}" type="password" bind:value={password}
        class={inputClass}
        data-testid="service-password-input" />
    </div>
  </div>
{:else if hasApiKeyOnly(serviceTypeDefs, serviceType)}
  <div>
    <label for="svc-apikey-{idSuffix}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('services.apiKey')}</label>
    <input id="svc-apikey-{idSuffix}" type="password" bind:value={token}
      placeholder={$t('services.apiKey')}
      class={inputClass} />
  </div>
{:else if hasPasswordOnly(serviceTypeDefs, serviceType)}
  <div>
    <label for="svc-pass-{idSuffix}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.password')}</label>
    <input id="svc-pass-{idSuffix}" type="password" bind:value={password}
      placeholder={$t('settings.password')}
      class={inputClass} />
  </div>
{:else if hasDualAuth(serviceTypeDefs, serviceType)}
  <div>
    <span class="mb-2 block text-xs text-[var(--color-text-secondary)]">{$t('settings.authMode')}</span>
    <div class="mb-3 flex gap-2">
      <button
        onclick={() => authMode = 'token'}
        class="rounded-lg px-3 py-1 text-sm transition-colors"
        class:bg-[var(--color-primary)]={authMode === 'token'}
        class:text-white={authMode === 'token'}
        class:bg-[var(--color-bg-tertiary)]={authMode !== 'token'}
        class:text-[var(--color-text-secondary)]={authMode !== 'token'}
        data-testid="service-auth-token"
      >{$t('settings.authToken')}</button>
      <button
        onclick={() => authMode = 'userpass'}
        class="rounded-lg px-3 py-1 text-sm transition-colors"
        class:bg-[var(--color-primary)]={authMode === 'userpass'}
        class:text-white={authMode === 'userpass'}
        class:bg-[var(--color-bg-tertiary)]={authMode !== 'userpass'}
        class:text-[var(--color-text-secondary)]={authMode !== 'userpass'}
        data-testid="service-auth-userpass"
      >{$t('settings.authUserPass')}</button>
    </div>

    {#if authMode === 'token'}
      {#if serviceType === 'proxmox'}
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label for="svc-pve-tokenid-{idSuffix}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.proxmoxTokenId')}</label>
            <input id="svc-pve-tokenid-{idSuffix}" type="text" bind:value={proxmoxTokenId}
              placeholder={$t('settings.proxmoxTokenIdHint')}
              class={inputClass}
              data-testid="service-pve-tokenid" />
          </div>
          <div>
            <label for="svc-pve-tokensecret-{idSuffix}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.proxmoxTokenSecret')}</label>
            <input id="svc-pve-tokensecret-{idSuffix}" type="password" bind:value={proxmoxTokenSecret}
              placeholder={$t('settings.proxmoxTokenSecretHint')}
              class={inputClass}
              data-testid="service-pve-tokensecret" />
          </div>
        </div>
      {:else}
        <div>
          <label for="svc-token-{idSuffix}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.token')}</label>
          <input id="svc-token-{idSuffix}" type="password" bind:value={token}
            placeholder={$t('settings.portainerTokenHint')}
            class={inputClass}
            data-testid="service-token-input" />
        </div>
      {/if}
    {:else}
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label for="svc-user-{idSuffix}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.username')}</label>
          <input id="svc-user-{idSuffix}" type="text" bind:value={username}
            class={inputClass}
            data-testid="service-username-input" />
        </div>
        <div>
          <label for="svc-pass-{idSuffix}" class="mb-1 block text-xs text-[var(--color-text-secondary)]">{$t('settings.password')}</label>
          <input id="svc-pass-{idSuffix}" type="password" bind:value={password}
            class={inputClass}
            data-testid="service-password-input" />
        </div>
      </div>
    {/if}
  </div>
{/if}
