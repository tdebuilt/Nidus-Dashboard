<script lang="ts">
  import { FileQuestion } from 'lucide-svelte';
  import { navigate } from '../lib/router';
</script>

<div class="flex items-center justify-center min-h-screen" style="background-color: var(--color-bg-primary);">
  <div class="text-center">
    <FileQuestion size={48} class="mx-auto mb-4" style="color: var(--color-text-muted);" />
    <h1 class="text-4xl font-bold mb-2">404</h1>
    <p class="mb-6" style="color: var(--color-text-secondary);">Page introuvable</p>
    <button
      class="px-4 py-2 rounded-lg transition-colors"
      style="background-color: var(--color-accent); color: white;"
      onclick={() => navigate('dashboard')}
    >
      Retour au dashboard
    </button>
  </div>
</div>
