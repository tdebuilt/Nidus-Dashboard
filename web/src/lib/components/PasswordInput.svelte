<script lang="ts">
  import { Eye, EyeOff } from 'lucide-svelte'

  interface Props {
    id: string
    value: string
    placeholder?: string
    autocomplete?: HTMLInputElement['autocomplete']
    required?: boolean
    testid?: string
  }

   
  let {
    id,
    value = $bindable(),
    placeholder = '',
    autocomplete = 'new-password',
    required = false,
    testid = '',
  }: Props = $props()
   

  let showPassword = $state(false)
</script>

<div class="relative">
  <input
    {id}
    type={showPassword ? 'text' : 'password'}
    bind:value
    {placeholder}
    {autocomplete}
    {required}
    data-testid={testid}
    class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] px-3 py-2 pe-10 text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
  />
  <button
    type="button"
    onclick={() => (showPassword = !showPassword)}
    class="absolute end-2 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
    aria-label={showPassword ? 'Hide password' : 'Show password'}
    data-testid={testid ? `${testid}-toggle` : ''}
  >
    {#if showPassword}
      <EyeOff size={18} />
    {:else}
      <Eye size={18} />
    {/if}
  </button>
</div>
