<script lang="ts">
	import { LorivoButton, LorivoPanel } from '$lib/components';
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

	function asText(value: unknown): string {
		return String(value ?? '').trim();
	}
</script>

{#if browser}
	<LorivoPanel title={title} subtitle={panelSubtitle}>
		<div class="browser-actions">
			{#if browser.parent}
				<LorivoButton variant="ghost" onclick={() => onBrowse(browser.parent || '')}>Up one folder</LorivoButton>
			{/if}
			{#if browser.path}
				<LorivoButton variant="secondary" onclick={() => onUsePath(browser.path || '')}>Use this folder</LorivoButton>
			{/if}
		</div>

		{#if browser.error}
			<p class="error">{browser.error}</p>
		{:else if (browser.entries || []).length === 0}
			<p class="muted">{emptyMessage}</p>
		{:else}
			<div class="folder-list">
				{#each browser.entries || [] as entry}
					<button type="button" class="folder-entry" onclick={() => onBrowse(asText(entry.path))}>
						<span>{asText(entry.name) || 'Folder'}</span>
						<small>{asText(entry.path)}</small>
					</button>
				{/each}
			</div>
		{/if}
	</LorivoPanel>
{/if}

<style>
	.browser-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.error {
		margin: 0;
		color: var(--lorivo-color-danger, #ff9f9f);
		font-size: 0.85rem;
	}

	.muted {
		margin: 0;
		color: var(--lorivo-color-text-muted);
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
		border: 0;
		border-top: 1px solid var(--lorivo-color-border-soft);
		background: transparent;
		color: var(--lorivo-color-text);
		padding: 10px 0;
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
		color: var(--lorivo-color-text-muted);
		font-size: 0.75rem;
	}
</style>
