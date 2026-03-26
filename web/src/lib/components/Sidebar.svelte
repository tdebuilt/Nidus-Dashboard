<script lang="ts">
  import { LayoutDashboard, Settings, HelpCircle, LogOut, ChevronLeft, ChevronRight, Info } from 'lucide-svelte'
  import { navigate, currentRoute } from '../stores/router'
  import { sidebarOpen, closeSidebar } from '../stores/sidebar'
  import { categories } from '../stores/categories'
  import { openAboutModal } from '../stores/aboutModal'
  import { t, isRTL } from '../i18n'
  import ThemeToggle from './ThemeToggle.svelte'
  import NidusLogo from './NidusLogo.svelte'
  import DynamicIcon from './DynamicIcon.svelte'

  function handleNav(path: string) {
    navigate(path)
    closeSidebar()
  }

  function handleLogout() {
    fetch('/api/auth/logout', { method: 'POST', credentials: 'include' })
      .then(() => {
        navigate('/login')
      })
      .catch(() => {
        navigate('/login')
      })
    closeSidebar()
  }

  /* ── Swipe to close sidebar ── */
  let touchStartX = 0
  let touchStartY = 0
  let swiping = false

  function handleTouchStart(e: TouchEvent) {
    touchStartX = e.touches[0].clientX
    touchStartY = e.touches[0].clientY
    swiping = true
  }

  function handleTouchEnd(e: TouchEvent) {
    if (!swiping) return
    swiping = false
    const deltaX = e.changedTouches[0].clientX - touchStartX
    const deltaY = Math.abs(e.changedTouches[0].clientY - touchStartY)
    // Only respond to horizontal swipes, not vertical scrolling
    if (deltaY > Math.abs(deltaX)) return
    const shouldClose = $isRTL ? deltaX > 50 : deltaX < -50
    if (shouldClose) {
      closeSidebar()
    }
  }
</script>

<!-- Mobile overlay -->
{#if $sidebarOpen}
  <button
    class="fixed inset-0 z-30 bg-black/50 transition-opacity duration-300"
    onclick={closeSidebar}
    aria-label={$t('sidebar.closeMenu')}
    data-testid="sidebar-overlay"
  ></button>
{/if}

<!-- Sidebar -->
<aside
  ontouchstart={handleTouchStart}
  ontouchend={handleTouchEnd}
  class="fixed top-0 start-0 z-40 flex h-full w-64 flex-col border-e border-[var(--color-border)] bg-[var(--color-sidebar-bg)] transition-transform duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]
    {$sidebarOpen ? 'translate-x-0' : ($isRTL ? 'translate-x-full' : '-translate-x-full')}"
  data-testid="sidebar"
>
  <!-- Header -->
  <div class="flex h-14 items-center justify-between border-b border-[var(--color-border)] px-4">
    <button onclick={() => handleNav('/')} class="flex items-center gap-2 text-lg font-bold text-[var(--color-text)]">
      <NidusLogo size={24} />
      Nidus
    </button>
    <button
      class="touch-action-btn rounded-lg p-2 transition-colors hover:bg-[var(--color-sidebar-hover)]"
      onclick={closeSidebar}
      aria-label={$t('sidebar.closeMenu')}
      data-testid="sidebar-close"
    >
      {#if $isRTL}
        <ChevronRight size={20} class="text-[var(--color-text-secondary)]" />
      {:else}
        <ChevronLeft size={20} class="text-[var(--color-text-secondary)]" />
      {/if}
    </button>
  </div>

  <!-- Navigation -->
  <nav class="flex-1 overflow-y-auto p-3" data-testid="sidebar-nav">
    <!-- Dashboard -->
    <button
      onclick={() => handleNav('/')}
      class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors
        {$currentRoute === 'dashboard'
          ? 'bg-[var(--color-primary)] text-white'
          : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-sidebar-hover)]'}"
    >
      <LayoutDashboard size={18} />
      <span>{$t('sidebar.dashboard')}</span>
    </button>

    <!-- Categories -->
    {#if $categories.length > 0}
      <div class="mt-4 mb-2 px-3 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">
        {$t('sidebar.categories')}
      </div>
      {#each $categories as cat (cat.id)}
        <button
          onclick={() => handleNav(`/dashboard/${cat.slug}`)}
          class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-sidebar-hover)]"
        >
          <DynamicIcon name={cat.icon} size={18} />
          <span>{cat.name}</span>
        </button>
      {/each}
    {/if}
  </nav>

  <!-- Footer -->
  <div class="border-t border-[var(--color-border)] p-3" data-testid="sidebar-footer">
    <button
      onclick={() => handleNav('/settings')}
      class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors
        {$currentRoute === 'settings'
          ? 'bg-[var(--color-primary)] text-white'
          : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-sidebar-hover)]'}"
      data-testid="settings-link"
    >
      <Settings size={18} />
      <span>{$t('sidebar.settings')}</span>
    </button>

    <button
      onclick={() => handleNav('/help')}
      class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors
        {$currentRoute === 'help'
          ? 'bg-[var(--color-primary)] text-white'
          : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-sidebar-hover)]'}"
      data-testid="help-link"
    >
      <HelpCircle size={18} />
      <span>{$t('sidebar.help')}</span>
    </button>

    <div class="mt-1 flex items-center justify-between">
      <button
        onclick={handleLogout}
        class="flex items-center gap-3 rounded-lg px-3 py-2 text-sm text-[var(--color-danger)] transition-colors hover:bg-[var(--color-sidebar-hover)]"
        data-testid="logout-button"
      >
        <LogOut size={18} />
        <span>{$t('sidebar.logout')}</span>
      </button>
      <div class="flex items-center gap-1">
        <button
          onclick={() => { openAboutModal(); closeSidebar() }}
          class="rounded-lg p-2 text-[var(--color-text-muted)] transition-colors hover:bg-[var(--color-sidebar-hover)] hover:text-[var(--color-text-secondary)]"
          aria-label={$t('settings.about')}
          data-testid="about-button"
        >
          <Info size={18} />
        </button>
        <ThemeToggle />
      </div>
    </div>
  </div>
</aside>
