<script lang="ts">
  import { onMount } from 'svelte'
  import { AlertCircle, RefreshCw } from 'lucide-svelte'
  import { t } from '../../i18n'
  import TerminalToolbar from './TerminalToolbar.svelte'

  interface Props {
    config?: string
    active?: boolean
    widgetId?: number
  }

  const { config = '{}', active = true, widgetId }: Props = $props()

  type Status = 'connecting' | 'connected' | 'disconnected'

  let containerEl: HTMLDivElement | undefined = $state()
  let status = $state<Status>('disconnected')
  let terminalRef = $state<import('@xterm/xterm').Terminal | null>(null)
  let terminal: import('@xterm/xterm').Terminal | null = null
  let fitAddon: import('@xterm/addon-fit').FitAddon | null = null
  let ws: WebSocket | null = null
  let resizeObserver: ResizeObserver | null = null

  interface TerminalCfg {
    font_size: number
    scrollback: number
    host: string
    snippets: Array<{ label: string; command: string }>
  }

  function getConfig(): TerminalCfg {
    try {
      const parsed = JSON.parse(config)
      return {
        font_size: parsed.font_size ?? 14,
        scrollback: parsed.scrollback ?? 1000,
        host: parsed.host ?? '',
        snippets: Array.isArray(parsed.snippets) ? parsed.snippets : [],
      }
    } catch {
      return { font_size: 14, scrollback: 1000, host: '', snippets: [] }
    }
  }

  const snippets = $derived(getConfig().snippets)

  function getWSUrl(): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${protocol}//${window.location.host}/api/terminal/ws?widget_id=${widgetId}`
  }

  async function connect() {
    if (!containerEl || !widgetId) return
    cleanup()

    const cfg = getConfig()
    if (!cfg.host) {
      status = 'disconnected'
      return
    }

    const { Terminal } = await import('@xterm/xterm')
    const { FitAddon } = await import('@xterm/addon-fit')
    const { WebLinksAddon } = await import('@xterm/addon-web-links')

    terminal = terminalRef = new Terminal({
      fontSize: cfg.font_size,
      scrollback: cfg.scrollback,
      cursorBlink: true,
      theme: {
        background: '#1a1b26',
        foreground: '#a9b1d6',
        cursor: '#c0caf5',
        selectionBackground: '#33467c',
      },
    })

    fitAddon = new FitAddon()
    terminal.loadAddon(fitAddon)
    terminal.loadAddon(new WebLinksAddon())
    terminal.open(containerEl)

    requestAnimationFrame(() => {
      fitAddon?.fit()
      connectWS()
    })

    resizeObserver = new ResizeObserver(() => {
      fitAddon?.fit()
      sendResize()
    })
    resizeObserver.observe(containerEl)
  }

  function connectWS() {
    status = 'connecting'
    ws = new WebSocket(getWSUrl())
    ws.binaryType = 'arraybuffer'

    ws.onopen = () => {
      status = 'connected'
      sendResize()
    }

    ws.onmessage = (event: MessageEvent) => {
      if (event.data instanceof ArrayBuffer) {
        terminal?.write(new Uint8Array(event.data))
      }
    }

    ws.onclose = () => {
      status = 'disconnected'
    }

    ws.onerror = () => {
      status = 'disconnected'
    }

    terminal?.onData((data: string) => {
      if (ws?.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'input', data: btoa(data) }))
      }
    })
  }

  function sendResize() {
    if (terminal && ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        type: 'resize',
        cols: terminal.cols,
        rows: terminal.rows,
      }))
    }
  }

  function cleanup() {
    resizeObserver?.disconnect()
    resizeObserver = null
    ws?.close()
    ws = null
    terminal?.dispose()
    terminal = terminalRef = null
    fitAddon = null
    status = 'disconnected'
  }

  function reconnect() {
    connect()
  }

  $effect(() => {
    if (active && containerEl) {
      connect()
    } else {
      cleanup()
    }
    return () => cleanup()
  })

  onMount(() => {
    import('@xterm/xterm/css/xterm.css')
  })
</script>

<div class="flex h-full min-h-[300px] flex-col" data-testid="terminal-widget">
  <div class="relative min-h-0 flex-1">
    <div
      bind:this={containerEl}
      class="h-full overflow-hidden rounded-t"
      style="background: #1a1b26;"
    ></div>

    {#if status === 'disconnected'}
      <div class="absolute inset-0 z-10 flex flex-col items-center justify-center gap-3 rounded bg-black/80">
        <AlertCircle size={24} class="text-[var(--color-danger)]" />
        <p class="text-sm text-[var(--color-text-muted)]">{$t('terminal.disconnected')}</p>
        <button
          onclick={reconnect}
          class="flex items-center gap-2 rounded-lg border border-[var(--color-border)] px-3 py-1.5 text-xs text-[var(--color-text)] hover:bg-[var(--color-bg-hover)]"
        >
          <RefreshCw size={12} />
          {$t('terminal.reconnect')}
        </button>
      </div>
    {:else if status === 'connecting'}
      <div class="absolute inset-0 z-10 flex items-center justify-center rounded bg-black/80">
        <p class="text-sm text-[var(--color-text-muted)]">{$t('terminal.connecting')}</p>
      </div>
    {/if}
  </div>

  <TerminalToolbar terminal={terminalRef} {snippets} />
</div>
