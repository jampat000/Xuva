<script lang="ts">
	import { onMount } from 'svelte';
	import { getAuthSession, type AuthSessionUser } from '$lib/api/auth';
	import {
		getMovies,
		refreshMetadataBatch,
		scanMovies,
		type MovieListItem
	} from '$lib/api/browse';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import { getLibraries, type LibraryRecord } from '$lib/api/home';
	import { resolvePreviewMode } from '$lib/home/model';
	import { previewPoster } from '$lib/preview/artwork';
	import {
		BrowseFilterGroup,
		BrowseGrid,
		BrowseHeader,
		MediaShell,
		BrowsePage,
		BrowseStatChip,
		BrowseToolbar,
		PosterCard,
		VyrdenButton,
		VyrdenEmptyState,
		VyrdenPanel
	} from '$lib/components';
	import {
		buildMovieCards,
		filterAndSortMovieCards,
		type MovieFilter,
		type MovieSort
	} from '$lib/browse/model';

	let isLoading = $state(true);
	let isScanning = $state(false);
	let isRefreshing = $state(false);
	let loadError = $state('');
	let actionMessage = $state('');
	let searchValue = $state('');
	let previewMode = $state(false);
	let movieFilter = $state<MovieFilter>('all');
	let movieSort = $state<MovieSort>('title');
	let user = $state<AuthSessionUser | null>(null);
	let libraries = $state<LibraryRecord[]>([]);
	let movieRows = $state<MovieListItem[]>([]);

	const movieCards = $derived.by(() => buildMovieCards(movieRows));
	const visibleCards = $derived.by(() =>
		filterAndSortMovieCards(movieCards, searchValue, movieFilter, movieSort)
	);
	const renderedCards = $derived.by(() =>
		previewMode
			? visibleCards.map((item) => ({
					...item,
					meta: item.year > 0 ? `${item.year} · Movie` : 'Movie'
				}))
			: visibleCards
	);
	const userDisplayName = $derived.by(() => user?.displayName || user?.username || 'Local User');
	const userInitials = $derived.by(() => initialsForName(userDisplayName));
	const reviewCount = $derived.by(() => movieCards.filter((item) => item.needsReview).length);
	const metadataPendingCount = $derived.by(() => movieCards.filter((item) => !item.hasMetadata).length);
	const multiVersionCount = $derived.by(() => movieCards.filter((item) => item.versionCount > 1).length);

	onMount(() => {
		try {
			const params = new URL(window.location.href).searchParams;
			previewMode = resolvePreviewMode(params);
			const q = params.get('q');
			if (q) searchValue = q;
		} catch {
			// Ignore URL parsing errors and use default search state.
		}
		void loadMovies();
	});

	async function loadMovies(): Promise<void> {
		isLoading = true;
		loadError = '';
		const activePreviewMode = resolvePreviewMode(new URL(window.location.href).searchParams);
		if (activePreviewMode) {
			previewMode = true;
			user = null;
			libraries = [];
			movieRows = previewMovieRows();
			isLoading = false;
			return;
		}
		try {
			const [sessionPayload, librariesPayload, moviesPayload] = await Promise.all([
				getAuthSession(apiClient).catch((error: unknown) => {
					if (isApiStatus(error, 401)) return { user: null };
					throw error;
				}),
				getLibraries(apiClient),
				getMovies(apiClient, 500)
			]);
			user = sessionPayload?.user || null;
			libraries = librariesPayload.libraries || [];
			movieRows = moviesPayload.movies || [];
			if (activePreviewMode && movieRows.length === 0) {
				movieRows = previewMovieRows();
			}
		} catch (error) {
			if (activePreviewMode && isApiStatus(error, 401)) {
				user = null;
				movieRows = previewMovieRows();
				loadError = '';
			} else {
				loadError = formatLoadError(error);
			}
		} finally {
			isLoading = false;
		}
	}

	async function startMovieScan(): Promise<void> {
		isScanning = true;
		actionMessage = '';
		try {
			await scanMovies(apiClient, 50);
			actionMessage = 'Movie scan started. Refreshing the wall...';
			await loadMovies();
		} catch (error) {
			actionMessage = formatLoadError(error);
		} finally {
			isScanning = false;
		}
	}

	async function runMetadataRefresh(): Promise<void> {
		isRefreshing = true;
		actionMessage = '';
		try {
			const result = await refreshMetadataBatch('movie', apiClient, 25);
			if (Array.isArray(result.warnings) && result.warnings.length > 0) {
				actionMessage = result.warnings.slice(0, 3).join(' | ');
			} else {
				actionMessage = 'Metadata refresh accepted.';
			}
			await loadMovies();
		} catch (error) {
			actionMessage = formatLoadError(error);
		} finally {
			isRefreshing = false;
		}
	}

	function formatCount(value: number): string {
		if (!Number.isFinite(value)) return '0';
		return new Intl.NumberFormat().format(Math.max(0, Math.round(value)));
	}

	function initialsForName(name: string): string {
		const words = asText(name).split(/\s+/).filter(Boolean);
		if (words.length === 0) return 'V';
		if (words.length === 1) return words[0].slice(0, 1).toUpperCase();
		return `${words[0][0] || ''}${words[1][0] || ''}`.toUpperCase();
	}

	function asText(value: unknown): string {
		return String(value ?? '').trim();
	}

	function isApiStatus(error: unknown, expectedStatus: number): boolean {
		if (error instanceof ApiClientError) return error.status === expectedStatus;
		if (typeof error !== 'object' || !error) return false;
		const candidate = (error as { status?: unknown }).status;
		return Number(candidate) === expectedStatus;
	}

	function formatLoadError(error: unknown): string {
		if (error instanceof ApiClientError) return error.userMessage || error.message;
		if (isApiStatus(error, 401)) return 'Your session is no longer active. Sign in again to continue.';
		if (error instanceof Error) return error.message;
		return 'Movies could not load.';
	}

	function previewMovieRows(): MovieListItem[] {
		return [
			{
				id: 'preview-movie-ember-harbor',
				title: 'Ember Harbor',
				year: 2025,
				versionCount: 2,
				needsReview: false,
				metadata: {
					title: 'Ember Harbor',
					year: 2025,
					overview: 'Preview movie item.',
					posterUrl: previewArtwork('Ember Harbor')
				}
			},
			{
				id: 'preview-movie-atlas-of-dawn',
				title: 'Atlas of Dawn',
				year: 2024,
				versionCount: 2,
				needsReview: false,
				metadata: { title: 'Atlas of Dawn', year: 2024, overview: 'Preview movie item.', posterUrl: previewArtwork('Atlas of Dawn') }
			},
			{
				id: 'preview-movie-hinterland',
				title: 'Hinterland',
				year: 2023,
				versionCount: 1,
				needsReview: true,
				metadata: { title: 'Hinterland', year: 2023, overview: 'Preview movie item.', posterUrl: previewArtwork('Hinterland') }
			},
			{
				id: 'preview-movie-coastline',
				title: 'Coastline',
				year: 2022,
				versionCount: 1,
				needsReview: false,
				metadata: { title: 'Coastline', year: 2022, overview: 'Preview movie item.', posterUrl: previewArtwork('Coastline') }
			},
			{
				id: 'preview-movie-polar-night',
				title: 'Polar Night',
				year: 2021,
				versionCount: 3,
				needsReview: false,
				metadata: { title: 'Polar Night', year: 2021, overview: 'Preview movie item.', posterUrl: previewArtwork('Polar Night') }
			},
			{
				id: 'preview-movie-night-archive',
				title: 'Night Archive',
				year: 2024,
				versionCount: 1,
				needsReview: false,
				metadata: { title: 'Night Archive', year: 2024, overview: 'Preview movie item.', posterUrl: previewArtwork('Night Archive') }
			},
			{
				id: 'preview-movie-return-vector',
				title: 'Return Vector',
				year: 2023,
				versionCount: 2,
				needsReview: false,
				metadata: { title: 'Return Vector', year: 2023, overview: 'Preview movie item.', posterUrl: previewArtwork('Return Vector') }
			},
			{
				id: 'preview-movie-last-orchard',
				title: 'The Last Orchard',
				year: 2020,
				versionCount: 1,
				needsReview: false,
				metadata: { title: 'The Last Orchard', year: 2020, overview: 'Preview movie item.', posterUrl: previewArtwork('The Last Orchard') }
			},
			{
				id: 'preview-movie-violet-signal',
				title: 'Violet Signal',
				year: 2024,
				versionCount: 1,
				needsReview: true,
				metadata: { title: 'Violet Signal', year: 2024, overview: 'Preview movie item.', posterUrl: previewArtwork('Violet Signal') }
			},
			{
				id: 'preview-movie-broken-current',
				title: 'Broken Current',
				year: 2021,
				versionCount: 2,
				needsReview: false,
				metadata: { title: 'Broken Current', year: 2021, overview: 'Preview movie item.', posterUrl: previewArtwork('Broken Current') }
			},
			{
				id: 'preview-movie-glass-canyon',
				title: 'Glass Canyon',
				year: 2019,
				versionCount: 1,
				needsReview: false,
				metadata: { title: 'Glass Canyon', year: 2019, overview: 'Preview movie item.', posterUrl: previewArtwork('Glass Canyon') }
			},
			{
				id: 'preview-movie-copper-sky',
				title: 'Copper Sky',
				year: 2022,
				versionCount: 1,
				needsReview: false,
				metadata: { title: 'Copper Sky', year: 2022, overview: 'Preview movie item.', posterUrl: previewArtwork('Copper Sky') }
			}
		];
	}

	function previewArtwork(title: string): string {
		return previewPoster(title);
	}
</script>

<MediaShell active="movies" bind:searchValue {userInitials}>
	<BrowsePage>
		{#if isLoading}
			<VyrdenPanel title="Loading Movies" subtitle="Fetching your movie library from the media APIs." />
		{:else if loadError}
			<VyrdenPanel title="Movies could not load" subtitle={loadError}>
				<div class="status-actions">
					<VyrdenButton variant="secondary" onclick={loadMovies}>Retry</VyrdenButton>
					<VyrdenButton variant="ghost" href="/">Back to Home</VyrdenButton>
				</div>
			</VyrdenPanel>
		{:else}
			<BrowseHeader title="Movies" subtitle="Browse your movie library.">
				{#snippet chips()}
					<BrowseStatChip label={`${formatCount(renderedCards.length)} visible`} />
					{#if !previewMode && reviewCount > 0}
						<BrowseStatChip label={`${formatCount(reviewCount)} review`} />
					{/if}
					{#if !previewMode && metadataPendingCount > 0}
						<BrowseStatChip label={`${formatCount(metadataPendingCount)} metadata pending`} />
					{/if}
					{#if !previewMode && multiVersionCount > 0}
						<BrowseStatChip label={`${formatCount(multiVersionCount)} multi-version`} />
					{/if}
				{/snippet}
			</BrowseHeader>

			<BrowseToolbar message={actionMessage}>
				{#snippet controls()}
					{#if !previewMode}
						<BrowseFilterGroup ariaLabel="Movie filters">
							<button
								type="button"
								class:selected={movieFilter === 'all'}
								onclick={() => (movieFilter = 'all')}
							>
								All
							</button>
							<button
								type="button"
								class:selected={movieFilter === 'review'}
								onclick={() => (movieFilter = 'review')}
							>
								Needs Review
							</button>
							<button
								type="button"
								class:selected={movieFilter === 'metadata'}
								onclick={() => (movieFilter = 'metadata')}
							>
								Metadata Pending
							</button>
							<button
								type="button"
								class:selected={movieFilter === 'versions'}
								onclick={() => (movieFilter = 'versions')}
							>
								Multiple Versions
							</button>
						</BrowseFilterGroup>
					{/if}
					<BrowseFilterGroup segmented ariaLabel="Movie sorting">
						<button type="button" class:selected={movieSort === 'title'} onclick={() => (movieSort = 'title')}>
							Title
						</button>
						<button type="button" class:selected={movieSort === 'year'} onclick={() => (movieSort = 'year')}>
							Year
						</button>
						{#if !previewMode}
							<button
								type="button"
								class:selected={movieSort === 'versions'}
								onclick={() => (movieSort = 'versions')}
							>
								Versions
							</button>
							<button
								type="button"
								class:selected={movieSort === 'review'}
								onclick={() => (movieSort = 'review')}
							>
								Review
							</button>
						{/if}
					</BrowseFilterGroup>
				{/snippet}
				{#snippet actions()}
					{#if !previewMode}
						<VyrdenButton variant="primary" onclick={startMovieScan} disabled={isScanning || isRefreshing}>
							{isScanning ? 'Scanning...' : 'Scan Movies'}
						</VyrdenButton>
						<VyrdenButton variant="secondary" onclick={runMetadataRefresh} disabled={isScanning || isRefreshing}>
							{isRefreshing ? 'Refreshing...' : 'Refresh Metadata'}
						</VyrdenButton>
					{/if}
				{/snippet}
			</BrowseToolbar>

			{#if movieCards.length === 0}
				<VyrdenEmptyState
					title="No movies found"
					message="Try adding a movie library or running a scan."
				/>
			{:else if renderedCards.length === 0}
				<VyrdenEmptyState title="No movies found" message="Try changing filters or search terms." />
			{:else}
				<BrowseGrid>
					{#each renderedCards as item (item.id)}
						<PosterCard
							title={item.title}
							meta={item.meta}
							imageUrl={item.posterUrl}
							href={`/movies/${encodeURIComponent(item.id)}${previewMode ? '?preview=1' : ''}`}
						/>
					{/each}
				</BrowseGrid>
			{/if}
		{/if}
	</BrowsePage>
</MediaShell>

<style>
	.status-actions {
		display: flex;
		flex-wrap: wrap;
		gap: var(--vyrden-space-2);
	}
</style>
