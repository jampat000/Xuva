<script lang="ts">
	type Status = 'healthy' | 'warning' | 'critical' | 'idle';

	let { status = 'idle', label } = $props<{ status?: Status; label?: string }>();

	const resolvedLabel = $derived(
		label ||
			(status === 'healthy'
				? 'Healthy'
				: status === 'warning'
					? 'Warning'
					: status === 'critical'
						? 'Critical'
						: 'Idle')
	);
</script>

<span class="live-status-badge" data-status={status}>{resolvedLabel}</span>

<style>
	.live-status-badge {
		display: inline-flex;
		align-items: center;
		min-height: var(--xuva-control-height-sm);
		padding: 0 var(--xuva-space-3);
		border-radius: var(--xuva-radius-pill);
		font-size: 0.76rem;
		font-weight: 700;
		letter-spacing: 0.02em;
		text-transform: uppercase;
	}

	.live-status-badge[data-status='healthy'] {
		background: rgb(123 200 146 / 18%);
		color: var(--xuva-color-good);
	}

	.live-status-badge[data-status='warning'] {
		background: rgb(226 191 115 / 18%);
		color: var(--xuva-color-warn);
	}

	.live-status-badge[data-status='critical'] {
		background: rgb(220 139 131 / 18%);
		color: var(--xuva-color-danger);
	}

	.live-status-badge[data-status='idle'] {
		background: rgb(255 246 229 / 8%);
		color: var(--xuva-color-text-muted);
	}
</style>
