<script lang="ts">
  import { Users } from 'lucide-svelte'
  import { api, ApiError } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { confirm } from '../../stores/confirm'
  import { t, translate } from '../../i18n'
  import UsersSection from './UsersSection.svelte'
  import InvitesSection from './InvitesSection.svelte'

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
  let usersList = $state.raw<UserEntry[]>([])
  let invitesList = $state.raw<InviteEntry[]>([])
  let usersLoading = $state(false)
  let generatedInviteCode = $state('')
  let generatedResetCode = $state('')
  let resetUsername = $state('')

  $effect(() => {
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

  async function createInvite(role: string) {
    try {
      const result = await api.post<{ code: string }>('/api/invites', { role })
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
      <UsersSection
        users={usersList}
        {generatedResetCode}
        {resetUsername}
        onUpdateRole={updateUserRole}
        onDeleteUser={deleteUser}
        onResetUser={resetUser}
        onCopyResetLink={copyResetLink}
      />

      <InvitesSection
        invites={invitesList}
        {generatedInviteCode}
        onCreateInvite={createInvite}
        onDeleteInvite={deleteInvite}
        onCopyInviteLink={copyInviteLink}
        onClearInviteCode={() => { generatedInviteCode = '' }}
      />
    {/if}
  </section>
</div>
