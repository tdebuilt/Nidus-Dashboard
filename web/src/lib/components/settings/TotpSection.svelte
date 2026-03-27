<script lang="ts">
  import { ShieldCheck } from 'lucide-svelte'
  import { api, ApiError } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { auth } from '../../stores/auth'
  import { t, translate } from '../../i18n'
  import { get } from 'svelte/store'

  let totpEnabled = $state(false)
  let totpQR = $state('')
  let totpSecret = $state('')
  let totpCode = $state('')
  let showTotpSetup = $state(false)

  const initialAuth = get(auth)
  if (initialAuth.totpEnabled !== undefined) totpEnabled = initialAuth.totpEnabled
  const stored = typeof localStorage !== 'undefined' ? localStorage.getItem('nidus-totp-enabled') : null
  if (stored !== null) totpEnabled = stored === 'true'

  async function handleTotpGenerate() {
    try {
      const result = await api.post<{ secret: string; url: string; qr: string }>(
        '/api/auth/totp/generate',
      )
      totpSecret = result.secret
      totpQR = result.qr
      showTotpSetup = true
    } catch (e) {
      if (e instanceof ApiError && e.status === 409) {
        totpEnabled = true
        if (typeof localStorage !== 'undefined') {
          localStorage.setItem('nidus-totp-enabled', 'true')
        }
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
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('nidus-totp-enabled', 'true')
      }
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
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('nidus-totp-enabled', 'false')
      }
      toasts.success(translate('settings.totpDisabled'))
    } catch {
      toasts.error(translate('settings.totpDisableError'))
    }
  }
</script>

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
