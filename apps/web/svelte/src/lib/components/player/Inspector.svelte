<script lang="ts">
  import { X, Cpu, Wifi, Film, Layers, Clock, Hash, AlertTriangle } from 'lucide-svelte';
  import type { PlaybackDecisionResponse, MediaSourceItem } from '$lib/api/details';

  interface Props {
    open: boolean;
    decision: PlaybackDecisionResponse | undefined;
    mediaSource: MediaSourceItem | undefined;
    positionSeconds: number;
    durationSeconds: number;
    onClose: () => void;
  }

  let { open, decision, mediaSource, positionSeconds, durationSeconds, onClose }: Props = $props();

  function fmt(seconds: number): string {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = Math.floor(seconds % 60);
    if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
    return `${m}:${String(s).padStart(2, '0')}`;
  }

  function fmtBitrate(bps: number | undefined): string {
    if (!bps) return '—';
    if (bps >= 1_000_000) return `${(bps / 1_000_000).toFixed(1)} Mbps`;
    return `${Math.round(bps / 1000)} kbps`;
  }

  function fmtSize(bytes: number | undefined): string {
    if (!bytes) return '—';
    if (bytes >= 1_073_741_824) return `${(bytes / 1_073_741_824).toFixed(2)} GB`;
    if (bytes >= 1_048_576) return `${(bytes / 1_048_576).toFixed(1)} MB`;
    return `${Math.round(bytes / 1024)} KB`;
  }

  function fmtResolution(w: number | undefined, h: number | undefined): string {
    if (!w || !h) return '—';
    return `${w}×${h}`;
  }

  function modeColor(mode: string | undefined): string {
    const m = (mode ?? '').toLowerCase();
    if (m === 'adaptive' || m === 'hls') return 'text-violet-400';
    if (m.includes('transcode') || m.includes('encode')) return 'text-rose-400';
    if (m.includes('remux')) return 'text-sky-400';
    if (m.includes('direct')) return 'text-emerald-400';
    if (m.includes('audio')) return 'text-amber-400';
    return 'text-white/60';
  }

  function handleKey(e: KeyboardEvent) {
    if (!open) return;
    if (e.key === 'Escape' || e.key === 'i' || e.key === 'I') {
      e.stopPropagation();
      onClose();
    }
  }
</script>

<svelte:window onkeydown={handleKey} />

<!-- Slide-in panel from right -->
<div
  class={`fixed right-0 top-0 z-50 flex h-full w-80 flex-col overflow-hidden border-l border-white/10 bg-black/95 backdrop-blur-xl transition-transform duration-300 ${open ? 'translate-x-0' : 'translate-x-full'}`}
  role="complementary"
  aria-label="Playback inspector"
  aria-hidden={!open}
>
  <!-- Header -->
  <div class="flex shrink-0 items-center justify-between border-b border-white/10 px-5 py-4">
    <div>
      <div class="text-xs font-semibold uppercase tracking-[0.2em] text-white/40">Inspector</div>
      {#if decision?.mode}
        <div class={`mt-0.5 text-sm font-semibold ${modeColor(decision.mode)}`}>
          {decision.mode.replace(/_/g, ' ')}
        </div>
      {/if}
    </div>
    <button
      onclick={onClose}
      class="flex h-8 w-8 items-center justify-center rounded-full text-white/40 transition-colors hover:bg-white/10 hover:text-white"
      aria-label="Close inspector"
    >
      <X class="h-4 w-4" />
    </button>
  </div>

  <!-- Scrollable content -->
  <div class="flex-1 overflow-y-auto px-5 py-4 space-y-5 text-sm">

    <!-- Playback position -->
    <section>
      <div class="mb-2 flex items-center gap-2 text-[10px] font-semibold uppercase tracking-[0.2em] text-white/30">
        <Clock class="h-3 w-3" />
        Position
      </div>
      <div class="grid grid-cols-2 gap-2">
        <div class="rounded-lg bg-white/[0.04] px-3 py-2">
          <div class="text-[10px] text-white/40">Current</div>
          <div class="font-mono text-white">{fmt(positionSeconds)}</div>
        </div>
        <div class="rounded-lg bg-white/[0.04] px-3 py-2">
          <div class="text-[10px] text-white/40">Duration</div>
          <div class="font-mono text-white">{fmt(durationSeconds)}</div>
        </div>
      </div>
    </section>

    <!-- Decision -->
    {#if decision}
      <section>
        <div class="mb-2 flex items-center gap-2 text-[10px] font-semibold uppercase tracking-[0.2em] text-white/30">
          <Layers class="h-3 w-3" />
          Decision
        </div>
        <div class="space-y-1.5">
          {#if decision.containerAction}
            <div class="flex items-center justify-between rounded-lg bg-white/[0.04] px-3 py-2">
              <span class="text-white/50">Container</span>
              <span class="font-mono text-xs text-white">{decision.containerAction}</span>
            </div>
          {/if}
          {#if decision.videoAction}
            <div class="flex items-center justify-between rounded-lg bg-white/[0.04] px-3 py-2">
              <span class="text-white/50">Video</span>
              <span class={`font-mono text-xs ${decision.videoAction.toLowerCase().includes('transcode') ? 'text-rose-400' : 'text-emerald-400'}`}>
                {decision.videoAction}
              </span>
            </div>
          {/if}
          {#if decision.audioAction}
            <div class="flex items-center justify-between rounded-lg bg-white/[0.04] px-3 py-2">
              <span class="text-white/50">Audio</span>
              <span class={`font-mono text-xs ${decision.audioAction.toLowerCase().includes('transcode') ? 'text-amber-400' : 'text-emerald-400'}`}>
                {decision.audioAction}
              </span>
            </div>
          {/if}
          {#if decision.subtitleAction}
            <div class="flex items-center justify-between rounded-lg bg-white/[0.04] px-3 py-2">
              <span class="text-white/50">Subtitles</span>
              <span class="font-mono text-xs text-white">{decision.subtitleAction}</span>
            </div>
          {/if}
        </div>
      </section>

      <!-- Server cost estimate -->
      {#if decision.estimatedCpuCost || decision.estimatedNetworkBitrate}
        <section>
          <div class="mb-2 flex items-center gap-2 text-[10px] font-semibold uppercase tracking-[0.2em] text-white/30">
            <Cpu class="h-3 w-3" />
            Server Impact
          </div>
          <div class="grid grid-cols-2 gap-2">
            {#if decision.estimatedCpuCost}
              <div class="rounded-lg bg-white/[0.04] px-3 py-2">
                <div class="text-[10px] text-white/40">CPU cost</div>
                <div class="text-white">{decision.estimatedCpuCost}</div>
              </div>
            {/if}
            {#if decision.estimatedNetworkBitrate}
              <div class="rounded-lg bg-white/[0.04] px-3 py-2">
                <div class="text-[10px] text-white/40">Network</div>
                <div class="text-white">{fmtBitrate(decision.estimatedNetworkBitrate)}</div>
              </div>
            {/if}
          </div>
        </section>
      {/if}

      <!-- Reason -->
      {#if decision.reasonText || decision.reasonCode}
        <section>
          <div class="mb-2 flex items-center gap-2 text-[10px] font-semibold uppercase tracking-[0.2em] text-white/30">
            <Hash class="h-3 w-3" />
            Reason
          </div>
          <div class="rounded-lg bg-white/[0.04] px-3 py-2 space-y-1">
            {#if decision.reasonCode}
              <div class="font-mono text-[11px] text-white/50">{decision.reasonCode}</div>
            {/if}
            {#if decision.reasonText}
              <div class="text-xs text-white/70 leading-relaxed">{decision.reasonText}</div>
            {/if}
          </div>
        </section>
      {/if}

      <!-- Suggested fixes -->
      {#if decision.suggestedFixes && decision.suggestedFixes.length > 0}
        <section>
          <div class="mb-2 flex items-center gap-2 text-[10px] font-semibold uppercase tracking-[0.2em] text-amber-500/70">
            <AlertTriangle class="h-3 w-3" />
            Suggestions
          </div>
          <ul class="space-y-1.5">
            {#each decision.suggestedFixes as fix (fix)}
              <li class="rounded-lg bg-amber-500/10 px-3 py-2 text-xs text-amber-300/90 leading-relaxed">
                {fix}
              </li>
            {/each}
          </ul>
        </section>
      {/if}

      <!-- Trace ID -->
      {#if decision.decisionTraceId}
        <section>
          <div class="rounded-lg bg-white/[0.04] px-3 py-2">
            <div class="text-[10px] text-white/30">Trace ID</div>
            <div class="font-mono text-[11px] text-white/50 break-all">{decision.decisionTraceId}</div>
          </div>
        </section>
      {/if}
    {/if}

    <!-- File info -->
    {#if mediaSource}
      <section>
        <div class="mb-2 flex items-center gap-2 text-[10px] font-semibold uppercase tracking-[0.2em] text-white/30">
          <Film class="h-3 w-3" />
          File
        </div>
        <div class="space-y-1.5">
          {#if mediaSource.container}
            <div class="flex items-center justify-between rounded-lg bg-white/[0.04] px-3 py-2">
              <span class="text-white/50">Container</span>
              <span class="font-mono text-xs text-white uppercase">{mediaSource.container}</span>
            </div>
          {/if}
          {#if mediaSource.videoCodec}
            <div class="flex items-center justify-between rounded-lg bg-white/[0.04] px-3 py-2">
              <span class="text-white/50">Video codec</span>
              <span class="font-mono text-xs text-white uppercase">{mediaSource.videoCodec}</span>
            </div>
          {/if}
          {#if mediaSource.width && mediaSource.height}
            <div class="flex items-center justify-between rounded-lg bg-white/[0.04] px-3 py-2">
              <span class="text-white/50">Resolution</span>
              <span class="font-mono text-xs text-white">{fmtResolution(mediaSource.width, mediaSource.height)}</span>
            </div>
          {/if}
          {#if mediaSource.bitrate}
            <div class="flex items-center justify-between rounded-lg bg-white/[0.04] px-3 py-2">
              <span class="text-white/50">Bitrate</span>
              <span class="font-mono text-xs text-white">{fmtBitrate(mediaSource.bitrate)}</span>
            </div>
          {/if}
          {#if mediaSource.sizeBytes}
            <div class="flex items-center justify-between rounded-lg bg-white/[0.04] px-3 py-2">
              <span class="text-white/50">File size</span>
              <span class="font-mono text-xs text-white">{fmtSize(mediaSource.sizeBytes)}</span>
            </div>
          {/if}
          {#if mediaSource.audioStreams !== undefined}
            <div class="flex items-center justify-between rounded-lg bg-white/[0.04] px-3 py-2">
              <span class="text-white/50">Audio tracks</span>
              <span class="font-mono text-xs text-white">{mediaSource.audioStreams}</span>
            </div>
          {/if}
          {#if mediaSource.subtitleStreams !== undefined}
            <div class="flex items-center justify-between rounded-lg bg-white/[0.04] px-3 py-2">
              <span class="text-white/50">Subtitle tracks</span>
              <span class="font-mono text-xs text-white">{mediaSource.subtitleStreams}</span>
            </div>
          {/if}
        </div>
      </section>
    {/if}

    <!-- Network -->
    <section>
      <div class="mb-2 flex items-center gap-2 text-[10px] font-semibold uppercase tracking-[0.2em] text-white/30">
        <Wifi class="h-3 w-3" />
        Network
      </div>
      <div class="rounded-lg bg-white/[0.04] px-3 py-2">
        <div class="text-[10px] text-white/40">Protocol</div>
        <div class="font-mono text-xs text-white">{decision?.selected?.protocol ?? 'http'}</div>
      </div>
    </section>

  </div>

  <!-- Footer hint -->
  <div class="shrink-0 border-t border-white/10 px-5 py-3">
    <div class="text-[10px] text-white/20 text-center">Press <kbd class="font-mono">I</kbd> to toggle</div>
  </div>
</div>
