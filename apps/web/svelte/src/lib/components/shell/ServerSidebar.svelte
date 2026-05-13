<script lang="ts">
	import SidebarUser from './SidebarUser.svelte';
	import LorivoSidebar from './LorivoSidebar.svelte';
	import SettingsBrand from './SettingsBrand.svelte';
	import SettingsNav from './SettingsNav.svelte';

	let {
		active = 'dashboard',
		userDisplayName = 'Local User',
		userRole = 'Local Account',
		showStorage = true
	} = $props<{
		active?:
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
			| 'owner-access'
			| 'about';
		userDisplayName?: string;
		userRole?: string;
		showStorage?: boolean;
	}>();
</script>

<LorivoSidebar>
	{#snippet brand()}
		<SettingsBrand />
	{/snippet}

	{#snippet primary()}
		<SettingsNav {active} {showStorage} />
	{/snippet}

	{#snippet secondary()}
		<SettingsNav section="secondary" />
	{/snippet}

	{#snippet profile()}
		<SidebarUser name={userDisplayName} subtitle={userRole} />
	{/snippet}
</LorivoSidebar>
