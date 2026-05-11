<script lang="ts">
	import { onMount } from 'svelte';
	import { resolvePreviewMode } from '$lib/home/model';
	import {
		BrowseGrid,
		BrowseHeader,
		MediaShell,
		BrowsePage,
		BrowseStatChip,
		BrowseToolbar,
		PosterCard,
		LorivoButton,
		LorivoEmptyState,
		LorivoPanel
	} from '$lib/components';
	import {
		asText,
		findRow,
		formatLoadError,
		initialsForName,
		itemDetailHref,
		loadSecondaryRouteContext
	} from '$lib/secondary/media-routes';

	let isLoading = $state(true);
	let loadError = $state('');
	let searchValue = $state('');
	let unavailable = $state(false);
	let userDisplayName = $state('Local User');
	let userRole = $state('Local Account');
	let userInitials = $state('V');
	let libraries = $state<Array<{ id?: string; name?: string; kind?: string }>>([]);
	let cards = $state<Array<{ id: string; title: string; meta: string; imageUrl: string; href?: string }>>([]);

	const visibleCards = $derived.by(() => {
		const needle = asText(searchValue).toLowerCase();
		if (!needle) return cards;
		return cards.filter((item) => `${item.title} ${item.meta}`.toLowerCase().includes(needle));
	});

	onMount(() => {
		void loadCollections();
	});

	async function loadCollections(): Promise<void> {
		isLoading = true;
		loadError = '';
		unavailable = false;
		try {
			const url = new URL(window.location.href);
			const previewMode = resolvePreviewMode(url.searchParams);
			const context = await loadSecondaryRouteContext({ previewMode, limit: 60 });
			userDisplayName =
				context.user?.displayName || context.user?.username || userDisplayName;
			userRole = context.user?.role || 'Local Account';
			userInitials = initialsForName(userDisplayName);
			libraries = context.libraries || [];

			const row = findRow(context.homePayload, 'collections');
			if (!row) {
				unavailable = true;
				cards = [];
				return;
			}

			const allowed = new Set(['movie', 'series']);
			const items = [
				...context.model.movieItems,
				...context.model.tvItems
			].filter((item) => allowed.has(asText(item.kind).toLowerCase()));
			const seen = new Set<string>();
			const nextCards: Array<{
				id: string;
				title: string;
				meta: string;
				imageUrl: string;
				href?: string;
			}> = [];
			for (const item of items) {
				const key = `${item.kind}:${item.id}`;
				if (seen.has(key)) continue;
				seen.add(key);
				nextCards.push({
					id: key,
					title: item.title,
					meta: item.subtitle || item.meta,
					imageUrl: item.posterUrl || item.backdropUrl,
					href: itemDetailHref(item)
				});
			}
			cards = nextCards;
		} catch (error) {
			loadError = formatLoadError(error, 'Collections');
		} finally {
			isLoading = false;
		}
	}
</script>

<MediaShell active="collections" bind:searchValue {userInitials}>
	<BrowsePage>
		{#if isLoading}
			<LorivoPanel
				title="Loading Collections"
				subtitle="Checking existing backend APIs for collection data."
			/>
		{:else if loadError}
			<LorivoPanel title="Collections could not load" subtitle={loadError}>
				<div class="status-actions">
					<LorivoButton variant="secondary" onclick={loadCollections}>Retry</LorivoButton>
					<LorivoButton variant="ghost" href="/">Back to Home</LorivoButton>
				</div>
			</LorivoPanel>
		{:else}
			<BrowseHeader title="Collections" subtitle="Curated groups from your library.">
				{#snippet chips()}
					<BrowseStatChip label={`${visibleCards.length} visible`} />
					{#if unavailable}
						<BrowseStatChip label="Not available yet" />
					{/if}
				{/snippet}
			</BrowseHeader>

			<BrowseToolbar />

			{#if unavailable}
				<LorivoEmptyState
					title="Collections are not available yet"
					message="The current backend APIs do not expose a dedicated collections feed for this route yet."
				>
					{#snippet action()}
						<LorivoButton variant="secondary" href="/">Back to Home</LorivoButton>
					{/snippet}
				</LorivoEmptyState>
			{:else if cards.length === 0}
				<LorivoEmptyState
					title="No collections found"
					message="Collections are enabled, but no collection items are currently available."
				/>
			{:else if visibleCards.length === 0}
				<LorivoEmptyState
					title="No collections match"
					message="Try changing search terms to find a collection."
				/>
			{:else}
				<BrowseGrid>
					{#each visibleCards as item (item.id)}
						<PosterCard
							title={item.title}
							meta={item.meta}
							imageUrl={item.imageUrl}
							href={item.href}
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
		gap: var(--lorivo-space-2);
	}
</style>
