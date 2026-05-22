<script lang="ts">
  import type { SessionItem } from '$lib/api/operator';

  let { session } = $props<{ session: SessionItem }>();

  const modeLabel = $derived(
    session.mode === 'direct'     ? 'Direct Play' :
    session.mode === 'transcode'  ? 'Transcoding'  :
    session.mode === 'adaptive'   ? 'Adaptive'     :
    session.mode ?? session.route ?? 'Streaming'
  );

  const modeBadgeClass = $derived(
    session.mode === 'direct'    ? 'bg-emerald-400/10 text-emerald-300' :
    session.mode === 'transcode' ? 'bg-amber-400/10 text-amber-300'     :
    'bg-primary-glow/10 text-primary-glow'
  );

  const title = $derived(session.title ?? session.sourceName ?? 'Unknown');
  const device = $derived(session.deviceId ?? 'Unknown device');
</script>

<div class="hairline flex items-center gap-4 rounded-xl bg-surface/50 px-4 py-3">
  <!-- Play indicator -->
  <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary-glow/10">
    <svg class="h-4 w-4 fill-primary-glow text-primary-glow" viewBox="0 0 24 24" aria-hidden="true">
      <path d="M8 5v14l11-7z"/>
    </svg>
  </div>

  <!-- Info -->
  <div class="min-w-0 flex-1">
    <div class="truncate text-sm font-medium">{title}</div>
    <div class="mt-0.5 flex items-center gap-2 text-[11px] text-muted-foreground">
      <span class="truncate">{device}</span>
      {#if session.state}
        <span class="shrink-0 capitalize text-muted-foreground/60">{session.state}</span>
      {/if}
    </div>
  </div>

  <!-- Mode badge -->
  <span class="shrink-0 rounded-full px-2.5 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] {modeBadgeClass}">
    {modeLabel}
  </span>
</div>
