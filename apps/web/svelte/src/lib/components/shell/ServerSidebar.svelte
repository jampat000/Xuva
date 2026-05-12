<script lang="ts">
	import SidebarItem from './SidebarItem.svelte';
	import SidebarUser from './SidebarUser.svelte';
	import LorivoBrand from './LorivoBrand.svelte';
	import LorivoSidebar from './LorivoSidebar.svelte';

	let {
		active = 'dashboard',
		userDisplayName = 'Local User',
		userRole = 'Local Account'
	} = $props<{
		active?: 'dashboard' | 'library' | 'scanning' | 'metadata' | 'playback' | 'access' | 'about';
		userDisplayName?: string;
		userRole?: string;
	}>();
</script>

<LorivoSidebar>
	{#snippet brand()}
		<div class="settings-brand">
			<LorivoBrand />
			<span>Settings</span>
		</div>
	{/snippet}

	{#snippet primary()}
		<SidebarItem label="Dashboard" href="/settings#dashboard" active={active === 'dashboard'}>
			{#snippet icon()}
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path d="M5 13.5h5.5V6H5zM13.5 18H19V6h-5.5zM5 18h5.5v-1.8H5z" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" />
				</svg>
			{/snippet}
		</SidebarItem>
		<SidebarItem label="Library" href="/settings#library" active={active === 'library'}>
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
		<SidebarItem label="Access" href="/settings#access" active={active === 'access'}>
			{#snippet icon()}
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<circle cx="12" cy="8.4" r="3.1" fill="none" stroke="currentColor" stroke-width="1.5" />
					<path d="M6.2 18.2c1.4-2.7 3.5-4 5.8-4s4.4 1.3 5.8 4" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
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
	{/snippet}

	{#snippet secondary()}
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
	{/snippet}

	{#snippet profile()}
		<SidebarUser name={userDisplayName} subtitle={userRole} />
	{/snippet}
</LorivoSidebar>

<style>
	.settings-brand {
		display: flex;
		align-items: center;
		gap: 12px;
		min-width: 0;
		min-height: 48px;
	}

	.settings-brand span {
		padding: 5px 9px;
		border: 1px solid rgb(154 167 255 / 16%);
		border-radius: 999px;
		background: rgb(154 167 255 / 7%);
		color: color-mix(in srgb, #9aa7ff 76%, white 24%);
		font-size: 0.7rem;
		font-weight: 760;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}
</style>
