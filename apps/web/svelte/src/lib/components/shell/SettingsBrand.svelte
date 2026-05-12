<script lang="ts">
	import LorivoBrand from './LorivoBrand.svelte';
</script>

<div class="settings-brand">
	<LorivoBrand />
	<span>Settings</span>
</div>

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
