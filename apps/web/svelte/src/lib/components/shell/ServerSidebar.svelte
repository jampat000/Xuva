<script lang="ts">
	import XuvaSidebar from './XuvaSidebar.svelte';
	import SettingsBrand from './SettingsBrand.svelte';
	import SettingsNav from './SettingsNav.svelte';

	let {
		active = 'dashboard',
		showStorage = true,
		serverGroupLabel = 'Xuva',
		showBrand = true
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
		showBrand?: boolean;
	}>();
</script>

<XuvaSidebar>
	{#if showBrand}
		{#snippet brand()}
			<SettingsBrand />
		{/snippet}
	{/if}

	{#snippet primary()}
		<SettingsNav {active} {showStorage} {serverGroupLabel} />
	{/snippet}
</XuvaSidebar>
