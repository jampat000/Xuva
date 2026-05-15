<script lang="ts">
	import { onMount } from 'svelte';
	import { XuvaButton, XuvaPanel } from '$lib/components';

	const wizardPath = '/signin?setup=1';

	onMount(() => {
		window.location.replace(wizardPath);
	});
</script>

<div class="setup-wizard-redirect">
	<div class="setup-wizard-redirect__card">
		<XuvaPanel title="Opening Setup Wizard" subtitle="Redirecting to initial setup.">
			<div class="setup-wizard-redirect__actions">
				<XuvaButton variant="primary" href={wizardPath}>Open Setup Wizard</XuvaButton>
				<XuvaButton variant="ghost" href="/signin">Open Sign In</XuvaButton>
			</div>
		</XuvaPanel>
	</div>
</div>

<style>
	.setup-wizard-redirect {
		min-height: 100vh;
		display: grid;
		place-items: center;
		padding: clamp(18px, 4vw, 40px);
		background:
			radial-gradient(circle at 8% 10%, rgb(126 183 169 / 14%), transparent 38%),
			linear-gradient(180deg, #0f1723 0%, #0c131d 55%, #0a1119 100%);
	}

	.setup-wizard-redirect__card {
		width: min(520px, 100%);
	}

	.setup-wizard-redirect__actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
</style>
