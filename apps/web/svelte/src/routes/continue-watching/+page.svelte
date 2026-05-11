<script lang="ts">
	import { onMount } from 'svelte';
	import { resolvePreviewMode } from '$lib/home/model';
	import {
		BrowseHeader,
		MediaShell,
		BrowsePage,
		BrowseStatChip,
		BrowseToolbar,
		ResumeTile,
		LorivoButton,
		LorivoEmptyState,
		LorivoPanel
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
	let usingPreviewData = $state(false);
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
		usingPreviewData = false;
		try {
			const url = new URL(window.location.href);
			const previewMode = resolvePreviewMode(url.searchParams);
			const context = await loadSecondaryRouteContext({ previewMode, limit: 60 });
			userDisplayName =
				context.user?.displayName || context.user?.username || userDisplayName;
			userRole = context.user?.role || 'Local Account';
			userInitials = initialsForName(userDisplayName);
			libraries = context.libraries || [];
			rowAvailable = Boolean(findRow(context.homePayload, 'continue'));
			usingPreviewData = context.model.usingPreviewData;
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
			<LorivoPanel
				title="Loading Continue Watching"
				subtitle="Fetching in-progress playback from existing APIs."
			/>
		{:else if loadError}
			<LorivoPanel title="Continue Watching could not load" subtitle={loadError}>
				<div class="status-actions">
					<LorivoButton variant="secondary" onclick={loadContinueWatching}>Retry</LorivoButton>
					<LorivoButton variant="ghost" href="/">Back to Home</LorivoButton>
				</div>
			</LorivoPanel>
		{:else}
			<BrowseHeader title="Continue Watching" subtitle="Resume in-progress movies and episodes.">
				{#snippet chips()}
					<BrowseStatChip label={`${visibleItems.length} visible`} />
					{#if !rowAvailable && !usingPreviewData}
						<BrowseStatChip label="Not available yet" />
					{:else if usingPreviewData}
						<BrowseStatChip label="Preview mode" />
					{/if}
				{/snippet}
			</BrowseHeader>

			<BrowseToolbar />

			{#if !rowAvailable && !usingPreviewData}
				<LorivoEmptyState
					title="Continue Watching is not available yet"
					message="The current backend payload does not expose a continue row for this route."
				>
					{#snippet action()}
						<LorivoButton variant="secondary" href="/">Back to Home</LorivoButton>
					{/snippet}
				</LorivoEmptyState>
			{:else if items.length === 0}
				<LorivoEmptyState
					title="Nothing in progress"
					message="Start playback and your in-progress titles will appear here."
				/>
			{:else if visibleItems.length === 0}
				<LorivoEmptyState
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
		gap: var(--lorivo-space-2);
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
