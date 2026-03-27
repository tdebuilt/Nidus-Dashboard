<script lang="ts">
  import { Video, Play, Square, RotateCw, Loader2 } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { t, translate } from '../../i18n'
  import { onMount, onDestroy } from 'svelte'

  interface Go2RTCStatus {
    available: boolean
    running: boolean
    uptime?: string
    cameras: number
  }

  let go2rtcStatus = $state<Go2RTCStatus>({ available: false, running: false, cameras: 0 })
  let go2rtcLoading = $state(false)
  let go2rtcPollTimer: ReturnType<typeof setInterval> | null = null

  onMount(() => {
    loadGo2rtcStatus()
    go2rtcPollTimer = setInterval(loadGo2rtcStatus, 10000)
  })

  onDestroy(() => {
    if (go2rtcPollTimer) clearInterval(go2rtcPollTimer)
  })

  async function loadGo2rtcStatus() {
    try { go2rtcStatus = await api.get<Go2RTCStatus>('/api/go2rtc/status') }
    catch { go2rtcStatus = { available: false, running: false, cameras: 0 } }
  }

  async function go2rtcAction(action: 'start' | 'stop' | 'restart') {
    go2rtcLoading = true
    try { go2rtcStatus = await api.post<Go2RTCStatus>(`/api/go2rtc/${action}`) }
    catch { toasts.error(translate('go2rtc.actionError')) }
    finally { go2rtcLoading = false }
  }
</script>

{#if go2rtcStatus.available}
<section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-go2rtc">
  <div class="mb-3 flex items-center gap-2">
    <Video size={18} class="text-[var(--color-text-secondary)]" />
    <h3 class="font-semibold text-[var(--color-text)]">{$t('go2rtc.title')}</h3>
  </div>

  <div class="space-y-3">
    <div class="flex items-center gap-3">
      <span class="inline-block h-2.5 w-2.5 rounded-full {go2rtcStatus.running ? 'bg-green-500' : 'bg-red-500'}"></span>
      <span class="text-sm text-[var(--color-text)]">
        {go2rtcStatus.running ? $t('go2rtc.running') : $t('go2rtc.stopped')}
      </span>
      {#if go2rtcStatus.running && go2rtcStatus.uptime}
        <span class="text-xs text-[var(--color-text-muted)]">({go2rtcStatus.uptime})</span>
      {/if}
      {#if go2rtcStatus.cameras > 0}
        <span class="text-xs text-[var(--color-text-muted)]">— {go2rtcStatus.cameras} {$t('go2rtc.cameras')}</span>
      {/if}
    </div>

    <div class="flex gap-2">
      {#if !go2rtcStatus.running}
        <button
          onclick={() => go2rtcAction('start')}
          disabled={go2rtcLoading}
          class="flex items-center gap-1.5 rounded-lg bg-green-600 px-3 py-1.5 text-xs text-white hover:bg-green-700 disabled:opacity-50"
        >
          {#if go2rtcLoading}<Loader2 size={12} class="animate-spin" />{:else}<Play size={12} />{/if}
          {$t('go2rtc.start')}
        </button>
      {:else}
        <button
          onclick={() => go2rtcAction('stop')}
          disabled={go2rtcLoading}
          class="flex items-center gap-1.5 rounded-lg bg-red-600 px-3 py-1.5 text-xs text-white hover:bg-red-700 disabled:opacity-50"
        >
          {#if go2rtcLoading}<Loader2 size={12} class="animate-spin" />{:else}<Square size={12} />{/if}
          {$t('go2rtc.stop')}
        </button>
        <button
          onclick={() => go2rtcAction('restart')}
          disabled={go2rtcLoading}
          class="flex items-center gap-1.5 rounded-lg bg-[var(--color-bg-tertiary)] px-3 py-1.5 text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-primary)] disabled:opacity-50"
        >
          {#if go2rtcLoading}<Loader2 size={12} class="animate-spin" />{:else}<RotateCw size={12} />{/if}
          {$t('go2rtc.restart')}
        </button>
      {/if}
    </div>

    <p class="text-xs text-[var(--color-text-muted)]">{$t('go2rtc.hint')}</p>
  </div>
</section>
{/if}
