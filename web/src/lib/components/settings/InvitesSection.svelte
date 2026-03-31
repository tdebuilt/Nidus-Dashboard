<script lang="ts">
  import { UserPlus, Trash2, Plus, Copy } from 'lucide-svelte'
  import { t } from '../../i18n'

  interface InviteEntry {
    id: number
    code: string
    role: string
    used_by: number | null
    expires_at: string
    created_at: string
  }

  interface Props {
    invites: InviteEntry[]
    generatedInviteCode: string
    onCreateInvite: (role: string) => void
    onDeleteInvite: (invId: number) => void
    onCopyInviteLink: (code: string) => void
    onClearInviteCode: () => void
  }

  const {
    invites,
    generatedInviteCode,
    onCreateInvite,
    onDeleteInvite,
    onCopyInviteLink,
    onClearInviteCode,
  }: Props = $props()

  let inviteRole = $state('viewer')
  let showInviteForm = $state(false)

  const pendingInvites = $derived(invites.filter((i) => !i.used_by))
</script>

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
          onclick={() => onCopyInviteLink(generatedInviteCode)}
          class="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-primary)]"
          title={$t('users.copyLink')}
          aria-label={$t('users.copyLink')}
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
        aria-label="Invite role"
        class="rounded border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-sm text-[var(--color-text)]"
      >
        <option value="viewer">{$t('users.roleViewer')}</option>
        <option value="editor">{$t('users.roleEditor')}</option>
        <option value="admin">{$t('users.roleAdmin')}</option>
      </select>
      <button
        onclick={() => onCreateInvite(inviteRole)}
        class="rounded-lg bg-[var(--color-primary)] px-3 py-1.5 text-sm text-white hover:bg-[var(--color-primary-hover)]"
      >
        {$t('users.generateInvite')}
      </button>
      <button
        onclick={() => { showInviteForm = false; onClearInviteCode() }}
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

  {#if pendingInvites.length > 0}
    <div class="mt-3">
      <p class="mb-1 text-xs font-semibold text-[var(--color-text-muted)]">{$t('users.pendingInvites')}</p>
      {#each pendingInvites as invite (invite.id)}
        <div class="flex items-center justify-between rounded border border-[var(--color-border)] px-2 py-1 text-xs">
          <div class="flex items-center gap-2">
            <code class="text-[var(--color-text-muted)]">{invite.code.slice(0, 8)}...</code>
            <span class="rounded bg-[var(--color-bg-tertiary)] px-1.5 py-0.5 text-[var(--color-text-secondary)]">{invite.role}</span>
          </div>
          <div class="flex items-center gap-2">
            <button
              onclick={() => onCopyInviteLink(invite.code)}
              class="text-[var(--color-text-muted)] hover:text-[var(--color-primary)]"
              title={$t('users.copyLink')}
              aria-label={$t('users.copyLink')}
            >
              <Copy size={12} />
            </button>
            <button
              onclick={() => onDeleteInvite(invite.id)}
              class="text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"
              title={$t('common.delete')}
              aria-label={$t('common.delete')}
            >
              <Trash2 size={12} />
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>
