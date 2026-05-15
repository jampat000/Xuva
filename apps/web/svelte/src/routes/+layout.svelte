<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import { onMount } from 'svelte';
	import { getClientBootstrap } from '$lib/api/auth';
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
			const payload = await getClientBootstrap();
			const server = (payload as { server?: { name?: string } }).server;
			serverName = normalizeServerName(server?.name);
		} catch {
			serverName = 'Xuva';
		} finally {
			applyTitle();
		}
	}
</script>

<svelte:head>
	<title>Xuva</title>
	<link rel="icon" type="image/svg+xml" href="/favicon.svg" />
	<link rel="icon" type="image/png" sizes="32x32" href="/favicon-32.png" />
	<link rel="icon" type="image/png" sizes="16x16" href="/favicon-16.png" />
	<link rel="apple-touch-icon" href="/apple-touch-icon.png" />
	<meta name="theme-color" content="#7C3AED" />
	<meta
		name="description" content="Xuva private cinema and personal media server."
	/>
    <meta property="og:title" content="Xuva - your media, beautifully played" />
	<meta property="og:description" content="A modern media player for your movies, shows, and music." />
</svelte:head>

{@render children()}
