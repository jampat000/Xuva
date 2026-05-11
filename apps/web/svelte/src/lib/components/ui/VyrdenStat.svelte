<script lang="ts">
	type Tone = 'neutral' | 'good' | 'warn' | 'danger';

	let {
		label,
		value,
		meta = '',
		tone = 'neutral'
	} = $props<{
		label: string;
		value: string;
		meta?: string;
		tone?: Tone;
	}>();
</script>

<article class="v-stat" data-tone={tone}>
	<span>{label}</span>
	<strong>{value}</strong>
	{#if meta}
		<p>{meta}</p>
	{/if}
</article>

<style>
	.v-stat {
		display: grid;
		gap: var(--vyrden-space-1);
		padding: var(--vyrden-space-3);
		border: 1px solid var(--vyrden-color-border-soft);
		border-radius: var(--vyrden-radius-md);
		background: linear-gradient(180deg, rgb(255 246 229 / 4%), rgb(255 246 229 / 2%));
	}

	span {
		color: var(--vyrden-color-text-soft);
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.01em;
	}

	strong {
		font-size: 1.08rem;
		line-height: 1.1;
	}

	p {
		margin: 0;
		font-size: 0.78rem;
		color: var(--vyrden-color-text-muted);
	}

	.v-stat[data-tone='good'] strong {
		color: var(--vyrden-color-good);
	}

	.v-stat[data-tone='warn'] strong {
		color: var(--vyrden-color-warn);
	}

	.v-stat[data-tone='danger'] strong {
		color: var(--vyrden-color-danger);
	}
</style>
