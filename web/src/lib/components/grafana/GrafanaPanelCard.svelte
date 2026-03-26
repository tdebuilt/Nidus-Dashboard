<script lang="ts">
  interface Props {
    url: string
    title: string
  }

  const { url, title }: Props = $props()
</script>

<div class="flex flex-col gap-1 overflow-hidden rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-secondary)]">
  <span class="truncate px-2 pt-1 text-xs font-medium text-[var(--color-text-muted)]">{title}</span>
  <iframe
    src={url}
    {title}
    class="w-full border-0"
    style="height: 200px;"
    loading="lazy"
    sandbox="allow-scripts allow-same-origin"
  ></iframe>
</div>
