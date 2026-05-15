<script lang="ts">
	import SidebarItem from './SidebarItem.svelte';

	type SettingsSection =
		| 'dashboard'
		| 'general'
		| 'libraries'
		| 'scanning'
		| 'metadata'
		| 'playback'
		| 'transcoding'
		| 'storage'
		| 'network'
		| 'pairing'
		| 'approved-devices'
		| 'discovery'
		| 'admin-access'
		| 'planned-tools'
		| 'about';

	let {
		active = 'dashboard',
		section = 'primary',
		showStorage = true,
		serverGroupLabel = 'Xuva'
	} = $props<{
		active?: SettingsSection;
		section?: 'primary' | 'secondary';
		showStorage?: boolean;
		serverGroupLabel?: string;
	}>();
</script>

{#if section === 'primary'}
	<p class="settings-nav-group settings-nav-group--server">{serverGroupLabel || 'Xuva'}</p>
	<SidebarItem label="Dashboard" href="/settings#dashboard" active={active === 'dashboard'}>
		{#snippet icon()}
			<svg viewBox="0 0 24 24" aria-hidden="true">
				<path d="M5 13.5h5.5V6H5zM13.5 18H19V6h-5.5zM5 18h5.5v-1.8H5z" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" />
			</svg>
		{/snippet}
	</SidebarItem>
	<SidebarItem label="General" href="/settings#general" active={active === 'general'}>
		{#snippet icon()}
			<svg viewBox="0 0 24 24" aria-hidden="true">
				<circle cx="12" cy="8.6" r="3.1" fill="none" stroke="currentColor" stroke-width="1.5" />
				<path d="M6.2 18.2c1.4-2.7 3.5-4 5.8-4s4.4 1.3 5.8 4" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
			</svg>
		{/snippet}
	</SidebarItem>
	<SidebarItem label="Libraries" href="/settings#libraries" active={active === 'libraries'}>
		{#snippet icon()}
			<svg viewBox="0 0 24 24" aria-hidden="true">
				<path
					d="M4.8 6.2h5.4l1.4 1.8h7.6v9.8a1.4 1.4 0 0 1-1.4 1.4h-13a1.4 1.4 0 0 1-1.4-1.4V7.6a1.4 1.4 0 0 1 1.4-1.4Z"
					fill="none"
					stroke="currentColor"
					stroke-width="1.5"
				/>
			</svg>
		{/snippet}
	</SidebarItem>
	<SidebarItem label="Scanning" href="/settings#scanning" active={active === 'scanning'}>
		{#snippet icon()}
			<svg viewBox="0 0 24 24" aria-hidden="true">
				<path
					d="M5 12a7 7 0 0 1 12-4.9M19 12a7 7 0 0 1-12 4.9"
					fill="none"
					stroke="currentColor"
					stroke-width="1.6"
					stroke-linecap="round"
				/>
				<path d="M17 4.8v3.5h-3.5M7 19.2v-3.5h3.5" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" />
			</svg>
		{/snippet}
	</SidebarItem>
	<SidebarItem label="Metadata" href="/settings#metadata" active={active === 'metadata'}>
		{#snippet icon()}
			<svg viewBox="0 0 24 24" aria-hidden="true">
				<path d="M5 5.5h14v13H5z" fill="none" stroke="currentColor" stroke-width="1.5" />
				<path d="M8 9h8M8 12h8M8 15h5" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
			</svg>
		{/snippet}
	</SidebarItem>
	<SidebarItem label="Playback" href="/settings#playback" active={active === 'playback'}>
		{#snippet icon()}
			<svg viewBox="0 0 24 24" aria-hidden="true">
				<circle cx="12" cy="12" r="8" fill="none" stroke="currentColor" stroke-width="1.5" />
				<path d="m10 8.8 5.2 3.2-5.2 3.2z" fill="currentColor" />
			</svg>
		{/snippet}
	</SidebarItem>
	{#if showStorage}
		<SidebarItem label="Transcoding" href="/settings#transcoding" active={active === 'transcoding'}>
			{#snippet icon()}
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path d="M4.8 8h14.4M4.8 12h10.2M4.8 16h6.2" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
					<path d="m14.5 14.5 3.8-2.5-3.8-2.5" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
				</svg>
			{/snippet}
		</SidebarItem>
		<SidebarItem label="Storage" href="/settings#storage" active={active === 'storage'}>
			{#snippet icon()}
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path d="M5 7.5h14v9H5z" fill="none" stroke="currentColor" stroke-width="1.5" />
					<path d="M8 11h8M8 14h5" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
				</svg>
			{/snippet}
		</SidebarItem>
		<SidebarItem label="Network" href="/settings#network" active={active === 'network'}>
			{#snippet icon()}
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path d="M4.8 12a10 10 0 0 1 14.4 0M7.8 15a6.2 6.2 0 0 1 8.4 0M11.1 18a2.1 2.1 0 0 1 1.8 0" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
				</svg>
			{/snippet}
		</SidebarItem>
		<SidebarItem label="Discovery" href="/settings#discovery" active={active === 'discovery'}>
			{#snippet icon()}
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path d="M12 7v10M8 11h8M6 7h12v10H6z" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" />
				</svg>
			{/snippet}
		</SidebarItem>
	{/if}
	<SidebarItem label="Users" href="/settings#admin-access" active={active === 'admin-access'}>
		{#snippet icon()}
			<svg viewBox="0 0 24 24" aria-hidden="true">
				<circle cx="12" cy="8.4" r="3.1" fill="none" stroke="currentColor" stroke-width="1.5" />
				<path d="M6.2 18.2c1.4-2.7 3.5-4 5.8-4s4.4 1.3 5.8 4" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
			</svg>
		{/snippet}
	</SidebarItem>

	<p class="settings-nav-group">Devices</p>
	<SidebarItem label="Pairing" href="/settings#pairing" active={active === 'pairing'}>
		{#snippet icon()}
			<svg viewBox="0 0 24 24" aria-hidden="true">
				<path d="M7.2 8.4a3.2 3.2 0 0 1 4.5 0l.3.3.3-.3a3.2 3.2 0 0 1 4.5 4.5l-4.8 4.8-4.8-4.8a3.2 3.2 0 0 1 0-4.5Z" fill="none" stroke="currentColor" stroke-width="1.5" />
			</svg>
		{/snippet}
	</SidebarItem>
	<SidebarItem label="Approved Devices" href="/settings#approved-devices" active={active === 'approved-devices'}>
		{#snippet icon()}
			<svg viewBox="0 0 24 24" aria-hidden="true">
				<rect x="6.2" y="5.6" width="11.6" height="12.8" rx="1.6" fill="none" stroke="currentColor" stroke-width="1.5" />
				<path d="m9.2 11.8 2.1 2.1 3.5-3.6" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" />
			</svg>
		{/snippet}
	</SidebarItem>
	<p class="settings-nav-group">Advanced</p>
	<SidebarItem label="Planned (Placeholder)" href="/settings#planned-tools" active={active === 'planned-tools'}>
		{#snippet icon()}
			<svg viewBox="0 0 24 24" aria-hidden="true">
				<path d="M6 6.5h12v11H6z" fill="none" stroke="currentColor" stroke-width="1.5" />
				<path d="M9 10h6M9 13h6" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
			</svg>
		{/snippet}
	</SidebarItem>
	<SidebarItem label="About" href="/settings#about" active={active === 'about'}>
		{#snippet icon()}
			<svg viewBox="0 0 24 24" aria-hidden="true">
				<circle cx="12" cy="12" r="8.2" fill="none" stroke="currentColor" stroke-width="1.5" />
				<path d="M12 10.8v5M12 8h.1" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
			</svg>
		{/snippet}
	</SidebarItem>
{:else}
	<SidebarItem label="Back to Media" href="/" variant="back">
		{#snippet icon()}
			<svg viewBox="0 0 24 24" aria-hidden="true">
				<path
					d="M14.5 5 8 12l6.5 7"
					fill="none"
					stroke="currentColor"
					stroke-width="1.7"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
			</svg>
		{/snippet}
	</SidebarItem>
{/if}

<style>
	.settings-nav-group {
		margin: 6px 10px 2px;
		color: color-mix(in srgb, var(--xuva-color-text-soft) 84%, transparent);
		font-size: 0.69rem;
		font-weight: 750;
		letter-spacing: 0.09em;
		text-transform: uppercase;
	}

	.settings-nav-group--server {
		letter-spacing: 0.04em;
		text-transform: none;
	}

	.settings-nav-group:first-child {
		margin-top: 6px;
	}
</style>
