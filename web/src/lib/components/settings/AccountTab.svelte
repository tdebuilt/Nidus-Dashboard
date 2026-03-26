<script lang="ts">
  import { UserCircle, KeyRound, ShieldCheck } from 'lucide-svelte'
  import { api, ApiError } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { auth } from '../../stores/auth'
  import { t, translate } from '../../i18n'
  import { onMount } from 'svelte'
  import PasswordInput from '../PasswordInput.svelte'

  // Username change
  let usernameCurrentPassword = $state('')
  let newUsername = $state($auth.username ?? '')
  let usernameLoading = $state(false)

  // Password change
  let passwordCurrentPassword = $state('')
  let newPassword = $state('')
  let confirmPassword = $state('')
  let passwordLoading = $state(false)

  // 2FA state
  let totpEnabled = $state(false)
  let totpQR = $state('')
  let totpSecret = $state('')
  let totpCode = $state('')
  let showTotpSetup = $state(false)

  onMount(() => {
    const unsub = auth.subscribe((s) => {
      if (s.totpEnabled !== undefined) totpEnabled = s.totpEnabled
    })
    unsub()
    const stored = typeof localStorage !== 'undefined' ? localStorage.getItem('nidus-totp-enabled') : null
    if (stored !== null) totpEnabled = stored === 'true'
  })

  async function handleUsernameChange() {
    if (!usernameCurrentPassword) {
      toasts.error(translate('account.currentPasswordRequired'))
      return
    }
    if (!newUsername.trim()) {
      toasts.error(translate('account.usernameRequired'))
      return
    }
    if (newUsername === $auth.username) {
      toasts.error(translate('account.noChanges'))
      return
    }

    usernameLoading = true
    try {
      const result = await api.put<{ message: string; user: { username: string } }>(
        '/api/auth/account',
        { current_password: usernameCurrentPassword, username: newUsername.trim() },
      )
      auth.update((s) => ({ ...s, username: result.user.username }))
      usernameCurrentPassword = ''
      toasts.success(translate('account.updated'))
    } catch (err) {
      toasts.error(err instanceof ApiError ? err.message : translate('account.error'))
    } finally {
      usernameLoading = false
    }
  }

  async function handlePasswordChange() {
    if (!passwordCurrentPassword) {
      toasts.error(translate('account.currentPasswordRequired'))
      return
    }
    if (newPassword.length < 8) {
      toasts.error(translate('account.passwordTooShort'))
      return
    }
    if (newPassword !== confirmPassword) {
      toasts.error(translate('account.passwordMismatch'))
      return
    }

    passwordLoading = true
    try {
      await api.put('/api/auth/account', {
        current_password: passwordCurrentPassword,
        new_password: newPassword,
      })
      passwordCurrentPassword = ''
      newPassword = ''
      confirmPassword = ''
      toasts.success(translate('account.updated'))
    } catch (err) {
      toasts.error(err instanceof ApiError ? err.message : translate('account.error'))
    } finally {
      passwordLoading = false
    }
  }

  async function handleTotpGenerate() {
    try {
      const result = await api.post<{ secret: string; url: string; qr: string }>('/api/auth/totp/generate')
      totpSecret = result.secret
      totpQR = result.qr
      showTotpSetup = true
    } catch (e) {
      if (e instanceof ApiError && e.status === 409) {
        totpEnabled = true
        if (typeof localStorage !== 'undefined') localStorage.setItem('nidus-totp-enabled', 'true')
      } else {
        toasts.error(translate('settings.totpGenError'))
      }
    }
  }

  async function handleTotpEnable() {
    if (!totpCode || totpCode.length !== 6) return
    try {
      await api.post('/api/auth/totp/enable', { code: totpCode })
      totpEnabled = true
      showTotpSetup = false
      totpCode = ''
      auth.update((s) => ({ ...s, totpEnabled: true }))
      if (typeof localStorage !== 'undefined') localStorage.setItem('nidus-totp-enabled', 'true')
      toasts.success(translate('settings.totpEnabled'))
    } catch {
      toasts.error(translate('settings.totpInvalidCode'))
    }
  }

  async function handleTotpDisable() {
    try {
      await api.delete('/api/auth/totp')
      totpEnabled = false
      auth.update((s) => ({ ...s, totpEnabled: false }))
      if (typeof localStorage !== 'undefined') localStorage.setItem('nidus-totp-enabled', 'false')
      toasts.success(translate('settings.totpDisabled'))
    } catch {
      toasts.error(translate('settings.totpDisableError'))
    }
  }
</script>

<div class="space-y-6" data-testid="settings-account">
  <!-- Change Username -->
  <section
    class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5"
    data-testid="account-username"
  >
    <div class="mb-4 flex items-center gap-2">
      <UserCircle size={18} class="text-[var(--color-text-secondary)]" />
      <h3 class="font-semibold text-[var(--color-text)]">{$t('account.changeUsername')}</h3>
    </div>

    <form
      onsubmit={(e) => {
        e.preventDefault()
        handleUsernameChange()
      }}
      class="space-y-3"
    >
      <div>
        <label for="username-current-pw" class="mb-1 block text-sm text-[var(--color-text-secondary)]"
          >{$t('account.currentPassword')}</label
        >
        <PasswordInput
          id="username-current-pw"
          bind:value={usernameCurrentPassword}
          autocomplete="current-password"
          required
          testid="account-username-current-pw"
        />
      </div>
      <div>
        <label for="new-username" class="mb-1 block text-sm text-[var(--color-text-secondary)]"
          >{$t('account.newUsername')}</label
        >
        <input
          id="new-username"
          type="text"
          bind:value={newUsername}
          required
          autocomplete="username"
          data-testid="account-new-username"
          class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
        />
      </div>
      <button
        type="submit"
        disabled={usernameLoading}
        class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white transition-colors hover:bg-[var(--color-primary-hover)] disabled:opacity-50"
        data-testid="account-username-submit"
      >
        {$t('common.save')}
      </button>
    </form>
  </section>

  <!-- Change Password -->
  <section
    class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5"
    data-testid="account-password"
  >
    <div class="mb-4 flex items-center gap-2">
      <KeyRound size={18} class="text-[var(--color-text-secondary)]" />
      <h3 class="font-semibold text-[var(--color-text)]">{$t('account.changePassword')}</h3>
    </div>

    <form
      onsubmit={(e) => {
        e.preventDefault()
        handlePasswordChange()
      }}
      class="space-y-3"
    >
      <div>
        <label for="password-current-pw" class="mb-1 block text-sm text-[var(--color-text-secondary)]"
          >{$t('account.currentPassword')}</label
        >
        <PasswordInput
          id="password-current-pw"
          bind:value={passwordCurrentPassword}
          autocomplete="current-password"
          required
          testid="account-password-current-pw"
        />
      </div>
      <div>
        <label for="new-password" class="mb-1 block text-sm text-[var(--color-text-secondary)]"
          >{$t('account.newPassword')}</label
        >
        <PasswordInput
          id="new-password"
          bind:value={newPassword}
          autocomplete="new-password"
          required
          testid="account-new-password"
        />
      </div>
      <div>
        <label for="confirm-password" class="mb-1 block text-sm text-[var(--color-text-secondary)]"
          >{$t('account.confirmPassword')}</label
        >
        <PasswordInput
          id="confirm-password"
          bind:value={confirmPassword}
          autocomplete="new-password"
          required
          testid="account-confirm-password"
        />
      </div>
      <button
        type="submit"
        disabled={passwordLoading}
        class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white transition-colors hover:bg-[var(--color-primary-hover)] disabled:opacity-50"
        data-testid="account-password-submit"
      >
        {$t('common.save')}
      </button>
    </form>
  </section>

  <!-- 2FA -->
  <section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-2fa">
    <div class="mb-3 flex items-center gap-2">
      <ShieldCheck size={18} class="text-[var(--color-text-secondary)]" />
      <h3 class="font-semibold text-[var(--color-text)]">{$t('settings.twoFaSection')}</h3>
    </div>
    {#if showTotpSetup}
      <div data-testid="totp-setup">
        <p class="mb-3 text-sm text-[var(--color-text-secondary)]">{$t('settings.totpScanHint')}</p>
        {#if totpQR}
          <div class="mb-3 flex justify-center">
            <img src={totpQR} alt={$t('settings.totpQrAlt')} class="rounded-lg" data-testid="totp-qr" />
          </div>
        {/if}
        <p class="mb-3 text-xs text-[var(--color-text-muted)]">{$t('settings.totpSecret', { secret: totpSecret })}</p>
        <div class="flex items-center gap-2">
          <input type="text" bind:value={totpCode} placeholder={$t('settings.totpCodePlaceholder')} maxlength="6" inputmode="numeric"
            class="w-32 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-center text-[var(--color-text)] outline-none"
            data-testid="totp-code-input" />
          <button onclick={handleTotpEnable}
            class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white hover:bg-[var(--color-primary-hover)]"
            data-testid="totp-enable-btn">{$t('common.enable')}</button>
          <button onclick={() => showTotpSetup = false}
            class="rounded-lg border border-[var(--color-border)] px-4 py-2 text-sm text-[var(--color-text-secondary)]"
            data-testid="totp-cancel-btn">{$t('common.cancel')}</button>
        </div>
      </div>
    {:else}
      <div class="flex items-center gap-4">
        <span class="text-sm text-[var(--color-text-secondary)]">{totpEnabled ? $t('common.enabled') : $t('common.disabled')}</span>
        {#if totpEnabled}
          <button onclick={handleTotpDisable}
            class="rounded-lg border border-[var(--color-danger)] px-4 py-1.5 text-sm text-[var(--color-danger)] hover:bg-[var(--color-error-bg)]"
            data-testid="totp-disable-btn">{$t('common.disable')}</button>
        {:else}
          <button onclick={handleTotpGenerate}
            class="rounded-lg bg-[var(--color-primary)] px-4 py-1.5 text-sm text-white hover:bg-[var(--color-primary-hover)]"
            data-testid="totp-setup-btn">{$t('common.configure')}</button>
        {/if}
      </div>
    {/if}
  </section>
</div>
