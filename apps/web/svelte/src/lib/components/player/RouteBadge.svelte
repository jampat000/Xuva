<script lang="ts">
  import type { PlaybackDecisionResponse } from '$lib/api/details';

  interface Props {
    decision: PlaybackDecisionResponse | undefined;
    class?: string;
  }

  let { decision, class: className = '' }: Props = $props();

  interface BadgeConfig {
    label: string;
    colorClass: string;
    dotClass: string;
    title: string;
  }

  const badge = $derived((): BadgeConfig => {
    const mode = (decision?.mode ?? '').toLowerCase();
    const va = (decision?.videoAction ?? '').toLowerCase();
    const ca = (decision?.containerAction ?? '').toLowerCase();

    if (mode === 'adaptive' || mode === 'hls') {
      return {
        label: 'Adaptive',
        colorClass: 'bg-violet-500/20 text-violet-300 border-violet-500/30',
        dotClass: 'bg-violet-400',
        title: `Adaptive HLS — ${decision?.reasonText ?? 'multi-bitrate ladder'}`
      };
    }
    if (va === 'transcode' || va === 'encode') {
      return {
        label: 'Transcoding',
        colorClass: 'bg-rose-500/20 text-rose-300 border-rose-500/30',
        dotClass: 'bg-rose-400',
        title: `Video transcode — ${decision?.reasonText ?? decision?.reasonCode ?? ''}`
      };
    }
    if (ca === 'remux' || ca === 'mux') {
      return {
        label: 'Remux',
        colorClass: 'bg-sky-500/20 text-sky-300 border-sky-500/30',
        dotClass: 'bg-sky-400',
        title: `Remux — ${decision?.reasonText ?? 'container repackage, no re-encode'}`
      };
    }
    if (mode === 'direct' || mode === 'direct_play' || mode === 'directplay') {
      return {
        label: 'Direct Play',
        colorClass: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30',
        dotClass: 'bg-emerald-400',
        title: 'Direct Play — zero server CPU, bit-perfect'
      };
    }
    const aa = (decision?.audioAction ?? '').toLowerCase();
    if (aa === 'transcode' || aa === 'encode') {
      return {
        label: 'Audio Tx',
        colorClass: 'bg-amber-500/20 text-amber-300 border-amber-500/30',
        dotClass: 'bg-amber-400',
        title: `Audio transcode — ${decision?.reasonText ?? decision?.reasonCode ?? ''}`
      };
    }
    // Unknown / loading
    return {
      label: '—',
      colorClass: 'bg-white/10 text-white/40 border-white/10',
      dotClass: 'bg-white/30',
      title: 'Determining playback route…'
    };
  });
</script>

<div
  class={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.12em] backdrop-blur-sm ${badge().colorClass} ${className}`}
  title={badge().title}
  role="status"
  aria-label={badge().label}
>
  <span class={`h-1.5 w-1.5 rounded-full ${badge().dotClass}`} aria-hidden="true"></span>
  {badge().label}
</div>
