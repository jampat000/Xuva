<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import { onMount } from 'svelte';
	import { getSettings } from '$lib/api/operator';
	import { xuvaTitle, normalizeServerName } from '$lib/server-name';
	import '../app.css';
	import '$lib/styles/tokens.css';
	import '$lib/styles/base.css';

	let { children } = $props();
	let serverName = $state('Xuva');

	function applyTitle(): void {
		if (typeof document === 'undefined') return;
		setTimeout(() => {
			document.title = xuvaTitle(serverName);
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
		window.addEventListener('xuva:server-name-changed', handleServerNameChanged);
		return () => window.removeEventListener('xuva:server-name-changed', handleServerNameChanged);
	});

	async function loadServerName(): Promise<void> {
		try {
			const payload = await getSettings();
			serverName = normalizeServerName(payload.config?.serverName);
		} catch {
			serverName = 'Xuva';
		} finally {
			applyTitle();
		}
	}
</script>

<svelte:head>
	<title>Xuva</title>
	<link rel="icon" href="/favicon.svg" />
	<meta
		name="description" content="Xuva private cinema and personal media server."
	/>
    <meta property="og:title" content="Xuva - your media, beautifully played" />
	<meta property="og:description" content="A modern media player for your movies, shows, and music." />
</svelte:head>

{@render children()}
