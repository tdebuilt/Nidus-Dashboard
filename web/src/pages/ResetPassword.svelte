<script lang="ts">
  import NidusLogo from '../lib/components/NidusLogo.svelte'
  import PasswordInput from '../lib/components/PasswordInput.svelte'
  import { api, ApiError } from '../lib/api/client'
  import { navigate } from '../lib/stores/router'
  import { t, translate } from '../lib/i18n'

  let newPassword = $state('')
  let confirmPassword = $state('')
  let error = $state('')
  let loading = $state(false)
  let success = $state(false)

  const code = $derived(
    typeof window !== 'undefined'
      ? new URLSearchParams(window.location.search).get('code') ?? ''
      : '',
  )

  async function handleSubmit(e: Event) {
    e.preventDefault()
    error = ''

    if (newPassword !== confirmPassword) {
      error = translate('resetPassword.passwordMismatch')
      return
    }
    if (newPassword.length < 8) {
      error = translate('resetPassword.passwordTooShort')
      return
    }
    if (!code) {
      error = translate('resetPassword.noCode')
      return
    }

    loading = true
    try {
      await api.post('/api/auth/reset-password', { code, new_password: newPassword })
      success = true
    } catch (err) {
      error = err instanceof ApiError ? err.message : translate('resetPassword.error')
    } finally {
      loading = false
    }
  }
</script>

<div class="flex min-h-screen items-center justify-center p-4" data-testid="reset-password-page">
  <div
    class="w-full max-w-sm rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-8"
  >
    <div class="mb-6 flex flex-col items-center gap-3">
      <NidusLogo size={48} />
      <h1 class="text-2xl font-bold text-[var(--color-text)]">{$t('resetPassword.title')}</h1>
      <p class="text-sm text-[var(--color-text-muted)]">{$t('resetPassword.subtitle')}</p>
    </div>

    {#if success}
      <div class="text-center">
        <div class="mb-4 rounded-lg border border-green-500/30 bg-green-500/10 p-4">
          <p class="text-sm text-[var(--color-text)]">{$t('resetPassword.success')}</p>
        </div>
        <button
          onclick={() => navigate('/login')}
          class="w-full rounded-lg bg-[var(--color-primary)] py-2 text-white transition-colors hover:bg-[var(--color-primary-hover)]"
        >
          {$t('resetPassword.goToLogin')}
        </button>
      </div>
    {:else}
      {#if !code}
        <div
          class="rounded-lg border border-[var(--color-danger)]/30 bg-[var(--color-danger)]/10 p-4 text-center text-sm text-[var(--color-text)]"
        >
          {$t('resetPassword.noCode')}
        </div>
      {:else}
        {#if error}
          <div class="alert-error mb-4 rounded-lg border px-4 py-2 text-sm">
            {error}
          </div>
        {/if}

        <form onsubmit={handleSubmit}>
          <div class="mb-4">
            <label
              for="reset-password"
              class="mb-1 block text-sm text-[var(--color-text-secondary)]"
              >{$t('resetPassword.newPassword')}</label
            >
            <PasswordInput
              id="reset-password"
              bind:value={newPassword}
              autocomplete="new-password"
              required
              testid="reset-new-password"
            />
          </div>

          <div class="mb-4">
            <label
              for="reset-confirm"
              class="mb-1 block text-sm text-[var(--color-text-secondary)]"
              >{$t('resetPassword.confirmPassword')}</label
            >
            <PasswordInput
              id="reset-confirm"
              bind:value={confirmPassword}
              autocomplete="new-password"
              required
              testid="reset-confirm-password"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            class="w-full rounded-lg bg-[var(--color-primary)] py-2 text-white transition-colors hover:bg-[var(--color-primary-hover)] disabled:opacity-50"
          >
            {loading ? $t('resetPassword.submitting') : $t('resetPassword.submit')}
          </button>
        </form>
      {/if}
    {/if}
  </div>
</div>
