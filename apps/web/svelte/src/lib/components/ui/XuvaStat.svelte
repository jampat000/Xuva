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
		gap: 0.2rem;
		padding: 0.8rem 0;
		border-top: 1px solid var(--xuva-color-border-soft);
		background: transparent;
	}

	.v-stat:first-child {
		border-top: 0;
		padding-top: 0;
	}

	span {
		color: var(--xuva-color-text-soft);
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
		color: var(--xuva-color-text-muted);
	}

	.v-stat[data-tone='good'] strong {
		color: var(--xuva-color-good);
	}

	.v-stat[data-tone='warn'] strong {
		color: var(--xuva-color-warn);
	}

	.v-stat[data-tone='danger'] strong {
		color: var(--xuva-color-danger);
	}
</style>
