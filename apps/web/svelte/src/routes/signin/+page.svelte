<script lang="ts">
	import { onMount } from 'svelte';
	import { XuvaButton, XuvaPanel } from '$lib/components';
	import {
		bootstrapAccount,
		getAuthSessionIfAvailable,
		getClientBootstrap,
		login
	} from '$lib/api/auth';
	import { updateSettings } from '$lib/api/operator';
	import { saveLibrary, startLibraryScan } from '$lib/api/setup';
	import { ApiClientError } from '$lib/api/client';

	type AuthMode = 'signin' | 'bootstrap';
	type BootstrapStep = 'language' | 'server' | 'account' | 'libraries';

	let isLoading = $state(true);
	let isSubmitting = $state(false);
	let mode = $state<AuthMode>('signin');
	let bootstrapStep = $state<BootstrapStep>('language');
	let statusMessage = $state('');
	let errorMessage = $state('');

	let username = $state('');
	let displayName = $state('');
	let password = $state('');
	let confirmPassword = $state('');

	let language = $state('en');
	let serverName = $state('Xuva');
	let moviesPath = $state('');
	let tvPath = $state('');
	let moviesName = $state('Movies');
	let tvName = $state('TV Shows');
	let runScanAfterSave = $state(true);

	const bootstrapStepNumber = $derived.by(() => {
		if (bootstrapStep === 'language') return 1;
		if (bootstrapStep === 'server') return 2;
		if (bootstrapStep === 'account') return 3;
		return 4;
	});

	onMount(() => {
		void initialize();
	});

	async function initialize(): Promise<void> {
		isLoading = true;
		errorMessage = '';
		statusMessage = '';

	try {
			const bootstrap = await getClientBootstrap();
			const auth = bootstrap.auth || {};
			const wantsSetup = new URL(window.location.href).searchParams.get('setup') === '1';
			mode = wantsSetup && auth.bootstrapAllowed ? 'bootstrap' : 'signin';
			bootstrapStep = 'language';
			username = asText(auth.defaultUsername) || username;
			serverName = asText((bootstrap as { server?: { name?: unknown } })?.server?.name) || 'Xuva';

			const session = await getAuthSessionIfAvailable().catch((error: unknown) => {
				if (isApiStatus(error, 401)) return null;
				throw error;
			});
			if (session?.user) {
				window.location.href = '/';
				return;
			}
		} catch (error) {
			errorMessage = formatAuthError(error, 'Sign-in status could not be loaded.');
		} finally {
			isLoading = false;
		}
	}

	async function submit(): Promise<void> {
		if (isSubmitting) return;
		errorMessage = '';
		statusMessage = '';

		if (mode === 'signin') {
			await submitSignIn();
			return;
		}

		if (bootstrapStep === 'language') {
			saveLanguageChoice();
			bootstrapStep = 'server';
			return;
		}
		if (bootstrapStep === 'server') {
			const serverNameValue = asText(serverName);
			if (!serverNameValue) {
				errorMessage = 'Enter a server name.';
				return;
			}
			if (serverNameValue.length > 50) {
				errorMessage = 'Server name must be 50 characters or fewer.';
				return;
			}
			bootstrapStep = 'account';
			return;
		}
		if (bootstrapStep === 'account') {
			await submitBootstrapAccount();
			return;
		}
		await finishBootstrap(false);
	}

	function saveLanguageChoice(): void {
		const value = asText(language) || 'en';
		try {
			localStorage.setItem('xuva.language', value);
		} catch {
			// Ignore storage failures in restricted browser contexts.
		}
		if (typeof document !== 'undefined') {
			document.documentElement.lang = value;
		}
	}

	async function submitSignIn(): Promise<void> {
		const usernameValue = asText(username);
		const passwordValue = asText(password);

		if (!usernameValue) {
			errorMessage = 'Enter your username.';
			return;
		}
		if (!passwordValue) {
			errorMessage = 'Enter your password.';
			return;
		}

		isSubmitting = true;
		try {
			await login({
				username: usernameValue,
				password: passwordValue
			});
			window.location.href = '/';
		} catch (error) {
			errorMessage = formatAuthError(error, 'Sign-in failed.');
		} finally {
			isSubmitting = false;
		}
	}

	async function submitBootstrapAccount(): Promise<void> {
		const usernameValue = asText(username);
		const passwordValue = asText(password);
		const displayNameValue = asText(displayName);
		const serverNameValue = asText(serverName);

		if (!usernameValue) {
			errorMessage = 'Enter an admin username.';
			return;
		}
		if (!displayNameValue) {
			errorMessage = 'Enter an admin display name.';
			return;
		}
		if (!passwordValue) {
			errorMessage = 'Enter a password.';
			return;
		}
		if (passwordValue !== confirmPassword) {
			errorMessage = 'Passwords do not match.';
			return;
		}

		isSubmitting = true;
		try {
			await bootstrapAccount({
				username: usernameValue,
				password: passwordValue,
				displayName: displayNameValue
			});

			let serverNameSaved = false;
			try {
				await updateSettings({ serverName: serverNameValue });
				serverNameSaved = true;
			} catch {
				statusMessage =
					'Admin account created. Server name could not be saved right now. You can update it in Settings > General.';
			}

			password = '';
			confirmPassword = '';
			bootstrapStep = 'libraries';
			if (serverNameSaved) {
				statusMessage = 'Admin account created. Server name saved. Add library paths now or skip for later.';
			}
		} catch (error) {
			errorMessage = formatAuthError(error, 'Account setup failed.');
		} finally {
			isSubmitting = false;
		}
	}

	async function finishBootstrap(skipLibraries: boolean): Promise<void> {
		if (isSubmitting) return;
		errorMessage = '';
		statusMessage = '';

		const moviesPathValue = asText(moviesPath);
		const tvPathValue = asText(tvPath);
		const moviesNameValue = asText(moviesName) || 'Movies';
		const tvNameValue = asText(tvName) || 'TV Shows';

		isSubmitting = true;
		try {
			if (!skipLibraries) {
				if (moviesPathValue) {
					const created = await saveLibrary({
						name: moviesNameValue,
						kind: 'movies',
						path: moviesPathValue
					});
					if (runScanAfterSave && created.id) {
						await startLibraryScan(created.id);
					}
				}
				if (tvPathValue) {
					const created = await saveLibrary({
						name: tvNameValue,
						kind: 'tv',
						path: tvPathValue
					});
					if (runScanAfterSave && created.id) {
						await startLibraryScan(created.id);
					}
				}
			}
			window.location.href = '/';
		} catch (error) {
			errorMessage = formatAuthError(error, 'Initial library setup failed.');
		} finally {
			isSubmitting = false;
		}
	}

	function asText(value: unknown): string {
		return String(value ?? '').trim();
	}

	function isApiStatus(error: unknown, expectedStatus: number): boolean {
		if (error instanceof ApiClientError) return error.status === expectedStatus;
		if (typeof error !== 'object' || !error) return false;
		return Number((error as { status?: unknown }).status) === expectedStatus;
	}

	function formatAuthError(error: unknown, fallback: string): string {
		if (error instanceof ApiClientError) return error.userMessage || error.message || fallback;
		if (error instanceof Error) return error.message || fallback;
		return fallback;
	}
</script>

<div class="auth-shell">
	<div class="auth-card">
		<XuvaPanel
			title={mode === 'bootstrap' ? 'Initial Setup' : 'Sign in to Xuva'}
			subtitle={mode === 'bootstrap'
				? `Step ${bootstrapStepNumber} of 4: ${bootstrapStep === 'language' ? 'Language' : bootstrapStep === 'server' ? 'Server name' : bootstrapStep === 'account' ? 'Admin account' : 'Library paths'}`
				: 'Use your account to open your library.'}
		>
			{#if isLoading}
				<p class="auth-copy">Checking server authentication status...</p>
			{:else if mode === 'signin'}
				<form class="auth-form" onsubmit={(event) => { event.preventDefault(); void submit(); }}>
					<label class="field">
						<span>Username</span>
						<input bind:value={username} autocomplete="username" />
					</label>
					<label class="field">
						<span>Password</span>
						<input type="password" bind:value={password} autocomplete="current-password" />
					</label>

					{#if errorMessage}
						<p class="auth-error">{errorMessage}</p>
					{/if}

					<div class="actions">
						<XuvaButton variant="primary" type="submit" disabled={isSubmitting}>
							{isSubmitting ? 'Signing in...' : 'Sign in'}
						</XuvaButton>
						<XuvaButton variant="ghost" href="/">Back to Home</XuvaButton>
					</div>
				</form>
			{:else}
				<form class="auth-form" onsubmit={(event) => { event.preventDefault(); void submit(); }}>
					{#if bootstrapStep === 'language'}
						<p class="auth-copy">Choose your preferred language for this browser session.</p>
						<label class="field">
							<span>Language</span>
							<select bind:value={language}>
								<option value="en">English</option>
								<option value="es">Spanish</option>
								<option value="fr">French</option>
								<option value="de">German</option>
								<option value="it">Italian</option>
								<option value="pt">Portuguese</option>
							</select>
						</label>
					{:else if bootstrapStep === 'server'}
						<p class="auth-copy">Choose the server name clients will see on your home network.</p>
						<label class="field">
							<span>Server name</span>
							<input bind:value={serverName} maxlength="50" placeholder="Xuva" />
						</label>
					{:else if bootstrapStep === 'account'}
						<p class="auth-copy">Create the first account. This account is the admin for this server.</p>
						<label class="field">
							<span>Admin username</span>
							<input bind:value={username} autocomplete="username" />
						</label>
						<label class="field">
							<span>Admin display name</span>
							<input bind:value={displayName} autocomplete="name" />
						</label>
						<label class="field">
							<span>Password</span>
							<input type="password" bind:value={password} autocomplete="new-password" />
						</label>
						<label class="field">
							<span>Confirm password</span>
							<input type="password" bind:value={confirmPassword} autocomplete="new-password" />
						</label>
					{:else}
						<p class="auth-copy">Add initial library paths now, or skip and do this later from settings.</p>
						<label class="field">
							<span>Movies library path (optional)</span>
							<input bind:value={moviesPath} placeholder="D:\\Media\\Movies" />
						</label>
						<label class="field">
							<span>Movies library name</span>
							<input bind:value={moviesName} placeholder="Movies" />
						</label>
						<label class="field">
							<span>TV library path (optional)</span>
							<input bind:value={tvPath} placeholder="D:\\Media\\TV" />
						</label>
						<label class="field">
							<span>TV library name</span>
							<input bind:value={tvName} placeholder="TV Shows" />
						</label>
						<label class="check">
							<input type="checkbox" bind:checked={runScanAfterSave} />
							<span>Start scan after saving library paths</span>
						</label>
					{/if}

					{#if errorMessage}
						<p class="auth-error">{errorMessage}</p>
					{/if}
					{#if statusMessage}
						<p class="auth-status">{statusMessage}</p>
					{/if}

					<div class="actions">
						<XuvaButton variant="primary" type="submit" disabled={isSubmitting}>
							{#if isSubmitting}
								{bootstrapStep === 'libraries' ? 'Finishing setup...' : 'Saving...'}
							{:else}
								{bootstrapStep === 'language' || bootstrapStep === 'server' ? 'Continue' : bootstrapStep === 'account' ? 'Create admin user' : 'Finish Setup'}
							{/if}
						</XuvaButton>
						{#if bootstrapStep === 'libraries'}
							<XuvaButton
								variant="ghost"
								type="button"
								disabled={isSubmitting}
								onclick={() => void finishBootstrap(true)}
							>
								Skip for now
							</XuvaButton>
						{:else}
							<XuvaButton variant="ghost" href="/">Back to Home</XuvaButton>
						{/if}
					</div>
				</form>
			{/if}
		</XuvaPanel>
	</div>
</div>

<style>
	.auth-shell {
		min-height: 100vh;
		display: grid;
		place-items: center;
		padding: clamp(18px, 4vw, 40px);
		background:
			radial-gradient(circle at 8% 10%, rgb(126 183 169 / 14%), transparent 38%),
			linear-gradient(180deg, #0f1723 0%, #0c131d 55%, #0a1119 100%);
	}

	.auth-card {
		width: min(540px, 100%);
	}

	.auth-copy {
		margin: 0;
		color: var(--xuva-color-text-muted);
		font-size: 0.94rem;
	}

	.auth-form {
		display: grid;
		gap: 10px;
	}

	.field {
		display: grid;
		gap: 5px;
	}

	.field span {
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--xuva-color-text-muted);
	}

	.field input,
	.field select {
		border: 1px solid var(--xuva-color-border-soft);
		background: var(--xuva-color-surface-elevated);
		color: var(--xuva-color-text);
		border-radius: var(--xuva-radius-md);
		min-height: 38px;
		padding: 0 11px;
		font: inherit;
	}

	.auth-error {
		margin: 0;
		font-size: 0.85rem;
		color: var(--xuva-color-danger, #ff9f9f);
	}

	.auth-status {
		margin: 0;
		font-size: 0.85rem;
		color: var(--xuva-color-accent-teal);
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-top: 4px;
	}

	.check {
		display: inline-flex;
		gap: 8px;
		align-items: center;
		font-size: 0.88rem;
		color: var(--xuva-color-text-muted);
	}
</style>
