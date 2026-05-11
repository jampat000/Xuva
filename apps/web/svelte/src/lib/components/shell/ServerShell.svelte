<script lang="ts">
	import type { Snippet } from 'svelte';
	import AppTopbar from './AppTopbar.svelte';
	import ServerSidebar from './ServerSidebar.svelte';
	import LorivoShell from './LorivoShell.svelte';

	let {
		active = 'library',
		searchValue = $bindable(''),
		userDisplayName = 'Local User',
		userRole = 'Local Account',
		userInitials = 'U',
		children
	} = $props<{
		active?: 'library' | 'scanning' | 'metadata' | 'playback' | 'server' | 'about';
		searchValue?: string;
		userDisplayName?: string;
		userRole?: string;
		userInitials?: string;
		children?: Snippet;
	}>();
</script>

<div data-shell="server">
	<LorivoShell density="default">
		{#snippet sidebar()}
			<ServerSidebar {active} {userDisplayName} {userRole} />
		{/snippet}

		{#snippet topbar()}
			<AppTopbar bind:searchValue {userInitials} onProfileClick={() => (window.location.href = '/settings')} />
		{/snippet}

		{@render children?.()}
	</LorivoShell>
</div>
