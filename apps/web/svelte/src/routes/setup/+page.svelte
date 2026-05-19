<script lang="ts">
	import { onMount } from 'svelte';
	import { bootstrapAccount, getAuthSessionIfAvailable } from '$lib/api/auth';
	import { updateSettings } from '$lib/api/operator';
	import { saveLibrary, startLibraryScan, completeSetup, getSetupStatus } from '$lib/api/setup';
	import { ApiClientError } from '$lib/api/client';
	import Logo from '$lib/components/Logo.svelte';

	// ── Step machine ───────────────────────────────────────────────────────────
	type Step = 'loading' | 'account' | 'region' | 'libraries' | 'done';
	let step         = $state<Step>('loading');
	let totalSteps   = $state(3); // account (if needed) + region + libraries
	let needsAccount = $state(true); // declare before stepNumber uses it

	let stepNumber = $derived.by(() => {
		if (step === 'account')   return needsAccount ? 1 : null;
		if (step === 'region')    return needsAccount ? 2 : 1;
		if (step === 'libraries') return needsAccount ? 3 : 2;
		return null;
	});

	// ── Account step ──────────────────────────────────────────────────────────
	let serverName   = $state('Xuva');
	let username     = $state('');
	let displayName  = $state('');
	let password     = $state('');
	let confirmPw    = $state('');

	// ── Region step ───────────────────────────────────────────────────────────
	let country  = $state('');
	let timezone = $state('');

	// ── Libraries step ────────────────────────────────────────────────────────
	let moviesPath   = $state('');
	let moviesName   = $state('Movies');
	let tvPath       = $state('');
	let tvName       = $state('TV Shows');
	let runScan      = $state(true);

	// ── UI state ──────────────────────────────────────────────────────────────
	let submitting = $state(false);
	let errorMsg   = $state('');

	// ── Initialise ────────────────────────────────────────────────────────────
	onMount(async () => {
		// If the user is already logged in and setup is complete, send them home.
		try {
			const session = await getAuthSessionIfAvailable().catch(() => null);
			if (session?.user) {
				const status = await getSetupStatus().catch(() => null);
				if (!status?.requiresSetup) {
					window.location.href = '/';
					return;
				}
				// Logged in but setup not done — skip the account step.
				needsAccount = false;
				totalSteps = 2;
				step = 'region';
				return;
			}
		} catch {
			// Ignore; fall through to account step
		}

		// Not logged in — start from account creation.
		needsAccount = true;
		totalSteps = 3;
		step = 'account';
	});

	// ── Submission handlers ───────────────────────────────────────────────────
	async function submitAccount(): Promise<void> {
		errorMsg = '';
		if (!username.trim())       { errorMsg = 'Enter a username.';           return; }
		if (!displayName.trim())    { errorMsg = 'Enter a display name.';       return; }
		if (!password)              { errorMsg = 'Enter a password.';           return; }
		if (password !== confirmPw) { errorMsg = 'Passwords do not match.';     return; }

		submitting = true;
		try {
			await bootstrapAccount({ username: username.trim(), password, displayName: displayName.trim() });
			try { await updateSettings({ serverName: serverName.trim() || 'Xuva' }); } catch { /* non-fatal */ }
			password = ''; confirmPw = '';
			step = 'region';
		} catch (e) {
			errorMsg = fmtError(e, 'Account creation failed.');
		} finally {
			submitting = false;
		}
	}

	async function submitRegion(): Promise<void> {
		errorMsg = '';
		// Region is optional — user can skip with empty values.
		submitting = true;
		try {
			await completeSetup({ country: country || undefined, timezone: timezone || undefined });
			step = 'libraries';
		} catch (e) {
			errorMsg = fmtError(e, 'Could not save region settings.');
		} finally {
			submitting = false;
		}
	}

	async function submitLibraries(skip: boolean): Promise<void> {
		errorMsg = '';
		submitting = true;
		try {
			if (!skip) {
				if (moviesPath.trim()) {
					const lib = await saveLibrary({ name: moviesName || 'Movies', kind: 'movies', path: moviesPath.trim() });
					if (runScan && lib.id) startLibraryScan(lib.id).catch(() => {});
				}
				if (tvPath.trim()) {
					const lib = await saveLibrary({ name: tvName || 'TV Shows', kind: 'tv', path: tvPath.trim() });
					if (runScan && lib.id) startLibraryScan(lib.id).catch(() => {});
				}
			}
			step = 'done';
		} catch (e) {
			errorMsg = fmtError(e, 'Could not save library.');
		} finally {
			submitting = false;
		}
	}

	function fmtError(e: unknown, fallback: string): string {
		if (e instanceof ApiClientError) return e.userMessage || e.message || fallback;
		if (e instanceof Error) return e.message || fallback;
		return fallback;
	}

	// ── Step progress bar ─────────────────────────────────────────────────────
	const COUNTRIES = [
		{ code: 'AU', label: 'Australia' },
		{ code: 'CA', label: 'Canada' },
		{ code: 'FR', label: 'France' },
		{ code: 'DE', label: 'Germany' },
		{ code: 'IN', label: 'India' },
		{ code: 'IT', label: 'Italy' },
		{ code: 'JP', label: 'Japan' },
		{ code: 'MX', label: 'Mexico' },
		{ code: 'NL', label: 'Netherlands' },
		{ code: 'NZ', label: 'New Zealand' },
		{ code: 'BR', label: 'Brazil' },
		{ code: 'PL', label: 'Poland' },
		{ code: 'PT', label: 'Portugal' },
		{ code: 'ES', label: 'Spain' },
		{ code: 'SE', label: 'Sweden' },
		{ code: 'CH', label: 'Switzerland' },
		{ code: 'GB', label: 'United Kingdom' },
		{ code: 'US', label: 'United States' },
	];

	const TIMEZONES = [
		{ id: 'Pacific/Honolulu',    label: 'Hawaii (UTC−10)' },
		{ id: 'America/Los_Angeles', label: 'Pacific Time (UTC−8/−7)' },
		{ id: 'America/Denver',      label: 'Mountain Time (UTC−7/−6)' },
		{ id: 'America/Chicago',     label: 'Central Time (UTC−6/−5)' },
		{ id: 'America/New_York',    label: 'Eastern Time (UTC−5/−4)' },
		{ id: 'America/Sao_Paulo',   label: 'Brasília (UTC−3)' },
		{ id: 'Atlantic/Reykjavik',  label: 'Reykjavik (UTC+0)' },
		{ id: 'Europe/London',       label: 'London (UTC+0/+1)' },
		{ id: 'Europe/Paris',        label: 'Paris / Berlin (UTC+1/+2)' },
		{ id: 'Europe/Helsinki',     label: 'Helsinki (UTC+2/+3)' },
		{ id: 'Europe/Moscow',       label: 'Moscow (UTC+3)' },
		{ id: 'Asia/Dubai',          label: 'Dubai (UTC+4)' },
		{ id: 'Asia/Kolkata',        label: 'India (UTC+5:30)' },
		{ id: 'Asia/Bangkok',        label: 'Bangkok (UTC+7)' },
		{ id: 'Asia/Singapore',      label: 'Singapore (UTC+8)' },
		{ id: 'Asia/Tokyo',          label: 'Tokyo (UTC+9)' },
		{ id: 'Australia/Sydney',    label: 'Sydney (UTC+10/+11)' },
		{ id: 'Pacific/Auckland',    label: 'Auckland (UTC+12/+13)' },
	];
</script>

<svelte:head>
	<title>Setup — Xuva</title>
</svelte:head>

<div class="relative flex min-h-screen flex-col items-center justify-center bg-background px-6 py-16">
	<!-- Atmospheric glow -->
	<div
		aria-hidden="true"
		class="pointer-events-none absolute inset-0 -z-10"
		style="background: radial-gradient(ellipse at 25% 15%, oklch(0.62 0.22 285 / 0.35), transparent 50%), radial-gradient(ellipse at 80% 85%, oklch(0.72 0.16 255 / 0.28), transparent 50%), radial-gradient(ellipse at 65% 5%, oklch(0.68 0.20 300 / 0.22), transparent 40%);"
	></div>
	<div class="grain pointer-events-none absolute inset-0 -z-10"></div>

	<div class="hairline w-full max-w-sm rounded-3xl bg-surface/60 px-8 py-10 shadow-elev backdrop-blur-xl">
		<!-- Logo + step indicator -->
		<div class="mb-8 flex flex-col items-center gap-3">
			<a href="/" aria-label="Xuva home"><Logo /></a>

			{#if step !== 'loading' && step !== 'done'}
				<!-- Progress dots -->
				<div class="mt-2 flex items-center gap-1.5">
					{#each { length: totalSteps } as _, i (i)}
						{@const n = i + 1}
						<span
							class={`h-1 rounded-full transition-all duration-300 ${
								n === stepNumber
									? 'w-6 bg-primary-glow shadow-glow'
									: n < (stepNumber ?? 0)
										? 'w-3 bg-primary-glow/60'
										: 'w-3 bg-foreground/15'
							}`}
						></span>
					{/each}
				</div>
				<div class="text-center">
					<div class="text-[10px] font-semibold uppercase tracking-[0.3em] text-primary-glow">
						Setup · Step {stepNumber} of {totalSteps}
					</div>
					<h1 class="font-serif-display mt-1 text-2xl tracking-tight">
						{#if step === 'account'}Create your account
						{:else if step === 'region'}Your region
						{:else if step === 'libraries'}Add your media
						{/if}
					</h1>
				</div>
			{/if}
		</div>

		<!-- ── Loading ── -->
		{#if step === 'loading'}
			<p class="text-center text-sm text-muted-foreground">Loading…</p>

		<!-- ── Account step ── -->
		{:else if step === 'account'}
			<p class="mb-4 text-sm text-muted-foreground">
				Create the admin account for this server. You can add more users later.
			</p>
			<form class="space-y-3" onsubmit={(e) => { e.preventDefault(); void submitAccount(); }}>
				<div class="space-y-1.5">
					<label for="sv-name" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Server name</label>
					<input id="sv-name" bind:value={serverName} placeholder="Xuva" maxlength="50"
						class="h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70" />
				</div>
				<div class="space-y-1.5">
					<label for="sv-username" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Username</label>
					<input id="sv-username" bind:value={username} autocomplete="username"
						class="h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70" />
				</div>
				<div class="space-y-1.5">
					<label for="sv-display" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Display name</label>
					<input id="sv-display" bind:value={displayName} autocomplete="name"
						class="h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70" />
				</div>
				<div class="space-y-1.5">
					<label for="sv-pw" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Password</label>
					<input id="sv-pw" type="password" bind:value={password} autocomplete="new-password"
						class="h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70" />
				</div>
				<div class="space-y-1.5">
					<label for="sv-pw2" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Confirm password</label>
					<input id="sv-pw2" type="password" bind:value={confirmPw} autocomplete="new-password"
						class="h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70" />
				</div>

				{#if errorMsg}
					<p class="rounded-xl bg-destructive/10 px-4 py-3 text-sm text-destructive">{errorMsg}</p>
				{/if}

				<button type="submit" disabled={submitting}
					class="mt-1 w-full rounded-full bg-gradient-primary py-3 text-sm font-semibold text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60">
					{submitting ? 'Creating account…' : 'Continue →'}
				</button>
			</form>

		<!-- ── Region step ── -->
		{:else if step === 'region'}
			<p class="mb-4 text-sm text-muted-foreground">
				Choose your country so Xuva can show you trending titles from your region. You can change this later in Settings.
			</p>
			<form class="space-y-3" onsubmit={(e) => { e.preventDefault(); void submitRegion(); }}>
				<div class="space-y-1.5">
					<label for="sv-country" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Country</label>
					<select id="sv-country" bind:value={country}
						class="h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none focus:border-primary/60 focus:bg-background/70">
						<option value="">— Select your country —</option>
						{#each COUNTRIES as c (c.code)}
							<option value={c.code}>{c.label}</option>
						{/each}
					</select>
				</div>
				<div class="space-y-1.5">
					<label for="sv-tz" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Timezone</label>
					<select id="sv-tz" bind:value={timezone}
						class="h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none focus:border-primary/60 focus:bg-background/70">
						<option value="">— Select your timezone —</option>
						{#each TIMEZONES as tz (tz.id)}
							<option value={tz.id}>{tz.label}</option>
						{/each}
					</select>
				</div>

				{#if errorMsg}
					<p class="rounded-xl bg-destructive/10 px-4 py-3 text-sm text-destructive">{errorMsg}</p>
				{/if}

				<div class="flex flex-wrap gap-2 pt-1">
					<button type="submit" disabled={submitting}
						class="flex-1 rounded-full bg-gradient-primary py-3 text-sm font-semibold text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60">
						{submitting ? 'Saving…' : 'Continue →'}
					</button>
					<button type="button" disabled={submitting} onclick={() => void submitRegion()}
						class="hairline rounded-full bg-foreground/[0.04] px-5 py-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground disabled:cursor-not-allowed disabled:opacity-60">
						Skip
					</button>
				</div>
			</form>

		<!-- ── Libraries step ── -->
		{:else if step === 'libraries'}
			<p class="mb-4 text-sm text-muted-foreground">
				Point Xuva at your media folders. You can add more libraries later from Settings.
			</p>
			<form class="space-y-3" onsubmit={(e) => { e.preventDefault(); void submitLibraries(false); }}>
				<div class="space-y-1.5">
					<label for="sv-movies-path" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Movies folder <span class="normal-case text-muted-foreground/60">(optional)</span></label>
					<input id="sv-movies-path" bind:value={moviesPath} placeholder="D:\Media\Movies"
						class="h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm font-mono outline-none placeholder:text-muted-foreground/40 focus:border-primary/60 focus:bg-background/70" />
				</div>
				{#if moviesPath.trim()}
					<div class="space-y-1.5">
						<label for="sv-movies-name" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Movies library name</label>
						<input id="sv-movies-name" bind:value={moviesName} placeholder="Movies"
							class="h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70" />
					</div>
				{/if}
				<div class="space-y-1.5">
					<label for="sv-tv-path" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">TV folder <span class="normal-case text-muted-foreground/60">(optional)</span></label>
					<input id="sv-tv-path" bind:value={tvPath} placeholder="D:\Media\TV"
						class="h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm font-mono outline-none placeholder:text-muted-foreground/40 focus:border-primary/60 focus:bg-background/70" />
				</div>
				{#if tvPath.trim()}
					<div class="space-y-1.5">
						<label for="sv-tv-name" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">TV library name</label>
						<input id="sv-tv-name" bind:value={tvName} placeholder="TV Shows"
							class="h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70" />
					</div>
				{/if}

				{#if moviesPath.trim() || tvPath.trim()}
					<label class="flex cursor-pointer items-center gap-3 pt-1 text-sm text-muted-foreground">
						<input type="checkbox" bind:checked={runScan} class="accent-primary-glow" />
						Start scanning after saving
					</label>
				{/if}

				{#if errorMsg}
					<p class="rounded-xl bg-destructive/10 px-4 py-3 text-sm text-destructive">{errorMsg}</p>
				{/if}

				<div class="flex flex-wrap gap-2 pt-1">
					<button type="submit" disabled={submitting}
						class="flex-1 rounded-full bg-gradient-primary py-3 text-sm font-semibold text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60">
						{submitting ? 'Saving…' : 'Finish setup'}
					</button>
					<button type="button" disabled={submitting} onclick={() => void submitLibraries(true)}
						class="hairline rounded-full bg-foreground/[0.04] px-5 py-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground disabled:cursor-not-allowed disabled:opacity-60">
						Skip for now
					</button>
				</div>
			</form>

		<!-- ── Done ── -->
		{:else if step === 'done'}
			<div class="flex flex-col items-center gap-4 py-4 text-center">
				<div class="flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/15 text-primary-glow">
					<svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
					</svg>
				</div>
				<div>
					<h1 class="font-serif-display text-2xl tracking-tight">You're all set!</h1>
					<p class="mt-2 text-sm text-muted-foreground">
						Your server is ready. Head to your library and start streaming.
					</p>
				</div>
				<a href="/"
					class="mt-2 w-full rounded-full bg-gradient-primary py-3 text-center text-sm font-semibold text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110">
					Go to my library →
				</a>
			</div>
		{/if}
	</div>

	<p class="mt-8 text-center text-[11px] uppercase tracking-[0.2em] text-muted-foreground/50">
		Xuva · Your home cinema, on every screen
	</p>
</div>
