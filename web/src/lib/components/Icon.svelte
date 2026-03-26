<script lang="ts">
  import type { Icon as LucideIcon } from 'lucide-svelte'
  import type { ComponentType } from 'svelte'

  interface Props {
    icon: ComponentType<LucideIcon>
    size?: number
    class?: string
    strokeWidth?: number
  }

  const { icon: IconComponent, size = 20, class: className = '', strokeWidth = 2 }: Props = $props()
</script>

<IconComponent {size} {strokeWidth} class={className} />
