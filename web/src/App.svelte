<script lang="ts">
  import { currentRoute } from './lib/stores/router'
  import { auth } from './lib/stores/auth'
  import { get } from 'svelte/store'
  import { pollingInterval } from './lib/stores/polling'
  import { appVersion } from './lib/stores/version'
  import { ws } from './lib/stores/websocket'
  import { categories } from './lib/stores/categories'
  import { api, ApiError, NetworkError } from './lib/api/client'
  import Sidebar from './lib/components/Sidebar.svelte'
  import Header from './lib/components/Header.svelte'
  import Login from './pages/Login.svelte'
  import Setup from './pages/Setup.svelte'
  import Dashboard from './pages/Dashboard.svelte'
  import SettingsPage from './pages/SettingsPage.svelte'
  import HelpPage from './pages/HelpPage.svelte'
  import Register from './pages/Register.svelte'
  import ResetPassword from './pages/ResetPassword.svelte'
  import KioskPage from './pages/KioskPage.svelte'
  import NotFound from './pages/NotFound.svelte'
  import NetworkErrorPage from './pages/NetworkError.svelte'
  import ToastContainer from './lib/components/ToastContainer.svelte'
  import ConfirmDialog from './lib/components/ConfirmDialog.svelte'
  import OfflineIndicator from './lib/components/OfflineIndicator.svelte'
  import KeyboardShortcutHandler from './lib/components/KeyboardShortcutHandler.svelte'
  import ShortcutHelpModal from './lib/components/ShortcutHelpModal.svelte'
  import AboutModal from './lib/components/AboutModal.svelte'
  import ThemeEditorModal from './lib/components/ThemeEditorModal.svelte'
  import { customThemes } from './lib/stores/customThemes'
  import { setKeyboardShortcutsEnabled } from './lib/stores/keyboardShortcuts'
  import { navigate } from './lib/stores/router'
  import { t, applyBrowserLocale, setLocale, isRTL } from './lib/i18n'
  import { setTheme } from './lib/stores/theme'
  import { openSidebar, sidebarOpen } from './lib/stores/sidebar'

  let networkError = $state(false)

  /* ── Swipe from edge to open sidebar (left edge in LTR, right edge in RTL) ── */
  let edgeTouchStartX = 0
  let edgeTouchStartY = 0
  let edgeSwipeActive = false

  function handleEdgeTouchStart(e: TouchEvent) {
    const startX = e.touches[0].clientX
    const rtl = get(isRTL)
    const isEdge = rtl ? startX > window.innerWidth - 24 : startX < 24
    if (isEdge && !$sidebarOpen) {
      edgeTouchStartX = startX
      edgeTouchStartY = e.touches[0].clientY
      edgeSwipeActive = true
    }
  }

  function handleEdgeTouchEnd(e: TouchEvent) {
    if (!edgeSwipeActive) return
    edgeSwipeActive = false
    const deltaX = e.changedTouches[0].clientX - edgeTouchStartX
    const deltaY = Math.abs(e.changedTouches[0].clientY - edgeTouchStartY)
    const rtl = get(isRTL)
    const isSwipeOpen = rtl ? deltaX < -50 && deltaY < Math.abs(deltaX) : deltaX > 50 && deltaY < deltaX
    if (isSwipeOpen) {
      openSidebar()
    }
  }

  async function checkAuth() {
    auth.set({ authenticated: false, setupCompleted: false, loading: true })
    try {
      const status = await api.get<{ setup_completed: boolean }>('/api/auth/status')
      if (!status.setup_completed) {
        auth.set({ authenticated: false, setupCompleted: false, loading: false })
        navigate('/setup')
        return
      }
      // Verify session by calling an authenticated endpoint
      try {
        await api.get('/api/settings')
        const savedRole = (typeof localStorage !== 'undefined' ? localStorage.getItem('nidus-role') : null) as 'admin' | 'editor' | 'viewer' | null
        auth.set({ authenticated: true, setupCompleted: true, loading: false, role: savedRole ?? 'admin' })
      } catch {
        auth.set({ authenticated: false, setupCompleted: true, loading: false })
        return
      }
      networkError = false
    } catch (e) {
      if (e instanceof NetworkError) {
        networkError = true
        auth.set({ authenticated: false, setupCompleted: false, loading: false })
      } else if (e instanceof ApiError && e.status === 401) {
        auth.set({ authenticated: false, setupCompleted: true, loading: false })
      } else {
        auth.set({ authenticated: false, setupCompleted: true, loading: false })
      }
    }
  }

  function applyCustomCSS(css: string) {
    let el = document.getElementById('nidus-custom-css')
    if (!css) {
      el?.remove()
      return
    }
    if (!el) {
      el = document.createElement('style')
      el.id = 'nidus-custom-css'
      document.head.appendChild(el)
    }
    el.textContent = css
  }

  async function loadUserSettings() {
    try {
      // Load global settings (custom CSS only)
      const global = await api.get<{ custom_css: string }>('/api/settings')
      applyCustomCSS(global.custom_css || '')

      // Load user preferences (theme, language, etc.)
      // Load custom themes before setting theme (so custom IDs are in the registry)
      await customThemes.load()

      const prefs = await api.get<{ theme: string; language: string; refresh_interval: number; accent_color: string; enable_keyboard_shortcuts: boolean }>('/api/preferences')
      setTheme(prefs.theme)
      setLocale(prefs.language)
      if (prefs.refresh_interval) pollingInterval.update(prefs.refresh_interval)
      if (prefs.accent_color) {
        document.documentElement.style.setProperty('--color-primary', prefs.accent_color)
      }
      setKeyboardShortcutsEnabled(prefs.enable_keyboard_shortcuts !== false)
    } catch {
      // Settings not available (not logged in, etc.) — use defaults
      applyBrowserLocale()
    }
  }

  $effect(() => {
    const handler = (e: PromiseRejectionEvent) => {
      console.error('Unhandled rejection:', e.reason)
    }
    window.addEventListener('unhandledrejection', handler)
    return () => window.removeEventListener('unhandledrejection', handler)
  })

  $effect(() => {
    applyBrowserLocale()
    checkAuth().then(() => {
      if (get(auth).authenticated) {
        categories.load()
        loadUserSettings()
        ws.connect()
        pollingInterval.load()
      }
      appVersion.load()
    })
  })

  function handleRetry() {
    networkError = false
    checkAuth()
  }
</script>

{#if $auth.loading}
  <div class="flex min-h-screen items-center justify-center" data-testid="loading">
    <p class="text-[var(--color-text-secondary)]">{$t('common.loading')}</p>
  </div>
{:else if networkError}
  <NetworkErrorPage onRetry={handleRetry} />
{:else if $currentRoute === 'setup'}
  <Setup />
{:else if $currentRoute === 'login'}
  <Login />
{:else if $currentRoute === 'register'}
  <Register />
{:else if $currentRoute === 'reset-password'}
  <ResetPassword />
{:else if $currentRoute === 'kiosk'}
  <KioskPage />
{:else}
  <!-- Main layout with sidebar -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="flex min-h-screen"
    ontouchstart={handleEdgeTouchStart}
    ontouchend={handleEdgeTouchEnd}
  >
    <Sidebar />
    <div class="flex min-w-0 flex-1 flex-col lg:ms-0">
      <Header />
      <main class="flex-1 overflow-x-hidden p-4 sm:p-6 2xl:p-10" data-testid="main-content">
        {#if $currentRoute === 'dashboard'}
          <Dashboard />
        {:else if $currentRoute === 'settings'}
          <SettingsPage />
        {:else if $currentRoute === 'help'}
          <HelpPage />
        {:else}
          <NotFound />
        {/if}
      </main>
    </div>
  </div>
{/if}

<ToastContainer />
<ConfirmDialog />
<OfflineIndicator />
<KeyboardShortcutHandler />
<ShortcutHelpModal />
<AboutModal />
<ThemeEditorModal />
