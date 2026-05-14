<script lang="ts">
	import { XuvaButton, XuvaPanel } from '$lib/components';
	import type { FolderBrowseResponse } from '$lib/api/setup';

	let {
		browser,
		title = 'Folder browser',
		subtitle = '',
		emptyMessage = 'No child folders here. Use this folder path if it contains media.',
		onBrowse,
		onUsePath
	} = $props<{
		browser: FolderBrowseResponse | null;
		title?: string;
		subtitle?: string;
		emptyMessage?: string;
		onBrowse: (path: string) => void | Promise<void>;
		onUsePath: (path: string) => void;
	}>();

	const panelSubtitle = $derived.by(() => subtitle || browser?.path || 'Select a folder path.');
	const pathSegments = $derived.by(() => splitPathSegments(browser?.path));
	let manualPath = $state('');

	$effect(() => {
		manualPath = asText(browser?.path);
	});

	function asText(value: unknown): string {
		return String(value ?? '').trim();
	}

	function splitPathSegments(pathValue: unknown): string[] {
		const raw = asText(pathValue);
		if (!raw) return [];
		const normalized = raw.replace(/\\/g, '/');
		if (normalized.startsWith('//')) {
			const uncParts = normalized.split('/').filter(Boolean);
			if (uncParts.length === 0) return ['\\\\'];
			const root = `\\\\${uncParts[0]}${uncParts[1] ? `\\${uncParts[1]}` : ''}`;
			const rest = uncParts.slice(2);
			return [root, ...rest];
		}
		const parts = normalized.split('/').filter(Boolean);
		if (/^[A-Za-z]:$/.test(parts[0] || '')) return parts;
		return normalized.startsWith('/') ? ['/', ...parts] : parts;
	}

	function pathToSegment(index: number): string {
		if (!pathSegments.length) return '';
		const first = pathSegments[0];
		if (first.startsWith('\\\\')) {
			let built = first;
			for (let i = 1; i <= index; i += 1) built += `\\${pathSegments[i]}`;
			return built;
		}
		if (first === '/') {
			const tail = pathSegments.slice(1, index + 1).join('/');
			return tail ? `/${tail}` : '/';
		}
		if (/^[A-Za-z]:$/.test(first)) {
			const tail = pathSegments.slice(1, index + 1).join('\\');
			return tail ? `${first}\\${tail}` : first;
		}
		return pathSegments.slice(0, index + 1).join('/');
	}
</script>

{#if browser}
	<XuvaPanel title={title} subtitle={panelSubtitle}>
		<div class="network-notice">
			<strong>Network paths can be entered manually.</strong>
			<p>
				If mapped drives do not appear, enter a UNC path directly, for example
				<code>\\server\media</code> or <code>\\192.168.1.101\media</code>.
			</p>
			<div class="network-notice__quick">
				<button type="button" class="path-crumb" onclick={() => (manualPath = '\\\\server\\share')}>Use UNC template</button>
				<button type="button" class="path-crumb" onclick={() => (manualPath = '\\\\192.168.1.101\\share')}>Use IP template</button>
			</div>
		</div>

		<div class="manual-path">
			<label class="manual-path__field">
				<span>Enter folder path</span>
				<input bind:value={manualPath} placeholder="D:\Media\Movies or \\NAS\Media\Movies" />
			</label>
			<div class="manual-path__actions">
				<XuvaButton variant="secondary" onclick={() => onBrowse(asText(manualPath))}>
					Go to path
				</XuvaButton>
				{#if asText(browser.path)}
					<XuvaButton variant="ghost" onclick={() => onBrowse(asText(browser.path))}>
						Refresh
					</XuvaButton>
				{/if}
				<XuvaButton variant="primary" onclick={() => onUsePath(asText(manualPath))}>
					Use this path
				</XuvaButton>
			</div>
		</div>

		{#if asText(browser.path)}
			<div class="path-bar">
				<span class="path-bar__label">Path</span>
				<code class="path-bar__value">{asText(browser.path)}</code>
			</div>
		{/if}

		{#if pathSegments.length > 1}
			<div class="path-crumbs" aria-label="Folder path breadcrumbs">
				{#each pathSegments as segment, index (index)}
					<button type="button" class="path-crumb" onclick={() => onBrowse(pathToSegment(index))}>
						{segment}
					</button>
				{/each}
			</div>
		{/if}

		<div class="browser-actions">
			{#if browser.parent}
				<XuvaButton variant="ghost" onclick={() => onBrowse(browser.parent || '')}>Up one folder</XuvaButton>
			{/if}
			{#if browser.path}
				<XuvaButton variant="secondary" onclick={() => onUsePath(browser.path || '')}>Use this folder</XuvaButton>
			{/if}
		</div>

		{#if browser.error}
			<p class="error">{browser.error}</p>
		{:else if (browser.entries || []).length === 0}
			<p class="muted">{emptyMessage}</p>
		{:else}
			<div class="folder-list">
				{#each browser.entries || [] as entry}
					<div class="folder-entry">
						<button type="button" class="folder-entry__open" onclick={() => onBrowse(asText(entry.path))}>
							<span>{asText(entry.name) || 'Folder'}</span>
							<small>{asText(entry.path)}</small>
						</button>
						<XuvaButton
							variant="ghost"
							size="sm"
							onclick={() => onUsePath(asText(entry.path))}
						>
							Use
						</XuvaButton>
					</div>
				{/each}
			</div>
		{/if}
	</XuvaPanel>
{/if}

<style>
	.browser-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.manual-path {
		display: grid;
		gap: 8px;
		padding: 10px;
		border: 1px solid var(--xuva-color-border-soft);
		border-radius: 8px;
		background: color-mix(in srgb, var(--xuva-color-bg-panel-elevated) 96%, #e7edf7 4%);
	}

	.network-notice {
		display: grid;
		gap: 6px;
		padding: 10px;
		border: 1px solid var(--xuva-color-border-soft);
		border-radius: 8px;
		background: color-mix(in srgb, var(--xuva-color-bg-panel-elevated) 92%, #e2e8f3 8%);
	}

	.network-notice strong {
		font-size: 0.82rem;
		color: var(--xuva-color-text);
	}

	.network-notice p {
		margin: 0;
		color: var(--xuva-color-text-soft);
		font-size: 0.8rem;
		line-height: 1.4;
	}

	.network-notice code {
		font-size: 0.76rem;
		color: var(--xuva-color-text);
	}

	.network-notice__quick {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}

	.manual-path__field {
		display: grid;
		gap: 6px;
	}

	.manual-path__field span {
		color: var(--xuva-color-text-muted);
		font-size: 0.78rem;
		font-weight: 700;
	}

	.manual-path__field input {
		min-height: 40px;
		border: 1px solid var(--xuva-color-border-soft);
		border-radius: 6px;
		background: #fff;
		color: var(--xuva-color-text);
		font: inherit;
		padding: 0 10px;
		width: 100%;
	}

	.manual-path__actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.path-bar {
		display: grid;
		gap: 4px;
	}

	.path-bar__label {
		color: var(--xuva-color-text-soft);
		font-size: 0.76rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}

	.path-bar__value {
		display: block;
		padding: 8px 10px;
		border: 1px solid var(--xuva-color-border-soft);
		border-radius: 6px;
		background: color-mix(in srgb, var(--xuva-color-bg-panel-elevated) 94%, #e8eef7 6%);
		color: var(--xuva-color-text);
		font-size: 0.8rem;
		line-height: 1.35;
		overflow-wrap: anywhere;
	}

	.path-crumbs {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}

	.path-crumb {
		border: 1px solid var(--xuva-color-border-soft);
		border-radius: 6px;
		background: transparent;
		color: var(--xuva-color-text-muted);
		padding: 5px 9px;
		font-size: 0.76rem;
		line-height: 1.25;
	}

	.path-crumb:hover,
	.path-crumb:focus-visible {
		color: var(--xuva-color-text);
		border-color: color-mix(in srgb, var(--xuva-color-accent-purple) 36%, var(--xuva-color-border-soft));
		outline: none;
	}

	.error {
		margin: 0;
		color: var(--xuva-color-danger, #ff9f9f);
		font-size: 0.85rem;
	}

	.muted {
		margin: 0;
		color: var(--xuva-color-text-muted);
		font-size: 0.88rem;
	}

	.folder-list {
		display: grid;
		gap: 0;
		max-height: 260px;
		overflow: auto;
		padding-right: 2px;
	}

	.folder-entry {
		border-top: 1px solid var(--xuva-color-border-soft);
		background: transparent;
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		align-items: center;
		gap: 10px;
		padding: 10px 0;
	}

	.folder-entry__open {
		border: 0;
		background: transparent;
		color: var(--xuva-color-text);
		padding: 0;
		text-align: left;
		display: grid;
		gap: 3px;
		cursor: pointer;
	}

	.folder-list > :first-child {
		border-top: 0;
		padding-top: 0;
	}


	.folder-entry small {
		color: var(--xuva-color-text-muted);
		font-size: 0.75rem;
	}
</style>
