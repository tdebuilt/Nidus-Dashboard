<script lang="ts">
  import { Trash2, ShieldCheck, Crown, PenTool, Eye, RotateCcw, Copy } from 'lucide-svelte'
  import { auth } from '../../stores/auth'
  import { t } from '../../i18n'

  interface UserEntry {
    id: number
    username: string
    role: string
    totp_enabled: boolean
    created_at: string
  }

  interface Props {
    users: UserEntry[]
    generatedResetCode: string
    resetUsername: string
    onUpdateRole: (userId: number, role: string) => void
    onDeleteUser: (userId: number, username: string) => void
    onResetUser: (userId: number, username: string) => void
    onCopyResetLink: (code: string) => void
  }

  const {
    users,
    generatedResetCode,
    resetUsername,
    onUpdateRole,
    onDeleteUser,
    onResetUser,
    onCopyResetLink,
  }: Props = $props()

  const _roleIcons: Record<string, typeof Crown> = {
    admin: Crown,
    editor: PenTool,
    viewer: Eye,
  }
</script>

<div class="space-y-2">
  {#each users as user (user.id)}
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
            onchange={(e) => onUpdateRole(user.id, (e.target as HTMLSelectElement).value)}
            class="rounded border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1 text-sm text-[var(--color-text)]"
          >
            <option value="admin">{$t('users.roleAdmin')}</option>
            <option value="editor">{$t('users.roleEditor')}</option>
            <option value="viewer">{$t('users.roleViewer')}</option>
          </select>
          <button
            onclick={() => onResetUser(user.id, user.username)}
            class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]"
            title={$t('users.resetUser')}
          >
            <RotateCcw size={16} />
          </button>
          <button
            onclick={() => onDeleteUser(user.id, user.username)}
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
        onclick={() => onCopyResetLink(generatedResetCode)}
        class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]"
        title={$t('users.copyResetLink')}
      >
        <Copy size={16} />
      </button>
    </div>
  </div>
{/if}
