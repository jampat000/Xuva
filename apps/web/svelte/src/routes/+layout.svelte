<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import { onMount } from 'svelte';
	import { getAuthSession, getClientBootstrap } from '$lib/api/auth';
	import { xuvaTitle, normalizeServerName } from '$lib/server-name';
	import '../app.css';
	import '$lib/styles/tokens.css';
	import '$lib/styles/base.css';

	let { children } = $props();
	let serverName = $state('Xuva');
	let accessCheckInFlight = false;

	function applyTitle(): void {
		if (typeof document === 'undefined') return;
		setTimeout(() => {
			document.title = xuvaTitle(serverName);
		}, 0);
	}

	afterNavigate(() => {
		applyTitle();
		void enforceRouteAccess();
	});

	onMount(() => {
		void loadServerName();
		void enforceRouteAccess();
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

	async function enforceRouteAccess(): Promise<void> {
		if (typeof window === 'undefined') return;
		if (accessCheckInFlight) return;
		accessCheckInFlight = true;
		try {
			const path = window.location.pathname || '/';
			const isSignIn = path.startsWith('/signin');
			const isSetupWizard = path.startsWith('/setup-wizard');
			const bootstrap = await getClientBootstrap().catch(() => null);
			const authRequired = Boolean((bootstrap as { auth?: { required?: boolean } } | null)?.auth?.required);
			const bootstrapAllowed = Boolean(
				(bootstrap as { auth?: { bootstrapAllowed?: boolean } } | null)?.auth?.bootstrapAllowed
			);
			if (!authRequired) return;

			const session = await getAuthSession().catch(() => null);
			const role = asText(session?.user?.role).toLowerCase();
			const isAdmin = role === 'admin';
			const isSignedIn = Boolean(session?.user);

			if (!isSignedIn) {
				if (isSetupWizard && !bootstrapAllowed) {
					redirectTo('/signin');
					return;
				}
				if (!isSignIn && !(isSetupWizard && bootstrapAllowed)) {
					redirectTo('/signin');
					return;
				}
				return;
			}

			if (isSignIn || isSetupWizard) {
				redirectTo('/');
				return;
			}

			if (path.startsWith('/settings') && !isAdmin) {
				redirectTo('/');
			}
		} catch {
			// Keep layout resilient if auth checks fail transiently.
		} finally {
			accessCheckInFlight = false;
		}
	}

	function redirectTo(targetPath: string): void {
		if (typeof window === 'undefined') return;
		const currentPath = window.location.pathname || '/';
		if (currentPath === targetPath) return;
		window.location.replace(targetPath);
	}

	function asText(value: unknown): string {
		return String(value ?? '').trim();
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
