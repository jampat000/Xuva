<script lang="ts">
  /**
   * SubtitleSelector — fetches embedded + sidecar subtitle tracks for a given
   * mediaSourceId and lets the user pick one before playback.
   *
   * Emits the selected play URL via the `playHref` bindable so the parent can
   * update its Play button.
   */
  import { onMount } from 'svelte';
  import { Subtitles, ChevronDown, Check } from 'lucide-svelte';
  import { getMediaSourceTracks, getMediaSourceSubtitles } from '$lib/api/details';
  import type { ProbeTrack, SubtitleSidecar } from '$lib/api/details';

  interface Props {
    mediaSourceId: string;
    /** The base play URL — updated whenever the subtitle selection changes */
    playHref?: string;
    /** Override the base play path (e.g. to include title/back params) */
    basePlayUrl?: string;
  }

  let { mediaSourceId, playHref = $bindable(`/play/${mediaSourceId}`), basePlayUrl }: Props = $props();

  interface TrackOption {
    kind: 'embedded' | 'sidecar';
    index: number;
    label: string;
    language?: string;
    forced?: boolean;
    hi?: boolean;
  }

  let loading = $state(true);
  let open = $state(false);
  let tracks = $state<TrackOption[]>([]);
  let selected = $state<TrackOption | null>(null);

  function buildPlayHref(track: TrackOption | null): string {
    const base = basePlayUrl ?? `/play/${mediaSourceId}`;
    if (!track) return base;
    // Preserve any existing params (e.g. title, back) and layer subtitle params on top
    const sep = base.includes('?') ? '&' : '?';
    return `${base}${sep}subtitleTrackActive=true&subtitleTrackIndex=${encodeURIComponent(String(track.index))}`;
  }

  function selectTrack(t: TrackOption | null) {
    selected = t;
    open = false;
    playHref = buildPlayHref(t);
  }

  function languageLabel(lang: string | undefined): string {
    if (!lang) return '';
    try {
      return new Intl.DisplayNames(['en'], { type: 'language' }).of(lang) ?? lang;
    } catch {
      return lang;
    }
  }

  onMount(async () => {
    try {
      const [tracksResp, sidecarResp] = await Promise.all([
        getMediaSourceTracks(mediaSourceId).catch(() => ({ subtitleTracks: [] })),
        getMediaSourceSubtitles(mediaSourceId).catch(() => ({ sidecars: [] })),
      ]);

      const embedded: TrackOption[] = (tracksResp.subtitleTracks ?? []).map((t: ProbeTrack, i: number) => ({
        kind: 'embedded' as const,
        index: t.index ?? i,
        label: t.title || languageLabel(t.language) || `Track ${i + 1}`,
        language: t.language,
        forced: t.forced,
        hi: false,
      }));

      const sidecars: TrackOption[] = (sidecarResp.sidecars ?? []).map((s: SubtitleSidecar, i: number) => ({
        kind: 'sidecar' as const,
        index: embedded.length + i,
        label: languageLabel(s.language) || s.format?.toUpperCase() || `Sidecar ${i + 1}`,
        language: s.language,
        forced: s.forced,
        hi: s.hearingImpaired,
      }));

      tracks = [...embedded, ...sidecars];
    } finally {
      loading = false;
    }
  });
</script>

{#if !loading && tracks.length > 0}
  <div class="relative inline-block">
    <button
      type="button"
      onclick={() => (open = !open)}
      class={`hairline inline-flex items-center gap-2 rounded-full px-4 py-2.5 text-sm font-medium transition-colors ${
        selected
          ? 'bg-foreground/10 text-foreground'
          : 'bg-foreground/[0.04] text-muted-foreground hover:bg-foreground/[0.08] hover:text-foreground'
      }`}
    >
      <Subtitles class="h-4 w-4" />
      {selected ? selected.label : 'Subtitles'}
      <ChevronDown class={`h-3.5 w-3.5 transition-transform ${open ? 'rotate-180' : ''}`} />
    </button>

    {#if open}
      <div class="absolute left-0 top-[calc(100%+6px)] z-40 min-w-[200px] overflow-hidden rounded-xl border border-border bg-background/95 shadow-2xl backdrop-blur-xl">
        <ul class="py-1">
          <!-- Off option -->
          <li>
            <button
              type="button"
              onclick={() => selectTrack(null)}
              class="flex w-full items-center gap-3 px-4 py-2 text-left text-sm transition-colors hover:bg-surface/60"
            >
              <span class={`flex h-4 w-4 shrink-0 items-center justify-center ${!selected ? 'text-foreground' : 'text-transparent'}`}>
                <Check class="h-3.5 w-3.5" />
              </span>
              <span class={!selected ? 'text-foreground font-medium' : 'text-muted-foreground'}>Off</span>
            </button>
          </li>
          {#each tracks as t (t.kind + t.index)}
            <li>
              <button
                type="button"
                onclick={() => selectTrack(t)}
                class="flex w-full items-center gap-3 px-4 py-2 text-left text-sm transition-colors hover:bg-surface/60"
              >
                <span class={`flex h-4 w-4 shrink-0 items-center justify-center ${selected?.index === t.index && selected?.kind === t.kind ? 'text-foreground' : 'text-transparent'}`}>
                  <Check class="h-3.5 w-3.5" />
                </span>
                <span class={selected?.index === t.index && selected?.kind === t.kind ? 'text-foreground font-medium' : 'text-muted-foreground'}>
                  {t.label}
                  {#if t.forced}<span class="ml-1 text-[10px] uppercase text-muted-foreground/60">Forced</span>{/if}
                  {#if t.hi}<span class="ml-1 text-[10px] uppercase text-muted-foreground/60">CC</span>{/if}
                </span>
                <span class="ml-auto text-[10px] uppercase text-muted-foreground/40">{t.kind}</span>
              </button>
            </li>
          {/each}
        </ul>
      </div>
    {/if}
  </div>
{/if}
