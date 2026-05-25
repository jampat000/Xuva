<script lang="ts">
  /**
   * ActivityRing — animated SVG donut showing probed vs unprobed files.
   * Purple fill = probed; dim track = unprobed.
   */
  let {
    probed = 0,
    total = 1,
    size = 120,
  } = $props<{ probed?: number; total?: number; size?: number }>();

  const R = 44;
  const C = $derived(2 * Math.PI * R);
  const pct = $derived(total > 0 ? Math.min(1, Math.max(0, probed / total)) : 0);
  const fill = $derived(pct * C);
  const gap  = $derived(C - fill);
  const color = $derived(
    pct >= 1   ? 'oklch(0.78 0.22 145)' :  // green — fully probed
    pct >= 0.5 ? 'oklch(0.85 0.22 75)'  :  // amber — more than half done
                 'oklch(0.68 0.26 22)'      // red   — less than half done
  );
</script>

<svg
  viewBox="0 0 100 100"
  width={size}
  height={size}
  style="transform: rotate(-90deg); overflow: visible;"
  aria-hidden="true"
>
  <!-- Track -->
  <circle
    cx="50" cy="50" r={R}
    fill="none"
    stroke="oklch(1 0 0 / 0.08)"
    stroke-width="10"
  />
  <!-- Filled portion -->
  {#if pct > 0}
    <circle
      cx="50" cy="50" r={R}
      fill="none"
      stroke={color}
      stroke-width="10"
      stroke-dasharray="{fill} {gap}"
      stroke-linecap="round"
      style="transition: stroke-dasharray 0.8s cubic-bezier(0.4,0,0.2,1)"
    />
  {/if}
</svg>
