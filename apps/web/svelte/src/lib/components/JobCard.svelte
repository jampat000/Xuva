<script lang="ts">
  /**
   * JobCard — reusable card for one automation job (scan / metadata / probe).
   * Shows status badge, optional progress bar, last/next run, and action buttons.
   */
  type Props = {
    name: string;
    emoji?: string;
    status?: string;           // "idle" | "running" | "paused" | "disabled"
    lastRunAt?: string | null;
    nextRunAt?: string | null;
    intervalMin?: number;
    error?: string | null;
    progress?: number | null;  // 0–100
    progressDetail?: string | null;  // e.g. "3,440 / 4,412"
    throughput?: string | null;      // e.g. "48 files/min"
    eta?: string | null;             // e.g. "~20 min"
    canRunNow?: boolean;
    canStop?: boolean;
    runBusy?: boolean;
    onRunNow?: () => void;
    onStop?: () => void;
  };

  let {
    name,
    emoji = '⚙',
    status = 'idle',
    lastRunAt = null,
    nextRunAt = null,
    intervalMin,
    error = null,
    progress = null,
    progressDetail = null,
    throughput = null,
    eta = null,
    canRunNow = true,
    canStop = false,
    runBusy = false,
    onRunNow,
    onStop,
  }: Props = $props();

  // ── Derived display values ─────────────────────────────────────────────────

  function relativeTime(iso: string | null | undefined): string {
    if (!iso) return 'Never';
    const d = new Date(iso);
    if (isNaN(d.getTime()) || d.getFullYear() < 2000) return 'Never';
    const now = Date.now();
    const diff = now - d.getTime();
    if (diff < 0) {
      // Future
      const sec = Math.abs(diff) / 1000;
      if (sec < 90)   return `in ${Math.round(sec)}s`;
      if (sec < 3600) return `in ${Math.round(sec / 60)}m`;
      if (sec < 86400) return `in ${Math.round(sec / 3600)}h`;
      return `in ${Math.round(sec / 86400)}d`;
    }
    const sec = diff / 1000;
    if (sec < 90)   return 'just now';
    if (sec < 3600) return `${Math.round(sec / 60)}m ago`;
    if (sec < 86400) return `${Math.round(sec / 3600)}h ago`;
    return `${Math.round(sec / 86400)}d ago`;
  }

  function intervalLabel(mins: number | undefined): string {
    if (!mins) return '';
    if (mins < 60) return `Every ${mins}m`;
    if (mins === 60) return 'Every hour';
    const h = Math.floor(mins / 60);
    const m = mins % 60;
    return m === 0 ? `Every ${h}h` : `Every ${h}h ${m}m`;
  }

  const statusLabel = $derived(
    status === 'running'  ? 'Running'  :
    status === 'paused'   ? 'Paused'   :
    status === 'disabled' ? 'Disabled' :
    'Idle'
  );

  const dotClass = $derived(
    status === 'running'  ? 'bg-primary-glow animate-pulse' :
    status === 'paused'   ? 'bg-amber-400' :
    status === 'disabled' ? 'bg-foreground/20' :
    'bg-foreground/30'
  );

  const lastLabel  = $derived(relativeTime(lastRunAt));
  const nextLabel  = $derived(status === 'running' ? 'Now' : relativeTime(nextRunAt));
  const intLabel   = $derived(intervalLabel(intervalMin));

  const pct = $derived(
    progress != null ? Math.min(100, Math.max(0, progress)) : null
  );
</script>

<article
  class="hairline flex flex-col rounded-2xl bg-surface/40 p-5 transition-colors hover:bg-surface/60 {status === 'running' ? 'ring-1 ring-primary-glow/30' : ''}"
>
  <!-- Header -->
  <div class="mb-4 flex items-center gap-3">
    <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-foreground/[0.06] text-xl" aria-hidden="true">
      {emoji}
    </span>
    <div class="min-w-0 flex-1">
      <div class="truncate font-semibold text-sm">{name}</div>
      <div class="mt-0.5 flex items-center gap-1.5">
        <span class="inline-block h-1.5 w-1.5 rounded-full {dotClass}"></span>
        <span class="text-[11px] text-muted-foreground">{statusLabel}</span>
      </div>
    </div>
  </div>

  <!-- Progress bar (when running) -->
  {#if pct != null}
    <div class="mb-4 space-y-2">
      <div class="h-2 w-full overflow-hidden rounded-full bg-foreground/10">
        <div
          class="h-full rounded-full bg-primary-glow transition-all duration-500"
          style="width: {pct}%"
        ></div>
      </div>
      <div class="flex items-center justify-between text-[11px] text-muted-foreground">
        {#if progressDetail}
          <span>{progressDetail}</span>
        {:else}
          <span>{Math.round(pct)}%</span>
        {/if}
        {#if throughput}
          <span>{throughput}</span>
        {/if}
      </div>
      {#if eta}
        <div class="text-[11px] text-muted-foreground">ETA: {eta}</div>
      {/if}
    </div>
  {/if}

  <!-- Meta: last / next / interval -->
  <div class="mb-4 space-y-1 text-[11px] text-muted-foreground">
    <div class="flex justify-between">
      <span class="uppercase tracking-[0.15em] font-semibold text-muted-foreground/60">Last</span>
      <span>{lastLabel}</span>
    </div>
    {#if status !== 'disabled' && nextRunAt}
      <div class="flex justify-between">
        <span class="uppercase tracking-[0.15em] font-semibold text-muted-foreground/60">Next</span>
        <span>{nextLabel}</span>
      </div>
    {/if}
    {#if intLabel}
      <div class="flex justify-between">
        <span class="uppercase tracking-[0.15em] font-semibold text-muted-foreground/60">Schedule</span>
        <span>{intLabel}</span>
      </div>
    {/if}
  </div>

  <!-- Error -->
  {#if error}
    <p class="mb-3 truncate text-[11px] text-red-400">{error}</p>
  {/if}

  <!-- Actions -->
  <div class="mt-auto flex gap-2">
    {#if canStop && status === 'running' && onStop}
      <button
        type="button"
        onclick={onStop}
        class="hairline flex-1 rounded-xl bg-foreground/[0.04] py-2.5 text-xs font-semibold uppercase tracking-[0.15em] text-muted-foreground transition-colors hover:bg-red-400/10 hover:text-red-300"
      >
        Stop
      </button>
    {:else if canRunNow && onRunNow}
      <button
        type="button"
        onclick={onRunNow}
        disabled={status === 'running' || status === 'disabled' || runBusy}
        class="hairline flex-1 rounded-xl bg-foreground/[0.04] py-2.5 text-xs font-semibold uppercase tracking-[0.15em] text-muted-foreground transition-colors hover:bg-primary-glow/10 hover:text-primary-glow disabled:cursor-not-allowed disabled:opacity-40"
      >
        {#if runBusy}
          <span class="inline-block h-3 w-3 animate-spin rounded-full border border-current border-t-transparent"></span>
        {:else}
          Run now
        {/if}
      </button>
    {/if}
  </div>
</article>
