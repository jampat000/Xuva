<script lang="ts">
  import type { Snippet } from 'svelte';
  let { children }: { children: Snippet } = $props();
</script>

<!-- Strip all chrome — pure black fullscreen canvas for the player -->
{@render children()}
