<script lang="ts">
	import { onMount } from 'svelte';

	onMount(() => {
		window.location.replace('/settings#dashboard');
	});
</script>

<svelte:head>
	<title>Settings - Lorivo</title>
</svelte:head>

<div class="redirect-page">
	<div>
		<h1>Opening Settings</h1>
		<p>This area has moved into Settings.</p>
		<a href="/settings#dashboard">Open Settings</a>
	</div>
</div>

<style>
	.redirect-page {
		display: grid;
		min-height: 100dvh;
		place-items: center;
		padding: 24px;
		background: #0b1120;
		color: white;
		font-family: var(--lorivo-font-body, system-ui, sans-serif);
	}

	.redirect-page > div {
		display: grid;
		gap: 12px;
		max-width: 420px;
		padding: 24px;
		border: 1px solid rgb(255 255 255 / 10%);
		border-radius: 16px;
		background: rgb(255 255 255 / 4%);
	}

	h1,
	p {
		margin: 0;
	}

	p {
		color: rgb(255 255 255 / 64%);
	}

	a {
		color: white;
		font-weight: 700;
	}
</style>
