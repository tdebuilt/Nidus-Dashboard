<script lang="ts">
  import { UserPlus, FolderPlus, ChevronRight, ChevronLeft, Check, Globe } from 'lucide-svelte'
  import NidusLogo from '../lib/components/NidusLogo.svelte'
  import { api, ApiError } from '../lib/api/client'
  import { auth } from '../lib/stores/auth'
  import { navigate } from '../lib/stores/router'
  import { toasts } from '../lib/stores/toast'
  import { t, translate, locale, setLocale, getAvailableLocales } from '../lib/i18n'

  let step = $state(0)
  let loading = $state(false)
  let error = $state('')

  // Step 0: Admin account
  let adminUsername = $state('')
  let adminPassword = $state('')
  let adminPasswordConfirm = $state('')

  // Step 1: First category
  let categoryName = $state('')
  const categoryIcon = $state('folder')

  const stepKeys = ['setup.stepAdmin', 'setup.stepCategory']
  const _stepIcons = [UserPlus, FolderPlus]

  async function handleNext() {
    error = ''
    loading = true

    try {
      if (step === 0) {
        if (adminPassword !== adminPasswordConfirm) {
          error = translate('setup.passwordMismatch')
          loading = false
          return
        }
        if (adminPassword.length < 8) {
          error = translate('setup.passwordTooShort')
          loading = false
          return
        }
        await api.post('/api/auth/setup', { username: adminUsername, password: adminPassword })
        // Login after setup
        await api.post('/api/auth/login', { username: adminUsername, password: adminPassword })
      } else if (step === 1) {
        if (categoryName) {
          await api.post('/api/categories', { name: categoryName, icon: categoryIcon })
        }
        // Finish setup
        auth.set({ authenticated: true, setupCompleted: true, loading: false, role: 'admin' })
        toasts.success(translate('setup.setupComplete'))
        navigate('/')
        loading = false
        return
      }

      step++
    } catch (err) {
      error = err instanceof ApiError ? err.message : 'Une erreur est survenue'
    } finally {
      loading = false
    }
  }

  function handleBack() {
    if (step > 0) step--
    error = ''
  }
</script>

<div class="flex min-h-screen items-center justify-center p-4" data-testid="setup-page">
  <div class="w-full max-w-lg rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-8">
    <!-- Language selector -->
    <div class="mb-4 flex justify-end">
      <div class="flex items-center gap-1 rounded-lg border border-[var(--color-border)] p-1" data-testid="setup-language">
        <Globe size={14} class="ms-1 text-[var(--color-text-muted)]" />
        {#each getAvailableLocales() as loc (loc.code)}
          <button
            onclick={() => setLocale(loc.code)}
            class="rounded px-2 py-1 text-xs font-medium transition-colors {$locale === loc.code ? 'bg-[var(--color-primary)] text-white' : 'text-[var(--color-text-muted)] hover:text-[var(--color-text)]'}"
            data-testid="setup-lang-{loc.code}"
          >{loc.code.toUpperCase()}</button>
        {/each}
      </div>
    </div>

    <!-- Logo -->
    <div class="mb-6 flex flex-col items-center gap-2">
      <NidusLogo size={48} />
      <span class="text-xl font-bold text-[var(--color-text)]">Nidus</span>
    </div>

    <!-- Progress -->
    <div class="mb-8 flex items-center justify-center gap-2" data-testid="setup-progress">
      {#each stepKeys as _, i (i)}
        <div class="flex items-center gap-2">
          <div
            class="flex h-8 w-8 items-center justify-center rounded-full text-xs font-bold
              {i < step ? 'bg-[var(--color-success)] text-white' :
               i === step ? 'bg-[var(--color-primary)] text-white' :
               'bg-[var(--color-bg-tertiary)] text-[var(--color-text-muted)]'}"
          >
            {#if i < step}
              <Check size={14} />
            {:else}
              {i + 1}
            {/if}
          </div>
          {#if i < stepKeys.length - 1}
            <div class="h-0.5 w-6 {i < step ? 'bg-[var(--color-success)]' : 'bg-[var(--color-bg-tertiary)]'}"></div>
          {/if}
        </div>
      {/each}
    </div>

    <h2 class="mb-1 text-xl font-bold text-[var(--color-text)]">{$t(stepKeys[step])}</h2>
    <p class="mb-6 text-sm text-[var(--color-text-muted)]">{$t('setup.progress', { current: step + 1, total: stepKeys.length })}</p>

    {#if error}
      <div class="alert-error mb-4 rounded-lg border px-4 py-2 text-sm" data-testid="setup-error">
        {error}
      </div>
    {/if}

    <!-- Step content -->
    {#if step === 0}
      <div data-testid="setup-step-admin">
        <div class="mb-4">
          <label for="admin-user" class="mb-1 block text-sm text-[var(--color-text-secondary)]">{$t('setup.username')}</label>
          <input id="admin-user" type="text" bind:value={adminUsername} required autocomplete="username"
            class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
            data-testid="setup-username" />
        </div>
        <div class="mb-4">
          <label for="admin-pass" class="mb-1 block text-sm text-[var(--color-text-secondary)]">{$t('setup.password')}</label>
          <input id="admin-pass" type="password" bind:value={adminPassword} required autocomplete="new-password"
            class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
            data-testid="setup-password" />
        </div>
        <div class="mb-4">
          <label for="admin-pass-confirm" class="mb-1 block text-sm text-[var(--color-text-secondary)]">{$t('setup.passwordConfirm')}</label>
          <input id="admin-pass-confirm" type="password" bind:value={adminPasswordConfirm} required autocomplete="new-password"
            class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
            data-testid="setup-password-confirm" />
        </div>
      </div>
    {:else if step === 1}
      <div data-testid="setup-step-category">
        <p class="mb-4 text-sm text-[var(--color-text-secondary)]">{$t('setup.categoryHint')}</p>
        <div class="mb-4">
          <label for="cat-name" class="mb-1 block text-sm text-[var(--color-text-secondary)]">{$t('setup.categoryName')}</label>
          <input id="cat-name" type="text" bind:value={categoryName} placeholder={$t('setup.categoryPlaceholder')}
            class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
            data-testid="setup-category-name" />
        </div>
      </div>
    {/if}

    <!-- Navigation -->
    <div class="mt-6 flex justify-between">
      <div>
        {#if step > 0}
          <button onclick={handleBack}
            class="flex items-center gap-1 rounded-lg border border-[var(--color-border)] px-4 py-2 text-sm text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
            data-testid="setup-back">
            <ChevronLeft size={16} /> {$t('common.back')}
          </button>
        {/if}
      </div>
      <div class="flex gap-2">
        <button onclick={handleNext} disabled={loading || (step === 0 && (!adminUsername || !adminPassword))}
          class="flex items-center gap-1 rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white transition-colors hover:bg-[var(--color-primary-hover)] disabled:opacity-50"
          data-testid="setup-next">
          {step === 1 ? $t('common.finish') : $t('common.next')} <ChevronRight size={16} />
        </button>
      </div>
    </div>
  </div>
</div>
