<script lang="ts">
  import { Plus, Trash2 } from 'lucide-svelte'
  import { t } from '../../i18n'

  interface Snippet {
    label: string
    command: string
  }

  interface Props {
    value?: string
    onchange?: (value: string) => void
  }

  const { value = '{}', onchange }: Props = $props()

  const inputClass = 'mt-1 w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]'

  function parseConfig(json: string) {
    try {
      const p = JSON.parse(json)
      return {
        host: (p.host as string) ?? '',
        port: (p.port as number) ?? 22,
        username: (p.username as string) ?? '',
        password: (p.password as string) ?? '',
        fontSize: (p.font_size as number) ?? 14,
        scrollback: (p.scrollback as number) ?? 1000,
        snippets: Array.isArray(p.snippets) ? (p.snippets as Snippet[]) : [],
      }
    } catch {
      return { host: '', port: 22, username: '', password: '', fontSize: 14, scrollback: 1000, snippets: [] as Snippet[] }
    }
  }

  const initial = parseConfig(value)
  let host = $state(initial.host)
  let port = $state(initial.port)
  let username = $state(initial.username)
  let password = $state(initial.password)
  let fontSize = $state(initial.fontSize)
  let scrollback = $state(initial.scrollback)
  let snippets = $state<Snippet[]>(initial.snippets)
  let newLabel = $state('')
  let newCommand = $state('')

  function emit() {
    const cfg: Record<string, unknown> = {
      host, port, username, password,
      font_size: fontSize, scrollback,
    }
    if (snippets.length > 0) cfg.snippets = snippets
    onchange?.(JSON.stringify(cfg))
  }

  function addSnippet() {
    if (!newLabel.trim() || !newCommand.trim()) return
    snippets = [...snippets, { label: newLabel.trim(), command: newCommand.trim() }]
    newLabel = ''
    newCommand = ''
    emit()
  }

  function removeSnippet(index: number) {
    snippets = snippets.filter((_, i) => i !== index)
    emit()
  }
</script>

<div class="space-y-3">
  <div class="grid grid-cols-3 gap-2">
    <div class="col-span-2">
      <label for="terminal-host" class="block text-sm text-[var(--color-text-secondary)]">
        {$t('terminal.connHost')}
      </label>
      <input id="terminal-host" type="text" bind:value={host} oninput={emit}
        placeholder="192.168.1.100" class={inputClass} />
    </div>
    <div>
      <label for="terminal-port" class="block text-sm text-[var(--color-text-secondary)]">
        {$t('terminal.connPort')}
      </label>
      <input id="terminal-port" type="number" min="1" max="65535"
        bind:value={port} oninput={emit} class={inputClass} />
    </div>
  </div>

  <div class="grid grid-cols-2 gap-2">
    <div>
      <label for="terminal-user" class="block text-sm text-[var(--color-text-secondary)]">
        {$t('settings.username')}
      </label>
      <input id="terminal-user" type="text" bind:value={username} oninput={emit}
        class={inputClass} />
    </div>
    <div>
      <label for="terminal-pass" class="block text-sm text-[var(--color-text-secondary)]">
        {$t('settings.password')}
      </label>
      <input id="terminal-pass" type="password" bind:value={password} oninput={emit}
        class={inputClass} />
    </div>
  </div>

  <hr class="border-[var(--color-border)]" />

  <div>
    <label for="terminal-fontsize" class="block text-sm text-[var(--color-text-secondary)]">
      {$t('terminal.fontSize')}
    </label>
    <div class="mt-1 flex items-center gap-3">
      <input id="terminal-fontsize" type="range" min="10" max="20"
        bind:value={fontSize} oninput={emit} class="flex-1" />
      <span class="w-8 text-center text-sm text-[var(--color-text)]">{fontSize}</span>
    </div>
  </div>

  <div>
    <label for="terminal-scrollback" class="block text-sm text-[var(--color-text-secondary)]">
      {$t('terminal.scrollback')}
    </label>
    <input id="terminal-scrollback" type="number" min="100" max="10000" step="100"
      bind:value={scrollback} oninput={emit} class={inputClass} />
  </div>

  <hr class="border-[var(--color-border)]" />

  <div>
    <p class="text-sm font-medium text-[var(--color-text-secondary)]">{$t('terminal.snippets')}</p>

    {#if snippets.length > 0}
      <div class="mt-2 space-y-1">
        {#each snippets as snippet, i (i)}
          <div class="flex items-center gap-2 rounded border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5">
            <span class="text-xs font-medium text-[var(--color-text)]">{snippet.label}</span>
            <span class="flex-1 truncate text-xs font-mono text-[var(--color-text-muted)]">{snippet.command}</span>
            <button onclick={() => removeSnippet(i)}
              class="p-0.5 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]">
              <Trash2 size={12} />
            </button>
          </div>
        {/each}
      </div>
    {/if}

    <div class="mt-2 flex gap-2">
      <input type="text" bind:value={newLabel} placeholder={$t('terminal.snippetLabel')}
        class="w-24 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-xs text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]" />
      <input type="text" bind:value={newCommand} placeholder={$t('terminal.snippetCommand')}
        class="flex-1 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-2 py-1.5 text-xs font-mono text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]" />
      <button onclick={addSnippet}
        disabled={!newLabel.trim() || !newCommand.trim()}
        class="flex items-center gap-1 rounded-lg border border-[var(--color-border)] px-2 py-1.5 text-xs text-[var(--color-text-secondary)] hover:border-[var(--color-primary)] hover:text-[var(--color-primary)] disabled:opacity-50 disabled:cursor-not-allowed">
        <Plus size={12} />
      </button>
    </div>
  </div>
</div>
