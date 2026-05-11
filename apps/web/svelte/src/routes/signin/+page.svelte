<script lang="ts">
	import { onMount } from 'svelte';
	import { VyrdenButton, VyrdenPanel } from '$lib/components';
	import {
		bootstrapAccount,
		getAuthSession,
		getClientBootstrap,
		login,
		type ClientBootstrapResponse
	} from '$lib/api/auth';
	import { ApiClientError } from '$lib/api/client';
	import { getLibraries } from '$lib/api/home';

	let isLoading = $state(true);
	let isSubmitting = $state(false);
	let mode = $state<'signin' | 'bootstrap'>('signin');
	let statusMessage = $state('');
	let errorMessage = $state('');

	let username = $state('');
	let displayName = $state('');
	let password = $state('');
	let confirmPassword = $state('');

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
			mode = auth.bootstrapAllowed ? 'bootstrap' : 'signin';
			username = asText(auth.defaultUsername) || username;

			const session = await getAuthSession().catch((error: unknown) => {
				if (isApiStatus(error, 401)) return null;
				throw error;
			});
			if (session?.user) {
				await redirectAfterSignIn();
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

		const usernameValue = asText(username);
		const passwordValue = asText(password);
		const displayNameValue = asText(displayName);

		if (!usernameValue) {
			errorMessage = 'Enter your username.';
			return;
		}
		if (!passwordValue) {
			errorMessage = 'Enter your password.';
			return;
		}
		if (mode === 'bootstrap') {
			if (!displayNameValue) {
				errorMessage = 'Enter your display name.';
				return;
			}
			if (passwordValue !== confirmPassword) {
				errorMessage = 'Passwords do not match.';
				return;
			}
		}

		isSubmitting = true;
		try {
			if (mode === 'bootstrap') {
				await bootstrapAccount({
					username: usernameValue,
					password: passwordValue,
					displayName: displayNameValue
				});
				statusMessage = 'Account created. Loading your library setup...';
			} else {
				await login({
					username: usernameValue,
					password: passwordValue
				});
				statusMessage = 'Signed in. Loading your library...';
			}
			await redirectAfterSignIn();
		} catch (error) {
			errorMessage = formatAuthError(
				error,
				mode === 'bootstrap' ? 'Account setup failed.' : 'Sign-in failed.'
			);
		} finally {
			isSubmitting = false;
		}
	}

	async function redirectAfterSignIn(): Promise<void> {
		const libraries = await getLibraries().catch((error: unknown) => {
			if (isApiStatus(error, 401)) return { libraries: [] };
			throw error;
		});
		const hasLibraries = (libraries.libraries || []).length > 0;
		window.location.href = hasLibraries ? '/' : '/setup';
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
		<VyrdenPanel
			title={mode === 'bootstrap' ? 'Create your Lorivo account' : 'Sign in to Lorivo'}
			subtitle={mode === 'bootstrap'
				? 'Set up the first local account to start your server.'
				: 'Use your account to open your library.'}
		>
			{#if isLoading}
				<p class="auth-copy">Checking server authentication status…</p>
			{:else}
				<form class="auth-form" onsubmit={(event) => { event.preventDefault(); void submit(); }}>
					<label class="field">
						<span>Username</span>
						<input bind:value={username} autocomplete="username" />
					</label>
					{#if mode === 'bootstrap'}
						<label class="field">
							<span>Display name</span>
							<input bind:value={displayName} autocomplete="name" />
						</label>
					{/if}
					<label class="field">
						<span>Password</span>
						<input type="password" bind:value={password} autocomplete="current-password" />
					</label>
					{#if mode === 'bootstrap'}
						<label class="field">
							<span>Confirm password</span>
							<input type="password" bind:value={confirmPassword} autocomplete="new-password" />
						</label>
					{/if}

					{#if errorMessage}
						<p class="auth-error">{errorMessage}</p>
					{/if}
					{#if statusMessage}
						<p class="auth-status">{statusMessage}</p>
					{/if}

					<div class="actions">
						<VyrdenButton variant="primary" type="submit" disabled={isSubmitting}>
							{isSubmitting
								? mode === 'bootstrap'
									? 'Creating account...'
									: 'Signing in...'
								: mode === 'bootstrap'
									? 'Create account'
									: 'Sign in'}
						</VyrdenButton>
						<VyrdenButton variant="ghost" href="/">Back to Home</VyrdenButton>
					</div>
				</form>
			{/if}
		</VyrdenPanel>
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
		width: min(460px, 100%);
	}

	.auth-copy {
		margin: 0;
		color: var(--vyrden-color-text-muted);
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
		color: var(--vyrden-color-text-muted);
	}

	.field input {
		border: 1px solid var(--vyrden-color-border-soft);
		background: var(--vyrden-color-surface-elevated);
		color: var(--vyrden-color-text);
		border-radius: var(--vyrden-radius-md);
		min-height: 38px;
		padding: 0 11px;
		font: inherit;
	}

	.auth-error {
		margin: 0;
		font-size: 0.85rem;
		color: var(--vyrden-color-danger, #ff9f9f);
	}

	.auth-status {
		margin: 0;
		font-size: 0.85rem;
		color: var(--vyrden-color-accent-teal);
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-top: 4px;
	}
</style>
