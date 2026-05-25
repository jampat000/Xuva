<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import {
    getSystemStatus, getCatalogHealth, getSessions, getJobs, getPerformanceSettings,
    getCatalogPlayabilityAudit,
    scanAllLibraries, startProbeJob,
    type SystemStatusResponse, type CatalogHealthResponse,
    type SessionItem, type JobsStatusResponse, type PerformanceSettingsResponse,
    type PlayabilityAuditResponse,
  } from '$lib/api/operator';
  import { startBackfill, stopBackfill } from '$lib/api/browse';
  import { createEventStream } from '$lib/events/stream';
  import { appState } from '$lib/stores/appState.svelte';
  import Header from '$lib/components/Header.svelte';
  import JobCard from '$lib/components/JobCard.svelte';
  import NowPlayingCard from '$lib/components/NowPlayingCard.svelte';
  import ActivityRing from '$lib/components/ActivityRing.svelte';

  let { data } = $props();

  // ── Live state ─────────────────────────────────────────────────────────────
  let sys      = $state<SystemStatusResponse | null>(data.sys);
  let health   = $state<CatalogHealthResponse | null>(data.health);
  let sessions = $state<SessionItem[]>(data.sessions);
  let jobs     = $state<JobsStatusResponse | null>(data.jobs);
  let perf     = $state<PerformanceSettingsResponse | null>(data.perf);
  let audit    = $state<PlayabilityAuditResponse | null>(null);
  let auditLoading = $state(false);
  let lastUpdatedAt = $state<string>(new Date().toLocaleTimeString());

  // ── Job busy flags ─────────────────────────────────────────────────────────
  let scanBusy   = $state(false);
  let metaBusy   = $state(false);
  let probeBusy  = $state(false);
  let runAllBusy = $state(false);

  // ── Derived values ─────────────────────────────────────────────────────────
  const probeRunning = $derived(
    jobs?.probe?.status === 'running' ||
    (jobs?.probe?.activeJobs?.some(j => j.status === 'running') ?? false)
  );

  const activeProbeJob = $derived(
    jobs?.probe?.activeJobs?.find(j => j.status === 'running') ?? null
  );

  const probeProgress = $derived.by(() => {
    if (!activeProbeJob) return null;
    const total = activeProbeJob.total ?? 0;
    const done  = activeProbeJob.completed ?? 0;
    if (total === 0) return null;
    return { pct: Math.round((done / total) * 100), detail: `${done.toLocaleString()} / ${total.toLocaleString()}` };
  });

  const metaProgress = $derived.by(() => {
    const bf = jobs?.metadata?.backfill;
    if (!bf?.running || bf.total === 0) return null;
    const done = bf.refreshed + bf.failed;
    return { pct: Math.round((done / bf.total) * 100), detail: `${done.toLocaleString()} / ${bf.total.toLocaleString()}` };
  });

  const totalFiles = $derived(health?.summary?.mediaSources ?? 0);
  const unprobed   = $derived(health?.unprobed ?? 0);
  const probed     = $derived(Math.max(0, totalFiles - unprobed));
  const cpuPct     = $derived(Math.round(sys?.cpu?.percent ?? 0));
  const memPct     = $derived(Math.round(sys?.memory?.usedPercent ?? 0));

  const gpuQueue    = $derived(perf?.queues?.find(q => q.name === 'gpu') ?? null);
  const gpuWorkers  = $derived(gpuQueue?.workers ?? perf?.limits?.gpuWorkers ?? 0);
  const gpuActive   = $derived(gpuQueue?.active ?? 0);
  const hwAvail     = $derived(perf?.hardwareAcceleration?.available ?? false);
  const hwEncoder   = $derived(
    perf?.hardwareAcceleration?.selectedEncoder?.label ??
    perf?.hardwareAcceleration?.encoders?.[0]?.label ??
    null
  );
  // Real GPU hardware stats (from nvidia-smi / WMI); may be absent.
  const gpuHW       = $derived(sys?.gpu ?? null);
  const gpuHasReal  = $derived(gpuHW != null && gpuHW.utilizationPct != null);
  const gpuUtil     = $derived(
    gpuHasReal
      ? Math.round(gpuHW!.utilizationPct!)
      : gpuWorkers > 0 ? Math.round((gpuActive / gpuWorkers) * 100) : 0
  );
  const gpuAdapterName = $derived(
    gpuHW?.adapterName ??
    perf?.hardwareAcceleration?.selectedEncoder?.label ??
    null
  );

  // ── SVG arc gauge helpers ──────────────────────────────────────────────────
  const _R    = 36;
  const _CIRC = 2 * Math.PI * _R; // ≈ 226.2

  function arcDash(pct: number): string {
    const p = Math.min(100, Math.max(0, pct));
    const filled = _CIRC * (p / 100);
    return `${filled} ${_CIRC - filled}`;
  }

  function gaugeColor(pct: number, warnAt = 70, critAt = 90): string {
    if (pct >= critAt)  return 'oklch(0.65 0.22 25)';   // red
    if (pct >= warnAt)  return 'oklch(0.80 0.18 65)';   // amber
    return 'oklch(0.62 0.22 285)';                       // purple (primary)
  }

  function diskColor(pct: number): string {
    if (pct >= 90) return 'bg-red-400';
    if (pct >= 75) return 'bg-amber-400';
    return 'bg-primary-glow';
  }

  function diskTextColor(pct: number): string {
    if (pct >= 90) return 'text-red-400';
    if (pct >= 75) return 'text-amber-400';
    return 'text-foreground';
  }

  // ── Formatters ─────────────────────────────────────────────────────────────
  function fmtBytes(bytes: number | undefined | null): string {
    if (!bytes || bytes === 0) return '0 B';
    if (bytes >= 1099511627776) return `${(bytes / 1099511627776).toFixed(1)} TB`;
    if (bytes >= 1073741824)    return `${(bytes / 1073741824).toFixed(1)} GB`;
    if (bytes >= 1048576)       return `${(bytes / 1048576).toFixed(1)} MB`;
    if (bytes >= 1024)          return `${(bytes / 1024).toFixed(0)} KB`;
    return `${bytes} B`;
  }

  function fmtBps(bps: number | undefined | null): string {
    if (!bps || bps === 0) return '0 B/s';
    if (bps >= 1073741824) return `${(bps / 1073741824).toFixed(1)} GB/s`;
    if (bps >= 1048576)    return `${(bps / 1048576).toFixed(1)} MB/s`;
    if (bps >= 1024)       return `${(bps / 1024).toFixed(1)} KB/s`;
    return `${bps.toFixed(0)} B/s`;
  }

  function fmtLinkSpeed(bps: number | undefined | null): string {
    if (!bps) return '';
    if (bps >= 1_000_000_000) return `${(bps / 1_000_000_000).toFixed(0)} Gbps`;
    if (bps >= 1_000_000)     return `${(bps / 1_000_000).toFixed(0)} Mbps`;
    return `${(bps / 1_000).toFixed(0)} Kbps`;
  }

  function netRagColor(bps: number | undefined | null, linkBps: number | undefined | null): string {
    const green = 'oklch(0.78 0.22 145)';
    if (!linkBps || bps == null) return green; // no link data — can't judge, show green
    if (bps <= 0) return green;                // no traffic — green (nothing to worry about)
    const pct = bps / linkBps;
    if (pct >= 0.8) return 'oklch(0.68 0.26 22)';  // red  — >80%
    if (pct >= 0.5) return 'oklch(0.85 0.22 75)';  // amber — 50–80%
    return green;                                    // green — <50%
  }

  function netPct(bps: number | undefined | null, linkBps: number | undefined | null): string {
    if (!linkBps || !bps) return '';
    return `${Math.min(100, (bps / linkBps) * 100).toFixed(1)}%`;
  }

  // ── Data refresh ───────────────────────────────────────────────────────────
  function stamp() { lastUpdatedAt = new Date().toLocaleTimeString(); }

  async function refreshSys() {
    try { sys = await getSystemStatus(); stamp(); } catch { /* silent */ }
  }

  async function refreshJobs() {
    try {
      const [j, s] = await Promise.allSettled([getJobs(), getSessions()]);
      if (j.status === 'fulfilled') jobs = j.value;
      if (s.status === 'fulfilled') sessions = s.value.sessions ?? [];
      stamp();
    } catch { /* silent */ }
  }

  async function refreshHealth() {
    try { health = await getCatalogHealth(); stamp(); } catch { /* silent */ }
  }

  async function refreshPerf() {
    try { perf = await getPerformanceSettings(); } catch { /* silent */ }
  }

  async function loadAudit() {
    if (auditLoading) return;
    auditLoading = true;
    try { audit = await getCatalogPlayabilityAudit(); } catch { /* silent */ } finally { auditLoading = false; }
  }

  async function refreshAll() {
    await Promise.allSettled([refreshSys(), refreshJobs(), refreshHealth(), refreshPerf()]);
  }

  // ── Job actions ────────────────────────────────────────────────────────────
  async function handleScanNow() {
    scanBusy = true;
    try { await scanAllLibraries(); await refreshJobs(); } catch { /* ignore */ } finally { scanBusy = false; }
  }

  async function handleMetaNow() {
    metaBusy = true;
    try { await startBackfill('tmdb'); await refreshJobs(); } catch { /* ignore */ } finally { metaBusy = false; }
  }

  async function handleMetaStop() {
    metaBusy = true;
    try { await stopBackfill(); await refreshJobs(); } catch { /* ignore */ } finally { metaBusy = false; }
  }

  async function handleProbeNow() {
    probeBusy = true;
    try { await startProbeJob(0); await refreshJobs(); } catch { /* ignore */ } finally { probeBusy = false; }
  }

  async function handleRunAll() {
    runAllBusy = true;
    try {
      await scanAllLibraries();
      await new Promise(r => setTimeout(r, 400));
      await startBackfill('tmdb');
      await new Promise(r => setTimeout(r, 400));
      await startProbeJob(0);
      await refreshJobs();
    } catch { /* ignore */ } finally { runAllBusy = false; }
  }

  // ── SSE ────────────────────────────────────────────────────────────────────
  const stream = createEventStream();

  $effect(() => {
    const unsub = stream.subscribeAny(({ type }) => {
      if (
        type.startsWith('automation.') ||
        type.startsWith('scan.')       ||
        type.startsWith('probe.')      ||
        type.startsWith('metadata.')   ||
        type === 'api.session.accepted' ||
        type === 'api.session.stopped'
      ) {
        refreshJobs();
      }
    });
    return unsub;
  });

  // ── Polling ────────────────────────────────────────────────────────────────
  let timers: ReturnType<typeof setInterval>[] = [];

  onMount(() => {
    stream.connect();
    timers = [
      setInterval(refreshSys,    5_000),   // system stats — fast
      setInterval(refreshJobs,   5_000),   // jobs + sessions — fast
      setInterval(refreshPerf,   15_000),  // worker queues
      setInterval(refreshHealth, 60_000),  // catalog health — slow
    ];
  });

  onDestroy(() => {
    stream.disconnect();
    for (const t of timers) clearInterval(t);
  });
</script>

<svelte:head>
  <title>Dashboard — {appState.serverName}</title>
  <meta name="description" content="Live server status — system resources, library health, automation jobs, and active sessions." />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  <main class="px-6 pb-32 pt-24 md:px-12 md:pt-28 lg:px-20">

    <!-- ── Page header ────────────────────────────────────────────────────── -->
    <header class="relative mb-12 flex flex-wrap items-end justify-between gap-4">
      <div
        aria-hidden="true"
        class="pointer-events-none absolute -inset-x-6 -top-10 -z-10 h-[180px] opacity-40 md:-inset-x-12 lg:-inset-x-20"
        style="background: radial-gradient(50% 100% at 15% 0%, oklch(0.62 0.22 285 / 0.25), transparent 70%);"
      ></div>
      <div>
        <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">Server</div>
        <h1 class="font-serif-display text-[clamp(2rem,4vw,3.25rem)] leading-[1] tracking-tight">Dashboard</h1>
      </div>
      <div class="flex items-center gap-3">
        <span class="flex items-center gap-1.5 text-xs text-muted-foreground">
          <span
            class="inline-block h-2 w-2 rounded-full bg-green-400"
            style="box-shadow: 0 0 6px oklch(0.72 0.20 155 / 0.8)"
          ></span>
          Live · {lastUpdatedAt}
        </span>
        <button
          type="button"
          onclick={refreshAll}
          class="hairline rounded-full px-3 py-1.5 text-[11px] font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.06] hover:text-foreground"
        >
          Refresh
        </button>
        <button
          type="button"
          onclick={handleRunAll}
          disabled={runAllBusy}
          class="inline-flex items-center gap-1.5 rounded-full bg-gradient-primary px-4 py-1.5 text-[10px] font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110 disabled:opacity-60"
        >
          {#if runAllBusy}
            <span class="h-2.5 w-2.5 animate-spin rounded-full border border-white/40 border-t-white"></span>
          {:else}
            <svg class="h-2.5 w-2.5" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M8 5v14l11-7z"/></svg>
          {/if}
          Run All Jobs
        </button>
      </div>
    </header>

    <!-- ── SYSTEM RESOURCES ──────────────────────────────────────────────── -->
    <section class="mb-10">
      <h2 class="mb-4 text-[10px] font-semibold uppercase tracking-[0.3em] text-muted-foreground">System Resources</h2>
      <div class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-5">

        <!-- CPU -->
        <div class="hairline flex flex-col items-center gap-3 rounded-2xl bg-surface/40 p-5">
          <div class="w-full text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">CPU</div>
          <div class="relative flex items-center justify-center">
            <svg viewBox="0 0 100 100" class="h-24 w-24 -rotate-90" aria-hidden="true">
              <circle cx="50" cy="50" r={_R} fill="none"
                style="stroke: oklch(1 0 0 / 0.08); stroke-width: 10;" />
              <circle cx="50" cy="50" r={_R} fill="none"
                style="stroke: {gaugeColor(cpuPct)}; stroke-width: 10; stroke-linecap: round;
                       stroke-dasharray: {arcDash(cpuPct)};
                       transition: stroke-dasharray 1s ease, stroke 0.5s ease;" />
            </svg>
            <div class="absolute inset-0 flex rotate-90 flex-col items-center justify-center">
              <span class="text-2xl font-bold leading-none tabular-nums">{cpuPct}%</span>
            </div>
          </div>
          <div class="w-full space-y-1 text-center">
            <div class="text-xs font-medium">{cpuPct}% utilisation</div>
            <div class="text-[11px] text-muted-foreground">{sys?.cpu?.cores ?? '—'} cores</div>
          </div>
        </div>

        <!-- Memory -->
        <div class="hairline flex flex-col items-center gap-3 rounded-2xl bg-surface/40 p-5">
          <div class="w-full text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">Memory</div>
          <div class="relative flex items-center justify-center">
            <svg viewBox="0 0 100 100" class="h-24 w-24 -rotate-90" aria-hidden="true">
              <circle cx="50" cy="50" r={_R} fill="none"
                style="stroke: oklch(1 0 0 / 0.08); stroke-width: 10;" />
              <circle cx="50" cy="50" r={_R} fill="none"
                style="stroke: {gaugeColor(memPct, 75, 90)}; stroke-width: 10; stroke-linecap: round;
                       stroke-dasharray: {arcDash(memPct)};
                       transition: stroke-dasharray 1s ease, stroke 0.5s ease;" />
            </svg>
            <div class="absolute inset-0 flex rotate-90 flex-col items-center justify-center">
              <span class="text-2xl font-bold leading-none tabular-nums">{memPct}%</span>
            </div>
          </div>
          <div class="w-full space-y-1 text-center">
            <div class="text-xs font-medium">{fmtBytes(sys?.memory?.usedBytes)} used</div>
            <div class="text-[11px] text-muted-foreground">of {fmtBytes(sys?.memory?.totalBytes)}</div>
          </div>
        </div>

        <!-- Network I/O -->
        <div class="hairline flex flex-col gap-3 rounded-2xl bg-surface/40 p-5">
          <div class="flex items-center justify-between gap-2">
            <div class="text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">Network I/O</div>
            {#if sys?.network?.linkSpeedBps}
              <div class="text-[10px] tabular-nums text-muted-foreground">{fmtLinkSpeed(sys.network.linkSpeedBps)} link</div>
            {/if}
          </div>
          <div class="flex flex-1 flex-col justify-center gap-4 py-2">
            <div class="flex items-center gap-3">
              <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl text-sm font-bold"
                style="background: oklch(0.72 0.20 155 / 0.12); color: oklch(0.72 0.20 155)">↓</div>
              <div class="min-w-0 flex-1">
                <div class="flex items-baseline gap-2">
                  <div class="font-semibold leading-none tabular-nums"
                    style="color: {netRagColor(sys?.network?.receiveBps, sys?.network?.linkSpeedBps)}">
                    {fmtBps(sys?.network?.receiveBps)}
                  </div>
                  {#if sys?.network?.linkSpeedBps}
                    <div class="text-[10px] tabular-nums text-muted-foreground">{netPct(sys.network.receiveBps, sys.network.linkSpeedBps)}</div>
                  {/if}
                </div>
                <div class="mt-0.5 text-[11px] text-muted-foreground">RECV</div>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl text-sm font-bold"
                style="background: oklch(0.65 0.20 255 / 0.12); color: oklch(0.70 0.18 255)">↑</div>
              <div class="min-w-0 flex-1">
                <div class="flex items-baseline gap-2">
                  <div class="font-semibold leading-none tabular-nums"
                    style="color: {netRagColor(sys?.network?.transmitBps, sys?.network?.linkSpeedBps)}">
                    {fmtBps(sys?.network?.transmitBps)}
                  </div>
                  {#if sys?.network?.linkSpeedBps}
                    <div class="text-[10px] tabular-nums text-muted-foreground">{netPct(sys.network.transmitBps, sys.network.linkSpeedBps)}</div>
                  {/if}
                </div>
                <div class="mt-0.5 text-[11px] text-muted-foreground">XMIT</div>
              </div>
            </div>
          </div>
        </div>

        <!-- GPU -->
        <div class="hairline flex flex-col gap-3 rounded-2xl bg-surface/40 p-5">
          <div class="text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">GPU</div>
          {#if hwAvail || gpuHasReal}
            <!-- Utilisation gauge — real hardware % when available, worker-slot % otherwise -->
            <div class="flex items-center gap-4">
              <div class="relative flex shrink-0 items-center justify-center">
                <svg viewBox="0 0 100 100" class="h-20 w-20 -rotate-90" aria-hidden="true">
                  <circle cx="50" cy="50" r={_R} fill="none"
                    style="stroke: oklch(1 0 0 / 0.08); stroke-width: 10;" />
                  <circle cx="50" cy="50" r={_R} fill="none"
                    style="stroke: {gaugeColor(gpuUtil)}; stroke-width: 10; stroke-linecap: round;
                           stroke-dasharray: {arcDash(gpuUtil)};
                           transition: stroke-dasharray 1s ease, stroke 0.5s ease;" />
                </svg>
                <div class="absolute inset-0 flex rotate-90 flex-col items-center justify-center">
                  <span class="text-xl font-bold leading-none tabular-nums">{gpuUtil}%</span>
                  <span class="text-[9px] text-muted-foreground/60">{gpuHasReal ? 'GPU' : 'workers'}</span>
                </div>
              </div>
              <div class="min-w-0 flex-1 space-y-1.5">
                {#if gpuAdapterName}
                  <div class="truncate text-[12px] font-medium leading-snug" title={gpuAdapterName}>{gpuAdapterName}</div>
                {/if}
                {#if hwEncoder && hwEncoder !== gpuAdapterName}
                  <div class="text-[11px] text-muted-foreground">{hwEncoder}</div>
                {/if}
                <div class="text-[11px] text-muted-foreground/70">{gpuActive} / {gpuWorkers} worker{gpuWorkers !== 1 ? 's' : ''}</div>
                {#if gpuWorkers === 0}
                  <div class="text-[10px] text-amber-400/70">Enable in Settings → Performance</div>
                {/if}
              </div>
            </div>
            <!-- VRAM bar (only when real stats available) -->
            {#if gpuHW && gpuHW.vramTotalBytes && gpuHW.vramTotalBytes > 0}
              {@const vramPct = Math.min(100, ((gpuHW.vramUsedBytes ?? 0) / gpuHW.vramTotalBytes) * 100)}
              <div class="space-y-1">
                <div class="flex justify-between text-[10px] text-muted-foreground">
                  <span>VRAM</span>
                  <span class="tabular-nums">{fmtBytes(gpuHW.vramUsedBytes)} / {fmtBytes(gpuHW.vramTotalBytes)}</span>
                </div>
                <div class="h-1.5 overflow-hidden rounded-full bg-white/8">
                  <div class="h-full rounded-full transition-[width] duration-1000"
                    style="width: {vramPct}%; background: {gaugeColor(vramPct, 75, 90)}"></div>
                </div>
              </div>
            {/if}
          {:else}
            <div class="flex flex-1 flex-col items-center justify-center gap-2 py-4">
              <svg class="h-8 w-8 text-muted-foreground/20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 3v1.5M4.5 8.25H3m18 0h-1.5M4.5 12H3m18 0h-1.5m-15 3.75H3m18 0h-1.5M8.25 19.5V21M12 3v1.5m0 15V21m3.75-18v1.5m0 15V21m-9-1.5h10.5a2.25 2.25 0 0 0 2.25-2.25V6.75a2.25 2.25 0 0 0-2.25-2.25H6.75A2.25 2.25 0 0 0 4.5 6.75v10.5a2.25 2.25 0 0 0 2.25 2.25Zm.75-12h9v9h-9v-9Z" />
              </svg>
              <div class="text-xs text-muted-foreground">Not available</div>
            </div>
          {/if}
        </div>

        <!-- Process -->
        <div class="hairline flex flex-col gap-3 rounded-2xl bg-surface/40 p-5">
          <div class="text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">Process</div>
          <div class="flex flex-1 flex-col justify-center gap-4 py-1">
            <div>
              <div class="text-3xl font-bold leading-none tabular-nums">{sys?.process?.goroutines ?? '—'}</div>
              <div class="mt-1 text-[11px] text-muted-foreground">Goroutines</div>
            </div>
            <div class="space-y-2 border-t border-border pt-3">
              <div class="flex items-center justify-between text-[11px]">
                <span class="text-muted-foreground">Heap</span>
                <span class="font-medium tabular-nums">{fmtBytes(sys?.process?.goAllocBytes)}</span>
              </div>
              <div class="flex items-center justify-between text-[11px]">
                <span class="text-muted-foreground">Reserved</span>
                <span class="font-medium tabular-nums">{fmtBytes(sys?.process?.goSysBytes)}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ── STORAGE ────────────────────────────────────────────────────────── -->
    {#if sys?.disks && sys.disks.length > 0}
    <section class="mb-10">
      <h2 class="mb-4 text-[10px] font-semibold uppercase tracking-[0.3em] text-muted-foreground">Storage</h2>
      <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {#each sys.disks as disk (disk.path ?? disk.name)}
          {@const dPct = Math.round(disk.usedPercent ?? 0)}
          <div class="hairline rounded-2xl bg-surface/40 p-4 {dPct >= 90 ? 'border-red-400/30' : dPct >= 75 ? 'border-amber-400/30' : ''}">
            <div class="mb-3 flex items-start justify-between gap-2">
              <div class="min-w-0">
                <div class="truncate text-sm font-medium font-mono">{disk.path ?? disk.name ?? 'Unknown'}</div>
                <div class="mt-0.5 text-[11px] text-muted-foreground tabular-nums">
                  {fmtBytes(disk.usedBytes)} / {fmtBytes(disk.totalBytes)}
                </div>
              </div>
              <div class="shrink-0 text-right">
                <div class="text-sm font-bold tabular-nums {diskTextColor(dPct)}">{dPct}%</div>
                <div class="text-[11px] text-muted-foreground tabular-nums">{fmtBytes(disk.freeBytes)} free</div>
              </div>
            </div>
            <div class="h-1.5 w-full overflow-hidden rounded-full bg-foreground/10">
              <div
                class="h-full rounded-full transition-all duration-700 {diskColor(dPct)}"
                style="width: {dPct}%"
              ></div>
            </div>
            {#if disk.error}
              <p class="mt-2 text-[10px] text-red-400">{disk.error}</p>
            {:else if disk.sharedWithData}
              <p class="mt-2 text-[10px] text-muted-foreground/50">Shared with data directory</p>
            {/if}
          </div>
        {/each}
      </div>
    </section>
    {/if}

    <!-- ── LIBRARY ─────────────────────────────────────────────────────────── -->
    <section class="mb-10">
      <h2 class="mb-4 text-[10px] font-semibold uppercase tracking-[0.3em] text-muted-foreground">Library</h2>

      <!-- Stat strip -->
      <div class="mb-4 grid grid-cols-3 gap-3 sm:grid-cols-6">
        {#each [
          { label: 'Movies',      value: health?.summary?.movies      ?? 0, warn: false },
          { label: 'Series',      value: health?.summary?.series      ?? 0, warn: false },
          { label: 'Episodes',    value: health?.summary?.episodes    ?? 0, warn: false },
          { label: 'Files',       value: totalFiles,                         warn: false },
          { label: 'Unprobed',    value: unprobed,                          warn: unprobed > 0 },
          { label: 'Need Review', value: health?.needsReview          ?? 0, warn: (health?.needsReview ?? 0) > 0 },
        ] as stat}
          <div class="hairline rounded-2xl bg-surface/40 p-4 {stat.warn ? 'border-amber-400/30 bg-amber-400/[0.04]' : ''}">
            <div class="text-2xl font-bold leading-none tabular-nums {stat.warn ? 'text-amber-400' : ''}">
              {stat.value.toLocaleString()}
            </div>
            <div class="mt-1.5 text-[11px] text-muted-foreground">{stat.label}</div>
          </div>
        {/each}
      </div>

      <!-- Analysis progress -->
      {#if totalFiles > 0}
        <div class="hairline flex flex-wrap items-center gap-5 rounded-2xl bg-surface/40 px-5 py-4">
          <div class="shrink-0">
            <ActivityRing {probed} total={totalFiles} size={80} />
          </div>
          <div class="min-w-0 flex-1">
            <div class="text-sm font-semibold">
              {probed.toLocaleString()}
              <span class="font-normal text-muted-foreground">/ {totalFiles.toLocaleString()} files analysed</span>
            </div>
            <div class="mt-2.5 h-1.5 w-full overflow-hidden rounded-full bg-foreground/10">
              <div
                class="h-full rounded-full bg-primary-glow transition-all duration-700"
                style="width: {totalFiles > 0 ? Math.round((probed / totalFiles) * 100) : 0}%"
              ></div>
            </div>
            <div class="mt-1.5 text-[11px] text-muted-foreground">
              {totalFiles > 0 ? Math.round((probed / totalFiles) * 100) : 0}% complete
              {#if unprobed > 0}· <span class="text-amber-400">{unprobed.toLocaleString()} remaining</span>{/if}
            </div>
          </div>
          {#if health?.unsupported && health.unsupported > 0}
            <div class="shrink-0 rounded-xl border border-border bg-foreground/[0.04] px-3 py-2 text-center">
              <div class="text-lg font-bold leading-none tabular-nums">{health.unsupported}</div>
              <div class="mt-1 text-[10px] text-muted-foreground">Unsupported</div>
            </div>
          {/if}
          {#if health?.highBitrate && health.highBitrate > 0}
            <div class="shrink-0 rounded-xl border border-border bg-foreground/[0.04] px-3 py-2 text-center">
              <div class="text-lg font-bold leading-none tabular-nums">{health.highBitrate}</div>
              <div class="mt-1 text-[10px] text-muted-foreground">High Bitrate</div>
            </div>
          {/if}
          {#if unprobed > 0 && !probeRunning}
            <button
              type="button"
              onclick={handleProbeNow}
              disabled={probeBusy}
              class="shrink-0 rounded-full bg-amber-400/15 px-4 py-2 text-xs font-semibold text-amber-300 transition-colors hover:bg-amber-400/25 disabled:opacity-50"
            >
              {probeBusy ? 'Starting…' : 'Probe now'}
            </button>
          {/if}
        </div>
      {/if}

      <!-- Metadata backfill detail (shown when running) -->
      {#if jobs?.metadata?.backfill?.running}
        {@const bf = jobs.metadata.backfill}
        <div class="hairline mt-3 flex items-center gap-4 rounded-xl border-primary/20 bg-primary-glow/[0.06] px-4 py-3">
          <span class="inline-block h-1.5 w-1.5 shrink-0 animate-pulse rounded-full bg-primary-glow"></span>
          <div class="min-w-0 flex-1 text-xs">
            <span class="font-medium text-foreground">Metadata backfill running</span>
            {#if bf.lastTitle}
              <span class="ml-2 truncate text-muted-foreground">— {bf.lastTitle}</span>
            {/if}
          </div>
          <div class="shrink-0 text-right text-[11px] text-muted-foreground tabular-nums">
            {(bf.refreshed + bf.failed).toLocaleString()} / {bf.total.toLocaleString()}
          </div>
        </div>
      {/if}
    </section>

    <!-- ── PLAYABILITY AUDIT ────────────────────────────────────────────────── -->
    <section class="mb-10">
      <div class="mb-4 flex items-center justify-between gap-3">
        <h2 class="text-[10px] font-semibold uppercase tracking-[0.3em] text-muted-foreground">Playability Audit</h2>
        <button
          type="button"
          onclick={loadAudit}
          disabled={auditLoading}
          class="inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.06] px-3 py-1 text-[10px] font-semibold uppercase tracking-[0.15em] text-muted-foreground transition hover:bg-foreground/[0.10] hover:text-foreground disabled:opacity-50"
        >
          {#if auditLoading}
            <span class="h-2 w-2 animate-spin rounded-full border border-muted-foreground/40 border-t-muted-foreground"></span>
            Analysing…
          {:else}
            Run Audit
          {/if}
        </button>
      </div>

      {#if !audit && !auditLoading}
        <div class="hairline flex flex-col items-center justify-center gap-3 rounded-2xl bg-surface/40 py-10 text-center">
          <svg class="h-8 w-8 text-muted-foreground/20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 0 0 2.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 0 0-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75 2.25 2.25 0 0 0-.1-.664m-5.8 0A2.251 2.251 0 0 1 13.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25Z" />
          </svg>
          <p class="text-sm text-muted-foreground">Click <span class="font-medium text-foreground">Run Audit</span> to analyse playback compatibility across device profiles</p>
        </div>
      {:else if audit}
        {@const profiles = [
          { id: 'web',         label: 'Web Browser' },
          { id: 'apple-tv',    label: 'Apple TV' },
          { id: 'android-tv',  label: 'Android TV' },
        ]}
        <div class="grid gap-4 md:grid-cols-2">
          <!-- Profile breakdown cards -->
          <div class="hairline rounded-2xl bg-surface/40 p-5">
            <div class="mb-4 text-[10px] font-semibold uppercase tracking-[0.2em] text-muted-foreground">By Device Profile</div>
            <div class="space-y-5">
              {#each profiles as { id, label }}
                {@const bd = audit.byProfile[id]}
                {#if bd && bd.total > 0}
                  {@const pct = (n: number) => Math.round((n / bd.total) * 100)}
                  <div>
                    <div class="mb-1.5 flex items-center justify-between text-xs">
                      <span class="font-medium">{label}</span>
                      <span class="tabular-nums text-muted-foreground">{bd.total.toLocaleString()} files</span>
                    </div>
                    <div class="flex h-3 w-full overflow-hidden rounded-full">
                      {#if bd.directPlay > 0}
                        <div class="h-full" style="width:{pct(bd.directPlay)}%; background: oklch(0.78 0.22 145);" title="Direct play: {pct(bd.directPlay)}%"></div>
                      {/if}
                      {#if bd.remux > 0}
                        <div class="h-full" style="width:{pct(bd.remux)}%; background: oklch(0.62 0.22 285);" title="Remux: {pct(bd.remux)}%"></div>
                      {/if}
                      {#if bd.audioTranscode > 0}
                        <div class="h-full" style="width:{pct(bd.audioTranscode)}%; background: oklch(0.85 0.22 75);" title="Audio transcode: {pct(bd.audioTranscode)}%"></div>
                      {/if}
                      {#if bd.videoTranscode > 0}
                        <div class="h-full" style="width:{pct(bd.videoTranscode)}%; background: oklch(0.68 0.26 22);" title="Video transcode: {pct(bd.videoTranscode)}%"></div>
                      {/if}
                    </div>
                    <div class="mt-1.5 flex flex-wrap gap-x-3 gap-y-1 text-[10px] tabular-nums text-muted-foreground">
                      <span style="color: oklch(0.78 0.22 145)">Direct {pct(bd.directPlay)}%</span>
                      {#if bd.remux > 0}<span style="color: oklch(0.72 0.20 285)">Remux {pct(bd.remux)}%</span>{/if}
                      {#if bd.audioTranscode > 0}<span style="color: oklch(0.85 0.22 75)">Audio {pct(bd.audioTranscode)}%</span>{/if}
                      {#if bd.videoTranscode > 0}<span style="color: oklch(0.68 0.26 22)">Video {pct(bd.videoTranscode)}%</span>{/if}
                    </div>
                  </div>
                {/if}
              {/each}
            </div>
          </div>

          <!-- Top reason codes -->
          {#if audit.topReasons && audit.topReasons.length > 0}
            <div class="hairline rounded-2xl bg-surface/40 p-5">
              <div class="mb-4 text-[10px] font-semibold uppercase tracking-[0.2em] text-muted-foreground">Top Transcode Triggers</div>
              <div class="space-y-2.5">
                {#each audit.topReasons.slice(0, 8) as reason}
                  <div class="flex items-center gap-3">
                    <div class="min-w-0 flex-1">
                      <div class="truncate text-xs font-medium">{reason.reasonText || reason.reasonCode}</div>
                      <div class="text-[10px] text-muted-foreground capitalize">{reason.profile}</div>
                    </div>
                    <div class="shrink-0 rounded-full bg-foreground/[0.06] px-2 py-0.5 text-[10px] font-semibold tabular-nums">
                      {reason.count.toLocaleString()}
                    </div>
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        </div>
      {/if}
    </section>

    <!-- ── AUTOMATION JOBS ──────────────────────────────────────────────────── -->
    <section class="mb-10">
      <h2 class="mb-4 text-[10px] font-semibold uppercase tracking-[0.3em] text-muted-foreground">Automation Jobs</h2>
      <div class="grid gap-4 sm:grid-cols-3">
        <JobCard
          name="Library Scan"
          emoji="📁"
          status={jobs?.scan?.status ?? 'idle'}
          lastRunAt={jobs?.scan?.lastRunAt}
          nextRunAt={jobs?.scan?.nextRunAt}
          intervalMin={jobs?.scan?.intervalMins}
          error={jobs?.scan?.lastRunError || null}
          canRunNow={true}
          canStop={false}
          runBusy={scanBusy}
          onRunNow={handleScanNow}
        />
        <JobCard
          name="Metadata"
          emoji="🎬"
          status={jobs?.metadata?.status ?? 'idle'}
          lastRunAt={jobs?.metadata?.lastRunAt}
          nextRunAt={jobs?.metadata?.nextRunAt}
          intervalMin={jobs?.metadata?.intervalMins}
          error={jobs?.metadata?.lastRunError || null}
          progress={metaProgress?.pct ?? null}
          progressDetail={metaProgress?.detail ?? null}
          canRunNow={true}
          canStop={true}
          runBusy={metaBusy}
          onRunNow={handleMetaNow}
          onStop={handleMetaStop}
        />
        <JobCard
          name="Probe"
          emoji="🔬"
          status={jobs?.probe?.status ?? 'idle'}
          lastRunAt={jobs?.probe?.lastRunAt}
          nextRunAt={jobs?.probe?.nextRunAt}
          error={jobs?.probe?.lastRunError || null}
          progress={probeProgress?.pct ?? null}
          progressDetail={probeProgress?.detail ?? null}
          canRunNow={true}
          canStop={false}
          runBusy={probeBusy}
          onRunNow={handleProbeNow}
        />
      </div>
    </section>

    <!-- ── NOW PLAYING + WORKER QUEUES ────────────────────────────────────── -->
    <div class="grid gap-6 lg:grid-cols-[1fr_360px]">

      <!-- Sessions / Now Playing -->
      <section>
        <div class="mb-4 flex items-center gap-3">
          <h2 class="text-[10px] font-semibold uppercase tracking-[0.3em] text-muted-foreground">Now Playing</h2>
          {#if sessions.length > 0}
            <span class="rounded-full bg-primary-glow/15 px-2 py-0.5 text-[9px] font-semibold uppercase tracking-[0.15em] text-primary-glow">
              {sessions.length} active
            </span>
          {/if}
        </div>

        {#if sessions.length === 0}
          <div class="hairline flex flex-col items-center justify-center gap-3 rounded-2xl bg-surface/40 py-14 text-center">
            <svg class="h-10 w-10 text-muted-foreground/20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z" />
            </svg>
            <p class="text-sm text-muted-foreground">Nothing playing right now</p>
          </div>
        {:else}
          {#if sessions.some(s => s.state === 'playing')}
            <div class="hairline mb-3 flex items-center gap-3 rounded-xl border-amber-400/20 bg-amber-400/[0.06] px-4 py-2.5">
              <span class="h-1.5 w-1.5 shrink-0 animate-pulse rounded-full bg-amber-400"></span>
              <p class="text-xs text-amber-300">
                {sessions.length === 1 ? '1 active session' : `${sessions.length} active sessions`} —
                automation jobs pause during playback
              </p>
            </div>
          {/if}
          <div class="space-y-2">
            {#each sessions as session (session.id)}
              <NowPlayingCard {session} />
            {/each}
          </div>
        {/if}
      </section>

      <!-- Worker Queues -->
      <section>
        <h2 class="mb-4 text-[10px] font-semibold uppercase tracking-[0.3em] text-muted-foreground">Worker Queues</h2>
        <div class="hairline rounded-2xl bg-surface/40 p-5">
          {#if perf?.queues && perf.queues.length > 0}
            <div class="space-y-5">
              {#each perf.queues as q (q.name)}
                {@const active  = q.active  ?? 0}
                {@const workers = q.workers ?? 0}
                {@const queued  = q.queued  ?? 0}
                {@const util    = workers > 0 ? Math.round((active / workers) * 100) : 0}
                <div>
                  <div class="mb-2 flex items-center justify-between gap-2">
                    <div class="flex items-center gap-2">
                      <span class="text-sm font-medium capitalize">{q.name}</span>
                      {#if active > 0}
                        <span class="rounded-full bg-primary-glow/15 px-1.5 py-0.5 text-[9px] font-bold text-primary-glow">{active}</span>
                      {/if}
                    </div>
                    <span class="text-[11px] text-muted-foreground tabular-nums">
                      {active} / {workers} workers
                    </span>
                  </div>
                  <div class="h-2 overflow-hidden rounded-full bg-foreground/10">
                    <div
                      class="h-full rounded-full transition-all duration-700 {util >= 100 ? 'bg-amber-400' : 'bg-primary-glow'}"
                      style="width: {util}%"
                    ></div>
                  </div>
                  {#if queued > 0}
                    <p class="mt-1.5 text-[11px] text-muted-foreground tabular-nums">
                      {queued} job{queued !== 1 ? 's' : ''} queued
                    </p>
                  {/if}
                </div>
              {/each}
            </div>
          {:else}
            <!-- Fallback: show limits from config -->
            {#if perf?.limits}
              <div class="space-y-5">
                {#each [
                  { name: 'Scan',       workers: perf.limits.scanWorkers      ?? 0 },
                  { name: 'Probe',      workers: perf.limits.probeWorkers     ?? 0 },
                  { name: 'Transcode',  workers: perf.limits.transcodeWorkers ?? 0 },
                  { name: 'GPU',        workers: perf.limits.gpuWorkers       ?? 0 },
                ].filter(q => q.workers > 0) as q (q.name)}
                  <div>
                    <div class="mb-2 flex items-center justify-between">
                      <span class="text-sm font-medium">{q.name}</span>
                      <span class="text-[11px] text-muted-foreground tabular-nums">{q.workers} workers</span>
                    </div>
                    <div class="h-2 overflow-hidden rounded-full bg-foreground/10">
                      <div class="h-full w-0 rounded-full bg-primary-glow"></div>
                    </div>
                  </div>
                {/each}
              </div>
            {:else}
              <div class="flex flex-col items-center gap-2 py-8 text-center">
                <p class="text-sm text-muted-foreground">No queue data available</p>
              </div>
            {/if}
          {/if}

          <!-- Hardware acceleration badge -->
          {#if perf?.hardwareAcceleration}
            {@const hw = perf.hardwareAcceleration}
            <div class="mt-5 border-t border-border pt-4">
              <div class="flex items-center justify-between text-[11px]">
                <span class="text-muted-foreground">Hardware acceleration</span>
                <span class="font-medium {hw.available ? 'text-green-400' : 'text-muted-foreground'}">
                  {hw.available ? hw.status ?? 'Available' : 'Not available'}
                </span>
              </div>
              {#if hw.available && hw.selectedEncoder}
                <div class="mt-2 flex items-center justify-between text-[11px]">
                  <span class="text-muted-foreground">Selected encoder</span>
                  <span class="font-medium text-amber-300">{hw.selectedEncoder.label}</span>
                </div>
              {/if}
              {#if hw.available && hw.encoders && hw.encoders.length > 0}
                {@const passingIds = new Set((hw.lastTest?.tests ?? []).filter(t => t.ok).map(t => t.id ?? ''))}
                {@const hasTests = (hw.lastTest?.tests?.length ?? 0) > 0}
                <div class="mt-2 flex flex-wrap gap-1.5">
                  {#each hw.encoders as enc (enc.id)}
                    {@const passing = !hasTests || passingIds.has(enc.id ?? '')}
                    <span class="rounded-full border px-2 py-0.5 text-[10px] {passing
                      ? 'border-border bg-foreground/[0.04] text-muted-foreground'
                      : 'border-border/40 bg-foreground/[0.02] text-muted-foreground/30'}">
                      {enc.label ?? enc.codec}{!passing ? ' · N/A' : ''}
                    </span>
                  {/each}
                </div>
              {/if}
            </div>
          {/if}
        </div>
      </section>
    </div>

  </main>
</div>
