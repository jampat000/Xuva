<script lang="ts">
	import type { Snippet } from 'svelte';
	import TopBar from '$lib/lorivo/TopBar.svelte';

	let {
		children
	}: {
		children: Snippet;
	} = $props();
</script>

<div class="min-h-screen bg-[#0B1120] font-sans text-white antialiased">
	<TopBar />
	{@render children()}
	<div class="h-16"></div>
</div>
