<script lang="ts">
  import { UserCircle, KeyRound } from 'lucide-svelte'
  import { api, ApiError } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { auth } from '../../stores/auth'
  import { t, translate } from '../../i18n'
  import PasswordInput from '../PasswordInput.svelte'
  import TotpSection from './TotpSection.svelte'

  // Username change
  let usernameCurrentPassword = $state('')
  let newUsername = $state($auth.username ?? '')
  let usernameLoading = $state(false)

  // Password change
  let passwordCurrentPassword = $state('')
  let newPassword = $state('')
  let confirmPassword = $state('')
  let passwordLoading = $state(false)

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
  <TotpSection />
</div>
