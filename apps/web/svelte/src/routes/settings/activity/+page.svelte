<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { goto } from '$app/navigation';
  import { getAuthSession } from '$lib/api/auth';
  import {
    getJobs,
    getCatalogHealth,
    getSessions,
    startProbeJob,
    scanAllLibraries,
    type JobsStatusResponse,
    type SessionItem,
  } from '$lib/api/operator';
  import { startBackfill, stopBackfill } from '$lib/api/browse';
  import { createEventStream } from '$lib/events/stream';
  import JobCard from '$lib/components/JobCard.svelte';
  import NowPlayingCard from '$lib/components/NowPlayingCard.svelte';
  import ActivityRing from '$lib/components/ActivityRing.svelte';

  // ── State ──────────────────────────────────────────────────────────────────
  let jobs       = $state<JobsStatusResponse | null>(null);
  let sessions   = $state<SessionItem[]>([]);
  let unprobed   = $state(0);
  let totalFiles = $state(0);
  let loading    = $state(true);
  let error      = $state<string | null>(null);

  // Per-job busy flags (while a manual trigger request is in-flight)
  let scanBusy     = $state(false);
  let metaBusy     = $state(false);
  let probeBusy    = $state(false);
  let runAllBusy   = $state(false);

  // Probe stop flag
  let probeRunning = $derived(
    (jobs?.probe?.status === 'running') ||
    (jobs?.probe?.activeJobs?.some(j => j.status === 'running') ?? false)
  );

  // ── Probe progress derived from active job ────────────────────────────────
  const activeProbeJob = $derived(
    jobs?.probe?.activeJobs?.find(j => j.status === 'running') ?? null
  );
  const probeProgress = $derived(() => {
    if (!activeProbeJob) return null;
    const total = activeProbeJob.total ?? 0;
    const done  = activeProbeJob.completed ?? 0;
    if (total === 0) return null;
    return {
      pct:    Math.round((done / total) * 100),
      detail: `${done.toLocaleString()} / ${total.toLocaleString()}`,
    };
  });

  // ── Metadata backfill progress ─────────────────────────────────────────────
  const metaBackfill = $derived(jobs?.metadata?.backfill ?? null);
  const metaProgress = $derived(() => {
    const bf = metaBackfill;
    if (!bf?.running || bf.total === 0) return null;
    const done = bf.refreshed + bf.failed;
    return {
      pct:    Math.round((done / bf.total) * 100),
      detail: `${done.toLocaleString()} / ${bf.total.toLocaleString()}`,
    };
  });

  // ── Data loading ──────────────────────────────────────────────────────────
  async function loadAll() {
    try {
      const [jobsResp, healthResp, sessionsResp] = await Promise.allSettled([
        getJobs(),
        getCatalogHealth(),
        getSessions(),
      ]);
      if (jobsResp.status === 'fulfilled') jobs = jobsResp.value;
      if (healthResp.status === 'fulfilled') {
        unprobed   = healthResp.value.unprobed ?? 0;
        totalFiles = healthResp.value.summary?.mediaSources ?? 0;
      }
      if (sessionsResp.status === 'fulfilled') {
        sessions = sessionsResp.value.sessions ?? [];
      }
      error = null;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Could not load activity data';
    } finally {
      loading = false;
    }
  }

  // Refresh only jobs + sessions (skip catalog health which is expensive)
  async function refreshJobs() {
    try {
      const [jobsResp, sessionsResp] = await Promise.allSettled([
        getJobs(),
        getSessions(),
      ]);
      if (jobsResp.status === 'fulfilled')    jobs = jobsResp.value;
      if (sessionsResp.status === 'fulfilled') sessions = sessionsResp.value.sessions ?? [];
    } catch { /* silent */ }
  }

  // ── Job action handlers ───────────────────────────────────────────────────
  async function handleScanNow() {
    scanBusy = true;
    try {
      await scanAllLibraries();
      await refreshJobs();
    } catch { /* ignore */ } finally {
      scanBusy = false;
    }
  }

  async function handleMetaNow() {
    metaBusy = true;
    try {
      await startBackfill('tmdb');
      await refreshJobs();
    } catch { /* ignore */ } finally {
      metaBusy = false;
    }
  }

  async function handleMetaStop() {
    metaBusy = true;
    try {
      await stopBackfill();
      await refreshJobs();
    } catch { /* ignore */ } finally {
      metaBusy = false;
    }
  }

  async function handleProbeNow() {
    probeBusy = true;
    try {
      await startProbeJob(0); // 0 = probe all unprobed
      await refreshJobs();
    } catch { /* ignore */ } finally {
      probeBusy = false;
    }
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
    } catch { /* ignore */ } finally {
      runAllBusy = false;
    }
  }

  // ── SSE live updates ──────────────────────────────────────────────────────
  const stream = createEventStream();

  $effect(() => {
    stream.subscribeAny(({ type }) => {
      if (
        type.startsWith('automation.') ||
        type.startsWith('probe.')      ||
        type.startsWith('metadata.backfill') ||
        type.startsWith('scan.')       ||
        type === 'api.session.accepted' ||
        type === 'api.session.stopped'
      ) {
        refreshJobs();
      }
    });
  });

  // Polling while any job is running (5 s cadence)
  let pollInterval: ReturnType<typeof setInterval> | null = null;

  $effect(() => {
    const anyRunning =
      jobs?.scan?.status === 'running' ||
      jobs?.metadata?.status === 'running' ||
      jobs?.probe?.status === 'running' ||
      probeRunning;

    if (anyRunning) {
      if (!pollInterval) {
        pollInterval = setInterval(refreshJobs, 5000);
      }
    } else {
      if (pollInterval) {
        clearInterval(pollInterval);
        pollInterval = null;
      }
    }
    return () => {
      if (pollInterval) { clearInterval(pollInterval); pollInterval = null; }
    };
  });

  // ── Lifecycle ─────────────────────────────────────────────────────────────
  onMount(async () => {
    // Admin guard
    try {
      const sess = await getAuthSession();
      if (!sess.authDisabled && sess.user?.role !== 'admin') {
        goto('/settings');
        return;
      }
    } catch {
      goto('/settings');
      return;
    }

    stream.connect();
    await loadAll();
  });

  onDestroy(() => {
    stream.disconnect();
    if (pollInterval) { clearInterval(pollInterval); pollInterval = null; }
  });

  // ── Display helpers ───────────────────────────────────────────────────────
  const probed = $derived(Math.max(0, totalFiles - unprobed));
</script>

<svelte:head>
  <title>Activity & Jobs — Xuva</title>
</svelte:head>

<div class="min-h-screen bg-background">
  <div class="mx-auto max-w-5xl px-6 py-10">

    <!-- Page header -->
    <div class="mb-8 flex flex-col gap-5 sm:flex-row sm:items-start sm:justify-between">
      <div>
        <a
          href="/settings"
          class="mb-3 inline-flex items-center gap-1.5 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
        >
          <svg class="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
          </svg>
          Settings
        </a>
        <h1 class="font-serif-display text-3xl tracking-tight">Activity &amp; Jobs</h1>
        <p class="mt-1.5 text-sm text-muted-foreground">
          Library automation runs in the background. Monitor progress or trigger a manual run here.
        </p>
      </div>

      <button
        type="button"
        onclick={handleRunAll}
        disabled={runAllBusy}
        class="inline-flex shrink-0 items-center gap-2 rounded-full bg-gradient-primary px-6 py-2.5 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110 disabled:opacity-60 sm:self-start"
      >
        {#if runAllBusy}
          <span class="h-3.5 w-3.5 animate-spin rounded-full border border-white/40 border-t-white"></span>
        {:else}
          <svg class="h-3.5 w-3.5" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
            <path d="M8 5v14l11-7z"/>
          </svg>
        {/if}
        Run All Jobs
      </button>
    </div>

    {#if loading}
      <!-- Loading state -->
      <div class="flex min-h-[40vh] items-center justify-center">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div>
      </div>

    {:else if error}
      <!-- Error state -->
      <div class="hairline flex items-center gap-4 rounded-2xl bg-red-500/10 border-red-400/20 px-6 py-5">
        <svg class="h-5 w-5 shrink-0 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
        </svg>
        <p class="text-sm text-red-300">{error}</p>
        <button
          type="button"
          onclick={loadAll}
          class="ml-auto text-xs font-medium text-red-300 underline hover:no-underline"
        >Retry</button>
      </div>

    {:else}

      <!-- ── Unprobed files warning ─────────────────────────────────────────── -->
      {#if unprobed > 0}
        <div class="hairline mb-8 flex items-center gap-4 rounded-2xl border-amber-400/25 bg-amber-500/8 px-5 py-4">
          <svg class="h-5 w-5 shrink-0 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
          </svg>
          <div class="min-w-0 flex-1">
            <span class="font-semibold text-amber-300">{unprobed.toLocaleString()} file{unprobed === 1 ? '' : 's'} awaiting analysis</span>
            <span class="ml-1.5 text-sm text-amber-300/70">— unanalysed files cannot play.</span>
          </div>
          {#if !probeRunning}
            <button
              type="button"
              onclick={handleProbeNow}
              disabled={probeBusy}
              class="shrink-0 rounded-full bg-amber-400/15 px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.15em] text-amber-300 transition-colors hover:bg-amber-400/25 disabled:opacity-50"
            >
              {probeBusy ? 'Starting…' : 'Probe now'}
            </button>
          {/if}
        </div>
      {/if}

      <!-- ── Job cards ──────────────────────────────────────────────────────── -->
      <div class="mb-8 grid gap-4 sm:grid-cols-3">

        <!-- Scan -->
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

        <!-- Metadata -->
        <JobCard
          name="Metadata"
          emoji="🎬"
          status={jobs?.metadata?.status ?? 'idle'}
          lastRunAt={jobs?.metadata?.lastRunAt}
          nextRunAt={jobs?.metadata?.nextRunAt}
          intervalMin={jobs?.metadata?.intervalMins}
          error={jobs?.metadata?.lastRunError || null}
          progress={metaProgress()?.pct ?? null}
          progressDetail={metaProgress()?.detail ?? null}
          canRunNow={true}
          canStop={true}
          runBusy={metaBusy}
          onRunNow={handleMetaNow}
          onStop={handleMetaStop}
        />

        <!-- Probe -->
        <JobCard
          name="Probe"
          emoji="🔬"
          status={jobs?.probe?.status ?? 'idle'}
          lastRunAt={jobs?.probe?.lastRunAt}
          nextRunAt={jobs?.probe?.nextRunAt}
          error={jobs?.probe?.lastRunError || null}
          progress={probeProgress()?.pct ?? null}
          progressDetail={probeProgress()?.detail ?? null}
          canRunNow={true}
          canStop={false}
          runBusy={probeBusy}
          onRunNow={handleProbeNow}
        />
      </div>

      <!-- ── Library analysis ring ──────────────────────────────────────────── -->
      {#if totalFiles > 0}
        <div class="hairline mb-8 flex items-center gap-6 rounded-2xl bg-surface/40 p-6">
          <div class="shrink-0">
            <ActivityRing {probed} total={totalFiles} size={96} />
          </div>
          <div class="min-w-0 flex-1">
            <div class="font-serif-display text-2xl tracking-tight">
              {probed.toLocaleString()}
              <span class="text-base text-muted-foreground">/ {totalFiles.toLocaleString()}</span>
            </div>
            <div class="mt-0.5 text-sm text-muted-foreground">files analysed</div>
            {#if totalFiles > 0}
              <div class="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-foreground/10">
                <div
                  class="h-full rounded-full bg-primary-glow transition-all duration-700"
                  style="width: {Math.round((probed / totalFiles) * 100)}%"
                ></div>
              </div>
              <div class="mt-1 text-[11px] text-muted-foreground">
                {Math.round((probed / totalFiles) * 100)}% complete
                {#if unprobed > 0}
                  · {unprobed.toLocaleString()} remaining
                {/if}
              </div>
            {/if}
          </div>
        </div>
      {/if}

      <!-- ── Now Playing ─────────────────────────────────────────────────────── -->
      {#if sessions.length > 0}
        <div>
          <div class="mb-3 flex items-center gap-3">
            <h2 class="font-serif-display text-xl tracking-tight">Now Playing</h2>
            <span class="rounded-full bg-primary-glow/15 px-2.5 py-0.5 text-[10px] font-semibold uppercase tracking-[0.18em] text-primary-glow">
              {sessions.length} active
            </span>
          </div>

          <div class="hairline mb-4 flex items-center gap-3 rounded-xl border-amber-400/20 bg-amber-500/8 px-4 py-3">
            <svg class="h-4 w-4 shrink-0 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25v13.5m-7.5-13.5v13.5" />
            </svg>
            <p class="text-xs text-amber-300">
              {sessions.length === 1 ? '1 active session' : `${sessions.length} active sessions`} —
              automation jobs pause while playback is detected.
            </p>
          </div>

          <div class="space-y-2">
            {#each sessions as session (session.id)}
              <NowPlayingCard {session} />
            {/each}
          </div>
        </div>
      {/if}

    {/if}
  </div>
</div>
