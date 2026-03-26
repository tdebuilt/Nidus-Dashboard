<script lang="ts">
  interface Props {
    size?: number
  }

  const { size = 24 }: Props = $props()
</script>

<img src="/icons/icon-192.png" alt="Nidus" width={size} height={size} class="inline-block" />
