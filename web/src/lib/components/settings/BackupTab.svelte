<script lang="ts">
  import { Upload, Download } from 'lucide-svelte'
  import { api } from '../../api/client'
  import { toasts } from '../../stores/toast'
  import { t, translate } from '../../i18n'

  interface Props {
    onSettingsReload?: () => void
  }

  const { onSettingsReload }: Props = $props()

  let showExportPassword = $state(false)
  let showImportPassword = $state(false)
  let exportPassword = $state('')
  let importPassword = $state('')
  let importFile = $state<File | null>(null)

  function openExportDialog() {
    exportPassword = ''
    showExportPassword = true
  }

  async function confirmExport() {
    if (!exportPassword) {
      toasts.error(translate('settings.passwordRequired'))
      return
    }
    try {
      const result = await api.post<{ data: string; salt?: string; kdf?: string }>('/api/config/export', { password: exportPassword })
      const fileContent = JSON.stringify({ data: result.data, salt: result.salt, kdf: result.kdf })
      const blob = new Blob([fileContent], { type: 'application/octet-stream' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'nidus-backup.enc'
      a.click()
      URL.revokeObjectURL(url)
      showExportPassword = false
      toasts.success(translate('settings.exportSuccess'))
    } catch {
      toasts.error(translate('settings.exportError'))
    }
  }

  function handleImportFile(e: Event) {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) return
    importFile = file
    importPassword = ''
    showImportPassword = true
    input.value = ''
  }

  async function confirmImport() {
    if (!importPassword) {
      toasts.error(translate('settings.passwordRequired'))
      return
    }
    if (!importFile) return
    try {
      const text = (await importFile.text()).trim()
      let payload: { password: string; data: string; salt?: string; kdf?: string }
      try {
        const parsed = JSON.parse(text)
        payload = { password: importPassword, data: parsed.data, salt: parsed.salt, kdf: parsed.kdf }
      } catch {
        payload = { password: importPassword, data: text }
      }
      await api.post('/api/config/import', payload)
      showImportPassword = false
      importFile = null
      toasts.success(translate('settings.importSuccess'))
      onSettingsReload?.()
    } catch {
      toasts.error(translate('settings.importError'))
    }
  }

  async function exportYAML() {
    try {
      const res = await fetch('/api/config/yaml', { credentials: 'include' })
      if (!res.ok) throw new Error()
      const text = await res.text()
      const blob = new Blob([text], { type: 'application/x-yaml' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'nidus-config.yaml'
      a.click()
      URL.revokeObjectURL(url)
      toasts.success(translate('settings.yamlExportSuccess'))
    } catch {
      toasts.error(translate('settings.yamlExportError'))
    }
  }

  function handleYAMLImportFile(e: Event) {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) return
    input.value = ''
    importYAMLFile(file)
  }

  async function importYAMLFile(file: File) {
    try {
      const text = await file.text()
      const res = await fetch('/api/config/yaml', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/x-yaml' },
        body: text,
      })
      if (!res.ok) {
        const data = await res.json().catch(() => ({ error: 'Import failed' }))
        throw new Error(data.error)
      }
      toasts.success(translate('settings.yamlImportSuccess'))
      onSettingsReload?.()
    } catch {
      toasts.error(translate('settings.yamlImportError'))
    }
  }
</script>

<div class="space-y-6">
  <h3 class="text-lg font-semibold text-[var(--color-text)]">{$t('settings.tabs.backup')}</h3>

  <!-- Encrypted Export/Import -->
  <section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-export-import">
    <div class="mb-3 flex items-center gap-2">
      <Upload size={18} class="text-[var(--color-text-secondary)]" />
      <h3 class="font-semibold text-[var(--color-text)]">{$t('settings.exportImportSection')}</h3>
    </div>
    <div class="flex gap-3">
      <button onclick={openExportDialog}
        class="flex items-center gap-2 rounded-lg border border-[var(--color-border)] px-4 py-2 text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]"
        data-testid="settings-export-btn">
        <Download size={16} /> {$t('common.export')}
      </button>
      <label class="flex cursor-pointer items-center gap-2 rounded-lg border border-[var(--color-border)] px-4 py-2 text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]" data-testid="settings-import-btn">
        <Upload size={16} /> {$t('common.import')}
        <input type="file" accept=".enc,.bin" onchange={handleImportFile} class="hidden" data-testid="settings-import-input" />
      </label>
    </div>

    {#if showExportPassword}
      <div class="mt-4 rounded-lg border border-[var(--color-border)] p-4" data-testid="export-password-dialog">
        <p class="mb-2 text-sm font-medium text-[var(--color-text)]">{$t('settings.exportPasswordTitle')}</p>
        <p class="mb-3 text-xs text-[var(--color-text-secondary)]">{$t('settings.exportPasswordHint')}</p>
        <input type="password" bind:value={exportPassword} placeholder="••••••••"
          class="mb-3 w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
          data-testid="export-password-input" />
        <div class="flex justify-end gap-2">
          <button onclick={() => showExportPassword = false}
            class="rounded-lg border border-[var(--color-border)] px-4 py-1.5 text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]"
            data-testid="export-password-cancel">{$t('common.cancel')}</button>
          <button onclick={confirmExport}
            class="rounded-lg bg-[var(--color-primary)] px-4 py-1.5 text-sm text-white hover:bg-[var(--color-primary-hover)]"
            data-testid="export-password-confirm">{$t('common.export')}</button>
        </div>
      </div>
    {/if}

    {#if showImportPassword}
      <div class="mt-4 rounded-lg border border-[var(--color-border)] p-4" data-testid="import-password-dialog">
        <p class="mb-2 text-sm font-medium text-[var(--color-text)]">{$t('settings.importPasswordTitle')}</p>
        <p class="mb-3 text-xs text-[var(--color-text-secondary)]">{$t('settings.importPasswordHint')}</p>
        <input type="password" bind:value={importPassword} placeholder="••••••••"
          class="mb-3 w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
          data-testid="import-password-input" />
        <div class="flex justify-end gap-2">
          <button onclick={() => { showImportPassword = false; importFile = null }}
            class="rounded-lg border border-[var(--color-border)] px-4 py-1.5 text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]"
            data-testid="import-password-cancel">{$t('common.cancel')}</button>
          <button onclick={confirmImport}
            class="rounded-lg bg-[var(--color-primary)] px-4 py-1.5 text-sm text-white hover:bg-[var(--color-primary-hover)]"
            data-testid="import-password-confirm">{$t('common.import')}</button>
        </div>
      </div>
    {/if}
  </section>

  <!-- YAML Config -->
  <section class="rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5" data-testid="settings-yaml">
    <div class="mb-3 flex items-center gap-2">
      <Download size={18} class="text-[var(--color-text-secondary)]" />
      <h3 class="font-semibold text-[var(--color-text)]">{$t('settings.yamlSection')}</h3>
    </div>
    <p class="mb-3 text-xs text-[var(--color-text-secondary)]">{$t('settings.yamlHint')}</p>
    <div class="flex gap-3">
      <button onclick={exportYAML}
        class="flex items-center gap-2 rounded-lg border border-[var(--color-border)] px-4 py-2 text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]"
        data-testid="settings-yaml-export-btn">
        <Download size={16} /> {$t('settings.yamlExport')}
      </button>
      <label class="flex cursor-pointer items-center gap-2 rounded-lg border border-[var(--color-border)] px-4 py-2 text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]" data-testid="settings-yaml-import-btn">
        <Upload size={16} /> {$t('settings.yamlImport')}
        <input type="file" accept=".yaml,.yml" onchange={handleYAMLImportFile} class="hidden" data-testid="settings-yaml-import-input" />
      </label>
    </div>
  </section>
</div>
