<script lang="ts">
  import { Loader2 } from 'lucide-svelte'
  import { t } from '../i18n'
  import { getWidget, type WidgetDefinition } from '../widgetRegistry'
  import type { GridWidget } from './widgets/gridEngine'
  import type { Component } from 'svelte'

  interface Props {
    widget: GridWidget
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    loadedComponent: Component<any> | null
    active: boolean
  }

  const { widget, loadedComponent, active }: Props = $props()

  const widgetDef: WidgetDefinition | undefined = $derived(getWidget(widget.type))

  function getWidgetProps(w: GridWidget, def: WidgetDefinition): Record<string, unknown> {
    const props: Record<string, unknown> = { config: w.config, active }
    for (const key of def.extraProps ?? []) {
      if (key === 'widgetId') props[key] = w.id
      else if (key === 'widgetType') props[key] = w.type
      else if (key === 'widgetTitle') props[key] = w.title
    }
    return props
  }
</script>

{#if widgetDef && loadedComponent}
  {@const WidgetComp = loadedComponent}
  <div class="mt-2 flex-1 overflow-y-auto rounded-lg bg-[var(--color-bg)] p-2 text-xs text-[var(--color-text-muted)]">
    <WidgetComp {...getWidgetProps(widget, widgetDef)} />
  </div>
{:else if widgetDef}
  <div class="mt-2 flex flex-1 items-center justify-center rounded-lg bg-[var(--color-bg)] p-2">
    <Loader2 size={16} class="animate-spin text-[var(--color-text-muted)]" />
  </div>
{:else}
  <div class="mt-2 flex-1 overflow-y-auto rounded-lg bg-[var(--color-bg)] p-2 text-xs text-[var(--color-text-muted)]">
    {$t('widget.placeholder')}
  </div>
{/if}
