<script lang="ts">
  import type { SessionItem } from '$lib/api/operator';
  import { formatBitrate, formatCodec } from '$lib/utils/mediaFormat';

  let { session } = $props<{ session: SessionItem }>();

  const isTranscoding = $derived(session.mode === 'transcode' || session.mode === 'adaptive');

  const modeLabel = $derived(
    session.mode === 'direct'    ? 'Direct Play' :
    session.mode === 'transcode' ? 'Transcoding'  :
    session.mode === 'adaptive'  ? 'HLS Stream'   :
    session.mode ?? session.route ?? 'Streaming'
  );

  const modeBadgeClass = $derived(
    session.mode === 'direct'    ? 'bg-emerald-400/10 text-emerald-300 ring-emerald-400/20' :
    session.mode === 'transcode' ? 'bg-amber-400/10 text-amber-300 ring-amber-400/20'       :
    'bg-primary-glow/10 text-primary-glow ring-primary-glow/20'
  );

  const title  = $derived(session.title ?? session.sourceName ?? 'Unknown');
  const device = $derived(
    [session.deviceId, session.clientProfile].filter(Boolean).join(' · ') || 'Unknown device'
  );

  const progress    = $derived(session.progressSeconds ?? 0);
  const duration    = $derived(session.durationSeconds ?? 0);
  const progressPct = $derived(duration > 0 ? Math.min(100, (progress / duration) * 100) : 0);

  function fmtTime(s: number): string {
    const h = Math.floor(s / 3600);
    const m = Math.floor((s % 3600) / 60);
    if (h > 0) return `${h}h ${m}m`;
    return `${m}m`;
  }

  const techPills = $derived(
    [
      session.videoCodec ? formatCodec(session.videoCodec) : null,
      session.container  ? session.container.toUpperCase() : null,
      session.qualityLabel ?? null,
      session.bitrate    ? formatBitrate(session.bitrate)  : null,
    ].filter((x): x is string => x !== null)
  );

  const impactBadge = $derived(
    session.serverImpact === 'high'   ? { label: 'High CPU',    cls: 'text-red-400 ring-red-400/20 bg-red-400/10' }   :
    session.serverImpact === 'medium' ? { label: 'Medium CPU',  cls: 'text-amber-400 ring-amber-400/20 bg-amber-400/10' } :
    null
  );
</script>

<div class="hairline overflow-hidden rounded-xl bg-surface/50">
  <div class="flex items-start gap-3 px-4 py-3">
    <!-- Icon -->
    <div class="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg {isTranscoding ? 'bg-amber-400/10' : 'bg-primary-glow/10'}">
      {#if isTranscoding}
        <svg class="h-4 w-4 animate-spin text-amber-400" fill="none" viewBox="0 0 24 24" aria-hidden="true">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
        </svg>
      {:else}
        <svg class="h-4 w-4 fill-primary-glow" viewBox="0 0 24 24" aria-hidden="true">
          <path d="M8 5v14l11-7z"/>
        </svg>
      {/if}
    </div>

    <!-- Main content -->
    <div class="min-w-0 flex-1 space-y-1">
      <!-- Title row -->
      <div class="flex items-start justify-between gap-3">
        <div class="truncate text-sm font-medium leading-snug">{title}</div>
        <span class="shrink-0 rounded-full px-2.5 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] ring-1 {modeBadgeClass}">
          {modeLabel}
        </span>
      </div>

      <!-- Device / state row -->
      <div class="flex items-center gap-1.5 text-[11px] text-muted-foreground">
        <span class="truncate">{device}</span>
        {#if session.state}
          <span class="shrink-0 capitalize text-muted-foreground/40">· {session.state}</span>
        {/if}
      </div>

      <!-- Tech pills -->
      {#if techPills.length > 0 || impactBadge}
        <div class="flex flex-wrap items-center gap-1 pt-0.5">
          {#each techPills as pill}
            <span class="rounded bg-white/5 px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">{pill}</span>
          {/each}
          {#if impactBadge}
            <span class="rounded px-1.5 py-0.5 font-mono text-[10px] ring-1 {impactBadge.cls}">{impactBadge.label}</span>
          {/if}
        </div>
      {/if}

      <!-- Transcode reason -->
      {#if isTranscoding && session.reasonText}
        <div class="text-[11px] text-amber-300/60 leading-snug">{session.reasonText}</div>
      {/if}
    </div>
  </div>

  <!-- Progress bar -->
  {#if duration > 0}
    <div class="px-4 pb-3">
      <div class="relative h-1 overflow-hidden rounded-full bg-white/8">
        <div
          class="absolute inset-y-0 left-0 rounded-full transition-[width] duration-1000 {isTranscoding ? 'bg-amber-400' : 'bg-primary-glow'}"
          style="width: {progressPct}%"
        ></div>
      </div>
      <div class="mt-1 flex justify-between text-[10px] text-muted-foreground/40">
        <span>{fmtTime(progress)}</span>
        <span>{fmtTime(duration)}</span>
      </div>
    </div>
  {/if}
</div>
