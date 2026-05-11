<script lang="ts">
	import { onMount } from 'svelte';
	import {
		BrowseHeader,
		MediaRow,
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
		asText,
		findRow,
		formatLoadError,
		initialsForName,
		itemDetailHref,
		loadSecondaryRouteContextSafe
	} from '$lib/secondary/media-routes';
	import type { HomeDisplayItem } from '$lib/home/model';

	let isLoading = $state(true);
	let loadError = $state('');
	let authMessage = $state('');
	let searchValue = $state('');
	let rowAvailable = $state(false);
	let userDisplayName = $state('Local User');
	let userRole = $state('Local Account');
	let userInitials = $state('V');
	let libraries = $state<Array<{ id?: string; name?: string; kind?: string }>>([]);
	let movieItems = $state<HomeDisplayItem[]>([]);
	let tvItems = $state<HomeDisplayItem[]>([]);

	const visibleMovies = $derived.by(() => {
		const needle = asText(searchValue).toLowerCase();
		if (!needle) return movieItems;
		return movieItems.filter((item) => item.searchText.includes(needle));
	});

	const visibleTV = $derived.by(() => {
		const needle = asText(searchValue).toLowerCase();
		if (!needle) return tvItems;
		return tvItems.filter((item) => item.searchText.includes(needle));
	});

	const totalVisible = $derived.by(() => visibleMovies.length + visibleTV.length);

	onMount(() => {
		void loadRecentlyAdded();
	});

	async function loadRecentlyAdded(): Promise<void> {
		isLoading = true;
		loadError = '';
		authMessage = '';
		rowAvailable = false;
		try {
			const outcome = await loadSecondaryRouteContextSafe({ limit: 60 });
			if (outcome.kind === 'auth') {
				authMessage = outcome.message;
				movieItems = [];
				tvItems = [];
				return;
			}
			if (outcome.kind === 'error') {
				throw outcome.error;
			}
			const context = outcome.context;
			userDisplayName =
				context.user?.displayName || context.user?.username || userDisplayName;
			userRole = context.user?.role || 'Local Account';
			userInitials = initialsForName(userDisplayName);
			libraries = context.libraries || [];

			const row = findRow(context.homePayload, 'recently-added');
			rowAvailable = Boolean(row);

			if (row && Array.isArray(row.items) && row.items.length > 0) {
				const movieIDs = new Set<string>();
				const tvIDs = new Set<string>();
				for (const item of row.items) {
					const id = asText(item?.id || item?.mediaSourceId);
					const kind = asText(item?.kind).toLowerCase();
					if (!id) continue;
					if (kind === 'movie') movieIDs.add(id);
					if (kind === 'series' || kind === 'tv') tvIDs.add(id);
				}
				movieItems = context.model.movieItems.filter((item) => movieIDs.has(item.id));
				tvItems = context.model.tvItems.filter((item) => tvIDs.has(item.id));
			} else {
				movieItems = context.model.movieItems;
				tvItems = context.model.tvItems;
			}
		} catch (error) {
			loadError = formatLoadError(error, 'Recently Added');
		} finally {
			isLoading = false;
		}
	}
</script>

<MediaShell active="recently-added" bind:searchValue {userInitials}>
	<BrowsePage>
		{#if isLoading}
			<VyrdenPanel
				title="Loading Recently Added"
				subtitle="Fetching recent movie and TV additions from existing APIs."
			/>
		{:else if loadError}
			<VyrdenPanel title="Recently Added could not load" subtitle={loadError}>
				<div class="status-actions">
					<VyrdenButton variant="secondary" onclick={loadRecentlyAdded}>Retry</VyrdenButton>
					<VyrdenButton variant="ghost" href="/">Back to Home</VyrdenButton>
				</div>
			</VyrdenPanel>
		{:else if authMessage}
			<VyrdenPanel title="Sign in required" subtitle={authMessage}>
				<div class="status-actions">
					<VyrdenButton variant="secondary" onclick={loadRecentlyAdded}>Retry</VyrdenButton>
					<VyrdenButton variant="ghost" href="/signin">Open Sign In</VyrdenButton>
				</div>
			</VyrdenPanel>
		{:else}
			<BrowseHeader title="Recently Added" subtitle="Latest additions across movies and TV.">
				{#snippet chips()}
					<BrowseStatChip label={`${totalVisible} visible`} />
					<BrowseStatChip label={`${visibleMovies.length} movies`} />
					<BrowseStatChip label={`${visibleTV.length} shows`} />
					{#if !rowAvailable}
						<BrowseStatChip label="Not available yet" />
					{/if}
				{/snippet}
			</BrowseHeader>

			<BrowseToolbar />

			{#if !rowAvailable}
				<VyrdenEmptyState
					title="Recently Added is not available yet"
					message="The current backend payload does not expose a recently added row for this route."
				>
					{#snippet action()}
						<VyrdenButton variant="secondary" href="/">Back to Home</VyrdenButton>
					{/snippet}
				</VyrdenEmptyState>
			{:else if movieItems.length === 0 && tvItems.length === 0}
				<VyrdenEmptyState
					title="Nothing recently added"
					message="Scan your libraries to populate recent additions."
				/>
			{:else if totalVisible === 0}
				<VyrdenEmptyState
					title="No recent matches"
					message="Try changing search terms to find recent titles."
				/>
			{:else}
				{#if visibleMovies.length > 0}
					<section class="row">
						<MediaRow title="Recently Added Movies" linkLabel="View all" linkHref="/movies" trackGap={10}>
							{#each visibleMovies as item (item.id)}
								<PosterCard
									title={item.title}
									meta={item.subtitle || item.meta}
									imageUrl={item.posterUrl || item.backdropUrl}
									href={itemDetailHref(item)}
								/>
							{/each}
						</MediaRow>
					</section>
				{/if}
				{#if visibleTV.length > 0}
					<section class="row">
						<MediaRow title="Recently Added TV" linkLabel="View all" linkHref="/tv" trackGap={10}>
							{#each visibleTV as item (item.id)}
								<PosterCard
									title={item.title}
									meta={item.subtitle || item.meta}
									imageUrl={item.posterUrl || item.backdropUrl}
									href={itemDetailHref(item)}
								/>
							{/each}
						</MediaRow>
					</section>
				{/if}
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

	.row {
		display: grid;
		gap: 12px;
	}

</style>

