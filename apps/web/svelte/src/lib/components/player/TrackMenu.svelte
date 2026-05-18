<script lang="ts">
  import { Check, X } from 'lucide-svelte';

  interface TrackOption {
    index: number;
    label: string;
    sublabel?: string;
    isNone?: boolean;
  }

  interface Props {
    open: boolean;
    title: string;
    tracks: TrackOption[];
    activeIndex: number | null;
    onSelect: (index: number | null) => void;
    onClose: () => void;
  }

  let { open, title, tracks, activeIndex, onSelect, onClose }: Props = $props();

  function handleKey(e: KeyboardEvent) {
    if (!open) return;
    if (e.key === 'Escape') {
      e.stopPropagation();
      onClose();
    }
  }
</script>

<svelte:window onkeydown={handleKey} />

{#if open}
  <!-- Backdrop -->
  <div
    class="fixed inset-0 z-40"
    role="presentation"
    onclick={onClose}
  ></div>

  <!-- Panel -->
  <div
    class="absolute bottom-20 right-4 z-50 w-64 overflow-hidden rounded-2xl border border-white/10 bg-black/90 shadow-2xl backdrop-blur-xl"
    role="dialog"
    aria-modal="true"
    aria-label={title}
  >
    <div class="flex items-center justify-between border-b border-white/10 px-4 py-3">
      <span class="text-xs font-semibold uppercase tracking-[0.15em] text-white/50">{title}</span>
      <button
        onclick={onClose}
        class="flex h-6 w-6 items-center justify-center rounded-full text-white/40 transition-colors hover:bg-white/10 hover:text-white"
        aria-label="Close"
      >
        <X class="h-3.5 w-3.5" />
      </button>
    </div>

    <ul class="max-h-72 overflow-y-auto py-1" role="listbox" aria-label={title}>
      {#each tracks as track (track.index)}
        {@const isActive = track.isNone ? activeIndex === null : activeIndex === track.index}
        <li>
          <button
            class={`flex w-full items-center gap-3 px-4 py-2.5 text-left text-sm transition-colors ${isActive ? 'bg-white/10 text-white' : 'text-white/70 hover:bg-white/[0.06] hover:text-white'}`}
            role="option"
            aria-selected={isActive}
            onclick={() => onSelect(track.isNone ? null : track.index)}
          >
            <span class="flex h-4 w-4 shrink-0 items-center justify-center">
              {#if isActive}
                <Check class="h-3.5 w-3.5 text-primary-glow" />
              {/if}
            </span>
            <span class="flex flex-col">
              <span class="leading-snug">{track.label}</span>
              {#if track.sublabel}
                <span class="text-[11px] text-white/40">{track.sublabel}</span>
              {/if}
            </span>
          </button>
        </li>
      {/each}
    </ul>
  </div>
{/if}
