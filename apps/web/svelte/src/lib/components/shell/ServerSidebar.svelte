<script lang="ts">
	import XuvaSidebar from './XuvaSidebar.svelte';
	import SettingsBrand from './SettingsBrand.svelte';
	import SettingsNav from './SettingsNav.svelte';

	let {
		active = 'dashboard',
		showStorage = true,
		serverGroupLabel = 'Xuva'
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
			| 'admin-access'
			| 'planned-tools'
			| 'about';
		showStorage?: boolean;
		serverGroupLabel?: string;
	}>();
</script>

<XuvaSidebar>
	{#snippet brand()}
		<SettingsBrand />
	{/snippet}

	{#snippet primary()}
		<SettingsNav {active} {showStorage} {serverGroupLabel} />
	{/snippet}

	{#snippet secondary()}
		<SettingsNav section="secondary" />
	{/snippet}
</XuvaSidebar>
