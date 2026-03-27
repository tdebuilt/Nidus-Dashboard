<script lang="ts">
  interface Props {
    jsonText: string
    jsonError: string
    onInput: (e: Event) => void
  }

  const { jsonText, jsonError, onInput }: Props = $props()
</script>

<div class="mb-5">
  <textarea
    value={jsonText}
    oninput={onInput}
    rows="16"
    spellcheck="false"
    class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 font-mono text-xs text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
  ></textarea>
  {#if jsonError}
    <p class="mt-1 text-xs text-[var(--color-danger)]">{jsonError}</p>
  {/if}
</div>
