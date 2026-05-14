<script lang="ts">
	let {
		value = $bindable(''),
		placeholder = 'Search movies, shows, people...',
		label = 'Search',
		shortcut = 'Ctrl K'
	} = $props<{
		value?: string;
		placeholder?: string;
		label?: string;
		shortcut?: string;
	}>();
</script>

<label class="v-search">
	<span class="sr-only">{label}</span>
	<span class="v-search__icon" aria-hidden="true">
		<svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="6.2" fill="none" stroke="currentColor" stroke-width="1.7" /><path d="m16 16 4.2 4.2" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" /></svg>
	</span>
	<input bind:value {placeholder} />
	<span class="v-search__shortcut" aria-hidden="true">{shortcut}</span>
</label>

<style>
	.v-search {
		display: flex;
		align-items: center;
		gap: 12px;
		width: min(100%, 468px);
		height: 38px;
		padding: 0 13px;
		border: 1px solid rgb(255 255 255 / 11%);
		border-radius: 999px;
		background:
			linear-gradient(180deg, rgb(255 255 255 / 5%), rgb(255 255 255 / 1%)),
			rgb(17 24 39 / 82%);
		box-shadow:
			inset 0 1px 0 rgb(255 255 255 / 6%),
			0 10px 24px rgb(0 0 0 / 18%);
		backdrop-filter: blur(14px);
	}

	.v-search:focus-within {
		border-color: var(--xuva-color-border-strong);
		box-shadow: var(--xuva-focus-ring);
	}

	.v-search__icon {
		display: inline-flex;
		width: 18px;
		height: 18px;
		color: color-mix(in srgb, var(--xuva-color-text) 72%, transparent);
	}

	.v-search__icon svg {
		width: 100%;
		height: 100%;
	}

	input {
		flex: 1;
		border: 0;
		background: transparent;
		color: color-mix(in srgb, var(--xuva-color-text) 96%, transparent);
		font-size: 0.95rem;
		letter-spacing: 0.008em;
	}

	input::placeholder {
		color: color-mix(in srgb, var(--xuva-color-text-muted) 74%, transparent);
	}

	.v-search__shortcut {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 49px;
		height: 24px;
		padding: 0 8px;
		border-radius: 8px;
		background: rgb(255 255 255 / 7%);
		color: color-mix(in srgb, var(--xuva-color-text-muted) 72%, transparent);
		font-size: 0.8rem;
		font-weight: 600;
		letter-spacing: 0.01em;
	}

	@media (max-width: 560px) {
		.v-search {
			height: 36px;
			gap: 9px;
			padding: 0 10px;
		}

		input {
			font-size: 0.88rem;
		}

		.v-search__shortcut {
			min-width: 44px;
			font-size: 0.74rem;
			padding: 0 6px;
		}
	}

	@media (max-width: 430px) {
		.v-search__shortcut {
			display: none;
		}
	}
</style>
