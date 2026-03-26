<script lang="ts">
  import { t } from '../../i18n'

  interface Props {
    value?: string
    onchange?: (value: string) => void
  }

  const { value = '{}', onchange }: Props = $props()

  let selectedType = $state('jellyfin')

  $effect(() => {
    try {
      const parsed = JSON.parse(value)
      selectedType = parsed.server_type ?? 'jellyfin'
    } catch {
      selectedType = 'jellyfin'
    }
  })

  function handleChange(e: Event) {
    const select = e.target as HTMLSelectElement
    selectedType = select.value
    onchange?.(JSON.stringify({ server_type: selectedType }))
  }
</script>

<div class="space-y-2">
  <label for="mediaserver-type-select" class="block text-sm text-[var(--color-text-secondary)]">
    {$t('mediaserver.serverType')}
  </label>
  <select
    id="mediaserver-type-select"
    onchange={handleChange}
    class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
    data-testid="mediaserver-config-type"
  >
    <option value="plex" selected={selectedType === 'plex'}>Plex</option>
    <option value="jellyfin" selected={selectedType === 'jellyfin'}>Jellyfin</option>
  </select>
  <p class="text-xs text-[var(--color-text-muted)]">{$t('mediaserver.configHint')}</p>
</div>
