<script lang="ts">
  import { navigating } from '$app/state';

  // Threshold below which we hide the bar entirely — fast cache hits resolve
  // in <50ms and a flash of progress reads worse than no progress at all.
  const SHOW_AFTER_MS = 150;

  let visible = $state(false);
  let timer: ReturnType<typeof setTimeout> | null = null;

  $effect(() => {
    // $app/state.navigating exposes reactive getters (to/from/type) that are
    // null when no navigation is in flight. We arm a delay timer when a nav
    // starts and clear it when it ends, so cache-hit navigations <150ms never
    // flash the bar.
    if (navigating.to) {
      if (timer) clearTimeout(timer);
      timer = setTimeout(() => { visible = true; }, SHOW_AFTER_MS);
    } else {
      if (timer) { clearTimeout(timer); timer = null; }
      visible = false;
    }
    return () => {
      if (timer) clearTimeout(timer);
      timer = null;
    };
  });
</script>

{#if visible}
  <div
    aria-hidden="true"
    class="pointer-events-none fixed inset-x-0 top-0 z-[60] h-0.5 overflow-hidden"
  >
    <div class="h-full w-full origin-left animate-[nav-progress_1.4s_ease-out_infinite] bg-gradient-to-r from-primary-glow via-primary to-primary-glow"></div>
  </div>
{/if}

<style>
  @keyframes nav-progress {
    0%   { transform: translateX(-100%) scaleX(0.4); }
    50%  { transform: translateX(20%) scaleX(0.6); }
    100% { transform: translateX(100%) scaleX(0.4); }
  }
</style>
