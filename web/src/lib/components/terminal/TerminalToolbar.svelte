<script lang="ts">
  import type { Terminal } from '@xterm/xterm'

  interface Snippet {
    label: string
    command: string
  }

  interface Props {
    terminal: Terminal | null
    snippets?: Snippet[]
  }

  const { terminal, snippets = [] }: Props = $props()

  let ctrlActive = $state(false)
  let altActive = $state(false)

  function send(data: string) {
    if (!terminal) return
    terminal.input(data, true)
    terminal.focus()
  }

  function sendKey(key: string) {
    if (ctrlActive) {
      const code = key.toUpperCase().charCodeAt(0) - 64
      if (code >= 0 && code <= 31) send(String.fromCharCode(code))
      ctrlActive = false
      altActive = false
      return
    }
    if (altActive) {
      send('\x1b' + key)
      altActive = false
      return
    }
    send(key)
  }

  function toggleCtrl() {
    ctrlActive = !ctrlActive
    altActive = false
  }

  function toggleAlt() {
    altActive = !altActive
    ctrlActive = false
  }

  function sendSnippet(command: string) {
    send(command + '\r')
  }

  const specialKeys: Array<{ label: string; action: () => void }> = [
    { label: 'Esc', action: () => sendKey('\x1b') },
    { label: 'Tab', action: () => sendKey('\t') },
    { label: 'Ctrl', action: toggleCtrl },
    { label: 'Alt', action: toggleAlt },
    { label: '|', action: () => sendKey('|') },
    { label: '/', action: () => sendKey('/') },
    { label: '-', action: () => sendKey('-') },
    { label: '~', action: () => sendKey('~') },
  ]

  const arrowKeys: Array<{ label: string; seq: string; gridArea: string }> = [
    { label: '↑', seq: '\x1b[A', gridArea: '1 / 2' },
    { label: '←', seq: '\x1b[D', gridArea: '2 / 1' },
    { label: '↓', seq: '\x1b[B', gridArea: '2 / 2' },
    { label: '→', seq: '\x1b[C', gridArea: '2 / 3' },
  ]

  const btnClass = 'flex items-center justify-center rounded border border-white/20 bg-white/10 text-xs font-mono text-white/90 active:bg-white/25 select-none'
  const modActiveClass = 'bg-[var(--color-primary)] border-[var(--color-primary)] text-white'
</script>

<!-- Snippets: always visible -->
{#if snippets.length > 0}
  <div class="flex gap-1 overflow-x-auto rounded-b bg-black/80 px-2 py-1" style="scrollbar-width: none;">
    {#each snippets as snippet (snippet.label)}
      <button
        onclick={() => sendSnippet(snippet.command)}
        class="{btnClass} shrink-0 px-2.5 py-1 text-[10px] text-amber-300/90 border-amber-400/30 bg-amber-400/10"
      >{snippet.label}</button>
    {/each}
  </div>
{/if}

<!-- Special keys: touch devices only -->
<div class="terminal-touch-keys hidden flex-col gap-1 bg-black/80 px-2 py-1.5 backdrop-blur"
  class:rounded-b={snippets.length === 0}>
  <div class="flex items-center gap-1">
    <div class="flex flex-1 flex-wrap gap-1">
      {#each specialKeys as key (key.label)}
        {@const isCtrl = key.label === 'Ctrl'}
        {@const isAlt = key.label === 'Alt'}
        {@const isActive = (isCtrl && ctrlActive) || (isAlt && altActive)}
        <button
          onclick={key.action}
          class="{btnClass} min-h-[36px] min-w-[36px] px-2 {isActive ? modActiveClass : ''}"
        >{key.label}</button>
      {/each}
    </div>

    <div class="grid grid-cols-3 grid-rows-2 gap-0.5" style="width: 90px;">
      {#each arrowKeys as arrow (arrow.label)}
        <button
          onclick={() => send(arrow.seq)}
          class="{btnClass} h-[30px] w-[28px]"
          style="grid-area: {arrow.gridArea};"
        >{arrow.label}</button>
      {/each}
    </div>
  </div>
</div>

<style>
  @media (pointer: coarse) {
    .terminal-touch-keys {
      display: flex;
    }
  }
</style>
