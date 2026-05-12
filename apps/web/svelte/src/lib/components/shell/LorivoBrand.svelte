<script lang="ts">
	let { wordmark = 'Lorivo' } = $props<{ compact?: boolean; wordmark?: string }>();
</script>

<div class="v-brand">
	<svg class="v-brand__mark" width="22" height="22" viewBox="0 0 22 22" aria-hidden="true">
		<path d="M3 2.5h3.2v13.3h9.3V19H3V2.5z" fill="#7C5CFF" />
		<path d="M13.4 6.2 19.5 10l-6.1 3.8V6.2z" fill="#fff" />
	</svg>
	<strong class="v-brand__wordmark">{wordmark}</strong>
</div>

<style>
	.v-brand {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 10px;
		width: auto;
		min-height: 44px;
		padding: 0;
	}

	.v-brand__mark {
		flex: 0 0 auto;
		filter: drop-shadow(0 8px 18px rgb(124 92 255 / 20%));
	}

	.v-brand__wordmark {
		font-family: var(--lorivo-font-sans);
		font-size: 1.16rem;
		font-weight: 800;
		letter-spacing: 0.26em;
		line-height: 1;
		text-transform: uppercase;
		color: #ffffff;
		text-shadow: 0 1px 0 rgb(0 0 0 / 20%);
		font-variation-settings: 'wght' 680;
		text-rendering: geometricPrecision;
		white-space: nowrap;
	}
</style>
