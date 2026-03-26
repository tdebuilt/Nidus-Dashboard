<script lang="ts">
  import { icons } from 'lucide-svelte';
  import type { ComponentType } from 'svelte';

  interface Props {
    name: string
    size?: number
    color?: string
    strokeWidth?: number
    [key: string]: unknown
  }

  const { name, size = 24, color = 'currentColor', strokeWidth = 2, ...rest }: Props = $props()

  // Convert any format (lowercase, kebab-case) to PascalCase to match lucide-svelte exports
  function toPascalCase(s: string): string {
    return s.replace(/(^|[-_ ])(\w)/g, (_, _sep, c) => c.toUpperCase())
  }

  // Aliases for renamed/removed icons
  const aliases: Record<string, string> = {
    home: 'House',
    Home: 'House',
  }

  const iconsMap = icons as Record<string, ComponentType>
  const iconComponent = $derived(
    iconsMap[name] ?? iconsMap[toPascalCase(name)] ?? iconsMap[aliases[name]] ?? null
  )
</script>

{#if iconComponent}
  {@const Icon = iconComponent}
  <Icon {size} {color} {strokeWidth} {...rest} />
{:else}
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    stroke={color}
    stroke-width={strokeWidth}
    stroke-linecap="round"
    stroke-linejoin="round"
    class="lucide-icon lucide-fallback"
    aria-hidden="true"
  >
    <circle cx="12" cy="12" r="10" />
    <line x1="12" y1="8" x2="12" y2="12" />
    <line x1="12" y1="16" x2="12.01" y2="16" />
  </svg>
{/if}
