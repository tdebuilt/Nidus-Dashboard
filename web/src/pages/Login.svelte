<script lang="ts">
  import NidusLogo from '../lib/components/NidusLogo.svelte'
  import { api, ApiError } from '../lib/api/client'
  import { auth } from '../lib/stores/auth'
  import { navigate } from '../lib/stores/router'
  import { t, translate } from '../lib/i18n'

  let username = $state('')
  let password = $state('')
  let totpCode = $state('')
  let error = $state('')
  let loading = $state(false)
  let showTotp = $state(false)

  function handleLoginError(err: unknown) {
    if (err instanceof ApiError) {
      const msg = err.message.toLowerCase()
      if (msg.includes('totp') || msg.includes('2fa')) {
        showTotp = true
        error = translate('login.totpRequired')
      } else {
        error = err.message
      }
    } else {
      error = translate('login.error')
    }
  }

  function handleLoginSuccess(result: { user?: { id?: number; username?: string; role?: string; totp_enabled?: boolean } }) {
    const user = result?.user
    const userTotpEnabled = user?.totp_enabled ?? false
    const userRole = (user?.role ?? 'admin') as 'admin' | 'editor' | 'viewer'
    auth.set({
      authenticated: true,
      setupCompleted: true,
      loading: false,
      totpEnabled: userTotpEnabled,
      role: userRole,
      userId: user?.id,
      username: user?.username,
    })
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('nidus-totp-enabled', String(userTotpEnabled))
      localStorage.setItem('nidus-role', userRole)
    }
    navigate('/')
  }

  async function handleSubmit(e: Event) {
    e.preventDefault()
    error = ''
    loading = true
    try {
      const result = await api.post<{ user?: { id?: number; username?: string; role?: string; totp_enabled?: boolean } }>('/api/auth/login', {
        username,
        password,
        totp_code: showTotp ? totpCode : undefined,
      })
      handleLoginSuccess(result)
    } catch (err) {
      handleLoginError(err)
    } finally {
      loading = false
    }
  }
</script>

<div class="flex min-h-screen items-center justify-center p-4" data-testid="login-page">
  <div class="w-full max-w-sm rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-8">
    <div class="mb-6 flex flex-col items-center gap-3">
      <NidusLogo size={48} />
      <h1 class="text-2xl font-bold text-[var(--color-text)]">Nidus</h1>
    </div>

    {#if error}
      <div class="alert-error mb-4 rounded-lg border px-4 py-2 text-sm" data-testid="login-error">
        {error}
      </div>
    {/if}

    <form onsubmit={handleSubmit} data-testid="login-form">
      <div class="mb-4">
        <label for="username" class="mb-1 block text-sm text-[var(--color-text-secondary)]">{$t('login.username')}</label>
        <input
          id="username"
          type="text"
          bind:value={username}
          required
          autocomplete="username"
          class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
          data-testid="login-username"
        />
      </div>

      <div class="mb-4">
        <label for="password" class="mb-1 block text-sm text-[var(--color-text-secondary)]">{$t('login.password')}</label>
        <input
          id="password"
          type="password"
          bind:value={password}
          required
          autocomplete="current-password"
          class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
          data-testid="login-password"
        />
      </div>

      {#if showTotp}
        <div class="mb-4" data-testid="totp-field">
          <label for="totp" class="mb-1 block text-sm text-[var(--color-text-secondary)]">{$t('login.totpCode')}</label>
          <input
            id="totp"
            type="text"
            bind:value={totpCode}
            autocomplete="one-time-code"
            inputmode="numeric"
            maxlength="6"
            class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-center text-lg tracking-widest text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
            data-testid="login-totp"
          />
        </div>
      {/if}

      <button
        type="submit"
        disabled={loading}
        class="w-full rounded-lg bg-[var(--color-primary)] py-2 text-white transition-colors hover:bg-[var(--color-primary-hover)] disabled:opacity-50"
        data-testid="login-submit"
      >
        {loading ? $t('login.submitting') : $t('login.submit')}
      </button>
    </form>
  </div>
</div>
