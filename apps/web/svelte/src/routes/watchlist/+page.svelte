<script lang="ts">
	import { onMount } from 'svelte';
	import {
		BrowseGrid,
		BrowseHeader,
		MediaShell,
		BrowsePage,
		BrowseStatChip,
		BrowseToolbar,
		PosterCard,
		XuvaButton,
		XuvaEmptyState,
		XuvaPanel
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
	let items = $state<HomeDisplayItem[]>([]);

	const visibleItems = $derived.by(() => {
		const needle = asText(searchValue).toLowerCase();
		if (!needle) return items;
		return items.filter((item) => item.searchText.includes(needle));
	});

	onMount(() => {
		void loadWatchlist();
	});

	async function loadWatchlist(): Promise<void> {
		isLoading = true;
		loadError = '';
		authMessage = '';
		rowAvailable = false;
		try {
			const outcome = await loadSecondaryRouteContextSafe({ limit: 60 });
			if (outcome.kind === 'auth') {
				authMessage = outcome.message;
				items = [];
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
			rowAvailable = Boolean(findRow(context.homePayload, 'watchlist'));
			items = context.model.watchlistItems;
		} catch (error) {
			loadError = formatLoadError(error, 'Watchlist');
		} finally {
			isLoading = false;
		}
	}
</script>

<MediaShell active="watchlist" bind:searchValue {userInitials}>
	<BrowsePage>
		{#if isLoading}
			<XuvaPanel
				title="Loading Watchlist"
				subtitle="Reading watchlist availability from existing backend APIs."
			/>
		{:else if loadError}
			<XuvaPanel title="Watchlist could not load" subtitle={loadError}>
				<div class="status-actions">
					<XuvaButton variant="secondary" onclick={loadWatchlist}>Retry</XuvaButton>
					<XuvaButton variant="ghost" href="/">Back to Home</XuvaButton>
				</div>
			</XuvaPanel>
		{:else if authMessage}
			<XuvaPanel title="Sign in required" subtitle={authMessage}>
				<div class="status-actions">
					<XuvaButton variant="secondary" onclick={loadWatchlist}>Retry</XuvaButton>
					<XuvaButton variant="ghost" href="/signin">Open Sign In</XuvaButton>
				</div>
			</XuvaPanel>
		{:else}
			<BrowseHeader title="Watchlist" subtitle="Saved titles to watch later.">
				{#snippet chips()}
					<BrowseStatChip label={`${visibleItems.length} visible`} />
					{#if !rowAvailable}
						<BrowseStatChip label="Not available yet" />
					{/if}
				{/snippet}
			</BrowseHeader>

			<BrowseToolbar />

			{#if !rowAvailable}
				<XuvaEmptyState
					title="Watchlist is not available yet"
					message="The current backend home payload does not expose a watchlist row for this route."
				>
					{#snippet action()}
						<XuvaButton variant="secondary" href="/">Back to Home</XuvaButton>
					{/snippet}
				</XuvaEmptyState>
			{:else if items.length === 0}
				<XuvaEmptyState
					title="No watchlist items yet"
					message="Save titles to your watchlist and they will appear here."
				/>
			{:else if visibleItems.length === 0}
				<XuvaEmptyState
					title="No watchlist matches"
					message="Try changing search terms to find a saved title."
				/>
			{:else}
				<BrowseGrid>
					{#each visibleItems as item (item.id)}
						<PosterCard
							title={item.title}
							meta={item.subtitle || item.meta}
							imageUrl={item.posterUrl || item.backdropUrl}
							href={itemDetailHref(item)}
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
		gap: var(--xuva-space-2);
	}
</style>

