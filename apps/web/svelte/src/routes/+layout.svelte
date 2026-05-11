<script lang="ts">
	import '../app.css';
	import '$lib/styles/tokens.css';
	import '$lib/styles/base.css';

	let { children } = $props();
</script>

<svelte:head>
	<title>Lorivo</title>
	<link rel="icon" href="/favicon.svg" />
	<meta
		name="description"
		content="Lorivo private cinema and personal media server."
	/>
</svelte:head>

{@render children()}
