<script lang="ts">
  import { Settings } from 'lucide-svelte'
  import { t } from '../../i18n'
  import CameraCard from './CameraCard.svelte'

  interface CameraEntry {
    name: string
    ip: string
    username: string
    password: string
    channel: number
    source: string
    entity_id: string
  }

  interface Props {
    config?: string
    active?: boolean
  }

  const { config = '{}', active = true }: Props = $props()

  const parsedConfig = $derived((() => {
    try { return JSON.parse(config) } catch { return {} }
  })())

  const configCameras = $derived<CameraEntry[]>(parsedConfig.cameras ?? [])
  const columns = $derived<number>(parsedConfig.columns ?? 2)

  // Generate stable camera ID matching backend logic (sha256 of ip:channel, first 8 bytes hex)
  async function generateCameraId(ip: string, channel: number): Promise<string> {
    const data = new TextEncoder().encode(`${ip}:${channel}`)
    const hashBuf = await crypto.subtle.digest('SHA-256', data)
    const bytes = new Uint8Array(hashBuf).slice(0, 8)
    return Array.from(bytes).map(b => b.toString(16).padStart(2, '0')).join('')
  }

  interface CameraWithId {
    id: string
    name: string
  }

  let cameras = $state<CameraWithId[]>([])

  $effect(() => {
    const entries = configCameras
    if (entries.length === 0) {
      cameras = []
      return
    }
    Promise.all(entries.map(async (cam) => {
      const key = cam.source === 'homeassistant' ? cam.entity_id : cam.ip
      const ch = cam.source === 'homeassistant' ? 0 : cam.channel
      const id = await generateCameraId(key, ch)
      return { id, name: cam.name }
    })).then(result => { cameras = result })
  })

</script>

<div class="h-full" data-testid="reolink-widget">
  {#if cameras.length === 0}
    <div class="flex h-full flex-col items-center justify-center gap-2 text-center text-sm text-[var(--color-text-muted)]">
      <Settings size={24} />
      <p>{$t('reolink.noCameras')}</p>
      <p class="text-xs">{$t('reolink.configureHint')}</p>
    </div>
  {:else}
    <div class="grid gap-2" style="grid-template-columns: repeat({columns}, 1fr);">
      {#each cameras as cam (cam.id)}
        <CameraCard id={cam.id} name={cam.name} {active} />
      {/each}
    </div>
  {/if}
</div>
