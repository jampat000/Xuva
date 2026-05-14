<script lang="ts">
	let {
		initials = 'U',
		name = 'User',
		role = 'User',
		changePasswordHref = '/settings#admin-access',
		canSignOut = false,
		signInHref = '/signin',
		onSignOut = undefined
	} = $props<{
		initials?: string;
		name?: string;
		role?: string;
		changePasswordHref?: string;
		canSignOut?: boolean;
		signInHref?: string;
		onSignOut?: (() => Promise<void> | void) | undefined;
	}>();

	let open = $state(false);
	let isSigningOut = $state(false);
	let rootEl: HTMLDivElement | null = null;

	function toggleMenu(): void {
		open = !open;
	}

	function closeMenu(): void {
		open = false;
	}

	async function signOut(): Promise<void> {
		if (!canSignOut || !onSignOut || isSigningOut) return;
		isSigningOut = true;
		try {
			await onSignOut();
			open = false;
		} finally {
			isSigningOut = false;
		}
	}

	function roleWithBrackets(value: string): string {
		const label = String(value || '').trim();
		if (!label) return '(User)';
		if (label.startsWith('(') && label.endsWith(')')) return label;
		return `(${label})`;
	}
</script>

<svelte:window
	onclick={(event) => {
		if (!open) return;
		if (rootEl && event.target instanceof Node && rootEl.contains(event.target)) return;
		closeMenu();
	}}
	onkeydown={(event) => {
		if (event.key === 'Escape') closeMenu();
	}}
/>

<div class="profile-menu" bind:this={rootEl}>
	<button
		class="profile-menu__trigger"
		type="button"
		aria-label="Open profile menu"
		aria-expanded={open}
		onclick={toggleMenu}
	>
		<span class="profile-menu__avatar">{initials}</span>
	</button>

	{#if open}
		<div class="profile-menu__panel" role="menu" aria-label="Profile options">
			<div class="profile-menu__header">
				<strong>{name || 'User'}</strong>
				<span>{roleWithBrackets(role || 'User')}</span>
			</div>
			<div class="profile-menu__actions">
				{#if canSignOut}
					<a href={changePasswordHref} role="menuitem">Change password</a>
					<button type="button" role="menuitem" disabled={isSigningOut} onclick={signOut}>
						{isSigningOut ? 'Signing out...' : 'Sign out'}
					</button>
				{:else}
					<a href={signInHref} role="menuitem">Sign in</a>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	.profile-menu {
		position: relative;
		display: inline-flex;
		--profile-menu-trigger-border: rgb(255 255 255 / 12%);
		--profile-menu-trigger-bg: rgb(255 255 255 / 8%);
		--profile-menu-avatar-bg: linear-gradient(180deg, rgb(94 212 186 / 30%), rgb(45 52 50 / 46%));
		--profile-menu-avatar-color: #f6f5f2;
		--profile-menu-panel-bg: var(--xuva-color-bg-panel-elevated);
		--profile-menu-panel-border: var(--xuva-color-border-soft);
		--profile-menu-panel-text: var(--xuva-color-text);
		--profile-menu-row-hover: rgb(99 102 241 / 10%);
	}

	.profile-menu__trigger {
		display: grid;
		place-items: center;
		width: 40px;
		height: 40px;
		padding: 0;
		border: 1px solid var(--profile-menu-trigger-border);
		border-radius: 999px;
		background: var(--profile-menu-trigger-bg);
	}

	.profile-menu__avatar {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 30px;
		height: 30px;
		border-radius: 999px;
		background: var(--profile-menu-avatar-bg);
		color: var(--profile-menu-avatar-color);
		font-size: 0.72rem;
		font-weight: 700;
	}

	.profile-menu__panel {
		position: absolute;
		top: calc(100% + 10px);
		right: 0;
		min-width: 208px;
		border: 1px solid var(--profile-menu-panel-border);
		border-radius: 10px;
		background: var(--profile-menu-panel-bg);
		color: var(--profile-menu-panel-text);
		box-shadow: 0 14px 28px rgb(15 23 42 / 22%);
		z-index: 90;
		overflow: hidden;
	}

	.profile-menu__header {
		display: grid;
		gap: 2px;
		padding: 10px 12px;
		border-bottom: 1px solid var(--xuva-color-border-soft);
	}

	.profile-menu__header strong {
		font-size: 0.86rem;
		font-weight: 760;
		line-height: 1.2;
	}

	.profile-menu__header span {
		font-size: 0.74rem;
		color: var(--xuva-color-text-soft);
	}

	.profile-menu__actions {
		display: grid;
	}

	.profile-menu__actions a,
	.profile-menu__actions button {
		display: flex;
		align-items: center;
		width: 100%;
		min-height: 36px;
		padding: 0 12px;
		border: 0;
		background: transparent;
		color: inherit;
		font: inherit;
		font-size: 0.82rem;
		text-decoration: none;
		text-align: left;
		cursor: pointer;
	}

	.profile-menu__actions a:hover,
	.profile-menu__actions button:hover {
		background: var(--profile-menu-row-hover);
	}

	.profile-menu__actions button:disabled {
		opacity: 0.6;
		cursor: default;
	}
</style>
