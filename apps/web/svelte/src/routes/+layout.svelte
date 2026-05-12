<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import { onMount } from 'svelte';
	import { getSettings } from '$lib/api/operator';
	import { lorivoTitle, normalizeServerName } from '$lib/server-name';
	import '../app.css';
	import '$lib/styles/tokens.css';
	import '$lib/styles/base.css';

	let { children } = $props();
	let serverName = $state('Lorivo');

	function applyTitle(): void {
		if (typeof document === 'undefined') return;
		setTimeout(() => {
			document.title = lorivoTitle(serverName);
		}, 0);
	}

	afterNavigate(applyTitle);

	onMount(() => {
		void loadServerName();
		const handleServerNameChanged = (event: Event) => {
			const detail = (event as CustomEvent<{ serverName?: string }>).detail;
			serverName = normalizeServerName(detail?.serverName);
			applyTitle();
		};
		window.addEventListener('lorivo:server-name-changed', handleServerNameChanged);
		return () => window.removeEventListener('lorivo:server-name-changed', handleServerNameChanged);
	});

	async function loadServerName(): Promise<void> {
		try {
			const payload = await getSettings();
			serverName = normalizeServerName(payload.config?.serverName);
		} catch {
			serverName = 'Lorivo';
		} finally {
			applyTitle();
		}
	}
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
