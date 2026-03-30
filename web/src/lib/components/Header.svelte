<script lang="ts">
  import { Menu, Pencil } from 'lucide-svelte'
  import { toggleSidebar } from '../stores/sidebar'
  import { editMode, toggleEditMode } from '../stores/editMode'
  import { isEditor } from '../stores/auth'
  import { t, translate } from '../i18n'
  import NidusLogo from './NidusLogo.svelte'
  import SearchBar from './SearchBar.svelte'
</script>

<header class="flex h-14 items-center justify-between border-b border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-4" data-testid="mobile-header">
  <div class="flex items-center">
    <button
      onclick={toggleSidebar}
      class="rounded-lg p-2 transition-colors hover:bg-[var(--color-bg-tertiary)]"
      aria-label={$t('sidebar.openMenu')}
      data-testid="burger-button"
    >
      <Menu size={22} class="text-[var(--color-text)]" />
    </button>
    <NidusLogo size={22} />
    <span class="ms-2 text-lg font-bold text-[var(--color-text)]">Nidus</span>
  </div>
  <div class="ms-auto flex items-center gap-2">
    {#if $isEditor}
      <button
        onclick={toggleEditMode}
        class="rounded-lg p-2 transition-colors"
        class:bg-[var(--color-primary)]={$editMode}
        class:text-white={$editMode}
        class:text-[var(--color-text-muted)]={!$editMode}
        class:hover:bg-[var(--color-bg-tertiary)]={!$editMode}
        title={$editMode ? translate('dashboard.editModeOff') : translate('dashboard.editModeOn')}
        aria-label={$editMode ? translate('dashboard.editModeOff') : translate('dashboard.editModeOn')}
        data-testid="edit-mode-toggle"
      >
        <Pencil size={18} />
      </button>
    {/if}
    <SearchBar />
  </div>
</header>
