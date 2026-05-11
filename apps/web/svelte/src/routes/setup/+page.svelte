<script lang="ts">
	import { onMount } from 'svelte';
	import { MediaShell, VyrdenButton, VyrdenEmptyState, VyrdenPanel } from '$lib/components';
	import { getAuthSession, type AuthSessionUser } from '$lib/api/auth';
	import { ApiClientError } from '$lib/api/client';
	import { getLibraries, type LibraryRecord } from '$lib/api/home';
	import { browseFolder, saveLibrary, startLibraryScan, type FolderBrowseResponse } from '$lib/api/setup';

	let isLoading = $state(true);
	let isSubmitting = $state(false);
	let isBrowsing = $state(false);
	let searchValue = $state('');

	let authRequired = $state(false);
	let loadError = $state('');
	let actionError = $state('');
	let actionMessage = $state('');

	let user = $state<AuthSessionUser | null>(null);
	let libraries = $state<LibraryRecord[]>([]);
	let folderBrowse = $state<FolderBrowseResponse | null>(null);

	let libraryName = $state('');
	let libraryKind = $state<'movies' | 'tv'>('movies');
	let libraryPath = $state('');
	let runScanAfterSave = $state(true);

	const hasLibraries = $derived.by(() => libraries.length > 0);
	const userDisplayName = $derived.by(() => user?.displayName || user?.username || 'Local User');
	const userInitials = $derived.by(() => {
		const words = userDisplayName
			.split(/\s+/)
			.filter(Boolean)
			.slice(0, 2)
			.map((item) => item[0]?.toUpperCase() || '');
		return words.join('') || 'C';
	});

	onMount(() => {
		void initialize();
	});

	async function initialize(): Promise<void> {
		isLoading = true;
		loadError = '';
		actionError = '';
		actionMessage = '';
		authRequired = false;

		try {
			const session = await getAuthSession().catch((error: unknown) => {
				if (isApiStatus(error, 401)) return null;
				throw error;
			});
			if (!session?.user) {
				authRequired = true;
				return;
			}
			user = session.user;
			const librariesPayload = await getLibraries();
			libraries = librariesPayload.libraries || [];
		} catch (error) {
			loadError = formatError(error, 'Setup could not load.');
		} finally {
			isLoading = false;
		}
	}

	async function openBrowser(path = ''): Promise<void> {
		actionError = '';
		isBrowsing = true;
		try {
			folderBrowse = await browseFolder(path || libraryPath);
		} catch (error) {
			actionError = formatError(error, 'Folder browser could not load.');
		} finally {
			isBrowsing = false;
		}
	}

	async function createLibrary(): Promise<void> {
		if (isSubmitting) return;
		actionError = '';
		actionMessage = '';

		const nameValue = asText(libraryName);
		const pathValue = asText(libraryPath);
		if (!nameValue) {
			actionError = 'Enter a library name.';
			return;
		}
		if (!pathValue) {
			actionError = 'Select a folder path.';
			return;
		}

		isSubmitting = true;
		try {
			const created = await saveLibrary({
				name: nameValue,
				kind: libraryKind,
				path: pathValue
			});
			actionMessage = `${created.name || 'Library'} saved.`;

			if (runScanAfterSave && created.id) {
				const scan = await startLibraryScan(created.id);
				actionMessage = `${created.name || 'Library'} saved. Scan started (${asText(scan.id) || 'queued'}).`;
			}

			libraryName = '';
			if (libraryKind === 'movies') {
				libraryKind = 'tv';
			}
			libraryPath = '';
			folderBrowse = null;

			const refreshed = await getLibraries();
			libraries = refreshed.libraries || [];
		} catch (error) {
			actionError = formatError(error, 'Library setup failed.');
		} finally {
			isSubmitting = false;
		}
	}

	function useEntryPath(path: string): void {
		libraryPath = asText(path);
	}

	function asText(value: unknown): string {
		return String(value ?? '').trim();
	}

	function isApiStatus(error: unknown, expectedStatus: number): boolean {
		if (error instanceof ApiClientError) return error.status === expectedStatus;
		if (typeof error !== 'object' || !error) return false;
		return Number((error as { status?: unknown }).status) === expectedStatus;
	}

	function formatError(error: unknown, fallback: string): string {
		if (error instanceof ApiClientError) return error.userMessage || error.message || fallback;
		if (error instanceof Error) return error.message || fallback;
		return fallback;
	}
</script>

<MediaShell active="setup" bind:searchValue userInitials={userInitials}>
	<div class="setup-page">
		{#if isLoading}
			<VyrdenPanel title="Loading setup" subtitle="Checking account and library status." />
		{:else if authRequired}
			<VyrdenPanel title="Sign in required" subtitle="Sign in to set up your first library.">
				<div class="actions">
					<VyrdenButton variant="primary" href="/signin">Open Sign In</VyrdenButton>
					<VyrdenButton variant="ghost" href="/">Back to Home</VyrdenButton>
				</div>
			</VyrdenPanel>
		{:else if loadError}
			<VyrdenPanel title="Setup could not load" subtitle={loadError}>
				<div class="actions">
					<VyrdenButton variant="secondary" onclick={initialize}>Retry</VyrdenButton>
				</div>
			</VyrdenPanel>
		{:else}
			<VyrdenPanel title="First library setup" subtitle="Add a Movies or TV folder to start building your Lorivo home.">
				<form class="setup-form" onsubmit={(event) => { event.preventDefault(); void createLibrary(); }}>
					<label class="field">
						<span>Library name</span>
						<input bind:value={libraryName} placeholder={libraryKind === 'movies' ? 'Movies' : 'TV Shows'} />
					</label>

					<label class="field">
						<span>Library type</span>
						<select bind:value={libraryKind}>
							<option value="movies">Movies</option>
							<option value="tv">TV Shows</option>
						</select>
					</label>

					<label class="field">
						<span>Folder path</span>
						<input bind:value={libraryPath} placeholder="D:\\Media\\Movies" />
					</label>

					<div class="actions">
						<VyrdenButton variant="secondary" type="button" onclick={() => openBrowser()}>
							{isBrowsing ? 'Loading folders...' : 'Browse folders'}
						</VyrdenButton>
						<VyrdenButton variant="primary" disabled={isSubmitting}>
							{isSubmitting ? 'Saving library...' : 'Save library'}
						</VyrdenButton>
					</div>

					<label class="check">
						<input type="checkbox" bind:checked={runScanAfterSave} />
						<span>Start a scan after saving</span>
					</label>

					{#if actionError}
						<p class="error">{actionError}</p>
					{/if}
					{#if actionMessage}
						<p class="status">{actionMessage}</p>
					{/if}
				</form>
			</VyrdenPanel>

			{#if folderBrowse}
				<VyrdenPanel
					title="Folder browser"
					subtitle={folderBrowse.path || 'Select a folder path for your library.'}
				>
					<div class="browser-actions">
						{#if folderBrowse.parent}
							<VyrdenButton variant="ghost" onclick={() => openBrowser(folderBrowse?.parent || '')}>Up one folder</VyrdenButton>
						{/if}
						{#if folderBrowse.path}
							<VyrdenButton variant="secondary" onclick={() => useEntryPath(folderBrowse?.path || '')}>Use this folder</VyrdenButton>
						{/if}
					</div>

					{#if folderBrowse.error}
						<p class="error">{folderBrowse.error}</p>
					{:else if (folderBrowse.entries || []).length === 0}
						<p class="muted">No child folders here. Use this folder path if it contains media.</p>
					{:else}
						<div class="folder-list">
							{#each folderBrowse.entries || [] as entry}
								<button
									type="button"
									class="folder-entry"
									onclick={() => openBrowser(asText(entry.path))}
								>
									<span>{asText(entry.name) || 'Folder'}</span>
									<small>{asText(entry.path)}</small>
								</button>
							{/each}
						</div>
					{/if}
				</VyrdenPanel>
			{/if}

			{#if hasLibraries}
				<VyrdenPanel title="Libraries configured" subtitle="Your server is ready for media browsing.">
					<VyrdenButton variant="primary" href="/">Open Home</VyrdenButton>
				</VyrdenPanel>
			{:else}
				<VyrdenEmptyState
					title="No libraries configured yet"
					message="Add at least one Movies or TV folder to populate your Home route."
				/>
			{/if}
		{/if}
	</div>
</MediaShell>

<style>
	.setup-page {
		display: grid;
		gap: 12px;
		padding-bottom: var(--vyrden-space-8);
	}

	.setup-form {
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

	.field input,
	.field select {
		border: 1px solid var(--vyrden-color-border-soft);
		background: var(--vyrden-color-surface-elevated);
		color: var(--vyrden-color-text);
		border-radius: var(--vyrden-radius-md);
		min-height: 38px;
		padding: 0 10px;
		font: inherit;
	}

	.actions,
	.browser-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.check {
		display: inline-flex;
		gap: 8px;
		align-items: center;
		font-size: 0.88rem;
		color: var(--vyrden-color-text-muted);
	}

	.error {
		margin: 0;
		color: var(--vyrden-color-danger, #ff9f9f);
		font-size: 0.85rem;
	}

	.status {
		margin: 0;
		color: var(--vyrden-color-accent-teal);
		font-size: 0.85rem;
	}

	.muted {
		margin: 0;
		color: var(--vyrden-color-text-muted);
		font-size: 0.88rem;
	}

	.folder-list {
		display: grid;
		gap: 6px;
		max-height: 260px;
		overflow: auto;
		padding-right: 2px;
	}

	.folder-entry {
		border: 1px solid var(--vyrden-color-border-soft);
		background: var(--vyrden-color-surface-elevated);
		color: var(--vyrden-color-text);
		border-radius: var(--vyrden-radius-md);
		padding: 8px 10px;
		text-align: left;
		display: grid;
		gap: 3px;
		cursor: pointer;
	}

	.folder-entry small {
		color: var(--vyrden-color-text-muted);
		font-size: 0.75rem;
	}
</style>
