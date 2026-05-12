<script lang="ts">
	import SidebarUser from './SidebarUser.svelte';
	import LorivoSidebar from './LorivoSidebar.svelte';
	import SettingsBrand from './SettingsBrand.svelte';
	import SettingsNav from './SettingsNav.svelte';

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
		<SettingsBrand />
	{/snippet}

	{#snippet primary()}
		<SettingsNav {active} />
	{/snippet}

	{#snippet secondary()}
		<SettingsNav section="secondary" />
	{/snippet}

	{#snippet profile()}
		<SidebarUser name={userDisplayName} subtitle={userRole} />
	{/snippet}
</LorivoSidebar>
