<script lang="ts">
	let { wordmark = 'Lorivo' } = $props<{ compact?: boolean; wordmark?: string }>();
</script>

<div class="v-brand">
	<strong class="v-brand__wordmark">{wordmark}</strong>
</div>

<style>
	.v-brand {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 100%;
		min-height: 44px;
		padding: 0;
	}

	.v-brand__wordmark {
		font-family: var(--vyrden-font-sans);
		font-size: 1.16rem;
		font-weight: 680;
		letter-spacing: 0.009em;
		line-height: 1;
		text-transform: none;
		color: #efe4d4;
		text-shadow: 0 1px 0 rgb(0 0 0 / 20%);
		font-variation-settings: 'wght' 680;
		text-rendering: geometricPrecision;
		white-space: nowrap;
	}
</style>
