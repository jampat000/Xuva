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

<article class="l-stat" data-tone={tone}>
	<span>{label}</span>
	<strong>{value}</strong>
	{#if meta}
		<p>{meta}</p>
	{/if}
</article>

<style>
	.l-stat {
		display: grid;
		gap: var(--lorivo-space-1);
		padding: var(--lorivo-space-3);
		border: 1px solid var(--lorivo-color-border-soft);
		border-radius: var(--lorivo-radius-md);
		background: linear-gradient(180deg, rgb(255 246 229 / 4%), rgb(255 246 229 / 2%));
	}

	span {
		color: var(--lorivo-color-text-soft);
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
		color: var(--lorivo-color-text-muted);
	}

	.l-stat[data-tone='good'] strong {
		color: var(--lorivo-color-good);
	}

	.l-stat[data-tone='warn'] strong {
		color: var(--lorivo-color-warn);
	}

	.l-stat[data-tone='danger'] strong {
		color: var(--lorivo-color-danger);
	}
</style>
