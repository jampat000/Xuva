<script lang="ts">
	import XuvaBrand from './XuvaBrand.svelte';
</script>

<div class="settings-brand">
	<XuvaBrand showBlurb={true} />
</div>

<style>
	.settings-brand {
		display: flex;
		align-items: center;
		justify-content: flex-start;
		gap: 0;
		width: 100%;
		min-width: 0;
		min-height: 28px;
	}
</style>
