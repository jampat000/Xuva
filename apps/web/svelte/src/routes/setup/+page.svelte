<script lang="ts">
	import { onMount } from 'svelte';
	import { XuvaButton, XuvaPanel } from '$lib/components';

	const setupWizardPath = '/signin';

	onMount(() => {
		window.location.replace(setupWizardPath);
	});
</script>

<div class="setup-redirect">
	<div class="setup-redirect__card">
		<XuvaPanel title="Opening Setup Wizard" subtitle="Redirecting to the Xuva setup flow.">
			<div class="setup-redirect__actions">
				<XuvaButton variant="primary" href={setupWizardPath}>Open setup wizard</XuvaButton>
				<XuvaButton variant="ghost" href="/">Back to Home</XuvaButton>
			</div>
		</XuvaPanel>
	</div>
</div>

<style>
	.setup-redirect {
		min-height: 100vh;
		display: grid;
		place-items: center;
		padding: clamp(18px, 4vw, 40px);
		background:
			radial-gradient(circle at 8% 10%, rgb(126 183 169 / 14%), transparent 38%),
			linear-gradient(180deg, #0f1723 0%, #0c131d 55%, #0a1119 100%);
	}

	.setup-redirect__card {
		width: min(520px, 100%);
	}

	.setup-redirect__actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
</style>
