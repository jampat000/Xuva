<script lang="ts">
	let {
		name,
		subtitle = '',
		avatarUrl = '',
		href = ''
	} = $props<{
		name: string;
		subtitle?: string;
		avatarUrl?: string;
		href?: string;
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

{#if href}
	<a class="sidebar-user" href={href} aria-label="Current profile">
		{#if avatarUrl}
			<img src={avatarUrl} alt="" loading="lazy" />
		{:else}
			<span class="sidebar-user__fallback" aria-hidden="true">{initials || 'V'}</span>
		{/if}
		<span class="sidebar-user__copy">
			<strong>{name}</strong>
			{#if subtitle}
				<span>{subtitle}</span>
			{/if}
		</span>
	</a>
{:else}
	<div class="sidebar-user" aria-label="Current profile">
		{#if avatarUrl}
			<img src={avatarUrl} alt="" loading="lazy" />
		{:else}
			<span class="sidebar-user__fallback" aria-hidden="true">{initials || 'V'}</span>
		{/if}
		<span class="sidebar-user__copy">
			<strong>{name}</strong>
			{#if subtitle}
				<span>{subtitle}</span>
			{/if}
		</span>
	</div>
{/if}

<style>
	.sidebar-user {
		display: grid;
		grid-template-columns: 38px minmax(0, 1fr);
		align-items: center;
		column-gap: 12px;
		width: 100%;
		margin-top: 8px;
		padding: 10px 13px 8px;
		color: color-mix(in srgb, var(--xuva-color-text) 92%, transparent);
		border-top: 1px solid rgb(255 255 255 / 9%);
		border-radius: 0;
		text-decoration: none;
		transition: background-color 150ms ease;
	}

	.sidebar-user:hover {
		background: rgb(255 246 229 / 3%);
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
		border-radius: 999px;
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
		color: var(--xuva-color-text-soft);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

</style>
