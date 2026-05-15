<script lang="ts">
	import SidebarItem from './SidebarItem.svelte';
	import SidebarSection from './SidebarSection.svelte';
	import SidebarUser from './SidebarUser.svelte';
	import XuvaBrand from './XuvaBrand.svelte';
	import XuvaSidebar from './XuvaSidebar.svelte';

	type ActiveRoute =
		| 'home'
		| 'movies'
		| 'tv'
		| 'continue-watching'
		| 'recently-added'
		| 'settings';

	interface LibraryItem {
		id?: string;
		name?: string;
		kind?: string;
	}

	let {
		active = 'home',
		libraryItems = [],
		userDisplayName = 'Local User',
		userRole = 'Local Account'
	} = $props<{
		active?: ActiveRoute;
		libraryItems?: LibraryItem[];
		userDisplayName?: string;
		userRole?: string;
	}>();

	function asText(value: unknown): string {
		return String(value ?? '').trim();
	}

	function libraryKindLabel(kind: string): string {
		const normalizedKind = asText(kind).toLowerCase();
		if (normalizedKind === 'movies' || normalizedKind === 'movie') return 'Movies Library';
		if (normalizedKind === 'tv' || normalizedKind === 'series') return 'TV Library';
		return 'Library';
	}

	function libraryHref(kind: string): string {
		const normalizedKind = asText(kind).toLowerCase();
		if (normalizedKind === 'movies' || normalizedKind === 'movie') return '/movies';
		if (normalizedKind === 'tv' || normalizedKind === 'series') return '/tv';
		return '/';
	}

	const hasLibraries = $derived.by(() => libraryItems.length > 0);
</script>

<XuvaSidebar>
	{#snippet brand()}
		<XuvaBrand />
	{/snippet}

	{#snippet primary()}
		<SidebarItem label="Home" href="/" active={active === 'home'}>
			{#snippet icon()}
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path
						d="M4 11.2 12 4l8 7.2v8.2a.9.9 0 0 1-.9.9H4.9a.9.9 0 0 1-.9-.9z"
						fill="currentColor"
					/>
				</svg>
			{/snippet}
		</SidebarItem>
		<SidebarItem label="Movies" href="/movies" active={active === 'movies'}>
			{#snippet icon()}
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path
						d="M5.3 4h13.4A1.3 1.3 0 0 1 20 5.3v13.4a1.3 1.3 0 0 1-1.3 1.3H5.3A1.3 1.3 0 0 1 4 18.7V5.3A1.3 1.3 0 0 1 5.3 4Z"
						fill="none"
						stroke="currentColor"
						stroke-width="1.6"
					/>
					<path d="M9 4v16m6-16v16" fill="none" stroke="currentColor" stroke-width="1.4" />
				</svg>
			{/snippet}
		</SidebarItem>
		<SidebarItem label="TV" href="/tv" active={active === 'tv'}>
			{#snippet icon()}
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path
						d="M5 7.5A2.5 2.5 0 0 1 7.5 5h9A2.5 2.5 0 0 1 19 7.5v6A2.5 2.5 0 0 1 16.5 16h-3l-3.5 3V16h-2.5A2.5 2.5 0 0 1 5 13.5z"
						fill="none"
						stroke="currentColor"
						stroke-width="1.5"
					/>
				</svg>
			{/snippet}
		</SidebarItem>
		<SidebarItem
			label="Continue Watching"
			href="/continue-watching"
			active={active === 'continue-watching'}
		>
			{#snippet icon()}
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<circle cx="12" cy="12" r="8.2" fill="none" stroke="currentColor" stroke-width="1.6" />
					<path
						d="M12 7.4v5l3.5 2.1"
						fill="none"
						stroke="currentColor"
						stroke-width="1.6"
						stroke-linecap="round"
					/>
				</svg>
			{/snippet}
		</SidebarItem>
		<SidebarItem
			label="Recently Added"
			href="/recently-added"
			active={active === 'recently-added'}
		>
			{#snippet icon()}
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path
						d="M12 4.2v15.6M4.2 12h15.6"
						fill="none"
						stroke="currentColor"
						stroke-width="1.5"
						stroke-linecap="round"
					/>
					<circle cx="12" cy="12" r="8.6" fill="none" stroke="currentColor" stroke-width="1.5" />
				</svg>
			{/snippet}
		</SidebarItem>
	{/snippet}

	{#snippet libraries()}
		<SidebarSection title="Libraries" actionLabel="+" actionAriaLabel="Add Library">
			{#if hasLibraries}
				{#each libraryItems.slice(0, 4) as library (library.id)}
					<SidebarItem
						label={asText(library.name) || libraryKindLabel(asText(library.kind))}
						href={libraryHref(asText(library.kind))}
					>
						{#snippet icon()}
							<span
								class="library-glyph"
								data-kind={asText(library.kind).toLowerCase() === 'movies' ? 'movies' : 'tv'}
							>
								{#if asText(library.kind).toLowerCase() === 'movies'}
									<svg viewBox="0 0 24 24" aria-hidden="true">
										<rect
											x="4.6"
											y="4.6"
											width="14.8"
											height="14.8"
											rx="2"
											fill="none"
											stroke="currentColor"
											stroke-width="1.5"
										/>
										<path
											d="M9.2 4.7v14.6m5.6-14.6v14.6"
											fill="none"
											stroke="currentColor"
											stroke-width="1.3"
										/>
									</svg>
								{:else}
									<svg viewBox="0 0 24 24" aria-hidden="true">
										<path
											d="M5 7.5A2.5 2.5 0 0 1 7.5 5h9A2.5 2.5 0 0 1 19 7.5v6A2.5 2.5 0 0 1 16.5 16h-3l-3.5 3V16h-2.5A2.5 2.5 0 0 1 5 13.5z"
											fill="none"
											stroke="currentColor"
											stroke-width="1.5"
										/>
									</svg>
								{/if}
							</span>
						{/snippet}
					</SidebarItem>
				{/each}
			{:else}
				<SidebarItem label="Add Library" href="/settings#libraries">
					{#snippet icon()}
						<span class="library-glyph library-glyph--empty">
							<svg viewBox="0 0 24 24" aria-hidden="true">
								<path
									d="M12 5v14M5 12h14"
									fill="none"
									stroke="currentColor"
									stroke-width="1.5"
									stroke-linecap="round"
								/>
							</svg>
						</span>
					{/snippet}
				</SidebarItem>
			{/if}
		</SidebarSection>
	{/snippet}

	{#snippet secondary()}
		<SidebarItem label="Settings" href="/settings" active={active === 'settings'}>
			{#snippet icon()}
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path
						d="M9.4 4.8h5.2l.7 2.1 2 .8 1.9-1.1 2.6 4.5-1.8 1.4.1 2.2 1.7 1.4-2.6 4.5-1.9-1.1-2 .8-.7 2.1H9.4l-.7-2.1-2-.8-1.9 1.1L2.2 16l1.7-1.4.1-2.2-1.8-1.4 2.6-4.5 1.9 1.1 2-.8z"
						fill="none"
						stroke="currentColor"
						stroke-width="1.3"
						stroke-linejoin="round"
					/>
					<circle cx="12" cy="12" r="2.6" fill="none" stroke="currentColor" stroke-width="1.3" />
				</svg>
			{/snippet}
		</SidebarItem>
	{/snippet}

	{#snippet profile()}
		<SidebarUser name={userDisplayName} subtitle={userRole} />
	{/snippet}
</XuvaSidebar>

<style>
	.library-glyph {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 21px;
		height: 21px;
		color: color-mix(in srgb, var(--xuva-color-text) 78%, transparent);
	}

	.library-glyph[data-kind='movies'] {
		color: var(--xuva-color-accent-teal);
	}

	.library-glyph[data-kind='tv'] {
		color: color-mix(in srgb, var(--xuva-color-text) 82%, transparent);
	}

	.library-glyph :global(svg) {
		width: 21px;
		height: 21px;
	}

	.library-glyph--empty {
		color: color-mix(in srgb, var(--xuva-color-text-muted) 72%, transparent);
	}
</style>
