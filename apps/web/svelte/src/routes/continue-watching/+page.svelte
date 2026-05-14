<script lang="ts">
	import { onMount } from 'svelte';
	import {
		BrowseHeader,
		MediaShell,
		BrowsePage,
		BrowseStatChip,
		BrowseToolbar,
		ResumeTile,
		XuvaButton,
		XuvaEmptyState,
		XuvaPanel
	} from '$lib/components';
	import {
		asText,
		findRow,
		formatLoadError,
		initialsForName,
		itemPlayHref,
		loadSecondaryRouteContext
	} from '$lib/secondary/media-routes';
	import type { HomeDisplayItem } from '$lib/home/model';

	let isLoading = $state(true);
	let loadError = $state('');
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
		void loadContinueWatching();
	});

	async function loadContinueWatching(): Promise<void> {
		isLoading = true;
		loadError = '';
		rowAvailable = false;
		try {
			const context = await loadSecondaryRouteContext({ limit: 60 });
			userDisplayName =
				context.user?.displayName || context.user?.username || userDisplayName;
			userRole = context.user?.role || 'Local Account';
			userInitials = initialsForName(userDisplayName);
			libraries = context.libraries || [];
			rowAvailable = Boolean(findRow(context.homePayload, 'continue'));
			items = context.model.continueItems;
		} catch (error) {
			loadError = formatLoadError(error, 'Continue Watching');
		} finally {
			isLoading = false;
		}
	}
</script>

<MediaShell active="continue-watching" bind:searchValue {userInitials}>
	<BrowsePage>
		{#if isLoading}
			<XuvaPanel
				title="Loading Continue Watching"
				subtitle="Fetching in-progress playback from existing APIs."
			/>
		{:else if loadError}
			<XuvaPanel title="Continue Watching could not load" subtitle={loadError}>
				<div class="status-actions">
					<XuvaButton variant="secondary" onclick={loadContinueWatching}>Retry</XuvaButton>
					<XuvaButton variant="ghost" href="/">Back to Home</XuvaButton>
				</div>
			</XuvaPanel>
		{:else}
			<BrowseHeader title="Continue Watching" subtitle="Resume in-progress movies and episodes.">
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
					title="Continue Watching is not available yet"
					message="The current backend payload does not expose a continue row for this route."
				>
					{#snippet action()}
						<XuvaButton variant="secondary" href="/">Back to Home</XuvaButton>
					{/snippet}
				</XuvaEmptyState>
			{:else if items.length === 0}
				<XuvaEmptyState
					title="Nothing in progress"
					message="Start playback and your in-progress titles will appear here."
				/>
			{:else if visibleItems.length === 0}
				<XuvaEmptyState
					title="No continue matches"
					message="Try changing search terms to find in-progress titles."
				/>
			{:else}
				<div class="resume-grid">
					{#each visibleItems as item (item.id)}
						<ResumeTile
							title={item.title}
							subtitle={item.subtitle}
							meta={item.meta}
							progress={item.progressPercent}
							imageUrl={item.backdropUrl || item.posterUrl}
							href={itemPlayHref(item)}
						/>
					{/each}
				</div>
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

	.resume-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(304px, 304px));
		gap: 15px;
		justify-content: start;
	}

	@media (max-width: 760px) {
		.resume-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
