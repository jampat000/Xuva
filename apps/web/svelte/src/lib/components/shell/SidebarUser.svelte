<script lang="ts">
	let {
		name,
		subtitle = '',
		avatarUrl = '',
		ariaLabel = 'Open profile menu'
	} = $props<{
		name: string;
		subtitle?: string;
		avatarUrl?: string;
		ariaLabel?: string;
	}>();

	const initials = $derived.by(() =>
		name
			.split(' ')
			.filter(Boolean)
			.slice(0, 2)
			.map((part: string) => part[0]?.toUpperCase() || '')
			.join('')
	);
</script>

<button class="sidebar-user" type="button" aria-label={ariaLabel}>
	{#if avatarUrl}
		<img src={avatarUrl} alt="" loading="lazy" />
	{:else}
		<span class="sidebar-user__fallback" aria-hidden="true">{initials || 'L'}</span>
	{/if}
	<span class="sidebar-user__copy">
		<strong>{name}</strong>
		{#if subtitle}
			<span>{subtitle}</span>
		{/if}
	</span>
	<span class="sidebar-user__chevron" aria-hidden="true">
		<svg viewBox="0 0 24 24"><path d="m8 10 4 4 4-4" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" /></svg>
	</span>
</button>

<style>
	.sidebar-user {
		display: grid;
		grid-template-columns: 38px minmax(0, 1fr) 16px;
		align-items: center;
		column-gap: 12px;
		width: 100%;
		margin-top: 8px;
		padding: 12px 13px 8px;
		color: color-mix(in srgb, var(--lorivo-color-text) 92%, transparent);
		border-radius: 12px;
		transition: background-color 150ms ease;
	}

	.sidebar-user:hover {
		background: rgb(255 246 229 / 5%);
	}

	img,
	.sidebar-user__fallback {
		flex: 0 0 auto;
		width: 38px;
		height: 38px;
		border-radius: 50%;
		object-fit: cover;
	}

	.sidebar-user__fallback {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		background: linear-gradient(180deg, rgb(88 201 176 / 28%), rgb(33 39 37 / 44%));
		color: #f4f1ea;
		font-size: 0.8rem;
		font-weight: 800;
	}

	.sidebar-user__copy {
		flex: 1;
		display: grid;
		min-width: 0;
	}

	strong {
		font-size: 0.96rem;
		font-weight: 650;
		letter-spacing: 0.01em;
		line-height: 1.2;
		text-align: left;
	}

	.sidebar-user__copy span {
		font-size: 0.72rem;
		color: var(--lorivo-color-text-soft);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.sidebar-user__chevron {
		display: inline-flex;
		width: 16px;
		height: 16px;
		color: rgb(255 255 255 / 58%);
	}

	.sidebar-user__chevron svg {
		width: 100%;
		height: 100%;
	}
</style>
