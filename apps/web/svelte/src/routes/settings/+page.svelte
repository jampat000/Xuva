<script lang="ts">
  import {
    ArrowRightLeft,
    BookMarked,
    Check,
    ChevronLeft,
    ChevronRight,
    Database,
    Film,
    Folder,
    FolderOpen,
    HardDrive,
    Info,
    KeyRound,
    LayoutDashboard,
    Library,
    Link,
    Link2,
    LogOut,
    Play,
    Plus,
    RefreshCw,
    ScanSearch,
    Search,
    Settings2,
    ShieldCheck,
    Sliders,
    Trash2,
    Tv,
    Unlink,
    User,
    Users,
    Wifi
  } from "lucide-svelte";
  import { onMount } from 'svelte';
  import Header from "$lib/components/Header.svelte";
  import {
    getSystemStatus,
    getCatalogSummary,
    getScans,
    getSessions,
    getSettings,
    getLibraries,
    saveLibrary,
    deleteLibrary,
    scanLibrary,
    browseFolders,
    type SystemStatusResponse,
    type CatalogSummaryResponse,
    type ScanJobItem,
    type SessionItem,
    type SettingsResponse,
    type LibraryItem,
    type FolderEntry
  } from '$lib/api/operator';

  type Group = "Account" | "Server" | "Devices" | "Advanced";
  type Mode = "balanced" | "automatic" | "advanced";

  type Section = {
    id: string;
    label: string;
    icon: typeof User;
    group: Group;
    hint?: string;
  };

  const sections: Section[] = [
    { id: "account", label: "My Account", icon: User, group: "Account", hint: "Profile, password & sign-out" },
    { id: "dashboard", label: "Dashboard", icon: LayoutDashboard, group: "Server", hint: "Overview & health" },
    { id: "general", label: "General", icon: Settings2, group: "Server", hint: "Server name, locale, theme" },
    { id: "libraries", label: "Libraries", icon: Library, group: "Server", hint: "Folders, library kinds, poster art" },
    { id: "scanning", label: "Scanning", icon: ScanSearch, group: "Server", hint: "Sync schedule & scan rules" },
    { id: "metadata", label: "Metadata", icon: Database, group: "Server", hint: "Providers, keys, matches" },
    { id: "playback", label: "Playback", icon: Play, group: "Server", hint: "Streaming & quality defaults" },
    { id: "transcoding", label: "Transcoding", icon: Sliders, group: "Server", hint: "Hardware acceleration & workers" },
    { id: "storage", label: "Storage", icon: HardDrive, group: "Server", hint: "Directories & disk usage" },
    { id: "network", label: "Network", icon: Wifi, group: "Server", hint: "Ports, mDNS discovery, remote access" },
    { id: "migration", label: "Migration", icon: ArrowRightLeft, group: "Server", hint: "Import from Plex, Emby & more" },
    { id: "watchlist-services", label: "Watchlist Services", icon: BookMarked, group: "Server", hint: "Sync with Trakt, Letterboxd & more" },
    { id: "users", label: "Users", icon: Users, group: "Server", hint: "Accounts & access roles" },
    { id: "pending-approvals", label: "Pending Approvals", icon: Link2, group: "Devices", hint: "Review & approve device requests" },
    { id: "approved-devices", label: "Approved Devices", icon: ShieldCheck, group: "Devices", hint: "Trusted & active devices" },
    { id: "about", label: "About", icon: Info, group: "Advanced", hint: "Build info & open-source licenses" }
  ];

  const modes: { id: Mode; title: string; desc: string }[] = [
    {
      id: "balanced",
      title: "Balanced",
      desc: "Xuva uses its recommended metadata and artwork fallbacks for the best overall coverage."
    },
    {
      id: "automatic",
      title: "Prefer local artwork",
      desc: "Keep metadata automatic but prefer artwork stored alongside your media before downloaded artwork."
    },
    {
      id: "advanced",
      title: "Advanced provider settings",
      desc: "Choose exact providers, order, and keys when you need full control."
    }
  ];

  const providers = [
    { id: "tmdb", name: "TMDB", blurb: "Movie, show, season, episode metadata and artwork" },
    { id: "tvdb", name: "TheTVDB", blurb: "TV and movie metadata, IDs, and ratings" },
    { id: "fanart", name: "Fanart.tv", blurb: "Logos, banners, thumbs, and extra artwork" }
  ];

  const groups: Group[] = ["Account", "Server", "Devices", "Advanced"];

  let active = $state("dashboard");
  let mode = $state<Mode>("advanced");
  let savedMode = $state<Mode>("advanced");
  let keys = $state<Record<string, string>>({});
  let savedKeys = $state<Record<string, string>>({});
  let q = $state("");
  let headerScrolled = $state(false);
  let mainRef = $state<HTMLDivElement | null>(null);
  let dirty = $derived(mode !== savedMode || JSON.stringify(keys) !== JSON.stringify(savedKeys));

  // Watchlist Services state (#11)
  interface WatchlistService {
    id: string;
    name: string;
    logo: string;
    description: string;
    authType: 'oauth' | 'apikey';
    docsUrl: string;
  }
  const watchlistServices: WatchlistService[] = [
    {
      id: 'trakt',
      name: 'Trakt',
      logo: 'T',
      description: 'Sync watch history, ratings, and watchlists with Trakt. Progress syncs automatically while you watch.',
      authType: 'oauth',
      docsUrl: 'https://trakt.tv/oauth/applications'
    },
    {
      id: 'letterboxd',
      name: 'Letterboxd',
      logo: 'L',
      description: 'Import your Letterboxd diary and watchlist. Mark films watched and keep ratings in sync.',
      authType: 'apikey',
      docsUrl: 'https://letterboxd.com/api-beta'
    },
    {
      id: 'simkl',
      name: 'Simkl',
      logo: 'S',
      description: 'Track what you watch across movies and TV on Simkl with automated scrobbling.',
      authType: 'oauth',
      docsUrl: 'https://simkl.com/apps/manage'
    }
  ];
  let wlConnected = $state<Record<string, boolean>>({});
  let wlKeys = $state<Record<string, string>>({});
  let wlConnecting = $state<Record<string, boolean>>({});

  let current = $derived(sections.find((section) => section.id === active) ?? sections[0]);

  // ─── Dashboard live data ───────────────────────────────────────────────────
  let settingsData = $state<SettingsResponse | null>(null);
  let sysStatus = $state<SystemStatusResponse | null>(null);
  let catalogSummary = $state<CatalogSummaryResponse | null>(null);
  let recentScans = $state<ScanJobItem[]>([]);
  let activeSessions = $state<SessionItem[]>([]);
  let dashLoading = $state(false);
  let dashError = $state<string | null>(null);

  // ─── Library CRUD state ────────────────────────────────────────────────────
  let libraries = $state<LibraryItem[]>([]);
  let libLoading = $state(false);
  let libSaving = $state(false);
  let libScanningId = $state<string | null>(null);
  let libDeletingId = $state<string | null>(null);
  let libError = $state<string | null>(null);

  // Add-library form
  let showAddLib = $state(false);
  let newLibName = $state('');
  let newLibKind = $state<'movies' | 'tv'>('movies');
  let newLibPath = $state('');
  let newLibStorageType = $state<'local' | 'network' | 'removable' | 'mounted'>('local');

  // Folder browser
  let showFolderBrowser = $state(false);
  let browserCurrentPath = $state('');
  let browserParentPath = $state<string | undefined>(undefined);
  let browserEntries = $state<FolderEntry[]>([]);
  let browserLoading = $state(false);

  // ─── Derived: server name for h1 ──────────────────────────────────────────
  const serverNameParts = $derived.by(() => {
    const name = settingsData?.config?.serverName ?? '';
    if (!name) return { first: 'Media', last: 'Server' };
    const parts = name.split(' ');
    if (parts.length === 1) return { first: '', last: parts[0] };
    return { first: parts.slice(0, -1).join(' '), last: parts[parts.length - 1] };
  });

  // ─── Derived: health badge ─────────────────────────────────────────────────
  const healthStatus = $derived(
    dashError
      ? 'Offline'
      : sysStatus
        ? ((sysStatus.cpu?.percent ?? 0) > 85 || (sysStatus.memory?.usedPercent ?? 0) > 90
            ? 'Warning'
            : 'Healthy')
        : dashLoading
          ? '…'
          : '—'
  );
  const healthAccent = $derived(
    healthStatus === 'Healthy' ? 'text-emerald-300' :
    healthStatus === 'Warning' ? 'text-amber-300' :
    healthStatus === 'Offline' ? 'text-red-300' :
    'text-muted-foreground'
  );
  const healthDotAccent = $derived(
    healthStatus === 'Healthy' ? 'bg-emerald-400' :
    healthStatus === 'Warning' ? 'bg-amber-400' :
    healthStatus === 'Offline' ? 'bg-red-400' :
    'bg-muted-foreground/30'
  );
  const updatedAtStr = $derived(
    sysStatus?.collectedAt
      ? new Date(sysStatus.collectedAt).toLocaleTimeString(undefined, {
          hour: '2-digit', minute: '2-digit', second: '2-digit'
        })
      : null
  );

  function discard(): void {
    mode = savedMode;
    keys = { ...savedKeys };
  }

  function saveChanges(): void {
    savedMode = mode;
    savedKeys = { ...keys };
  }

  function filtered(group: Group): Section[] {
    return sections
      .filter((section) => section.group === group)
      .filter((section) => !q || section.label.toLowerCase().includes(q.toLowerCase()));
  }

  function select(id: string): void {
    active = id;
    mainRef?.scrollTo({ top: 0, behavior: "smooth" });
  }

  $effect(() => {
    if (typeof window === "undefined") return;
    const onScroll = () => {
      headerScrolled = window.scrollY > 80;
    };
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  });

  // ─── Utility helpers ───────────────────────────────────────────────────────
  function formatBytes(bytes?: number): string {
    if (bytes === undefined || bytes === null) return '—';
    if (bytes >= 1_073_741_824) return `${(bytes / 1_073_741_824).toFixed(1)} GB`;
    if (bytes >= 1_048_576) return `${(bytes / 1_048_576).toFixed(0)} MB`;
    if (bytes >= 1_024) return `${(bytes / 1_024).toFixed(0)} KB`;
    return `${bytes} B`;
  }

  function formatBps(bps?: number): string {
    if (!bps) return '0 B/s';
    if (bps >= 1_000_000) return `${(bps / 1_000_000).toFixed(1)} MB/s`;
    if (bps >= 1_000) return `${(bps / 1_000).toFixed(0)} KB/s`;
    return `${bps} B/s`;
  }

  function computeParentPath(p: string): string | undefined {
    if (!p) return undefined;
    const hasBackslash = p.includes('\\');
    const sep = hasBackslash ? '\\' : '/';
    const trimmed = p.replace(new RegExp(`[${sep}]+$`), '');
    const idx = trimmed.lastIndexOf(sep);
    if (idx < 0) return undefined;
    const parent = trimmed.slice(0, idx) || sep;
    return parent === p ? undefined : parent;
  }

  // ─── Dashboard loading ─────────────────────────────────────────────────────
  async function loadDashboard() {
    dashLoading = true;
    dashError = null;
    try {
      const [statusRes, summaryRes, scansRes, sessionsRes, settingsRes] =
        await Promise.allSettled([
          getSystemStatus(),
          getCatalogSummary(),
          getScans(),
          getSessions(),
          getSettings()
        ]);
      if (statusRes.status === 'fulfilled') sysStatus = statusRes.value;
      if (summaryRes.status === 'fulfilled') catalogSummary = summaryRes.value;
      if (scansRes.status === 'fulfilled')
        recentScans = (scansRes.value.scans ?? []).slice(0, 6);
      if (sessionsRes.status === 'fulfilled')
        activeSessions = sessionsRes.value.sessions ?? [];
      if (settingsRes.status === 'fulfilled') {
        settingsData = settingsRes.value;
        if (libraries.length === 0)
          libraries = (settingsRes.value.libraries ?? []).map((l) => ({ ...l }));
      }
      if (statusRes.status === 'rejected' && summaryRes.status === 'rejected')
        dashError = 'Server unreachable';
    } finally {
      dashLoading = false;
    }
  }

  // ─── Library loading ───────────────────────────────────────────────────────
  async function loadLibraries() {
    libLoading = true;
    libError = null;
    try {
      const resp = await getLibraries();
      libraries = resp.libraries ?? [];
    } catch {
      if (settingsData?.libraries)
        libraries = [...(settingsData.libraries ?? [])];
    } finally {
      libLoading = false;
    }
  }

  async function handleSaveLibrary() {
    if (!newLibName.trim() || !newLibPath.trim()) return;
    libSaving = true;
    libError = null;
    try {
      const lib = await saveLibrary({
        name: newLibName.trim(),
        kind: newLibKind,
        path: newLibPath.trim(),
        storageType: newLibStorageType
      });
      libraries = [...libraries, lib];
      showAddLib = false;
      showFolderBrowser = false;
      newLibName = '';
      newLibPath = '';
      newLibKind = 'movies';
      newLibStorageType = 'local';
    } catch (e) {
      libError = e instanceof Error ? e.message : 'Failed to save library';
    } finally {
      libSaving = false;
    }
  }

  async function handleDeleteLibrary(id: string) {
    if (libDeletingId !== id) {
      libDeletingId = id;
      return;
    }
    try {
      await deleteLibrary(id);
      libraries = libraries.filter((l) => l.id !== id);
    } catch (e) {
      libError = e instanceof Error ? e.message : 'Failed to delete library';
    } finally {
      libDeletingId = null;
    }
  }

  async function handleScanLibrary(id: string) {
    libScanningId = id;
    try {
      await scanLibrary(id);
    } catch (e) {
      libError = e instanceof Error ? e.message : 'Failed to start scan';
    } finally {
      libScanningId = null;
    }
  }

  // ─── Folder browser ────────────────────────────────────────────────────────
  async function openFolderBrowser() {
    showFolderBrowser = true;
    await navigateFolder('');
  }

  async function navigateFolder(path: string) {
    browserLoading = true;
    try {
      const resp = await browseFolders(path || undefined);
      browserCurrentPath = resp.currentPath ?? path;
      browserParentPath = resp.parentPath ?? computeParentPath(browserCurrentPath);
      browserEntries = (resp.entries ?? []).filter((e) => e.isDir !== false);
    } catch {
      browserEntries = [];
    } finally {
      browserLoading = false;
    }
  }

  onMount(() => {
    loadDashboard();
    loadLibraries();
  });
</script>

<svelte:head>
  <title>Settings — Xuva</title>
  <meta
    name="description"
    content="Configure your Xuva media server — metadata, libraries, playback, devices, and advanced controls."
  />
  <meta property="og:title" content="Settings — Xuva" />
  <meta
    property="og:description"
    content="Tune how Xuva matches metadata, manages devices, and shapes the server experience."
  />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  <main class="px-6 pb-32 pt-24 md:px-12 md:pt-28 lg:px-20">
    <header class="relative mb-10">
      <div
        aria-hidden="true"
        class="pointer-events-none absolute -inset-x-6 -top-10 -z-10 h-[220px] opacity-60 md:-inset-x-12 lg:-inset-x-20"
        style="background: radial-gradient(50% 100% at 15% 0%, oklch(0.62 0.22 285 / 0.25), transparent 70%), radial-gradient(40% 100% at 90% 0%, oklch(0.72 0.16 255 / 0.18), transparent 70%);"
      ></div>
      <div class="flex flex-wrap items-end justify-between gap-6">
        <div>
          <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">
            Settings
          </div>
          <h1 class="font-serif-display text-[clamp(2rem,4vw,3.25rem)] leading-[1] tracking-tight">
            {#if serverNameParts.first}{serverNameParts.first}-{/if}<em>{serverNameParts.last}</em>
          </h1>
          <p class="mt-3 max-w-xl text-sm text-muted-foreground">
            {current.hint ?? "Configure your Xuva media server — libraries, metadata, playback, devices, and more."}
          </p>
        </div>
        <div class="flex items-center gap-3">
          <span class="hairline inline-flex items-center gap-2 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-[10px] font-semibold uppercase tracking-[0.25em] {healthAccent}">
            <span class="h-1.5 w-1.5 rounded-full {healthDotAccent} shadow-[0_0_10px_currentColor]"></span>
            {healthStatus}
          </span>
          {#if updatedAtStr}
            <div class="hidden text-right text-[11px] uppercase tracking-[0.22em] text-muted-foreground md:block">
              Updated <span class="text-foreground/90">{updatedAtStr}</span>
            </div>
          {/if}
        </div>
      </div>
    </header>

    <div class="grid gap-10 lg:grid-cols-[260px_minmax(0,1fr)] lg:gap-14">
      <aside class="lg:sticky lg:top-24 lg:self-start">
        <div class="relative mb-4">
          <Search class="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            bind:value={q}
            placeholder="Search settings..."
            autocomplete="off"
            class="h-10 w-full rounded-full border border-border bg-surface/60 pl-10 pr-4 text-sm outline-none placeholder:text-muted-foreground/70 focus:border-primary/60 focus:bg-surface"
          />
        </div>

        <nav class="scrollbar-none -mx-1 max-h-[calc(100vh-12rem)] space-y-6 overflow-y-auto px-1">
          {#each groups as group (group)}
            {@const items = filtered(group)}
            {#if items.length}
              <div>
                <div class="px-2 pb-2 text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground/70">
                  {group}
                </div>
                <ul class="space-y-0.5">
                  {#each items as section (section.id)}
                    <li>
                      <button
                        type="button"
                        onclick={() => select(section.id)}
                        class={`group/item relative flex w-full items-center gap-3 rounded-lg px-2.5 py-2 text-sm transition-all ${
                          section.id === active
                            ? "bg-foreground/[0.06] text-foreground"
                            : "text-muted-foreground hover:bg-foreground/[0.03] hover:text-foreground"
                        }`}
                      >
                        {#if section.id === active}
                          <span class="absolute inset-y-1.5 left-0 w-0.5 rounded-full bg-primary-glow shadow-glow"></span>
                        {/if}
                        <section.icon
                          class={`h-4 w-4 shrink-0 ${
                            section.id === active ? "text-primary-glow" : "text-muted-foreground/80"
                          }`}
                        />
                        <span class="truncate">{section.label}</span>
                      </button>
                    </li>
                  {/each}
                </ul>
              </div>
            {/if}
          {/each}
        </nav>
      </aside>

      <div bind:this={mainRef} class="min-w-0">
        <div
          class={`sticky top-16 z-20 -mx-6 mb-8 flex items-center justify-between gap-4 border-b px-6 py-4 backdrop-blur-xl transition-colors md:top-18 md:-mx-12 md:px-12 lg:-mx-0 lg:px-0 ${
            headerScrolled
              ? "border-border bg-background/80"
              : "border-transparent bg-transparent"
          }`}
        >
          <div class="min-w-0">
            <div class="text-[10px] font-semibold uppercase tracking-[0.3em] text-muted-foreground/80">
              {current.group}
            </div>
            <h2 class="font-serif-display mt-0.5 truncate text-2xl tracking-tight">
              {current.label}
            </h2>
            {#if current.hint}
              <p class="mt-0.5 truncate text-xs text-muted-foreground">
                {current.hint}
              </p>
            {/if}
          </div>
          {#if active === "metadata"}
            <div class="flex items-center gap-2">
              <button
                type="button"
                onclick={discard}
                disabled={!dirty}
                class="hairline rounded-full bg-foreground/[0.04] px-4 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40"
              >
                Discard
              </button>
              <button
                type="button"
                onclick={saveChanges}
                disabled={!dirty}
                class="rounded-full bg-gradient-primary px-5 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
              >
                Save changes
              </button>
            </div>
          {/if}
        </div>

        {#if active === "metadata"}
          <div class="space-y-12">
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Pick the level of control</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Most libraries should stay on Automatic. Switch to Advanced only when you need exact providers, ordering, and keys.
                </p>
              </div>
              <div class="grid gap-3 md:grid-cols-3">
                {#each modes as item (item.id)}
                  <button
                    type="button"
                    onclick={() => (mode = item.id)}
                    class={`hairline group relative overflow-hidden rounded-2xl p-5 text-left transition-all duration-300 ${
                      mode === item.id
                        ? "bg-surface-elevated/80 shadow-elev"
                        : "bg-surface/40 hover:bg-surface/70"
                    }`}
                  >
                    <div class="flex items-center justify-between gap-3">
                      <span
                        class={`flex h-4 w-4 items-center justify-center rounded-full border ${
                          mode === item.id
                            ? "border-primary-glow bg-primary-glow shadow-glow"
                            : "border-border"
                        }`}
                      >
                        {#if mode === item.id}
                          <Check class="h-2.5 w-2.5 text-black" />
                        {/if}
                      </span>
                      {#if mode === item.id}
                        <span class="text-[10px] font-semibold uppercase tracking-[0.25em] text-primary-glow">
                          Selected
                        </span>
                      {/if}
                    </div>
                    <div class="mt-4 text-base font-semibold">{item.title}</div>
                    <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                      {item.desc}
                    </p>
                  </button>
                {/each}
              </div>
            </section>

            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Provider keys</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Missing keys stay visible and safe. Providers without a working key remain disabled until you add one.
                </p>
              </div>
              <div class="grid gap-3">
                {#each providers as provider (provider.id)}
                  {@const value = keys[provider.id] ?? ""}
                  {@const hasKey = value.trim().length > 0}
                  <article class="hairline rounded-2xl bg-surface/40 p-6 transition-colors hover:bg-surface/60">
                    <div class="flex flex-wrap items-start justify-between gap-5">
                      <div class="min-w-0 flex-1">
                        <div class="flex items-center gap-3">
                          <h3 class="font-serif-display text-xl tracking-tight">
                            {provider.name}
                          </h3>
                          {#if hasKey}
                            <span class="hairline inline-flex items-center gap-1.5 rounded-full bg-emerald-400/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.2em] text-emerald-300">
                              <Check class="h-3 w-3" /> Active
                            </span>
                          {:else}
                            <span class="hairline inline-flex items-center gap-1.5 rounded-full bg-amber-400/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.2em] text-amber-300">
                              <KeyRound class="h-3 w-3" /> Key required
                            </span>
                          {/if}
                        </div>
                        <p class="mt-1 text-sm text-muted-foreground">{provider.blurb}</p>

                        <div class="mt-5 grid gap-3 sm:grid-cols-[1fr_auto]">
                          <div>
                            <label for={`${provider.id}-key`} class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                              {provider.name} API key
                            </label>
                            <input
                              id={`${provider.id}-key`}
                              type="password"
                              bind:value={keys[provider.id]}
                              placeholder="Paste key"
                              class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70"
                            />
                          </div>
                          <button
                            type="button"
                            onclick={() => (keys = { ...keys, [provider.id]: "" })}
                            disabled={!hasKey}
                            class="hairline mt-2 self-end rounded-xl bg-foreground/[0.04] px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40 sm:mt-7"
                          >
                            Clear key
                          </button>
                        </div>
                      </div>
                    </div>
                  </article>
                {/each}
              </div>
            </section>
          </div>
        {:else if active === "watchlist-services"}
          <div class="space-y-8">
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Connect your accounts</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Link external services to sync your watch history, ratings, and watchlists automatically as you play.
                </p>
              </div>

              <div class="grid gap-4">
                {#each watchlistServices as svc (svc.id)}
                  {@const connected = wlConnected[svc.id] ?? false}
                  {@const connecting = wlConnecting[svc.id] ?? false}
                  {@const keyVal = wlKeys[svc.id] ?? ''}

                  <article class="hairline rounded-2xl bg-surface/40 p-6 transition-colors hover:bg-surface/60">
                    <div class="flex items-start justify-between gap-4">
                      <div class="flex items-center gap-4">
                        <!-- Logo placeholder -->
                        <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-foreground/10 to-foreground/5 text-lg font-bold">
                          {svc.logo}
                        </div>
                        <div>
                          <div class="flex items-center gap-2">
                            <span class="font-semibold">{svc.name}</span>
                            {#if connected}
                              <span class="hairline inline-flex items-center gap-1.5 rounded-full bg-emerald-400/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.2em] text-emerald-300">
                                <Check class="h-3 w-3" /> Connected
                              </span>
                            {/if}
                          </div>
                          <p class="mt-0.5 text-xs text-muted-foreground max-w-sm">{svc.description}</p>
                        </div>
                      </div>

                      {#if connected}
                        <button
                          type="button"
                          onclick={() => { wlConnected = { ...wlConnected, [svc.id]: false }; wlKeys = { ...wlKeys, [svc.id]: '' }; }}
                          class="hairline shrink-0 inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-red-400/10 hover:text-red-300"
                        >
                          <Unlink class="h-3.5 w-3.5" /> Disconnect
                        </button>
                      {/if}
                    </div>

                    {#if !connected}
                      <div class="mt-5">
                        {#if svc.authType === 'apikey'}
                          <div class="grid gap-3 sm:grid-cols-[1fr_auto]">
                            <div>
                              <label for={`wl-${svc.id}-key`} class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                                {svc.name} API key
                              </label>
                              <input
                                id={`wl-${svc.id}-key`}
                                type="password"
                                value={keyVal}
                                oninput={(e) => { wlKeys = { ...wlKeys, [svc.id]: (e.currentTarget as HTMLInputElement).value }; }}
                                placeholder="Paste your API key"
                                class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70"
                              />
                            </div>
                            <button
                              type="button"
                              disabled={!keyVal.trim() || connecting}
                              onclick={() => {
                                wlConnecting = { ...wlConnecting, [svc.id]: true };
                                setTimeout(() => {
                                  wlConnected = { ...wlConnected, [svc.id]: true };
                                  wlConnecting = { ...wlConnecting, [svc.id]: false };
                                }, 800);
                              }}
                              class="hairline mt-2 self-end inline-flex items-center gap-2 rounded-xl bg-foreground/[0.06] px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground transition-colors hover:bg-foreground/[0.12] hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40 sm:mt-7"
                            >
                              {#if connecting}
                                <span class="h-3 w-3 animate-spin rounded-full border border-current border-t-transparent"></span>
                              {:else}
                                <Link class="h-3.5 w-3.5" />
                              {/if}
                              Connect
                            </button>
                          </div>
                        {:else}
                          <!-- OAuth flow -->
                          <div class="flex items-center gap-3">
                            <button
                              type="button"
                              onclick={() => {
                                wlConnecting = { ...wlConnecting, [svc.id]: true };
                                // In production this would open an OAuth window
                                setTimeout(() => {
                                  wlConnected = { ...wlConnected, [svc.id]: true };
                                  wlConnecting = { ...wlConnecting, [svc.id]: false };
                                }, 1200);
                              }}
                              disabled={connecting}
                              class="inline-flex items-center gap-2 rounded-xl bg-gradient-primary px-5 py-2.5 text-sm font-semibold text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
                            >
                              {#if connecting}
                                <span class="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white"></span>
                                Connecting…
                              {:else}
                                <Link class="h-4 w-4" /> Connect with {svc.name}
                              {/if}
                            </button>
                            <a
                              href={svc.docsUrl}
                              target="_blank"
                              rel="noopener noreferrer"
                              class="text-xs text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
                            >
                              Get API access →
                            </a>
                          </div>
                        {/if}
                      </div>
                    {:else}
                      <!-- Connected state: show sync controls -->
                      <div class="mt-4 flex flex-wrap items-center gap-3 border-t border-border pt-4">
                        <span class="text-xs text-muted-foreground">Last synced: just now</span>
                        <button
                          type="button"
                          class="hairline inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground"
                        >
                          Sync now
                        </button>
                      </div>
                    {/if}
                  </article>
                {/each}
              </div>
            </section>

            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Sync behaviour</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Choose what Xuva pushes and pulls from connected services.
                </p>
              </div>
              <div class="space-y-4">
                {#each [
                  { id: 'scrobble', label: 'Auto-scrobble', desc: 'Mark items watched when you reach 80% of playback' },
                  { id: 'ratings', label: 'Sync ratings', desc: 'Push ratings you set in Xuva to connected services' },
                  { id: 'watchlist_pull', label: 'Import watchlists', desc: 'Pull watchlists from services into your Xuva library wishlist' }
                ] as opt (opt.id)}
                  <label class="hairline flex cursor-pointer items-start gap-4 rounded-xl bg-surface/40 p-4 transition-colors hover:bg-surface/60">
                    <input type="checkbox" checked class="mt-0.5 h-4 w-4 cursor-pointer rounded accent-primary" />
                    <div>
                      <div class="text-sm font-medium">{opt.label}</div>
                      <p class="text-xs text-muted-foreground">{opt.desc}</p>
                    </div>
                  </label>
                {/each}
              </div>
            </section>
          </div>

        {:else if active === "dashboard"}
          {#if dashLoading && !sysStatus && !catalogSummary}
            <div class="flex min-h-[20vh] items-center justify-center">
              <div class="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div>
            </div>
          {:else}
            <div class="space-y-10">

              <!-- Refresh button -->
              <div class="flex justify-end">
                <button
                  type="button"
                  onclick={loadDashboard}
                  disabled={dashLoading}
                  class="hairline inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground disabled:opacity-40"
                >
                  <RefreshCw class="h-3 w-3 {dashLoading ? 'animate-spin' : ''}" /> Refresh
                </button>
              </div>

              <!-- Stat cards -->
              <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
                {#each ([
                  {
                    label: 'Movies',
                    value: catalogSummary?.movies != null ? String(catalogSummary.movies) : '—',
                    sub: !catalogSummary ? 'Loading…' : catalogSummary.movies === 0 ? 'None scanned yet' : 'In your library',
                    accent: catalogSummary?.movies ? 'text-foreground' : 'text-foreground/50'
                  },
                  {
                    label: 'TV Shows',
                    value: catalogSummary?.series != null ? String(catalogSummary.series) : '—',
                    sub: !catalogSummary ? 'Loading…' : catalogSummary.series === 0 ? 'None scanned yet' : 'In your library',
                    accent: catalogSummary?.series ? 'text-foreground' : 'text-foreground/50'
                  },
                  {
                    label: 'Episodes',
                    value: catalogSummary?.episodes != null ? String(catalogSummary.episodes) : '—',
                    sub: 'Across all seasons',
                    accent: 'text-foreground/80'
                  },
                  {
                    label: 'Sessions',
                    value: String(activeSessions.length),
                    sub: activeSessions.length === 1 ? 'Now playing' : activeSessions.length > 1 ? 'Now playing' : 'Nobody watching',
                    accent: activeSessions.length > 0 ? 'text-primary-glow' : 'text-foreground/50'
                  }
                ] as const) as stat (stat.label)}
                  <div class="hairline rounded-2xl bg-surface/40 p-5">
                    <div class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">{stat.label}</div>
                    <div class="font-serif-display mt-2 text-3xl {stat.accent}">{stat.value}</div>
                    <div class="mt-1 text-xs text-muted-foreground">{stat.sub}</div>
                  </div>
                {/each}
              </div>

              <!-- System resources -->
              {#if sysStatus}
                <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
                  <div>
                    <h3 class="font-serif-display text-lg tracking-tight">System resources</h3>
                    <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                      Live resource usage from the Xuva server process.
                    </p>
                  </div>
                  <div class="space-y-5">
                    <!-- CPU -->
                    {#if sysStatus.cpu}
                      {@const cpuPct = sysStatus.cpu.percent ?? 0}
                      <div>
                        <div class="flex items-center justify-between">
                          <span class="text-[10px] font-semibold uppercase tracking-[0.2em] text-muted-foreground">CPU</span>
                          <span class="text-xs {cpuPct > 85 ? 'text-amber-300' : 'text-foreground/80'}">{cpuPct.toFixed(1)}%{sysStatus.cpu.cores ? ` · ${sysStatus.cpu.cores} cores` : ''}</span>
                        </div>
                        <div class="mt-2 h-2 w-full overflow-hidden rounded-full bg-surface-elevated/60">
                          <div class="h-full rounded-full transition-all {cpuPct > 85 ? 'bg-amber-400' : 'bg-primary-glow'}" style="width: {Math.min(cpuPct, 100)}%"></div>
                        </div>
                      </div>
                    {/if}

                    <!-- Memory -->
                    {#if sysStatus.memory}
                      {@const memPct = sysStatus.memory.usedPercent ?? 0}
                      <div>
                        <div class="flex items-center justify-between">
                          <span class="text-[10px] font-semibold uppercase tracking-[0.2em] text-muted-foreground">Memory</span>
                          <span class="text-xs {memPct > 90 ? 'text-amber-300' : 'text-foreground/80'}">{formatBytes(sysStatus.memory.usedBytes)} / {formatBytes(sysStatus.memory.totalBytes)}</span>
                        </div>
                        <div class="mt-2 h-2 w-full overflow-hidden rounded-full bg-surface-elevated/60">
                          <div class="h-full rounded-full transition-all {memPct > 90 ? 'bg-amber-400' : 'bg-primary-glow/70'}" style="width: {Math.min(memPct, 100)}%"></div>
                        </div>
                      </div>
                    {/if}

                    <!-- Network -->
                    {#if sysStatus.network && (sysStatus.network.receiveBps || sysStatus.network.transmitBps)}
                      <div class="flex gap-6 text-xs text-muted-foreground/70">
                        <span>↓ {formatBps(sysStatus.network.receiveBps)}</span>
                        <span>↑ {formatBps(sysStatus.network.transmitBps)}</span>
                      </div>
                    {/if}

                    <!-- Disks -->
                    {#if sysStatus.disks && sysStatus.disks.length > 0}
                      <div class="space-y-3 border-t border-border pt-3">
                        {#each sysStatus.disks as disk (disk.path ?? disk.name)}
                          {@const diskPct = disk.usedPercent ?? 0}
                          <div>
                            <div class="flex items-center justify-between">
                              <span class="max-w-[55%] truncate font-mono text-[10px] text-muted-foreground">{disk.path ?? disk.name ?? 'Disk'}</span>
                              <span class="text-[11px] {diskPct > 90 ? 'text-amber-300' : 'text-foreground/60'}">{formatBytes(disk.usedBytes)} / {formatBytes(disk.totalBytes)}</span>
                            </div>
                            <div class="mt-1.5 h-1.5 w-full overflow-hidden rounded-full bg-surface-elevated/60">
                              <div class="h-full rounded-full {diskPct > 90 ? 'bg-amber-400' : 'bg-foreground/25'}" style="width: {Math.min(diskPct, 100)}%"></div>
                            </div>
                          </div>
                        {/each}
                      </div>
                    {/if}
                  </div>
                </section>
              {:else if dashError}
                <div class="hairline rounded-2xl bg-surface/30 px-6 py-8 text-center">
                  <p class="text-sm text-muted-foreground">Server unreachable — resource data unavailable.</p>
                </div>
              {/if}

              <!-- Active sessions -->
              {#if activeSessions.length > 0}
                <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
                  <div>
                    <h3 class="font-serif-display text-lg tracking-tight">Now playing</h3>
                    <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                      {activeSessions.length} active playback {activeSessions.length === 1 ? 'session' : 'sessions'}.
                    </p>
                  </div>
                  <div class="space-y-2">
                    {#each activeSessions as session (session.id)}
                      <div class="hairline flex items-center gap-4 rounded-xl bg-surface/40 px-4 py-3">
                        <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary-glow/10">
                          <Play class="h-3.5 w-3.5 fill-primary-glow text-primary-glow" />
                        </div>
                        <div class="min-w-0 flex-1">
                          <div class="truncate text-sm font-medium">{session.title ?? session.sourceName ?? 'Unknown'}</div>
                          <div class="text-[11px] text-muted-foreground">{session.mode ?? session.route ?? ''}{session.deviceId ? ` · ${session.deviceId}` : ''}</div>
                        </div>
                        <span class="shrink-0 rounded-full bg-emerald-400/10 px-2.5 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-emerald-300">Live</span>
                      </div>
                    {/each}
                  </div>
                </section>
              {/if}

              <!-- Recent scan activity -->
              <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
                <div>
                  <h3 class="font-serif-display text-lg tracking-tight">Recent activity</h3>
                  <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                    Library scans, metadata refreshes, and system events.
                  </p>
                </div>
                {#if recentScans.length > 0}
                  <div class="space-y-2">
                    {#each recentScans as scan (scan.id)}
                      <div class="hairline flex items-center gap-3 rounded-xl bg-surface/40 px-4 py-3">
                        <div class="min-w-0 flex-1">
                          <div class="flex items-center gap-2 text-sm">
                            <span class="font-medium capitalize">{scan.kind ?? 'Scan'}</span>
                            {#if scan.libraryId}<span class="font-mono text-[11px] text-muted-foreground">{scan.libraryId.slice(0, 8)}</span>{/if}
                          </div>
                          <div class="mt-0.5 text-[11px] text-muted-foreground">
                            {scan.status ?? ''}
                            {#if scan.updatedAt} · {new Date(scan.updatedAt).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}{/if}
                          </div>
                        </div>
                        {#if scan.status === 'running'}
                          <div class="h-4 w-4 animate-spin rounded-full border border-primary-glow border-t-transparent"></div>
                        {:else if scan.status === 'completed'}
                          <Check class="h-4 w-4 text-emerald-300" />
                        {/if}
                      </div>
                    {/each}
                  </div>
                {:else}
                  <div class="hairline flex flex-col items-center justify-center rounded-2xl bg-surface/30 px-8 py-16 text-center">
                    <p class="text-sm text-muted-foreground">No recent activity.</p>
                    <p class="mt-1 text-xs text-muted-foreground/60">Scan a library to see events here.</p>
                  </div>
                {/if}
              </section>
            </div>
          {/if}

        {:else if active === "libraries"}
          <div class="space-y-8">

            <!-- Header row -->
            <div class="flex items-center justify-between">
              <p class="text-sm text-muted-foreground">
                {libraries.length === 0 ? 'No libraries configured yet.' : `${libraries.length} librar${libraries.length === 1 ? 'y' : 'ies'}`}
              </p>
              <button
                type="button"
                onclick={() => { showAddLib = !showAddLib; if (!showAddLib) showFolderBrowser = false; }}
                class="inline-flex items-center gap-2 rounded-full bg-gradient-primary px-4 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110"
              >
                <Plus class="h-3.5 w-3.5" /> {showAddLib ? 'Cancel' : 'Add Library'}
              </button>
            </div>

            <!-- Error banner -->
            {#if libError}
              <div class="rounded-xl bg-red-400/10 px-4 py-3 text-sm text-red-300">{libError}</div>
            {/if}

            <!-- Add library form -->
            {#if showAddLib}
              <div class="hairline rounded-2xl bg-surface/50 p-6 space-y-5">
                <h3 class="font-semibold">New Library</h3>

                <div class="grid gap-5 sm:grid-cols-2">
                  <!-- Name -->
                  <div>
                    <label for="lib-name" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                      Library name
                    </label>
                    <input
                      id="lib-name"
                      type="text"
                      bind:value={newLibName}
                      placeholder="My Movies"
                      class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70"
                    />
                  </div>

                  <!-- Kind -->
                  <div>
                    <div class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Kind</div>
                    <div class="mt-2 flex gap-2">
                      {#each ([{ id: 'movies', label: 'Movies' }, { id: 'tv', label: 'TV Shows' }] as const) as opt (opt.id)}
                        <button
                          type="button"
                          onclick={() => (newLibKind = opt.id)}
                          class="flex-1 rounded-xl border py-2.5 text-sm font-medium transition-colors {newLibKind === opt.id ? 'border-primary/60 bg-primary-glow/10 text-foreground' : 'border-border bg-surface/40 text-muted-foreground hover:text-foreground'}"
                        >
                          {opt.label}
                        </button>
                      {/each}
                    </div>
                  </div>
                </div>

                <!-- Path + Browse -->
                <div>
                  <label for="lib-path" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                    Folder path
                  </label>
                  <div class="mt-2 flex gap-2">
                    <input
                      id="lib-path"
                      type="text"
                      bind:value={newLibPath}
                      placeholder="/media/movies"
                      class="h-11 flex-1 rounded-xl border border-border bg-background/40 px-4 font-mono text-sm outline-none placeholder:font-sans placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70"
                    />
                    <button
                      type="button"
                      onclick={openFolderBrowser}
                      class="hairline flex h-11 items-center gap-1.5 rounded-xl bg-foreground/[0.06] px-4 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.10] hover:text-foreground"
                    >
                      <Folder class="h-3.5 w-3.5" /> Browse
                    </button>
                  </div>

                  <!-- Inline folder browser -->
                  {#if showFolderBrowser}
                    <div class="mt-2 hairline overflow-hidden rounded-xl bg-background/40">
                      <!-- Browser toolbar -->
                      <div class="flex items-center gap-2 border-b border-border px-3 py-2">
                        <FolderOpen class="h-3.5 w-3.5 shrink-0 text-primary-glow" />
                        <span class="min-w-0 flex-1 truncate font-mono text-xs text-foreground/80">{browserCurrentPath || '/'}</span>
                        <button
                          type="button"
                          onclick={() => { newLibPath = browserCurrentPath; showFolderBrowser = false; }}
                          class="shrink-0 rounded-full bg-primary-glow/10 px-3 py-1 text-[11px] font-semibold text-primary-glow transition-colors hover:bg-primary-glow/20"
                        >
                          Select
                        </button>
                        <button
                          type="button"
                          onclick={() => (showFolderBrowser = false)}
                          class="shrink-0 rounded-full bg-foreground/[0.06] px-2 py-1 text-[11px] text-muted-foreground transition-colors hover:text-foreground"
                        >
                          ✕
                        </button>
                      </div>

                      {#if browserLoading}
                        <div class="flex items-center justify-center py-5">
                          <div class="h-5 w-5 animate-spin rounded-full border border-border border-t-primary-glow"></div>
                        </div>
                      {:else}
                        <ul class="max-h-52 overflow-y-auto py-1">
                          <!-- Up one level -->
                          {#if browserParentPath !== undefined}
                            <li>
                              <button
                                type="button"
                                onclick={() => navigateFolder(browserParentPath!)}
                                class="flex w-full items-center gap-3 px-3 py-2 text-sm text-muted-foreground hover:bg-foreground/[0.04] hover:text-foreground"
                              >
                                <ChevronLeft class="h-3.5 w-3.5 shrink-0" />
                                <span class="font-mono text-xs">..</span>
                              </button>
                            </li>
                          {/if}
                          <!-- Folder entries -->
                          {#each browserEntries as entry (entry.path ?? entry.name)}
                            <li>
                              <button
                                type="button"
                                onclick={() => navigateFolder(entry.path ?? '')}
                                class="flex w-full items-center gap-3 px-3 py-2 text-sm text-foreground/80 hover:bg-foreground/[0.04] hover:text-foreground"
                              >
                                <Folder class="h-3.5 w-3.5 shrink-0 text-muted-foreground/60" />
                                <span class="flex-1 truncate font-mono text-xs">{entry.name ?? entry.path}</span>
                                <ChevronRight class="h-3 w-3 shrink-0 text-muted-foreground/40" />
                              </button>
                            </li>
                          {:else}
                            <li class="px-4 py-3 text-xs text-muted-foreground/60">No sub-folders found.</li>
                          {/each}
                        </ul>
                      {/if}
                    </div>
                  {/if}
                </div>

                <!-- Storage type -->
                <div>
                  <div class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Storage type</div>
                  <div class="mt-2 flex flex-wrap gap-2">
                    {#each ([
                      { id: 'local', label: 'Local' },
                      { id: 'network', label: 'Network' },
                      { id: 'removable', label: 'Removable' },
                      { id: 'mounted', label: 'Mounted' }
                    ] as const) as opt (opt.id)}
                      <button
                        type="button"
                        onclick={() => (newLibStorageType = opt.id)}
                        class="rounded-full border px-3 py-1.5 text-xs font-medium transition-colors {newLibStorageType === opt.id ? 'border-primary/60 bg-primary-glow/10 text-foreground' : 'border-border bg-surface/40 text-muted-foreground hover:text-foreground'}"
                      >
                        {opt.label}
                      </button>
                    {/each}
                  </div>
                </div>

                <!-- Form actions -->
                <div class="flex items-center gap-3 border-t border-border pt-4">
                  <button
                    type="button"
                    onclick={handleSaveLibrary}
                    disabled={libSaving || !newLibName.trim() || !newLibPath.trim()}
                    class="inline-flex items-center gap-2 rounded-full bg-gradient-primary px-5 py-2.5 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {#if libSaving}
                      <span class="h-3.5 w-3.5 animate-spin rounded-full border border-white/30 border-t-white"></span>
                      Saving…
                    {:else}
                      Save Library
                    {/if}
                  </button>
                  <button
                    type="button"
                    onclick={() => { showAddLib = false; showFolderBrowser = false; }}
                    class="hairline rounded-full bg-foreground/[0.04] px-5 py-2.5 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            {/if}

            <!-- Library list -->
            {#if libLoading}
              <div class="flex items-center justify-center py-8">
                <div class="h-6 w-6 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div>
              </div>
            {:else if libraries.length > 0}
              <div class="space-y-3">
                {#each libraries as lib (lib.id ?? lib.path ?? lib.name)}
                  <div class="hairline flex flex-wrap items-center gap-4 rounded-2xl bg-surface/40 p-5 transition-colors hover:bg-surface/60">
                    <!-- Icon -->
                    <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-surface-elevated/60 text-primary-glow">
                      {#if lib.kind === 'tv'}
                        <Tv class="h-5 w-5" />
                      {:else}
                        <Film class="h-5 w-5" />
                      {/if}
                    </div>

                    <!-- Meta -->
                    <div class="min-w-0 flex-1">
                      <div class="flex flex-wrap items-center gap-2">
                        <span class="font-semibold">{lib.name ?? 'Unnamed'}</span>
                        <span class="rounded-full bg-foreground/[0.06] px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
                          {lib.kind === 'tv' ? 'TV' : 'Movies'}
                        </span>
                        {#if lib.storageType && lib.storageType !== 'local'}
                          <span class="rounded-full bg-foreground/[0.06] px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
                            {lib.storageType}
                          </span>
                        {/if}
                      </div>
                      <div class="mt-0.5 truncate font-mono text-xs text-muted-foreground/60">{lib.path ?? '—'}</div>
                    </div>

                    <!-- Actions -->
                    <div class="flex shrink-0 items-center gap-2">
                      <!-- Scan -->
                      <button
                        type="button"
                        onclick={() => lib.id && handleScanLibrary(lib.id)}
                        disabled={libScanningId === lib.id || !lib.id}
                        class="hairline inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground disabled:opacity-40"
                      >
                        {#if libScanningId === lib.id}
                          <span class="h-3 w-3 animate-spin rounded-full border border-current border-t-transparent"></span>
                          Scanning…
                        {:else}
                          <RefreshCw class="h-3 w-3" /> Scan
                        {/if}
                      </button>

                      <!-- Delete (arm → confirm) -->
                      {#if libDeletingId === lib.id}
                        <button
                          type="button"
                          onclick={() => lib.id && handleDeleteLibrary(lib.id)}
                          class="inline-flex items-center gap-1.5 rounded-full bg-red-400/10 px-3 py-1.5 text-xs font-semibold text-red-300 transition-colors hover:bg-red-400/20"
                        >
                          Confirm delete?
                        </button>
                        <button
                          type="button"
                          onclick={() => (libDeletingId = null)}
                          class="hairline rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08]"
                        >
                          Cancel
                        </button>
                      {:else}
                        <button
                          type="button"
                          onclick={() => lib.id && handleDeleteLibrary(lib.id)}
                          disabled={!lib.id}
                          class="hairline inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-red-400/10 hover:text-red-300 disabled:opacity-40"
                        >
                          <Trash2 class="h-3 w-3" /> Delete
                        </button>
                      {/if}
                    </div>
                  </div>
                {/each}
              </div>
            {:else if !showAddLib}
              <div class="hairline flex flex-col items-center justify-center rounded-3xl bg-surface/30 px-8 py-20 text-center">
                <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-surface-elevated/70 text-primary-glow shadow-elev">
                  <Library class="h-6 w-6" />
                </div>
                <div class="font-serif-display mt-5 text-xl tracking-tight">No libraries yet</div>
                <p class="mt-2 max-w-sm text-sm text-muted-foreground">
                  Add a library to point Xuva at a folder of movies or TV shows.
                </p>
                <button
                  type="button"
                  onclick={() => (showAddLib = true)}
                  class="mt-6 inline-flex items-center gap-2 rounded-full bg-gradient-primary px-5 py-2.5 text-sm font-semibold text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110"
                >
                  <Plus class="h-4 w-4" /> Add your first library
                </button>
              </div>
            {/if}
          </div>

        {:else if active === "general"}
          <div class="space-y-12">
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Server identity</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  The name your clients and devices see when they connect to this Xuva server.
                </p>
              </div>
              <div class="space-y-5">
                <div>
                  <label for="server-name" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                    Server name
                  </label>
                  <input
                    id="server-name"
                    type="text"
                    placeholder="Xuva"
                    class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70"
                  />
                </div>
                <div>
                  <label for="interface-lang" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                    Interface language
                  </label>
                  <select
                    id="interface-lang"
                    class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none focus:border-primary/60"
                  >
                    <option value="en">English</option>
                    <option value="es">Spanish</option>
                    <option value="fr">French</option>
                    <option value="de">German</option>
                    <option value="it">Italian</option>
                    <option value="pt">Portuguese</option>
                  </select>
                </div>
              </div>
            </section>

            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Theme</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Choose how Xuva appears. The app currently ships in dark-only mode.
                </p>
              </div>
              <div class="grid gap-3 sm:grid-cols-3">
                {#each [
                  { id: 'dark', label: 'Dark', sub: 'Cinema-first, easy on the eyes' },
                  { id: 'light', label: 'Light', sub: 'Coming soon', disabled: true },
                  { id: 'system', label: 'System', sub: 'Coming soon', disabled: true },
                ] as theme (theme.id)}
                  <div class={`hairline rounded-2xl p-4 text-left ${theme.disabled ? 'opacity-40 cursor-not-allowed' : 'bg-surface/40 hover:bg-surface/70 cursor-pointer'} ${theme.id === 'dark' ? 'bg-surface-elevated/70 ring-1 ring-primary-glow/30' : ''}`}>
                    <div class="text-sm font-semibold">{theme.label}</div>
                    <p class="mt-1 text-xs text-muted-foreground">{theme.sub}</p>
                  </div>
                {/each}
              </div>
            </section>
          </div>

        {:else if active === "about"}
          <div class="space-y-8">
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Build information</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Version details for this installation of Xuva.
                </p>
              </div>
              <div class="space-y-4">
                {#each [
                  { label: 'Application', value: 'Xuva' },
                  { label: 'Channel', value: 'Development build' },
                  { label: 'Web UI', value: 'SvelteKit + Svelte 5' },
                  { label: 'Server', value: 'Go' },
                ] as row (row.label)}
                  <div class="flex items-center justify-between border-b border-border py-3 last:border-0">
                    <span class="text-sm text-muted-foreground">{row.label}</span>
                    <span class="font-mono text-sm text-foreground/80">{row.value}</span>
                  </div>
                {/each}
              </div>
            </section>

            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Open source</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Xuva is built on the shoulders of open-source giants.
                </p>
              </div>
              <div class="space-y-3">
                {#each [
                  { name: 'SvelteKit', purpose: 'Web application framework', license: 'MIT' },
                  { name: 'Go', purpose: 'Server runtime and API layer', license: 'BSD-3-Clause' },
                  { name: 'FFmpeg', purpose: 'Media transcoding and probing', license: 'LGPL-2.1 / GPL-2.0' },
                  { name: 'Tailwind CSS', purpose: 'Utility-first styling', license: 'MIT' },
                  { name: 'Lucide', purpose: 'Icon library', license: 'ISC' },
                ] as lib (lib.name)}
                  <div class="hairline flex items-center justify-between rounded-xl bg-surface/40 px-4 py-3">
                    <div>
                      <div class="text-sm font-medium">{lib.name}</div>
                      <div class="text-xs text-muted-foreground">{lib.purpose}</div>
                    </div>
                    <span class="shrink-0 rounded-full bg-foreground/[0.06] px-2.5 py-1 font-mono text-[10px] text-muted-foreground">{lib.license}</span>
                  </div>
                {/each}
              </div>
            </section>
          </div>

        {:else if active === "account"}
          <div class="space-y-12">
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Profile</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Your display name and public identity on this server.
                </p>
              </div>
              <div class="space-y-5">
                <div>
                  <label for="display-name" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                    Display name
                  </label>
                  <input
                    id="display-name"
                    type="text"
                    placeholder="Your name"
                    class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70"
                  />
                </div>
                <div>
                  <label for="account-username" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                    Username
                  </label>
                  <input
                    id="account-username"
                    type="text"
                    disabled
                    placeholder="username"
                    class="mt-2 h-11 w-full cursor-not-allowed rounded-xl border border-border bg-background/20 px-4 text-sm text-muted-foreground/60 outline-none"
                  />
                  <p class="mt-1.5 text-[11px] text-muted-foreground/60">Username cannot be changed after setup.</p>
                </div>
                <button
                  type="button"
                  class="rounded-full bg-gradient-primary px-5 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110"
                >
                  Save profile
                </button>
              </div>
            </section>

            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Change password</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Keep your account secure with a strong, unique password.
                </p>
              </div>
              <div class="space-y-4">
                <div>
                  <label for="current-password" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                    Current password
                  </label>
                  <input
                    id="current-password"
                    type="password"
                    placeholder="••••••••"
                    class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70"
                  />
                </div>
                <div>
                  <label for="new-password" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                    New password
                  </label>
                  <input
                    id="new-password"
                    type="password"
                    placeholder="••••••••"
                    class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70"
                  />
                </div>
                <div>
                  <label for="confirm-password" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                    Confirm new password
                  </label>
                  <input
                    id="confirm-password"
                    type="password"
                    placeholder="••••••••"
                    class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70"
                  />
                </div>
                <button
                  type="button"
                  class="rounded-full bg-gradient-primary px-5 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110"
                >
                  Update password
                </button>
              </div>
            </section>

            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Session</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  End your current session on this server.
                </p>
              </div>
              <div>
                <button
                  type="button"
                  class="hairline inline-flex items-center gap-2 rounded-xl bg-foreground/[0.04] px-5 py-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-red-400/10 hover:text-red-300"
                >
                  <LogOut class="h-4 w-4" /> Sign out
                </button>
              </div>
            </section>
          </div>

        {:else}
          <div class="hairline flex flex-col items-center justify-center rounded-3xl bg-surface/30 px-8 py-24 text-center">
            <div class="hairline flex h-14 w-14 items-center justify-center rounded-2xl bg-surface-elevated/70 text-primary-glow shadow-elev">
              <current.icon class="h-6 w-6" />
            </div>
            <div class="font-serif-display mt-5 text-2xl tracking-tight">
              {current.label}
            </div>
            <p class="mt-2 max-w-sm text-sm text-muted-foreground">
              {current.hint ?? "This section is part of the Xuva settings surface."} — controls coming online soon.
            </p>
          </div>
        {/if}
      </div>
    </div>
  </main>
</div>
