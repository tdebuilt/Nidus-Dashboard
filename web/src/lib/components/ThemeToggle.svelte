<script lang="ts">
  import { Sun, Moon } from 'lucide-svelte'
  import { theme, toggleTheme } from '../stores/theme'
  import { t } from '../i18n'
</script>

<button
  onclick={toggleTheme}
  class="flex items-center justify-center rounded-lg p-2 transition-colors hover:bg-[var(--color-sidebar-hover)]"
  aria-label={$theme === 'dark' ? $t('theme.switchToLight') : $t('theme.switchToDark')}
  data-testid="theme-toggle"
>
  {#if $theme === 'dark'}
    <Sun size={20} class="text-[var(--color-text-secondary)]" />
  {:else}
    <Moon size={20} class="text-[var(--color-text-secondary)]" />
  {/if}
</button>
