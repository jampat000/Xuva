<script lang="ts">
  import type { PlaybackDecisionResponse, ProbeTrack } from '$lib/api/details';
  import type { MediaSourceItem } from '$lib/api/details';
  import { formatCodec, formatChannels, formatResolution } from '$lib/utils/mediaFormat';

  interface Props {
    decisions: Record<string, PlaybackDecisionResponse | undefined>;
    deviceProfiles: readonly string[];
    deviceLabels: Record<string, string>;
    mediaSource?: MediaSourceItem | null;
    audioTracks?: ProbeTrack[];
    subtitleTracks?: ProbeTrack[];
    encoderLabel?: string | null;
    /** compact = true trims padding for episode rows */
    compact?: boolean;
  }

  let {
    decisions,
    deviceProfiles,
    deviceLabels,
    mediaSource,
    audioTracks = [],
    subtitleTracks = [],
    encoderLabel,
    compact = false,
  }: Props = $props();

  // ── Plain-language action helpers ─────────────────────────────────────────

  type Cell = { label: string; detail: string; tone: 'green' | 'amber' | 'red' | 'muted' };

  function videoCell(action: string | undefined, videoCodec?: string, width?: number, height?: number): Cell {
    const codec = videoCodec ? formatCodec(videoCodec) : '';
    const res = (width && height) ? formatResolution(width, height) : '';
    const spec = [codec, res].filter(Boolean).join(' · ');
    switch (action) {
      case 'direct':      return { label: 'Direct play', detail: spec, tone: 'green' };
      case 'copy':        return { label: 'Stream copy',  detail: spec, tone: 'green' };
      case 'transcode':   return { label: 'Transcoding',  detail: spec, tone: 'red' };
      case 'adaptive':    return { label: 'Adaptive',     detail: spec, tone: 'amber' };
      case 'pending':
      case 'probe_required': return { label: 'Pending probe', detail: '', tone: 'muted' };
      default:            return { label: action ?? '—',  detail: spec, tone: 'muted' };
    }
  }

  function audioCell(action: string | undefined, track?: ProbeTrack): Cell {
    const codec = track?.codec ? formatCodec(track.codec) : '';
    const ch = track?.channels ? formatChannels(track.channels) : '';
    const lang = (track?.language && track.language !== 'und') ? track.language.toUpperCase() : '';
    const spec = [codec, ch, lang].filter(Boolean).join(' · ');
    switch (action) {
      case 'direct':            return { label: 'Direct',      detail: spec, tone: 'green' };
      case 'transcode':
      case 'copy_or_transcode': return { label: 'Converting',  detail: spec, tone: 'amber' };
      case 'pending':
      case 'probe_required':    return { label: 'Pending probe', detail: '', tone: 'muted' };
      default:                  return { label: action ?? '—', detail: spec, tone: 'muted' };
    }
  }

  function containerCell(action: string | undefined, container?: string): Cell {
    const fmt = container ? container.split(',')[0].toUpperCase() : '';
    switch (action) {
      case 'direct':
      case 'direct_or_remux':    return { label: 'Direct',      detail: fmt, tone: 'green' };
      case 'remux':              return { label: 'Remuxed',     detail: fmt, tone: 'amber' };
      case 'transcode_or_remux': return { label: 'Repackaged',  detail: fmt, tone: 'amber' };
      case 'transcode':          return { label: 'Converted',   detail: fmt, tone: 'red' };
      case 'pending':
      case 'probe_required':     return { label: 'Pending probe', detail: '', tone: 'muted' };
      default:                   return { label: action ?? '—', detail: fmt, tone: 'muted' };
    }
  }

  function subtitleCell(action: string | undefined): Cell | null {
    if (!action || action === 'none' || action === '') return null;
    switch (action) {
      case 'pass':
      case 'copy':    return { label: 'Passthrough', detail: '',               tone: 'green' };
      case 'burn':    return { label: 'Burned in',   detail: 'Video re-encoded', tone: 'red' };
      case 'convert': return { label: 'Converted',   detail: '',               tone: 'amber' };
      default:        return { label: action,         detail: '',               tone: 'muted' };
    }
  }

  function cpuLabel(cost?: string): string | null {
    switch (cost) {
      case 'light':  return 'Light server load';
      case 'medium': return 'Medium server load';
      case 'high':   return 'High server load';
      default:       return null;
    }
  }

  // ── Derived rows ─────────────────────────────────────────────────────────

  // Default audio track = first one marked default, fallback to index 0
  const defaultAudio = $derived(audioTracks.find(t => t.default) ?? audioTracks[0]);

  // Only show subtitle row if any device has a real subtitle action
  const showSubtitles = $derived(
    deviceProfiles.some(p => {
      const a = decisions[p]?.subtitleAction;
      return a && a !== 'none' && a !== '';
    })
  );

  // Tone → CSS classes
  function toneClass(tone: Cell['tone'], part: 'dot' | 'label' | 'detail') {
    const map: Record<Cell['tone'], Record<string, string>> = {
      green: { dot: 'bg-emerald-400', label: 'text-emerald-400', detail: 'text-emerald-400/55' },
      amber: { dot: 'bg-amber-400',   label: 'text-amber-300',   detail: 'text-amber-300/55' },
      red:   { dot: 'bg-red-400',     label: 'text-red-400',     detail: 'text-red-400/55' },
      muted: { dot: 'bg-foreground/25', label: 'text-muted-foreground', detail: 'text-muted-foreground/50' },
    };
    return map[tone][part] ?? '';
  }

  const pad = $derived(compact ? 'px-2.5 py-2' : 'px-3.5 py-3');
  const labelSize = $derived(compact ? 'text-[11px]' : 'text-[12px]');
  const detailSize = $derived(compact ? 'text-[9px]' : 'text-[10px]');
  const headerSize = $derived(compact ? 'text-[9px]' : 'text-[10px]');

  const hasNonDirect = $derived(deviceProfiles.some(p => {
    const d = decisions[p];
    return d?.videoAction === 'transcode' || d?.audioAction === 'transcode' ||
           d?.audioAction === 'copy_or_transcode' || d?.containerAction === 'remux' ||
           d?.containerAction === 'transcode' || d?.containerAction === 'transcode_or_remux' ||
           d?.subtitleAction === 'burn';
  }));
</script>

<div class="overflow-hidden rounded-xl border border-foreground/[0.10] bg-surface/20">

  <!-- ── Column headers ─────────────────────────────────────────────────── -->
  <div class="grid border-b border-foreground/[0.08] bg-surface/30"
    style="grid-template-columns: 5.5rem repeat({deviceProfiles.length}, 1fr);">
    <div class="{pad}"></div>
    {#each deviceProfiles as profile}
      <div class="{pad} {headerSize} font-semibold uppercase tracking-[0.18em] text-muted-foreground/60">
        {deviceLabels[profile] ?? profile}
      </div>
    {/each}
  </div>

  <!-- ── Video row ──────────────────────────────────────────────────────── -->
  <div class="grid border-b border-foreground/[0.05]"
    style="grid-template-columns: 5.5rem repeat({deviceProfiles.length}, 1fr);">
    <div class="{pad} {headerSize} font-medium uppercase tracking-[0.15em] text-muted-foreground/45 flex items-start pt-3">
      Video
    </div>
    {#each deviceProfiles as profile}
      {@const d = decisions[profile]}
      {@const cell = videoCell(d?.videoAction, mediaSource?.videoCodec, mediaSource?.width, mediaSource?.height)}
      <div class="{pad} border-l border-foreground/[0.04]">
        <div class="flex items-center gap-1.5">
          <span class="h-1.5 w-1.5 shrink-0 rounded-full {toneClass(cell.tone, 'dot')}"></span>
          <span class="{labelSize} font-medium {toneClass(cell.tone, 'label')}">{cell.label}</span>
        </div>
        {#if cell.detail}
          <div class="mt-0.5 pl-3 {detailSize} {toneClass(cell.tone, 'detail')} leading-snug">{cell.detail}</div>
        {/if}
        {#if (d?.mode === 'Video Transcode' || d?.mode === 'Adaptive Stream') && encoderLabel}
          <div class="mt-1 pl-3 {detailSize} text-muted-foreground/40">{encoderLabel}</div>
        {/if}
      </div>
    {/each}
  </div>

  <!-- ── Audio row ──────────────────────────────────────────────────────── -->
  <div class="grid border-b border-foreground/[0.05]"
    style="grid-template-columns: 5.5rem repeat({deviceProfiles.length}, 1fr);">
    <div class="{pad} {headerSize} font-medium uppercase tracking-[0.15em] text-muted-foreground/45 flex items-start pt-3">
      Audio
    </div>
    {#each deviceProfiles as profile}
      {@const d = decisions[profile]}
      {@const cell = audioCell(d?.audioAction, defaultAudio)}
      <div class="{pad} border-l border-foreground/[0.04]">
        <div class="flex items-center gap-1.5">
          <span class="h-1.5 w-1.5 shrink-0 rounded-full {toneClass(cell.tone, 'dot')}"></span>
          <span class="{labelSize} font-medium {toneClass(cell.tone, 'label')}">{cell.label}</span>
        </div>
        {#if cell.detail}
          <div class="mt-0.5 pl-3 {detailSize} {toneClass(cell.tone, 'detail')} leading-snug">{cell.detail}</div>
        {/if}
      </div>
    {/each}
  </div>

  <!-- ── Container row ──────────────────────────────────────────────────── -->
  <div class="grid {showSubtitles ? 'border-b border-foreground/[0.05]' : ''}"
    style="grid-template-columns: 5.5rem repeat({deviceProfiles.length}, 1fr);">
    <div class="{pad} {headerSize} font-medium uppercase tracking-[0.15em] text-muted-foreground/45 flex items-start pt-3">
      Container
    </div>
    {#each deviceProfiles as profile}
      {@const d = decisions[profile]}
      {@const cell = containerCell(d?.containerAction, mediaSource?.container)}
      <div class="{pad} border-l border-foreground/[0.04]">
        <div class="flex items-center gap-1.5">
          <span class="h-1.5 w-1.5 shrink-0 rounded-full {toneClass(cell.tone, 'dot')}"></span>
          <span class="{labelSize} font-medium {toneClass(cell.tone, 'label')}">{cell.label}</span>
        </div>
        {#if cell.detail}
          <div class="mt-0.5 pl-3 {detailSize} {toneClass(cell.tone, 'detail')} leading-snug">{cell.detail}</div>
        {/if}
      </div>
    {/each}
  </div>

  <!-- ── Subtitles row (only when relevant) ────────────────────────────── -->
  {#if showSubtitles}
    <div class="grid"
      style="grid-template-columns: 5.5rem repeat({deviceProfiles.length}, 1fr);">
      <div class="{pad} {headerSize} font-medium uppercase tracking-[0.15em] text-muted-foreground/45 flex items-start pt-3">
        Subtitles
      </div>
      {#each deviceProfiles as profile}
        {@const d = decisions[profile]}
        {@const cell = subtitleCell(d?.subtitleAction)}
        <div class="{pad} border-l border-foreground/[0.04]">
          {#if cell}
            <div class="flex items-center gap-1.5">
              <span class="h-1.5 w-1.5 shrink-0 rounded-full {toneClass(cell.tone, 'dot')}"></span>
              <span class="{labelSize} font-medium {toneClass(cell.tone, 'label')}">{cell.label}</span>
            </div>
            {#if cell.detail}
              <div class="mt-0.5 pl-3 {detailSize} {toneClass(cell.tone, 'detail')} leading-snug">{cell.detail}</div>
            {/if}
          {:else}
            <span class="{detailSize} text-muted-foreground/30">—</span>
          {/if}
        </div>
      {/each}
    </div>
  {/if}

  <!-- ── Impact / reason footer ─────────────────────────────────────────── -->
  {#if hasNonDirect}
    <div class="border-t border-foreground/[0.07] bg-surface/10 px-3.5 py-3 space-y-2">
      {#each deviceProfiles as profile}
        {@const d = decisions[profile]}
        {@const reason = d?.reasonText}
        {@const cost = cpuLabel(d?.estimatedCpuCost)}
        {@const isNonDirect = d?.videoAction === 'transcode' || d?.audioAction === 'transcode' ||
          d?.audioAction === 'copy_or_transcode' || d?.containerAction === 'remux' ||
          d?.containerAction === 'transcode' || d?.containerAction === 'transcode_or_remux' ||
          d?.subtitleAction === 'burn'}
        {#if isNonDirect && (reason || cost)}
          <div class="flex items-start gap-2">
            <span class="{detailSize} shrink-0 font-semibold uppercase tracking-[0.15em] text-muted-foreground/45 pt-px w-14 text-right">
              {deviceLabels[profile]?.split(' ')[0] ?? profile}
            </span>
            <div class="min-w-0">
              {#if reason}
                <p class="{detailSize} leading-relaxed text-muted-foreground/65">{reason}</p>
              {/if}
              {#if cost}
                <p class="{detailSize} mt-0.5 text-muted-foreground/45 italic">{cost}</p>
              {/if}
            </div>
          </div>
        {/if}
      {/each}
    </div>
  {/if}

</div>
