<script lang="ts">
  import { t } from '../../i18n'

  import type { ComponentType, SvelteComponent } from 'svelte'

  interface TabDef {
    id: string
    labelKey: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon: ComponentType<SvelteComponent<any>>
  }

  interface Props {
    tabs: TabDef[]
    activeTab: string
    onTabChange: (id: string) => void
  }

  const { tabs, activeTab, onTabChange }: Props = $props()
</script>

<!-- Desktop: pills -->
<nav class="mb-6 hidden flex-wrap gap-2 md:flex" data-testid="settings-nav">
  {#each tabs as tab (tab.id)}
    <button
      onclick={() => onTabChange(tab.id)}
      class="flex items-center gap-2 rounded-full px-4 py-2 text-sm font-medium transition-colors
        {activeTab === tab.id
          ? 'bg-[var(--color-primary)] text-white shadow-sm'
          : 'border border-[var(--color-border)] text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]'}"
      data-testid="settings-pill-{tab.id}"
    >
      <tab.icon size={16} />
      {$t(tab.labelKey)}
    </button>
  {/each}
</nav>

<!-- Mobile: dropdown -->
<div class="mb-6 md:hidden">
  <select
    value={activeTab}
    onchange={(e) => onTabChange((e.target as HTMLSelectElement).value)}
    aria-label="Settings"
    class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 text-sm text-[var(--color-text)]"
    data-testid="settings-nav-mobile"
  >
    {#each tabs as tab (tab.id)}
      <option value={tab.id}>{$t(tab.labelKey)}</option>
    {/each}
  </select>
</div>
