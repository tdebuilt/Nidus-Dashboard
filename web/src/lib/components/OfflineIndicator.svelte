<script lang="ts">
  import { onMount } from 'svelte'
  import { WifiOff } from 'lucide-svelte'
  import { t } from '../i18n'

  let offline = $state(false)

  onMount(() => {
    offline = !navigator.onLine
    const handleOnline = () => { offline = false }
    const handleOffline = () => { offline = true }
    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)
    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  })
</script>

{#if offline}
  <div class="fixed bottom-4 start-1/2 z-50 flex -translate-x-1/2 items-center gap-2 rounded-lg bg-[var(--color-warning)] px-4 py-2 text-sm font-medium text-white shadow-lg">
    <WifiOff size={16} />
    {$t('errors.offline')}
  </div>
{/if}
