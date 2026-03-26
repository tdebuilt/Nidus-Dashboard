<script lang="ts">
  import NidusLogo from '../lib/components/NidusLogo.svelte'
  import { api, ApiError } from '../lib/api/client'
  import { navigate } from '../lib/stores/router'
  import { t, translate } from '../lib/i18n'

  let username = $state('')
  let password = $state('')
  let confirmPassword = $state('')
  let error = $state('')
  let loading = $state(false)
  let success = $state(false)

  // Extract code from URL params
  const code = $derived(
    typeof window !== 'undefined'
      ? new URLSearchParams(window.location.search).get('code') ?? ''
      : '',
  )

  async function handleSubmit(e: Event) {
    e.preventDefault()
    error = ''

    if (password !== confirmPassword) {
      error = translate('register.passwordMismatch')
      return
    }
    if (password.length < 8) {
      error = translate('register.passwordTooShort')
      return
    }
    if (!code) {
      error = translate('register.noCode')
      return
    }

    loading = true
    try {
      await api.post('/api/auth/register', { username, password, code })
      success = true
    } catch (err) {
      error = err instanceof ApiError ? err.message : translate('register.error')
    } finally {
      loading = false
    }
  }
</script>

<div class="flex min-h-screen items-center justify-center p-4" data-testid="register-page">
  <div class="w-full max-w-sm rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-8">
    <div class="mb-6 flex flex-col items-center gap-3">
      <NidusLogo size={48} />
      <h1 class="text-2xl font-bold text-[var(--color-text)]">{$t('register.title')}</h1>
      <p class="text-sm text-[var(--color-text-muted)]">{$t('register.subtitle')}</p>
    </div>

    {#if success}
      <div class="text-center">
        <div class="mb-4 rounded-lg border border-green-500/30 bg-green-500/10 p-4">
          <p class="text-sm text-[var(--color-text)]">{$t('register.success')}</p>
        </div>
        <button
          onclick={() => navigate('/login')}
          class="w-full rounded-lg bg-[var(--color-primary)] py-2 text-white transition-colors hover:bg-[var(--color-primary-hover)]"
        >
          {$t('register.goToLogin')}
        </button>
      </div>
    {:else}
      {#if !code}
        <div class="rounded-lg border border-[var(--color-danger)]/30 bg-[var(--color-danger)]/10 p-4 text-center text-sm text-[var(--color-text)]">
          {$t('register.noCode')}
        </div>
      {:else}
        {#if error}
          <div class="alert-error mb-4 rounded-lg border px-4 py-2 text-sm">
            {error}
          </div>
        {/if}

        <form onsubmit={handleSubmit}>
          <div class="mb-4">
            <label for="reg-username" class="mb-1 block text-sm text-[var(--color-text-secondary)]">{$t('login.username')}</label>
            <input
              id="reg-username"
              type="text"
              bind:value={username}
              required
              autocomplete="username"
              class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
            />
          </div>

          <div class="mb-4">
            <label for="reg-password" class="mb-1 block text-sm text-[var(--color-text-secondary)]">{$t('login.password')}</label>
            <input
              id="reg-password"
              type="password"
              bind:value={password}
              required
              autocomplete="new-password"
              class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
            />
          </div>

          <div class="mb-4">
            <label for="reg-confirm" class="mb-1 block text-sm text-[var(--color-text-secondary)]">{$t('register.confirmPassword')}</label>
            <input
              id="reg-confirm"
              type="password"
              bind:value={confirmPassword}
              required
              autocomplete="new-password"
              class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            class="w-full rounded-lg bg-[var(--color-primary)] py-2 text-white transition-colors hover:bg-[var(--color-primary-hover)] disabled:opacity-50"
          >
            {loading ? $t('register.creating') : $t('register.submit')}
          </button>
        </form>

        <p class="mt-4 text-center text-xs text-[var(--color-text-muted)]">
          {$t('register.hasAccount')}
          <button onclick={() => navigate('/login')} class="text-[var(--color-primary)] hover:underline">{$t('register.login')}</button>
        </p>
      {/if}
    {/if}
  </div>
</div>
