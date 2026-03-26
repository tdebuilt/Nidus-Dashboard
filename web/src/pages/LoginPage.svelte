<script lang="ts">
  import { LogIn, Eye, EyeOff, Shield } from 'lucide-svelte';
  import { api, ApiRequestError } from '../lib/api';
  import { navigate, setAuthenticated } from '../lib/router';

  let username = '';
  let password = '';
  let totpCode = '';
  let showPassword = false;
  let totpRequired = false;
  let error = '';
  let loading = false;

  async function handleSubmit(): Promise<void> {
    error = '';
    loading = true;

    try {
      const body: Record<string, string> = { username, password };
      if (totpRequired && totpCode) {
        body.totp_code = totpCode;
      }
      await api.post('/api/auth/login', body);
      setAuthenticated(true);
      navigate('dashboard');
    } catch (e) {
      if (e instanceof ApiRequestError) {
        if (e.message === 'totp_required') {
          totpRequired = true;
        } else if (e.status === 401) {
          error = 'Identifiants incorrects';
        } else if (e.status === 429) {
          error = 'Trop de tentatives, réessayez plus tard';
        } else if (e.isNetwork) {
          error = 'Impossible de contacter le serveur';
        } else {
          error = e.message;
        }
      } else {
        error = 'Erreur inattendue';
      }
    }
    loading = false;
  }

  function handleFormSubmit(event: SubmitEvent): void {
    event.preventDefault();
    handleSubmit();
  }
</script>

<div class="flex items-center justify-center min-h-screen p-4">
  <div class="w-full max-w-sm">
    <div class="p-8 rounded-xl border" style="background-color: var(--color-bg-secondary); border-color: var(--color-border);">
      <div class="flex items-center gap-3 mb-6">
        <LogIn size={24} style="color: var(--color-accent);" />
        <h1 class="text-xl font-bold">Connexion</h1>
      </div>

      {#if error}
        <div class="alert-error mb-4 p-3 rounded-lg border text-sm">
          {error}
        </div>
      {/if}

      <form onsubmit={handleFormSubmit} class="space-y-4">
        <div>
          <label for="username" class="block text-sm font-medium mb-1" style="color: var(--color-text-secondary);">
            Nom d'utilisateur
          </label>
          <input
            id="username"
            type="text"
            bind:value={username}
            autocomplete="username"
            required
            disabled={loading}
            class="input-field w-full px-3 py-2 rounded-lg border text-sm outline-none transition-colors"
            style="background-color: var(--color-bg-primary); border-color: var(--color-border); color: var(--color-text-primary);"
            placeholder="admin"
          />
        </div>

        <div>
          <label for="password" class="block text-sm font-medium mb-1" style="color: var(--color-text-secondary);">
            Mot de passe
          </label>
          <div class="relative">
            <input
              id="password"
              type={showPassword ? 'text' : 'password'}
              bind:value={password}
              autocomplete="current-password"
              required
              disabled={loading}
              class="input-field w-full px-3 py-2 pe-10 rounded-lg border text-sm outline-none transition-colors"
              style="background-color: var(--color-bg-primary); border-color: var(--color-border); color: var(--color-text-primary);"
            />
            <button
              type="button"
              class="absolute end-2 top-1/2 -translate-y-1/2 p-1 rounded transition-colors"
              style="color: var(--color-text-muted);"
              onclick={() => showPassword = !showPassword}
              tabindex="-1"
              aria-label={showPassword ? 'Masquer le mot de passe' : 'Afficher le mot de passe'}
            >
              {#if showPassword}
                <EyeOff size={16} />
              {:else}
                <Eye size={16} />
              {/if}
            </button>
          </div>
        </div>

        {#if totpRequired}
          <div>
            <label for="totp" class="block text-sm font-medium mb-1" style="color: var(--color-text-secondary);">
              <span class="flex items-center gap-1.5">
                <Shield size={14} style="color: var(--color-accent);" />
                Code 2FA
              </span>
            </label>
            <input
              id="totp"
              type="text"
              inputmode="numeric"
              pattern="[0-9]*"
              maxlength="6"
              bind:value={totpCode}
              required
              disabled={loading}
              class="input-field w-full px-3 py-2 rounded-lg border text-sm outline-none transition-colors tracking-widest text-center font-mono"
              style="background-color: var(--color-bg-primary); border-color: var(--color-border); color: var(--color-text-primary);"
              placeholder="000000"
            />
          </div>
        {/if}

        <button
          type="submit"
          disabled={loading || !username || !password}
          class="w-full py-2.5 px-4 rounded-lg text-sm font-medium text-white transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          style="background-color: var(--color-accent);"
        >
          {#if loading}
            Connexion...
          {:else}
            Se connecter
          {/if}
        </button>
      </form>
    </div>
  </div>
</div>

<style>
  .input-field:focus {
    border-color: var(--color-accent);
  }

  .input-field:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  button[type="submit"]:not(:disabled):hover {
    background-color: var(--color-accent-hover) !important;
  }
</style>
