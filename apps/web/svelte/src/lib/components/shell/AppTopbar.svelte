<script lang="ts">
	import XuvaSearch from '../ui/XuvaSearch.svelte';

	let {
		searchValue = $bindable(''),
		userInitials = 'V',
		onProfileClick
	} = $props<{
		searchValue?: string;
		userInitials?: string;
		onProfileClick?: (() => void) | undefined;
	}>();
</script>

<div class="home-topbar">
	<div class="home-topbar__main">
		<XuvaSearch bind:value={searchValue} />
	</div>
	<div class="home-topbar__actions">
		<button class="profile-button" type="button" aria-label="Open profile settings" onclick={onProfileClick}>
			<span class="profile-button__avatar">{userInitials}</span>
			<span class="profile-button__presence" aria-hidden="true"></span>
		</button>
	</div>
</div>

<style>
	.home-topbar {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 284px;
		align-items: center;
		column-gap: 20px;
		width: 100%;
	}

	.home-topbar__main {
		width: 100%;
		display: flex;
		justify-content: center;
		min-width: 0;
	}

	.home-topbar__actions {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 12px;
		align-self: center;
	}

	.home-topbar__main :global(.v-search) {
		width: min(100%, 476px);
	}

	.profile-button {
		position: relative;
		display: grid;
		place-items: center;
		flex: 0 0 40px;
		width: 40px;
		height: 40px;
		padding: 0;
		aspect-ratio: 1 / 1;
		overflow: hidden;
		border: 1px solid rgb(255 255 255 / 9%);
		border-radius: 999px;
		background: rgb(28 29 27 / 74%);
		color: color-mix(in srgb, var(--xuva-color-text) 94%, transparent);
		font-weight: 700;
		line-height: 1;
		box-shadow:
			inset 0 1px 0 rgb(255 255 255 / 10%),
			0 18px 42px rgb(0 0 0 / 28%);
	}

	.profile-button__avatar {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 30px;
		height: 30px;
		border-radius: 999px;
		background: linear-gradient(180deg, rgb(88 201 176 / 26%), rgb(38 40 37 / 42%));
		color: #f4f1ea;
		font-size: 0.72rem;
	}

	.profile-button__presence {
		position: absolute;
		right: 4px;
		bottom: 4px;
		width: 8px;
		height: 8px;
		border: 2px solid #171815;
		border-radius: 50%;
		background: #58c9b0;
	}

	@media (max-width: 980px) {
		.home-topbar {
			grid-template-columns: minmax(0, 1fr) auto;
			column-gap: 12px;
		}

		.home-topbar__main {
			justify-content: stretch;
		}

		.home-topbar__main :global(.v-search) {
			width: 100%;
		}
	}

	@media (max-width: 560px) {
		.home-topbar {
			column-gap: 8px;
		}

		.profile-button {
			flex-basis: 36px;
			width: 36px;
			height: 36px;
		}

		.profile-button__avatar {
			width: 27px;
			height: 27px;
			font-size: 0.68rem;
		}

		.profile-button__presence {
			right: 3px;
			bottom: 3px;
			width: 7px;
			height: 7px;
		}
	}
</style>
