<script lang="ts">
  import { Users, UserPlus, Trash2, Plus, Copy, ShieldCheck, Crown, PenTool, Eye, RotateCcw } from 'lucide-svelte'
  import { api, ApiError } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { confirm } from '../../stores/confirm'
  import { auth } from '../../stores/auth'
  import { t, translate } from '../../i18n'
  import { onMount } from 'svelte'

  interface UserEntry {
    id: number
    username: string
    role: string
    totp_enabled: boolean
    created_at: string
  }
  interface InviteEntry {
    id: number
    code: string
    role: string
    used_by: number | null
    expires_at: string
    created_at: string
  }
  let usersList = $state<UserEntry[]>([])
  let invitesList = $state<InviteEntry[]>([])
  let usersLoading = $state(false)
  let inviteRole = $state('viewer')
  let showInviteForm = $state(false)
  let generatedInviteCode = $state('')
  let generatedResetCode = $state('')
  let resetUsername = $state('')

  const _roleIcons: Record<string, typeof Crown> = {
    admin: Crown,
    editor: PenTool,
    viewer: Eye,
  }

  onMount(() => {
    loadUsers()
  })

  async function loadUsers() {
    usersLoading = true
    try {
      const [users, invites] = await Promise.all([
        api.get<UserEntry[]>('/api/users'),
        api.get<InviteEntry[]>('/api/invites'),
      ])
      usersList = users ?? []
      invitesList = invites ?? []
    } catch {
      // error
    } finally {
      usersLoading = false
    }
  }

  async function updateUserRole(userId: number, role: string) {
    try {
      await api.put(`/api/users/${userId}/role`, { role })
      toasts.success(translate('users.roleUpdated'))
      await loadUsers()
    } catch (err) {
      toasts.error(err instanceof ApiError ? err.message : translate('users.error'))
    }
  }

  async function deleteUser(userId: number, username: string) {
    const ok = await confirm({
      title: translate('users.deleteConfirmTitle'),
      message: translate('users.deleteConfirmMessage', { username }),
      confirmLabel: translate('common.delete'),
      destructive: true,
    })
    if (!ok) return
    try {
      await api.delete(`/api/users/${userId}`)
      toasts.success(translate('users.deleted'))
      await loadUsers()
    } catch (err) {
      toasts.error(err instanceof ApiError ? err.message : translate('users.error'))
    }
  }

  async function resetUser(userId: number, username: string) {
    const ok = await confirm({
      title: translate('users.resetConfirmTitle'),
      message: translate('users.resetConfirmMessage', { username }),
      confirmLabel: translate('users.resetConfirm'),
      destructive: true,
    })
    if (!ok) return
    try {
      const result = await api.post<{ code: string }>(`/api/users/${userId}/reset`)
      generatedResetCode = result.code
      resetUsername = username
      toasts.success(translate('users.resetCreated'))
    } catch (err) {
      toasts.error(err instanceof ApiError ? err.message : translate('users.error'))
    }
  }

  function copyResetLink(code: string) {
    const url = `${window.location.origin}/reset-password?code=${code}`
    navigator.clipboard.writeText(url)
    toasts.success(translate('users.linkCopied'))
  }

  async function createInvite() {
    try {
      const result = await api.post<{ code: string }>('/api/invites', { role: inviteRole })
      generatedInviteCode = result.code
      toasts.success(translate('users.inviteCreated'))
      await loadUsers()
    } catch (err) {
      toasts.error(err instanceof ApiError ? err.message : translate('users.error'))
    }
  }

  async function deleteInvite(invId: number) {
    try {
      await api.delete(`/api/invites/${invId}`)
      toasts.success(translate('users.inviteDeleted'))
      await loadUsers()
    } catch {
      // ignore
    }
  }

  function copyInviteLink(code: string) {
    const url = `${window.location.origin}/register?code=${code}`
    navigator.clipboard.writeText(url)
    toasts.success(translate('users.linkCopied'))
  }
</script>

<div class="space-y-6">
  <section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-users">
    <div class="mb-3 flex items-center gap-2">
      <Users size={18} class="text-[var(--color-text-secondary)]" />
      <h3 class="font-semibold text-[var(--color-text)]">{$t('users.title')}</h3>
    </div>

    {#if usersLoading}
      <p class="text-sm text-[var(--color-text-muted)]">{$t('common.loading')}</p>
    {:else}
      <div class="space-y-2">
        {#each usersList as user (user.id)}
          <div class="flex items-center justify-between rounded-lg border border-[var(--color-border)] p-3">
            <div class="flex items-center gap-3">
              <div>
                <span class="font-medium text-[var(--color-text)]">{user.username}</span>
                {#if user.totp_enabled}
                  <ShieldCheck size={14} class="ms-1 inline text-green-500" />
                {/if}
              </div>
            </div>
            <div class="flex items-center gap-2">
              {#if user.id !== $auth.userId}
                <select
                  value={user.role}
                  onchange={(e) => updateUserRole(user.id, (e.target as HTMLSelectElement).value)}
                  class="rounded border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1 text-sm text-[var(--color-text)]"
                >
                  <option value="admin">{$t('users.roleAdmin')}</option>
                  <option value="editor">{$t('users.roleEditor')}</option>
                  <option value="viewer">{$t('users.roleViewer')}</option>
                </select>
                <button
                  onclick={() => resetUser(user.id, user.username)}
                  class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]"
                  title={$t('users.resetUser')}
                >
                  <RotateCcw size={16} />
                </button>
                <button
                  onclick={() => deleteUser(user.id, user.username)}
                  class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"
                  title={$t('common.delete')}
                >
                  <Trash2 size={16} />
                </button>
              {:else}
                <span class="rounded bg-[var(--color-primary)] px-2 py-0.5 text-xs text-white">{$t(`users.role${user.role.charAt(0).toUpperCase() + user.role.slice(1)}`)}</span>
                <span class="text-xs text-[var(--color-text-muted)]">({$t('users.you')})</span>
              {/if}
            </div>
          </div>
        {/each}
      </div>

      <!-- Reset code display -->
      {#if generatedResetCode}
        <div class="mt-3 rounded-lg border border-amber-500/30 bg-amber-500/10 p-3">
          <p class="mb-1 text-sm text-[var(--color-text)]">{$t('users.resetReady', { username: resetUsername })}</p>
          <div class="flex items-center gap-2">
            <code class="flex-1 rounded bg-[var(--color-bg)] px-2 py-1 text-xs text-[var(--color-text)]">{generatedResetCode}</code>
            <button
              onclick={() => copyResetLink(generatedResetCode)}
              class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]"
              title={$t('users.copyResetLink')}
            >
              <Copy size={16} />
            </button>
          </div>
        </div>
      {/if}

      <!-- Invite section -->
      <div class="mt-4 border-t border-[var(--color-border)] pt-4">
        <div class="mb-2 flex items-center gap-2">
          <UserPlus size={16} class="text-[var(--color-text-secondary)]" />
          <h4 class="text-sm font-semibold text-[var(--color-text)]">{$t('users.inviteTitle')}</h4>
        </div>

        {#if generatedInviteCode}
          <div class="mb-3 rounded-lg border border-green-500/30 bg-green-500/10 p-3">
            <p class="mb-1 text-sm text-[var(--color-text)]">{$t('users.inviteReady')}</p>
            <div class="flex items-center gap-2">
              <code class="flex-1 rounded bg-[var(--color-bg)] px-2 py-1 text-xs text-[var(--color-text)]">{generatedInviteCode}</code>
              <button
                onclick={() => copyInviteLink(generatedInviteCode)}
                class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]"
                title={$t('users.copyLink')}
              >
                <Copy size={16} />
              </button>
            </div>
          </div>
        {/if}

        {#if showInviteForm}
          <div class="flex items-center gap-2">
            <select
              bind:value={inviteRole}
              class="rounded border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]"
            >
              <option value="viewer">{$t('users.roleViewer')}</option>
              <option value="editor">{$t('users.roleEditor')}</option>
              <option value="admin">{$t('users.roleAdmin')}</option>
            </select>
            <button
              onclick={createInvite}
              class="rounded-lg bg-[var(--color-primary)] px-3 py-1.5 text-sm text-white hover:bg-[var(--color-primary-hover)]"
            >
              {$t('users.generateInvite')}
            </button>
            <button
              onclick={() => { showInviteForm = false; generatedInviteCode = '' }}
              class="rounded-lg px-3 py-1.5 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
            >
              {$t('common.cancel')}
            </button>
          </div>
        {:else}
          <button
            onclick={() => showInviteForm = true}
            class="flex items-center gap-1 text-sm text-[var(--color-primary)] hover:underline"
          >
            <Plus size={14} /> {$t('users.inviteUser')}
          </button>
        {/if}

        {#if invitesList.filter((i) => !i.used_by).length > 0}
          <div class="mt-3">
            <p class="mb-1 text-xs font-semibold text-[var(--color-text-muted)]">{$t('users.pendingInvites')}</p>
            {#each invitesList.filter((i) => !i.used_by) as invite (invite.id)}
              <div class="flex items-center justify-between rounded border border-[var(--color-border)] px-2 py-1 text-xs">
                <div class="flex items-center gap-2">
                  <code class="text-[var(--color-text-muted)]">{invite.code.slice(0, 8)}...</code>
                  <span class="rounded bg-[var(--color-bg-tertiary)] px-1.5 py-0.5 text-[var(--color-text-secondary)]">{invite.role}</span>
                </div>
                <div class="flex items-center gap-2">
                  <button
                    onclick={() => copyInviteLink(invite.code)}
                    class="text-[var(--color-text-muted)] hover:text-[var(--color-primary)]"
                  >
                    <Copy size={12} />
                  </button>
                  <button
                    onclick={() => deleteInvite(invite.id)}
                    class="text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"
                  >
                    <Trash2 size={12} />
                  </button>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/if}
  </section>
</div>
