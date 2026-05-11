<script lang="ts">
	import SidebarItem from './SidebarItem.svelte';
	import SidebarUser from './SidebarUser.svelte';
	import LorivoBrand from './LorivoBrand.svelte';
	import LorivoSidebar from './LorivoSidebar.svelte';

	let {
		active = 'admin',
		userDisplayName = 'Local User',
		userRole = 'Local Account'
	} = $props<{
		active?: 'admin' | 'settings';
		userDisplayName?: string;
		userRole?: string;
	}>();
</script>

<LorivoSidebar>
	{#snippet brand()}
		<LorivoBrand />
	{/snippet}

	{#snippet primary()}
		<SidebarItem label="Dashboard" href="/admin" active={active === 'admin'}>
			{#snippet icon()}
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path
						d="M4.5 4.5h6.8v6.8H4.5Zm8.2 0h6.8v4.2h-6.8Zm0 5.8h6.8v9h-6.8Zm-8.2 2.4h6.8v6.6H4.5Z"
						fill="none"
						stroke="currentColor"
						stroke-width="1.5"
					/>
				</svg>
			{/snippet}
		</SidebarItem>
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

	{#snippet secondary()}
		<SidebarItem label="Back to Media" href="/">
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
