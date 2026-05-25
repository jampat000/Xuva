<script lang="ts">
  import {
    ArrowRightLeft,
    ArchiveRestore,
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
    LayoutDashboard,
    Library,
    Link,
    Link2,
    LogOut,
    Play,
    Plus,
    QrCode,
    RefreshCw,
    ScanSearch,
    Search,
    Settings2,
    ShieldCheck,
    Sliders,
    Trash2,
    Tv,
    Unlink,
    Upload,
    User,
    Users,
    Wifi
  } from "lucide-svelte";
  import { onMount, onDestroy } from 'svelte';
  import { appState } from '$lib/stores/appState.svelte';
  import Header from "$lib/components/Header.svelte";
  import {
    getSystemStatus,
    getCatalogSummary,
    getScans,
    getSessions,
    getJobs,
    getCatalogHealth,
    getCatalogCodecs,
    startProbeJob,
    getSettings,
    updateSettings,
    updateMetadataSourcePreferences,
    getLibraries,
    saveLibrary,
    deleteLibrary,
    scanLibrary,
    browseFolders,
    getPerformanceSettings,
    getDiscoveryStatus,
    getPairingRequests,
    approvePairingRequest,
    denyPairingRequest,
    getApprovedDevices,
    revokeApprovedDevice,
    getUsers,
    createUser,
    deleteUser,
    updateUser,
    updateUserPassword,
    getDeviceProfiles,
    scanAllLibraries,
    runHardwareTest,
    exportBackup,
    importBackup,
    type SystemStatusResponse,
    type CatalogSummaryResponse,
    type ScanJobItem,
    type SessionItem,
    type JobsStatusResponse,
    type CatalogHealthResponse,
    type CodecBreakdownResponse,
    type SettingsResponse,
    type LibraryItem,
    type FolderEntry,
    type PerformanceSettingsResponse,
    type DiscoveryStatusResponse,
    type PairingRequestItem,
    type ApprovedDeviceItem,
    type UserItem,
    type HardwareTestResponse,
    type DeviceProfile,
    type BackupImportResponse,
    generateQRPairToken,
    type QRTokenResponse,
  } from '$lib/api/operator';
  import {
    getBackfillStatus,
    startBackfill,
    stopBackfill,
    type BackfillResponse,
  } from '$lib/api/browse';
  import { createEventStream } from '$lib/events/stream';
  import ActivityRing from '$lib/components/ActivityRing.svelte';
  import { getAuthSession, logout, type AuthSessionUser } from '$lib/api/auth';
  import { updateProfileSettings, setProfilePin, RATING_OPTIONS, AVATAR_PRESETS } from '$lib/api/profiles';

  type Group = "Account" | "Server" | "Devices" | "Advanced";

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
    { id: "backup", label: "Backup", icon: ArchiveRestore, group: "Server", hint: "Export & restore your catalog database" },
    { id: "network", label: "Network", icon: Wifi, group: "Server", hint: "Ports, mDNS discovery, remote access" },
    { id: "watchlist-services", label: "Watchlist Services", icon: BookMarked, group: "Server", hint: "Sync with Trakt, Letterboxd & more" },
    { id: "users", label: "Users", icon: Users, group: "Server", hint: "Accounts & access roles" },
    { id: "pending-approvals", label: "Pending Approvals", icon: Link2, group: "Devices", hint: "Review & approve device requests" },
    { id: "approved-devices", label: "Approved Devices", icon: ShieldCheck, group: "Devices", hint: "Trusted & active devices" },
    { id: "about", label: "About", icon: Info, group: "Advanced", hint: "Build info & open-source licenses" }
  ];

  const groups: Group[] = ["Account", "Server", "Devices", "Advanced"];

  let active = $state("dashboard");
  let q = $state("");
  let headerScrolled = $state(false);
  let mainRef = $state<HTMLDivElement | null>(null);

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

  // ─── Watchlist sync-behaviour options (persisted to localStorage) ─────────
  const WL_SYNC_KEY = 'xuva-wl-sync-opts';
  function loadWlSync(): Record<string, boolean> {
    try {
      const raw = typeof localStorage !== 'undefined' ? localStorage.getItem(WL_SYNC_KEY) : null;
      return raw ? JSON.parse(raw) : { scrobble: true, ratings: true, watchlist_pull: true };
    } catch { return { scrobble: true, ratings: true, watchlist_pull: true }; }
  }
  const _wlSyncInit = loadWlSync();
  let wlSyncSaved = $state<Record<string, boolean>>(_wlSyncInit);
  let wlSyncEdit  = $state<Record<string, boolean>>({ ..._wlSyncInit });
  const wlSyncDirty = $derived(
    Object.keys(wlSyncSaved).some(k => wlSyncEdit[k] !== wlSyncSaved[k])
  );
  function saveWlSync() {
    wlSyncSaved = { ...wlSyncEdit };
    try { localStorage.setItem(WL_SYNC_KEY, JSON.stringify(wlSyncSaved)); } catch { /* ignore */ }
  }
  function discardWlSync() {
    wlSyncEdit = { ...wlSyncSaved };
  }

  // ─── Backup state ─────────────────────────────────────────────────────────
  let backupExporting = $state(false);
  let backupExportError = $state<string | null>(null);
  let backupImporting = $state(false);
  let backupImportResult = $state<BackupImportResponse | null>(null);
  let backupImportError = $state<string | null>(null);

  async function handleExport() {
    backupExporting = true;
    backupExportError = null;
    try {
      await exportBackup();
    } catch (e) {
      backupExportError = e instanceof Error ? e.message : 'Export failed';
    } finally {
      backupExporting = false;
    }
  }

  async function handleImport(e: Event) {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    backupImporting = true;
    backupImportResult = null;
    backupImportError = null;
    try {
      backupImportResult = await importBackup(file);
    } catch (err) {
      backupImportError = err instanceof Error ? err.message : 'Import failed';
    } finally {
      backupImporting = false;
      input.value = '';
    }
  }

  let current = $derived(sections.find((section) => section.id === active) ?? sections[0]);

  // ─── Dashboard live data ───────────────────────────────────────────────────
  let settingsData = $state<SettingsResponse | null>(null);
  let sysStatus = $state<SystemStatusResponse | null>(null);
  let catalogSummary = $state<CatalogSummaryResponse | null>(null);
  let recentScans = $state<ScanJobItem[]>([]);
  let activeSessions = $state<SessionItem[]>([]);
  let dashLoading = $state(false);
  let dashError = $state<string | null>(null);

  // ─── Dashboard live extras (jobs, health, perf) ─────────────────────────
  let dashJobs  = $state<JobsStatusResponse | null>(null);
  let dashHealth = $state<CatalogHealthResponse | null>(null);
  let dashCodecs = $state<CodecBreakdownResponse | null>(null);
  let dashScanBusy  = $state(false);
  let dashProbeBusy = $state(false);
  let dashUpdatedAt = $state('');
  let dashPollTimers: ReturnType<typeof setInterval>[] = [];
  const dashStream = createEventStream();

  // Arc gauge helpers (used in dashboard SVG gauges)
  const _DR = 36;
  const _DC = 2 * Math.PI * _DR; // ≈ 226.2
  function dashArc(pct: number): string {
    const p = Math.min(100, Math.max(0, pct));
    return `${_DC * p / 100} ${_DC * (1 - p / 100)}`;
  }
  function dashGaugeColor(pct: number, warnAt = 70, critAt = 90): string {
    if (pct >= critAt) return 'oklch(0.68 0.26 22)';
    if (pct >= warnAt) return 'oklch(0.85 0.22 75)';
    return 'oklch(0.62 0.22 285)';
  }
  function dashDiskBarClass(pct: number): string {
    if (pct >= 90) return 'bg-red-400';
    if (pct >= 75) return 'bg-amber-400';
    return 'bg-foreground/25';
  }
  function dashDiskValClass(pct: number): string {
    if (pct >= 90) return 'text-red-400';
    if (pct >= 75) return 'text-amber-400';
    return 'text-foreground/70';
  }

  // ─── Flight deck HUD helpers ──────────────────────────────────────────────
  // Full-circle gauge dimensions. Switched from semicircle arc to a complete
  // ring because the semicircle's text overlay was hiding most of the arc on
  // some viewports — the ring keeps the % text inside the ring without ever
  // overlapping the stroke.
  const _RING_R    = 42;
  const _RING_CIRC = 2 * Math.PI * _RING_R; // ≈ 263.89 — full circle length
  // RAG (Red/Amber/Green) gauge fill — used for CPU/RAM where load semantics matter.
  function dashHudColor(pct: number, warnAt = 70, critAt = 90): string {
    if (pct >= critAt) return 'oklch(0.68 0.26 22)';   // critical red
    if (pct >= warnAt) return 'oklch(0.85 0.22 75)';   // caution amber
    return 'oklch(0.78 0.22 145)';                      // nominal green
  }
  const dashHudStatusStr = $derived.by(() => {
    if (!sysStatus) return '—';
    const cpu = sysStatus.cpu?.percent ?? 0;
    const mem = sysStatus.memory?.usedPercent ?? 0;
    if (cpu >= 90 || mem >= 90) return 'FAULT';
    if (cpu >= 70 || mem >= 70) return 'CAUTION';
    return 'NOMINAL';
  });

  // Derived dashboard metrics
  const dashCpuPct = $derived(Math.round(sysStatus?.cpu?.percent ?? 0));
  const dashMemPct = $derived(Math.round(sysStatus?.memory?.usedPercent ?? 0));
  const dashTotalFiles = $derived(dashHealth?.summary?.mediaSources ?? 0);
  const dashUnprobed   = $derived(dashHealth?.unprobed ?? 0);
  const dashProbed     = $derived(Math.max(0, dashTotalFiles - dashUnprobed));
  const dashProbeRunning = $derived(
    dashJobs?.probe?.status === 'running' ||
    (dashJobs?.probe?.activeJobs?.some(j => j.status === 'running') ?? false)
  );
  const dashMetaRunning = $derived(dashJobs?.metadata?.backfill?.running === true);

  // ─── Uptime ticker — re-evaluates every second so the dashboard clock ─────
  // shows live "1d 2h 14m" without polling /api/system/status more often.
  let dashNow = $state(Date.now());
  const dashUptimeStr = $derived.by(() => {
    const startedAt = sysStatus?.serverStartedAt;
    if (!startedAt) return '—';
    const elapsedMs = dashNow - new Date(startedAt).getTime();
    if (elapsedMs < 0 || !Number.isFinite(elapsedMs)) return '—';
    const s = Math.floor(elapsedMs / 1000);
    const days = Math.floor(s / 86400);
    const hours = Math.floor((s % 86400) / 3600);
    const mins = Math.floor((s % 3600) / 60);
    if (days > 0) return `${days}d ${hours}h ${mins}m`;
    if (hours > 0) return `${hours}h ${mins}m`;
    if (mins > 0) return `${mins}m`;
    return `${s}s`;
  });

  // ─── "Next run in …" countdown for a scheduled job ────────────────────────
  function dashCountdown(nextRunAt?: string): string {
    if (!nextRunAt) return '—';
    const ms = new Date(nextRunAt).getTime() - dashNow;
    if (ms <= 0) return 'now';
    const s = Math.floor(ms / 1000);
    if (s < 60) return `${s}s`;
    const m = Math.floor(s / 60);
    if (m < 60) return `${m}m`;
    const h = Math.floor(m / 60);
    return `${h}h ${m % 60}m`;
  }
  // Friendly interval label: "15m", "6h", "2d"
  function dashIntervalLabel(mins?: number): string {
    if (!mins || mins <= 0) return 'manual';
    if (mins < 60) return `${mins}m`;
    if (mins < 1440) return mins % 60 === 0 ? `${mins / 60}h` : `${Math.floor(mins / 60)}h ${mins % 60}m`;
    return `${Math.round(mins / 1440)}d`;
  }

  // ─── Inline scheduler editor (Scan only — backend supports syncIntervalMins) ─
  let scanIntervalEditing = $state(false);
  let scanIntervalDraft   = $state(15);
  let scanIntervalSaving  = $state(false);
  function openScanIntervalEditor() {
    scanIntervalDraft = dashJobs?.scan?.intervalMins ?? settingsData?.config?.syncIntervalMins ?? 15;
    scanIntervalEditing = true;
  }
  async function saveScanInterval() {
    const next = Math.max(5, Math.min(1440, Math.round(scanIntervalDraft)));
    scanIntervalSaving = true;
    try {
      const r = await updateSettings({ syncIntervalMins: next });
      settingsData = r; seedEditConfig(r);
      await refreshDashJobs();
      scanIntervalEditing = false;
    } catch { /* swallow — UI keeps editor open so user can retry */ }
    finally { scanIntervalSaving = false; }
  }

  async function refreshDashSys() {
    try {
      sysStatus = await getSystemStatus();
      dashUpdatedAt = new Date().toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    } catch { /* silent */ }
  }
  async function refreshDashJobs() {
    try {
      const [j, s] = await Promise.allSettled([getJobs(), getSessions()]);
      if (j.status === 'fulfilled') dashJobs = j.value;
      if (s.status === 'fulfilled') activeSessions = s.value.sessions ?? [];
    } catch { /* silent */ }
  }
  async function refreshDashHealth() {
    try { dashHealth = await getCatalogHealth(); } catch { /* silent */ }
  }
  async function refreshDashCodecs() {
    try { dashCodecs = await getCatalogCodecs(); } catch { /* silent */ }
  }

  // ─── Codec → playback-behaviour classification ────────────────────────────
  // The dashboard surfaces "1,200 H.264" as more than just a number — it tells
  // the user what each codec MEANS for playback (direct play vs needs remux
  // vs needs full transcode). These groupings reflect the actual playback
  // decision logic in server/internal/playback/decision.go (web profile).
  //
  // - direct: container-and-codec-compatible browsers play it untouched
  // - remux:  video stream is copied, only container is repackaged → fast
  // - audio:  video copies, audio re-encodes (HEVC + DTS = this path)
  // - transcode: video must be re-encoded — the only "slow" path
  type CodecClass = 'direct' | 'remux' | 'transcode';
  function classifyCodec(codec: string): { cls: CodecClass; label: string; explain: string } {
    const c = (codec || '').toLowerCase();
    // Direct-play on web (h264 in MP4/MOV/WebM is the baseline)
    if (c === 'h264') {
      return { cls: 'direct', label: 'Direct play', explain: 'Plays instantly — no server work.' };
    }
    // Remux territory — video stream is fine, server just repackages
    if (c === 'hevc' || c === 'av1' || c === 'vp9') {
      return { cls: 'remux', label: 'Fast remux', explain: 'Video stream is kept; the server only repackages the container. Starts in seconds.' };
    }
    // Everything else needs a full video re-encode
    return { cls: 'transcode', label: 'Needs transcoding', explain: 'No browser can decode this directly — the server has to re-encode the video, which is slow and CPU-heavy.' };
  }
  // Pretty codec name (same map the detail pages use, kept here to avoid
  // adding the util as a settings dep just for this one call).
  function dashCodecLabel(codec: string): string {
    const c = (codec || '').toLowerCase();
    const map: Record<string, string> = {
      h264: 'H.264', hevc: 'HEVC', av1: 'AV1', vp9: 'VP9', vp8: 'VP8',
      mpeg4: 'MPEG-4', mpeg2video: 'MPEG-2', mpeg1video: 'MPEG-1',
      wmv1: 'WMV', wmv2: 'WMV', wmv3: 'WMV', vc1: 'VC-1',
      rv30: 'RealVideo 3', rv40: 'RealVideo 4',
      flv1: 'Flash Video', h263: 'H.263',
      prores: 'ProRes', dnxhd: 'DNxHD', cinepak: 'Cinepak',
    };
    return map[c] ?? (codec || 'Unknown').toUpperCase();
  }
  async function handleDashScanNow() {
    dashScanBusy = true;
    try { await scanAllLibraries(); await refreshDashJobs(); } catch { /* ignore */ } finally { dashScanBusy = false; }
  }
  async function handleDashProbeNow() {
    dashProbeBusy = true;
    try { await startProbeJob(0); await refreshDashJobs(); } catch { /* ignore */ } finally { dashProbeBusy = false; }
  }

  // ─── Library CRUD state ────────────────────────────────────────────────────
  let libraries = $state<LibraryItem[]>([]);
  let libLoading = $state(false);
  let libSaving = $state(false);
  // Trailers toggle — stored as !disableTrailers so true = show trailers
  let trailersOn = $state(true);
  let trailersSaving = $state(false);
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

  // ─── Editable config snapshot ─────────────────────────────────────────────
  let editConfig = $state({
    serverName: '',
    canonicalWebOrigin: '',
    country: '',
    timezone: '',
    metadataLanguage: 'en-US',
    librarySyncMode: 'auto',
    syncIntervalMins: 60,
    watchDebounceSecs: 5,
    probeBatchLimit: 10,
    transcodeDir: '',
    downloadsDir: '',
    metadataDir: '',
    cacheDir: '',
    tempDir: '',
    hardwareUnlocked: false,
    playbackPolicy: 'original_only',
    preferTextSubtitles: false,
    originalQualityOnly: false,
    defaultSubtitlesMovies: false,
    defaultSubtitlesTV: false,
  });

  function seedEditConfig(s: SettingsResponse) {
    const c = s.config ?? {};
    editConfig = {
      serverName: c.serverName ?? '',
      canonicalWebOrigin: c.canonicalWebOrigin ?? '',
      country: c.country ?? '',
      timezone: c.timezone ?? '',
      metadataLanguage: c.metadataLanguage ?? 'en-US',
      librarySyncMode: c.librarySyncMode ?? 'auto',
      syncIntervalMins: c.syncIntervalMins ?? 60,
      watchDebounceSecs: c.watchDebounceSecs ?? 5,
      probeBatchLimit: c.probeBatchLimit ?? 10,
      transcodeDir: c.transcodeDir ?? '',
      downloadsDir: c.downloadsDir ?? '',
      metadataDir: c.metadataDir ?? '',
      cacheDir: c.cacheDir ?? '',
      tempDir: c.tempDir ?? '',
      hardwareUnlocked: c.hardwareUnlocked ?? false,
      playbackPolicy: c.playbackPolicy ?? 'original_only',
      preferTextSubtitles: c.preferTextSubtitles ?? false,
      originalQualityOnly: c.originalQualityOnly ?? false,
      defaultSubtitlesMovies: c.defaultSubtitlesMovies ?? false,
      defaultSubtitlesTV: c.defaultSubtitlesTV ?? false,
    };
    // Trailers toggle — separate reactive state (auto-saves, no Save bar needed)
    trailersOn = !(c.disableTrailers ?? false);
  }

  // ─── Per-section dirty checks ──────────────────────────────────────────────
  const generalDirty = $derived(
    editConfig.serverName        !== (settingsData?.config?.serverName        ?? '') ||
    editConfig.canonicalWebOrigin !== (settingsData?.config?.canonicalWebOrigin ?? '') ||
    editConfig.country           !== (settingsData?.config?.country           ?? '') ||
    editConfig.timezone          !== (settingsData?.config?.timezone          ?? '') ||
    editConfig.metadataLanguage  !== (settingsData?.config?.metadataLanguage  ?? 'en-US')
  );
  const scanningDirty = $derived(
    editConfig.librarySyncMode !== (settingsData?.config?.librarySyncMode ?? 'auto') ||
    editConfig.syncIntervalMins !== (settingsData?.config?.syncIntervalMins ?? 60) ||
    editConfig.watchDebounceSecs !== (settingsData?.config?.watchDebounceSecs ?? 5) ||
    editConfig.probeBatchLimit !== (settingsData?.config?.probeBatchLimit ?? 10)
  );
  const transcodingDirty = $derived(
    editConfig.hardwareUnlocked !== (settingsData?.config?.hardwareUnlocked ?? false)
  );
  const storageDirty = $derived(
    editConfig.transcodeDir !== (settingsData?.config?.transcodeDir ?? '') ||
    editConfig.downloadsDir !== (settingsData?.config?.downloadsDir ?? '') ||
    editConfig.metadataDir !== (settingsData?.config?.metadataDir ?? '') ||
    editConfig.cacheDir !== (settingsData?.config?.cacheDir ?? '') ||
    editConfig.tempDir !== (settingsData?.config?.tempDir ?? '')
  );
  const playbackDirty = $derived(
    editConfig.playbackPolicy         !== (settingsData?.config?.playbackPolicy         ?? 'original_only') ||
    editConfig.preferTextSubtitles    !== (settingsData?.config?.preferTextSubtitles    ?? false) ||
    editConfig.originalQualityOnly    !== (settingsData?.config?.originalQualityOnly    ?? false) ||
    editConfig.defaultSubtitlesMovies !== (settingsData?.config?.defaultSubtitlesMovies ?? false) ||
    editConfig.defaultSubtitlesTV     !== (settingsData?.config?.defaultSubtitlesTV     ?? false)
  );

  // ─── Metadata preferences state ───────────────────────────────────────────
  let editMetaPrefs = $state({ movie: [] as string[], series: [] as string[], movieArtwork: [] as string[], seriesArtwork: [] as string[] });
  let savedMetaPrefs = $state({ movie: [] as string[], series: [] as string[], movieArtwork: [] as string[], seriesArtwork: [] as string[] });
  const metaDirty = $derived(JSON.stringify(editMetaPrefs) !== JSON.stringify(savedMetaPrefs));

  // ── Metadata backfill (library-wide refresh of missing TMDB rows) ──────
  // Polls the server every 3s while running; stays quiet when idle.
  let backfill = $state<BackfillResponse | null>(null);
  let backfillError = $state<string | null>(null);
  let backfillPolling = $state<ReturnType<typeof setInterval> | null>(null);

  async function refreshBackfillStatus() {
    try {
      backfill = await getBackfillStatus();
      backfillError = null;
    } catch (e) {
      backfillError = e instanceof Error ? e.message : 'Could not load backfill status';
    }
  }
  async function triggerBackfill() {
    try {
      await startBackfill('tmdb');
      backfillError = null;
      await refreshBackfillStatus();
    } catch (e) {
      backfillError = e instanceof Error ? e.message : 'Could not start backfill';
    }
  }
  async function cancelBackfill() {
    try {
      await stopBackfill();
      await refreshBackfillStatus();
    } catch (e) {
      backfillError = e instanceof Error ? e.message : 'Could not stop backfill';
    }
  }

  // Start polling whenever the Metadata tab is open. Stop on leave/unmount.
  $effect(() => {
    if (active !== 'metadata') {
      if (backfillPolling) { clearInterval(backfillPolling); backfillPolling = null; }
      return;
    }
    refreshBackfillStatus();
    backfillPolling = setInterval(refreshBackfillStatus, 3000);
    return () => {
      if (backfillPolling) { clearInterval(backfillPolling); backfillPolling = null; }
    };
  });

  const backfillProgressPct = $derived.by(() => {
    const s = backfill?.status;
    if (!s || s.total === 0) return 0;
    const done = s.refreshed + s.failed;
    return Math.min(100, Math.round((done / s.total) * 100));
  });

  function moveMetaPref(list: keyof typeof editMetaPrefs, idx: number, dir: -1 | 1) {
    const arr = [...editMetaPrefs[list]];
    const swap = idx + dir;
    if (swap < 0 || swap >= arr.length) return;
    [arr[idx], arr[swap]] = [arr[swap], arr[idx]];
    editMetaPrefs = { ...editMetaPrefs, [list]: arr };
  }

  // ─── Per-section saving state ─────────────────────────────────────────────
  let generalSaving = $state(false);
  let scanningSaving = $state(false);
  let transcodingSaving = $state(false);
  let storageSaving = $state(false);
  let playbackSaving = $state(false);
  let metaSaving = $state(false);
  let sectionError = $state<string | null>(null);

  const sectionHasSaveDiscard = $derived(
    ['general','scanning','transcoding','storage','metadata','playback'].includes(active)
  );
  const currentDirty = $derived.by(() => {
    if (active === 'general') return generalDirty;
    if (active === 'scanning') return scanningDirty;
    if (active === 'transcoding') return transcodingDirty;
    if (active === 'storage') return storageDirty;
    if (active === 'playback') return playbackDirty;
    if (active === 'metadata') return metaDirty;
    return false;
  });
  const currentSaving = $derived.by(() => {
    if (active === 'general') return generalSaving;
    if (active === 'scanning') return scanningSaving;
    if (active === 'transcoding') return transcodingSaving;
    if (active === 'storage') return storageSaving;
    if (active === 'playback') return playbackSaving;
    if (active === 'metadata') return metaSaving;
    return false;
  });

  function discardSection() {
    if (!settingsData) return;
    const c = settingsData.config ?? {};
    sectionError = null;
    switch (active) {
      case 'general':
        editConfig.serverName       = c.serverName       ?? '';
        editConfig.canonicalWebOrigin = c.canonicalWebOrigin ?? '';
        editConfig.country          = c.country          ?? '';
        editConfig.timezone         = c.timezone         ?? '';
        editConfig.metadataLanguage = c.metadataLanguage ?? 'en-US';
        break;
      case 'scanning':
        editConfig.librarySyncMode = c.librarySyncMode ?? 'auto';
        editConfig.syncIntervalMins = c.syncIntervalMins ?? 60;
        editConfig.watchDebounceSecs = c.watchDebounceSecs ?? 5;
        editConfig.probeBatchLimit = c.probeBatchLimit ?? 10;
        break;
      case 'transcoding': editConfig.hardwareUnlocked = c.hardwareUnlocked ?? false; break;
      case 'storage':
        editConfig.transcodeDir = c.transcodeDir ?? '';
        editConfig.downloadsDir = c.downloadsDir ?? '';
        editConfig.metadataDir = c.metadataDir ?? '';
        editConfig.cacheDir = c.cacheDir ?? '';
        editConfig.tempDir = c.tempDir ?? '';
        break;
      case 'playback':
        editConfig.playbackPolicy         = c.playbackPolicy         ?? 'auto';
        editConfig.preferTextSubtitles    = c.preferTextSubtitles    ?? false;
        editConfig.originalQualityOnly    = c.originalQualityOnly    ?? false;
        editConfig.defaultSubtitlesMovies = c.defaultSubtitlesMovies ?? false;
        editConfig.defaultSubtitlesTV     = c.defaultSubtitlesTV     ?? false;
        break;
      case 'metadata': editMetaPrefs = { ...savedMetaPrefs }; break;
    }
  }

  async function saveSection() {
    sectionError = null;
    try {
      if (active === 'general') {
        generalSaving = true;
        const r = await updateSettings({
          serverName:       editConfig.serverName,
          canonicalWebOrigin: editConfig.canonicalWebOrigin || '',
          country:          editConfig.country          || '',
          timezone:         editConfig.timezone         || '',
          metadataLanguage: editConfig.metadataLanguage || 'en-US',
        });
        settingsData = r; seedEditConfig(r);
      } else if (active === 'scanning') {
        scanningSaving = true;
        const r = await updateSettings({
          librarySyncMode: editConfig.librarySyncMode,
          syncIntervalMins: editConfig.syncIntervalMins,
          watchDebounceSecs: editConfig.watchDebounceSecs,
          probeBatchLimit: editConfig.probeBatchLimit,
        });
        settingsData = r; seedEditConfig(r);
      } else if (active === 'transcoding') {
        transcodingSaving = true;
        const r = await updateSettings({ hardwareUnlocked: editConfig.hardwareUnlocked });
        settingsData = r; seedEditConfig(r);
      } else if (active === 'storage') {
        storageSaving = true;
        const r = await updateSettings({
          transcodeDir: editConfig.transcodeDir,
          downloadsDir: editConfig.downloadsDir,
          metadataDir: editConfig.metadataDir,
          cacheDir: editConfig.cacheDir,
          tempDir: editConfig.tempDir,
        });
        settingsData = r; seedEditConfig(r);
      } else if (active === 'playback') {
        playbackSaving = true;
        const r = await updateSettings({
          playbackPolicy:         editConfig.playbackPolicy,
          preferTextSubtitles:    editConfig.preferTextSubtitles,
          originalQualityOnly:    editConfig.originalQualityOnly,
          defaultSubtitlesMovies: editConfig.defaultSubtitlesMovies,
          defaultSubtitlesTV:     editConfig.defaultSubtitlesTV,
        });
        settingsData = r; seedEditConfig(r);
      } else if (active === 'metadata') {
        metaSaving = true;
        const r = await updateMetadataSourcePreferences(editMetaPrefs);
        settingsData = r;
        savedMetaPrefs = { ...editMetaPrefs };
      }
    } catch (e) {
      sectionError = e instanceof Error ? e.message : 'Failed to save';
    } finally {
      generalSaving = false; scanningSaving = false; transcodingSaving = false;
      storageSaving = false; playbackSaving = false; metaSaving = false;
    }
  }

  // ─── Performance / transcoding state ──────────────────────────────────────
  let perfSettings = $state<PerformanceSettingsResponse | null>(null);
  let perfLoading = $state(false);
  let hwTestResult = $state<HardwareTestResponse | null>(null);
  let hwTestRunning = $state(false);

  async function loadPerf() {
    perfLoading = true;
    try { perfSettings = await getPerformanceSettings(); } catch { /* ignore */ } finally { perfLoading = false; }
  }

  async function runHwTest() {
    hwTestRunning = true;
    hwTestResult = null;
    try { hwTestResult = await runHardwareTest(); } catch (e) {
      hwTestResult = { status: 'error', error: e instanceof Error ? e.message : 'Test failed' };
    } finally { hwTestRunning = false; }
  }

  // ─── Discovery / network state ────────────────────────────────────────────
  let discoveryStatus = $state<DiscoveryStatusResponse | null>(null);
  let discoveryLoading = $state(false);
  async function loadDiscovery() {
    discoveryLoading = true;
    try { discoveryStatus = await getDiscoveryStatus(); } catch { /* ignore */ } finally { discoveryLoading = false; }
  }

  // ─── Users state ──────────────────────────────────────────────────────────
  let usersList = $state<UserItem[]>([]);
  let usersLoading = $state(false);
  let usersError = $state<string | null>(null);
  let showAddUser = $state(false);
  let newUserName = $state('');
  let newUserDisplay = $state('');
  let newUserPass = $state('');
  let newUserRole = $state<'admin' | 'viewer'>('viewer');
  let userSaving = $state(false);
  let userDeletingId = $state<string | null>(null);

  async function loadUsers() {
    usersLoading = true; usersError = null;
    try { usersList = (await getUsers()).users ?? []; } catch (e) {
      usersError = e instanceof Error ? e.message : 'Failed to load users';
    } finally { usersLoading = false; }
  }

  async function handleCreateUser() {
    if (!newUserName.trim() || !newUserPass.trim()) return;
    userSaving = true; usersError = null;
    try {
      const u = await createUser({ username: newUserName.trim(), displayName: newUserDisplay.trim() || undefined, password: newUserPass, role: newUserRole });
      usersList = [...usersList, u];
      showAddUser = false; newUserName = ''; newUserDisplay = ''; newUserPass = ''; newUserRole = 'viewer';
    } catch (e) { usersError = e instanceof Error ? e.message : 'Failed to create user'; }
    finally { userSaving = false; }
  }

  async function handleDeleteUser(id: string) {
    if (userDeletingId !== id) { userDeletingId = id; return; }
    try { await deleteUser(id); usersList = usersList.filter(u => u.id !== id); }
    catch (e) { usersError = e instanceof Error ? e.message : 'Failed to delete user'; }
    finally { userDeletingId = null; }
  }

  // ─── Profile settings editing ──────────────────────────────────────────────
  let editingProfileUserId = $state<string | null>(null);
  let profileEditName = $state('');
  let profileEditPreset = $state('');
  let profileEditColor = $state('');
  let profileEditRestricted = $state(false);
  let profileEditMaxRating = $state('');
  let profileSaving = $state(false);
  let profileSaveError = $state<string | null>(null);

  // PIN editing
  let pinEditUserId = $state<string | null>(null);
  let pinEditValue = $state('');
  let pinSaving = $state(false);
  let pinSaveError = $state<string | null>(null);

  function startEditProfile(user: typeof usersList[number]) {
    editingProfileUserId = user.id ?? null;
    profileEditName = user.displayName ?? '';
    profileEditPreset = user.avatarPreset ?? '';
    profileEditColor = user.avatarColor ?? '';
    profileEditRestricted = user.isRestricted ?? false;
    profileEditMaxRating = user.maxRating ?? '';
    profileSaveError = null;
    pinEditUserId = null;
  }

  function cancelEditProfile() {
    editingProfileUserId = null;
    profileSaveError = null;
  }

  async function handleSaveProfile(userId: string) {
    profileSaving = true; profileSaveError = null;
    try {
      const resp = await updateProfileSettings(userId, {
        displayName: profileEditName,
        avatarPreset: profileEditPreset || undefined,
        avatarColor: profileEditColor || undefined,
        isRestricted: profileEditRestricted,
        maxRating: profileEditMaxRating || undefined,
      });
      const updated = (resp as { user: typeof usersList[number] }).user;
      usersList = usersList.map(u => u.id === userId ? { ...u, ...updated } : u);
      editingProfileUserId = null;
    } catch (e) {
      profileSaveError = e instanceof Error ? e.message : 'Failed to save profile';
    } finally {
      profileSaving = false;
    }
  }

  function startEditPin(userId: string) {
    pinEditUserId = userId;
    pinEditValue = '';
    pinSaveError = null;
    editingProfileUserId = null;
  }

  function cancelEditPin() {
    pinEditUserId = null;
    pinSaveError = null;
  }

  async function handleSavePin(userId: string) {
    pinSaving = true; pinSaveError = null;
    try {
      // Only digits; strip accidental spaces.
      const pin = pinEditValue.replace(/\D/g, '');
      if (pin && (pin.length < 4 || pin.length > 8)) {
        pinSaveError = 'PIN must be 4–8 digits (leave blank to remove).';
        return;
      }
      await setProfilePin(userId, pin);
      usersList = usersList.map(u => u.id === userId ? { ...u, hasPin: pin.length > 0 } : u);
      pinEditUserId = null;
    } catch (e) {
      pinSaveError = e instanceof Error ? e.message : 'Failed to save PIN';
    } finally {
      pinSaving = false;
    }
  }

  // ─── Pairing requests state ───────────────────────────────────────────────
  let pairingRequests = $state<PairingRequestItem[]>([]);
  let pairingLoading = $state(false);
  let pairingLoaded = $state(false);
  let pairingActionId = $state<string | null>(null);

  async function loadPairingRequests() {
    pairingLoading = true;
    try { pairingRequests = (await getPairingRequests()).requests ?? []; } catch { /* ignore */ }
    finally { pairingLoading = false; pairingLoaded = true; }
  }

  async function handleApprove(id: string) {
    pairingActionId = id;
    try { await approvePairingRequest(id); pairingRequests = pairingRequests.filter(r => r.id !== id); }
    catch { /* ignore */ } finally { pairingActionId = null; }
  }

  async function handleDeny(id: string) {
    pairingActionId = id;
    try { await denyPairingRequest(id); pairingRequests = pairingRequests.filter(r => r.id !== id); }
    catch { /* ignore */ } finally { pairingActionId = null; }
  }

  // ─── QR pair token state ──────────────────────────────────────────────────
  let qrToken = $state<QRTokenResponse | null>(null);
  let qrGenerating = $state(false);
  let qrSecondsLeft = $state(0);
  let qrTimerInterval = $state<ReturnType<typeof setInterval> | null>(null);

  async function generateQR() {
    qrGenerating = true;
    try {
      const resp = await generateQRPairToken();
      qrToken = resp;
      const expiresMs = new Date(resp.expiresAt).getTime() - Date.now();
      qrSecondsLeft = Math.max(0, Math.round(expiresMs / 1000));
      if (qrTimerInterval) clearInterval(qrTimerInterval);
      qrTimerInterval = setInterval(() => {
        qrSecondsLeft = Math.max(0, qrSecondsLeft - 1);
        if (qrSecondsLeft === 0) { clearInterval(qrTimerInterval!); qrToken = null; }
      }, 1000);
    } catch { /* ignore */ } finally { qrGenerating = false; }
  }

  function dismissQR() {
    if (qrTimerInterval) clearInterval(qrTimerInterval);
    qrToken = null;
  }

  // ─── Approved devices state ───────────────────────────────────────────────
  let approvedDevices = $state<ApprovedDeviceItem[]>([]);
  let devicesLoading = $state(false);
  let devicesLoaded = $state(false);
  let deviceRevokingId = $state<string | null>(null);

  async function loadApprovedDevices() {
    devicesLoading = true;
    try { approvedDevices = (await getApprovedDevices()).devices ?? []; } catch { /* ignore */ }
    finally { devicesLoading = false; devicesLoaded = true; }
  }

  async function handleRevoke(id: string) {
    if (deviceRevokingId !== id) { deviceRevokingId = id; return; }
    try { await revokeApprovedDevice(id); approvedDevices = approvedDevices.filter(d => d.id !== id); }
    catch { /* ignore */ } finally { deviceRevokingId = null; }
  }

  // ─── Device profiles state ────────────────────────────────────────────────
  let deviceProfiles = $state<DeviceProfile[]>([]);

  async function loadDeviceProfiles() {
    try { deviceProfiles = (await getDeviceProfiles()).profiles ?? []; } catch { /* ignore */ }
  }

  // ─── Scan-all ─────────────────────────────────────────────────────────────
  let scanAllRunning = $state(false);
  async function handleScanAll() {
    scanAllRunning = true;
    try { await scanAllLibraries(); } catch { /* ignore */ } finally { scanAllRunning = false; }
  }

  // ─── Storage folder browser context ───────────────────────────────────────
  let browserContext = $state<'library' | 'storage'>('library');
  let storageBrowserField = $state('');

  async function openStorageBrowser(field: string) {
    browserContext = 'storage';
    storageBrowserField = field;
    showFolderBrowser = true;
    const currentVal = editConfig[field as keyof typeof editConfig] as string;
    await navigateFolder(currentVal || '');
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
  // Binary (1024-based) scale — files, RAM, disk. PB ceiling because
  // beyond that we're in storage-array territory and 73 PB displaying as
  // "73 PB" reads better than "73000 TB" anyway.
  function formatBytes(bytes?: number): string {
    if (bytes === undefined || bytes === null) return '—';
    const PB = 1_125_899_906_842_624;     // 1024^5
    const TB = 1_099_511_627_776;          // 1024^4
    const GB = 1_073_741_824;              // 1024^3
    const MB = 1_048_576;
    const KB = 1_024;
    if (bytes >= PB) return `${(bytes / PB).toFixed(2)} PB`;
    if (bytes >= TB) return `${(bytes / TB).toFixed(2)} TB`;
    if (bytes >= GB) return `${(bytes / GB).toFixed(1)} GB`;
    if (bytes >= MB) return `${(bytes / MB).toFixed(0)} MB`;
    if (bytes >= KB) return `${(bytes / KB).toFixed(0)} KB`;
    return `${bytes} B`;
  }

  // SI (1000-based) scale — network throughput conventionally uses decimal
  // units. GB/s included for completeness on fast LAN / 10GbE.
  function formatBps(bps?: number): string {
    if (!bps) return '0 B/s';
    if (bps >= 1_000_000_000) return `${(bps / 1_000_000_000).toFixed(2)} GB/s`;
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
        seedEditConfig(settingsRes.value);
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

  async function toggleTrailers() {
    const next = !trailersOn;
    trailersOn = next;
    appState.trailersEnabled = next; // propagate immediately to live hero
    trailersSaving = true;
    try {
      const r = await updateSettings({ disableTrailers: !next });
      settingsData = r;
    } catch {
      // Revert optimistic update on error
      trailersOn = !next;
      appState.trailersEnabled = !next;
    } finally {
      trailersSaving = false;
    }
  }

  // ─── Folder browser ────────────────────────────────────────────────────────
  async function openFolderBrowser() {
    browserContext = 'library';
    storageBrowserField = '';
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

  // Seed edit config whenever settingsData changes
  $effect(() => {
    if (settingsData) {
      seedEditConfig(settingsData);
      const prefs = settingsData.metadataSourcePreferences ?? {};
      const prefObj = {
        movie: [...(prefs.movie ?? [])],
        series: [...(prefs.series ?? [])],
        movieArtwork: [...(prefs.movieArtwork ?? [])],
        seriesArtwork: [...(prefs.seriesArtwork ?? [])],
      };
      editMetaPrefs = prefObj;
      savedMetaPrefs = prefObj;
    }
  });

  // Lazy-load section data on first visit
  $effect(() => {
    if (active === 'transcoding' && !perfSettings && !perfLoading) loadPerf();
    if (active === 'network' && !discoveryStatus && !discoveryLoading) loadDiscovery();
    if (active === 'users' && usersList.length === 0 && !usersLoading) loadUsers();
    if (active === 'pending-approvals' && !pairingLoaded && !pairingLoading) loadPairingRequests();
    if (active === 'approved-devices' && !devicesLoaded && !devicesLoading) loadApprovedDevices();
    if (active === 'playback' && deviceProfiles.length === 0) loadDeviceProfiles();
    if ((active === 'transcoding' || active === 'playback' || active === 'scanning') && !perfSettings && !perfLoading) loadPerf();
    if (active !== current.id) sectionError = null;
  });

  let currentUser = $state<AuthSessionUser | null>(null);

  // ─── Self-service Account section state ──────────────────────────────────
  let acctDisplayName    = $state('');
  let acctProfileSaving  = $state(false);
  let acctProfileError   = $state<string | null>(null);
  let acctProfileSuccess = $state(false);

  let acctCurrentPw  = $state('');
  let acctNewPw      = $state('');
  let acctConfirmPw  = $state('');
  let acctPwSaving   = $state(false);
  let acctPwError    = $state<string | null>(null);
  let acctPwSuccess  = $state(false);

  // Seed the display name input once the session loads.
  $effect(() => {
    if (currentUser && !acctDisplayName) {
      acctDisplayName = currentUser.displayName ?? '';
    }
  });

  async function handleAccountSaveProfile() {
    if (!currentUser) return;
    acctProfileSaving = true;
    acctProfileError = null;
    acctProfileSuccess = false;
    try {
      const updated = await updateUser(currentUser.id, { displayName: acctDisplayName.trim() || currentUser.username });
      currentUser = { ...currentUser, displayName: updated.displayName ?? currentUser.displayName };
      acctProfileSuccess = true;
      setTimeout(() => { acctProfileSuccess = false; }, 3000);
    } catch (e) {
      acctProfileError = e instanceof Error ? e.message : 'Failed to save profile';
    } finally {
      acctProfileSaving = false;
    }
  }

  async function handleAccountUpdatePassword() {
    if (!currentUser) return;
    acctPwError = null;
    acctPwSuccess = false;
    if (!acctCurrentPw) { acctPwError = 'Enter your current password.'; return; }
    if (!acctNewPw)      { acctPwError = 'Enter a new password.'; return; }
    if (acctNewPw.length < 8) { acctPwError = 'New password must be at least 8 characters.'; return; }
    if (acctNewPw !== acctConfirmPw) { acctPwError = 'Passwords do not match.'; return; }
    acctPwSaving = true;
    try {
      await updateUserPassword(currentUser.id, { currentPassword: acctCurrentPw, password: acctNewPw });
      acctCurrentPw = '';
      acctNewPw = '';
      acctConfirmPw = '';
      acctPwSuccess = true;
      setTimeout(() => { acctPwSuccess = false; }, 3000);
    } catch (e) {
      acctPwError = e instanceof Error ? e.message : 'Failed to update password';
    } finally {
      acctPwSaving = false;
    }
  }

  async function handleSignOut() {
    try { await logout(); } catch { /* ignore */ }
    localStorage.removeItem('xuva-auth-token');
    localStorage.removeItem('xuva-profile-token');
    window.location.href = '/';
  }

  $effect(() => {
    const unsub = dashStream.subscribeAny(({ type }) => {
      if (
        type.startsWith('automation.') || type.startsWith('scan.') ||
        type.startsWith('probe.')      || type.startsWith('metadata.') ||
        type === 'api.session.accepted' || type === 'api.session.stopped'
      ) {
        refreshDashJobs();
      }
    });
    return unsub;
  });

  onMount(() => {
    loadDashboard();
    loadLibraries();
    getAuthSession().then(r => { if (r.user) currentUser = r.user; }).catch(() => {});
    // Dashboard live extras
    refreshDashSys();
    refreshDashJobs();
    refreshDashHealth();
    refreshDashCodecs();
    dashStream.connect();
    dashPollTimers = [
      setInterval(refreshDashSys,    5_000),
      setInterval(refreshDashJobs,   5_000),
      setInterval(refreshDashHealth, 60_000),
      setInterval(refreshDashCodecs, 120_000),
      // 1-second ticker drives the live uptime + countdown displays; pure
      // client-side, no network traffic.
      setInterval(() => { dashNow = Date.now(); }, 1_000),
    ];
  });

  onDestroy(() => {
    dashStream.disconnect();
    for (const t of dashPollTimers) clearInterval(t);
  });
</script>

<svelte:head>
  <title>Settings — {appState.serverName}</title>
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

  <main class="px-6 pb-32 pt-20 md:px-12 md:pt-24 lg:px-20">
    <header class="relative mb-4">
      <div
        aria-hidden="true"
        class="pointer-events-none absolute -inset-x-6 -top-6 -z-10 h-[140px] opacity-60 md:-inset-x-12 lg:-inset-x-20"
        style="background: radial-gradient(50% 100% at 15% 0%, oklch(0.62 0.22 285 / 0.25), transparent 70%), radial-gradient(40% 100% at 90% 0%, oklch(0.72 0.16 255 / 0.18), transparent 70%);"
      ></div>
      <div class="flex flex-wrap items-end justify-between gap-4">
        <div>
          <div class="mb-1 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">
            Settings
          </div>
          <h1 class="font-serif-display text-[clamp(1.5rem,3vw,2.25rem)] leading-[1] tracking-tight">
            {#if serverNameParts.first}{serverNameParts.first}-{/if}<em>{serverNameParts.last}</em>
          </h1>
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

    <div class="grid gap-6 lg:grid-cols-[260px_minmax(0,1fr)] lg:gap-10">
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
        <!-- Section header bar: pinned to min-h-[3.25rem] so it stays the
             exact same height whether the section has Save/Discard buttons
             (general/scanning/transcoding/storage/metadata/playback) or
             not (dashboard/libraries/users/etc.). Previously it bounced
             ~12px between sections because the button-bearing rows are
             taller than the bare H2 row. -->
        <div
          class={`sticky top-16 z-20 -mx-6 mb-4 flex min-h-[3.25rem] items-center justify-between gap-4 border-b px-6 py-2 backdrop-blur-xl transition-colors md:top-[72px] md:-mx-12 md:px-12 lg:-mx-0 lg:px-0 ${
            headerScrolled
              ? "border-border bg-background/80"
              : "border-transparent bg-transparent"
          }`}
        >
          <div class="min-w-0 flex items-baseline gap-3">
            <span class="text-[10px] font-semibold uppercase tracking-[0.3em] text-muted-foreground/70 shrink-0">
              {current.group}
            </span>
            <h2 class="font-serif-display truncate text-lg tracking-tight leading-none">
              {current.label}
            </h2>
            {#if current.hint}
              <span class="hidden md:inline truncate text-xs text-muted-foreground/60">
                · {current.hint}
              </span>
            {/if}
          </div>
          {#if sectionHasSaveDiscard}
            <div class="flex items-center gap-2">
              {#if sectionError}
                <span class="max-w-[180px] truncate text-[11px] text-red-300">{sectionError}</span>
              {/if}
              <button
                type="button"
                onclick={discardSection}
                disabled={!currentDirty || currentSaving}
                class="hairline rounded-full bg-foreground/[0.04] px-4 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40"
              >
                Discard
              </button>
              <button
                type="button"
                onclick={saveSection}
                disabled={!currentDirty || currentSaving}
                class="inline-flex items-center gap-2 rounded-full bg-gradient-primary px-5 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {#if currentSaving}
                  <span class="h-3 w-3 animate-spin rounded-full border border-white/30 border-t-white"></span>
                {/if}
                Save
              </button>
            </div>
          {/if}
        </div>

        {#if active === "metadata"}
          {@const movieSources = settingsData?.metadataSources?.movie ?? []}
          {@const seriesSources = settingsData?.metadataSources?.series ?? []}
          {@const allSources = [...movieSources, ...seriesSources].filter((s, i, arr) => arr.findIndex(x => x.id === s.id) === i)}
          {@const bfs = backfill?.status}
          {@const bfRunning = bfs?.running === true}
          {@const bfMissing = backfill?.missingTotal ?? 0}
          {@const bfTotal = bfs?.total ?? bfMissing}
          <div class="space-y-12">

            <!-- Library health: missing-metadata backfill -->
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Library health</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Xuva backfills missing TMDB metadata automatically in the background. You can also trigger it manually.
                </p>
              </div>
              <div class="hairline rounded-2xl bg-surface/40 p-5 space-y-4">
                {#if bfRunning}
                  <div class="flex items-start justify-between gap-4">
                    <div class="min-w-0">
                      <div class="flex items-center gap-2">
                        <span class="inline-block h-2 w-2 rounded-full bg-primary-glow animate-pulse"></span>
                        <span class="font-semibold">Backfilling {bfs?.provider ?? 'metadata'} ({bfs?.kind ?? '…'})</span>
                      </div>
                      <p class="mt-1 text-xs text-muted-foreground">
                        {bfs?.refreshed ?? 0} refreshed
                        {#if (bfs?.failed ?? 0) > 0}· {bfs?.failed} failed{/if}
                        {#if (bfs?.remaining ?? 0) > 0}· {bfs?.remaining} remaining{/if}
                      </p>
                      {#if bfs?.lastTitle}
                        <p class="mt-1 truncate text-xs text-muted-foreground/80">Last: {bfs.lastTitle}</p>
                      {/if}
                    </div>
                    <button
                      type="button"
                      onclick={cancelBackfill}
                      class="hairline shrink-0 rounded-full bg-foreground/[0.05] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-red-400/10 hover:text-red-300"
                    >
                      Stop
                    </button>
                  </div>
                  <div class="h-1.5 overflow-hidden rounded-full bg-foreground/[0.06]">
                    <div class="h-full rounded-full bg-gradient-primary transition-all duration-500" style={`width: ${Math.max(backfillProgressPct, 2)}%`}></div>
                  </div>
                {:else}
                  <div class="flex flex-wrap items-start justify-between gap-4">
                    <div class="min-w-0">
                      {#if bfMissing === 0}
                        <div class="flex items-center gap-2 text-emerald-300">
                          <Check class="h-4 w-4" />
                          <span class="font-semibold">All items have TMDB metadata</span>
                        </div>
                        <p class="mt-1 text-xs text-muted-foreground">Everything in your library is enriched. Nothing to backfill.</p>
                      {:else}
                        <div class="flex items-center gap-2">
                          <span class="inline-block h-2 w-2 rounded-full bg-amber-300"></span>
                          <span class="font-semibold">{bfMissing} item{bfMissing === 1 ? '' : 's'} missing TMDB metadata</span>
                        </div>
                        <p class="mt-1 text-xs text-muted-foreground">
                          {backfill?.missingMovies ?? 0} movies · {backfill?.missingSeries ?? 0} series
                        </p>
                        {#if bfs?.finishedAt && (bfs?.refreshed ?? 0) + (bfs?.failed ?? 0) > 0}
                          <p class="mt-1 text-xs text-muted-foreground/80">
                            Last run: {bfs.refreshed} refreshed
                            {#if (bfs?.failed ?? 0) > 0}, {bfs?.failed} failed{/if}
                          </p>
                        {/if}
                      {/if}
                    </div>
                    {#if bfMissing > 0}
                      <button
                        type="button"
                        onclick={triggerBackfill}
                        class="shrink-0 inline-flex items-center gap-2 rounded-full bg-foreground px-4 py-2 text-xs font-semibold text-background transition-all hover:bg-foreground/90"
                      >
                        Refresh missing metadata
                      </button>
                    {/if}
                  </div>
                {/if}
                {#if backfillError}
                  <p class="text-xs text-red-300">{backfillError}</p>
                {/if}
              </div>
            </section>

            <!-- Provider health -->
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Provider status</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Metadata sources configured on this server and their current health.
                </p>
              </div>
              <div class="space-y-3">
                {#if allSources.length === 0}
                  <p class="text-sm text-muted-foreground">No providers found. Reload settings to check.</p>
                {/if}
                {#each allSources as src (src.id)}
                  {@const h = src.providerHealth}
                  {@const healthy = h?.healthy === true}
                  {@const configured = h?.configured === true}
                  <div class="hairline flex items-start gap-4 rounded-2xl bg-surface/40 p-5">
                    <div class="min-w-0 flex-1">
                      <div class="flex flex-wrap items-center gap-2">
                        <span class="font-semibold">{src.name ?? src.id}</span>
                        {#if src.coverage}
                          <span class="rounded-full bg-foreground/[0.06] px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-muted-foreground">{src.coverage}</span>
                        {/if}
                        {#if healthy}
                          <span class="hairline inline-flex items-center gap-1.5 rounded-full bg-emerald-400/10 px-2.5 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-emerald-300"><Check class="h-3 w-3"/>Healthy</span>
                        {:else if configured}
                          <span class="hairline inline-flex items-center gap-1.5 rounded-full bg-amber-400/10 px-2.5 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-amber-300">Degraded</span>
                        {:else if src.local}
                          <span class="hairline inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.06] px-2.5 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-muted-foreground">Built in</span>
                        {:else}
                          <span class="hairline inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.06] px-2.5 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-muted-foreground">Not configured</span>
                        {/if}
                      </div>
                      {#if src.description}
                        <p class="mt-0.5 text-xs text-muted-foreground">{src.description}</p>
                      {/if}
                      {#if h?.error}
                        <p class="mt-1 text-xs text-red-300">{h.error}</p>
                      {/if}
                    </div>
                    <div class="shrink-0 text-right text-[10px] uppercase tracking-[0.15em] text-muted-foreground/60">
                      {#if src.supportsMetadata && src.supportsArtwork}Metadata + Artwork
                      {:else if src.supportsMetadata}Metadata only
                      {:else if src.supportsArtwork}Artwork only{/if}
                    </div>
                  </div>
                {/each}
              </div>
            </section>

            <!-- Provider order — Movies -->
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Movie provider order</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Xuva tries providers from top to bottom when fetching metadata for movies.
                </p>
              </div>
              <div class="space-y-2">
                {#each editMetaPrefs.movie as pid, i (pid)}
                  {@const src = movieSources.find(s => s.id === pid)}
                  <div class="hairline flex items-center gap-3 rounded-xl bg-surface/40 px-4 py-3">
                    <span class="w-5 text-center font-mono text-xs text-muted-foreground/60">{i + 1}</span>
                    <span class="flex-1 text-sm font-medium">{src?.name ?? pid}</span>
                    <button type="button" disabled={i === 0} onclick={() => moveMetaPref('movie', i, -1)}
                      class="rounded p-1 text-muted-foreground transition-colors hover:text-foreground disabled:opacity-30">▲</button>
                    <button type="button" disabled={i === editMetaPrefs.movie.length - 1} onclick={() => moveMetaPref('movie', i, 1)}
                      class="rounded p-1 text-muted-foreground transition-colors hover:text-foreground disabled:opacity-30">▼</button>
                  </div>
                {:else}
                  <p class="text-sm text-muted-foreground">No movie providers configured.</p>
                {/each}
              </div>
            </section>

            <!-- Provider order — TV -->
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">TV show provider order</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Provider preference order for TV series, seasons, and episodes.
                </p>
              </div>
              <div class="space-y-2">
                {#each editMetaPrefs.series as pid, i (pid)}
                  {@const src = seriesSources.find(s => s.id === pid)}
                  <div class="hairline flex items-center gap-3 rounded-xl bg-surface/40 px-4 py-3">
                    <span class="w-5 text-center font-mono text-xs text-muted-foreground/60">{i + 1}</span>
                    <span class="flex-1 text-sm font-medium">{src?.name ?? pid}</span>
                    <button type="button" disabled={i === 0} onclick={() => moveMetaPref('series', i, -1)}
                      class="rounded p-1 text-muted-foreground transition-colors hover:text-foreground disabled:opacity-30">▲</button>
                    <button type="button" disabled={i === editMetaPrefs.series.length - 1} onclick={() => moveMetaPref('series', i, 1)}
                      class="rounded p-1 text-muted-foreground transition-colors hover:text-foreground disabled:opacity-30">▼</button>
                  </div>
                {:else}
                  <p class="text-sm text-muted-foreground">No TV providers configured.</p>
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
                                autocomplete="new-password"
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
                    <input
                      type="checkbox"
                      checked={wlSyncEdit[opt.id] ?? true}
                      onchange={(e) => { wlSyncEdit = { ...wlSyncEdit, [opt.id]: (e.currentTarget as HTMLInputElement).checked }; }}
                      class="mt-0.5 h-4 w-4 cursor-pointer rounded accent-primary"
                    />
                    <div>
                      <div class="text-sm font-medium">{opt.label}</div>
                      <p class="text-xs text-muted-foreground">{opt.desc}</p>
                    </div>
                  </label>
                {/each}

                {#if wlSyncDirty}
                  <div class="flex items-center gap-3 pt-2">
                    <button
                      type="button"
                      onclick={saveWlSync}
                      class="rounded-full bg-gradient-primary px-5 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110"
                    >
                      Save
                    </button>
                    <button
                      type="button"
                      onclick={discardWlSync}
                      class="hairline rounded-full bg-foreground/[0.04] px-5 py-2 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground"
                    >
                      Discard
                    </button>
                  </div>
                {/if}
              </div>
            </section>
          </div>

        {:else if active === "dashboard"}
          <!-- ════════════════════════════ FLIGHT DECK ════════════════════════════ -->

          <!-- ── Command Header ──────────────────────────────────────────────── -->
          <div class="mb-4 flex flex-wrap items-center justify-between gap-3 rounded-xl border px-5 py-3.5"
            style="border-color: oklch(0.74 0.2 280 / 0.14); background: linear-gradient(135deg, oklch(0.16 0.025 285 / 0.7) 0%, oklch(0.12 0.018 285 / 0.5) 100%);">
            <div class="flex flex-wrap items-center gap-4">
              <!-- System status beacon -->
              <div class="flex items-center gap-2.5">
                <div class="relative flex h-2.5 w-2.5 shrink-0">
                  {#if dashHudStatusStr === 'NOMINAL'}
                    <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-50"></span>
                    <span class="relative inline-flex h-2.5 w-2.5 rounded-full bg-emerald-400 shadow-[0_0_14px_oklch(0.78_0.2_145)]"></span>
                  {:else if dashHudStatusStr === 'CAUTION'}
                    <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-amber-400 opacity-60"></span>
                    <span class="relative inline-flex h-2.5 w-2.5 animate-pulse rounded-full bg-amber-400 shadow-[0_0_14px_oklch(0.85_0.22_75)]"></span>
                  {:else if dashHudStatusStr === 'FAULT'}
                    <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-red-500 opacity-60"></span>
                    <span class="relative inline-flex h-2.5 w-2.5 animate-pulse rounded-full bg-red-500 shadow-[0_0_16px_oklch(0.68_0.26_22)]"></span>
                  {:else}
                    <span class="relative inline-flex h-2.5 w-2.5 rounded-full bg-foreground/20"></span>
                  {/if}
                </div>
                <span class="text-[13px] font-medium tracking-tight {dashHudStatusStr === 'NOMINAL' ? 'text-emerald-300' : dashHudStatusStr === 'CAUTION' ? 'text-amber-300' : dashHudStatusStr === 'FAULT' ? 'text-red-300' : 'text-muted-foreground'}"
                  style={dashHudStatusStr === 'NOMINAL' ? 'text-shadow: 0 0 10px oklch(0.78 0.2 145 / 0.32);'
                       : dashHudStatusStr === 'CAUTION' ? 'text-shadow: 0 0 10px oklch(0.85 0.22 75 / 0.4);'
                       : dashHudStatusStr === 'FAULT' ? 'text-shadow: 0 0 10px oklch(0.68 0.26 22 / 0.5);' : ''}>
                  {#if dashHudStatusStr === 'NOMINAL'}All systems healthy
                  {:else if dashHudStatusStr === 'CAUTION'}Heads up — system under load
                  {:else if dashHudStatusStr === 'FAULT'}Critical — system overloaded
                  {:else}Checking…{/if}
                </span>
              </div>
              <div class="h-3.5 w-px bg-foreground/10 hidden sm:block"></div>
              <div class="text-[12px] text-foreground/55 tracking-tight">
                {settingsData?.config?.serverName ?? 'Media Server'} <span class="mx-1.5 text-foreground/30">·</span>
                <span class="italic text-foreground/50">updated</span> <span class="text-foreground/70 tabular-nums">{dashUpdatedAt || '——:——:——'}</span>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <button type="button"
                onclick={() => { refreshDashSys(); loadDashboard(); refreshDashJobs(); refreshDashHealth(); refreshDashCodecs(); }}
                disabled={dashLoading}
                class="inline-flex items-center gap-1.5 rounded-lg border border-foreground/[0.08] bg-foreground/[0.03] px-3.5 py-1.5 text-[12px] text-foreground/65 transition-colors hover:border-foreground/15 hover:text-foreground/85 disabled:opacity-40">
                <RefreshCw class="h-3.5 w-3.5 {dashLoading ? 'animate-spin' : ''}" /> Refresh
              </button>
              <button type="button" onclick={handleDashScanNow} disabled={dashScanBusy}
                class="inline-flex items-center gap-1.5 rounded-lg border px-3.5 py-1.5 text-[12px] font-medium transition-all hover:brightness-110 disabled:opacity-50"
                style="border-color: oklch(0.74 0.2 280 / 0.35); color: oklch(0.80 0.2 280); background: oklch(0.74 0.2 280 / 0.08);">
                {#if dashScanBusy}
                  <span class="h-3.5 w-3.5 animate-spin rounded-full border border-current border-t-transparent"></span>
                {:else}
                  <ScanSearch class="h-3.5 w-3.5" />
                {/if}
                Scan libraries
              </button>
            </div>
          </div>

          <!-- ── Instrument Cluster ──────────────────────────────────────────── -->
          <div class="mb-4 grid grid-cols-2 gap-3 lg:grid-cols-4">

            <!-- CPU Gauge — full-circle ring, % in middle -->
            {#if true}
            {@const cpuOffset = _RING_CIRC * (1 - dashCpuPct / 100)}
            <div class="flex flex-col items-center rounded-xl border border-foreground/[0.12] bg-surface/20 px-3 py-4">
              <div class="mb-2 font-serif-display text-[15px] tracking-tight text-foreground/75">CPU Load</div>
              <svg viewBox="0 0 100 100" class="w-full max-w-[110px]" style="aspect-ratio: 1 / 1;">
                <!-- Background ring -->
                <circle cx="50" cy="50" r={_RING_R} fill="none" stroke="rgba(255,255,255,0.07)" stroke-width="8"/>
                <!-- Progress ring — rotated -90° so the fill starts at 12 o'clock and goes clockwise -->
                <circle cx="50" cy="50" r={_RING_R} fill="none" stroke-width="8" stroke-linecap="round" transform="rotate(-90 50 50)"
                  style="stroke: {dashHudColor(dashCpuPct)}; stroke-dasharray: {_RING_CIRC}; stroke-dashoffset: {cpuOffset}; filter: drop-shadow(0 0 6px {dashHudColor(dashCpuPct)}); transition: stroke-dashoffset 1s ease, stroke 0.5s ease, filter 0.5s ease;" />
                <text x="50" y="50" text-anchor="middle" dominant-baseline="central"
                  font-family='Geist, ui-sans-serif, system-ui, sans-serif' font-size="26" font-weight="600" letter-spacing="-1.2"
                  style="fill: {dashHudColor(dashCpuPct)}; transition: fill 0.5s ease;">{dashCpuPct}<tspan font-size="13" dx="1" dy="-4">%</tspan></text>
              </svg>
              {#if sysStatus?.cpu?.cores}
                <div class="mt-2 text-[12px] text-foreground/55">{sysStatus.cpu.cores} cores</div>
              {/if}
            </div>
            {/if}

            <!-- RAM Gauge — full-circle ring, % in middle -->
            {#if true}
            {@const memOffset = _RING_CIRC * (1 - dashMemPct / 100)}
            <div class="flex flex-col items-center rounded-xl border border-foreground/[0.12] bg-surface/20 px-3 py-4">
              <div class="mb-2 font-serif-display text-[15px] tracking-tight text-foreground/75">Memory</div>
              <svg viewBox="0 0 100 100" class="w-full max-w-[110px]" style="aspect-ratio: 1 / 1;">
                <circle cx="50" cy="50" r={_RING_R} fill="none" stroke="rgba(255,255,255,0.07)" stroke-width="8"/>
                <circle cx="50" cy="50" r={_RING_R} fill="none" stroke-width="8" stroke-linecap="round" transform="rotate(-90 50 50)"
                  style="stroke: {dashHudColor(dashMemPct)}; stroke-dasharray: {_RING_CIRC}; stroke-dashoffset: {memOffset}; filter: drop-shadow(0 0 6px {dashHudColor(dashMemPct)}); transition: stroke-dashoffset 1s ease, stroke 0.5s ease, filter 0.5s ease;" />
                <text x="50" y="50" text-anchor="middle" dominant-baseline="central"
                  font-family='Geist, ui-sans-serif, system-ui, sans-serif' font-size="26" font-weight="600" letter-spacing="-1.2"
                  style="fill: {dashHudColor(dashMemPct)}; transition: fill 0.5s ease;">{dashMemPct}<tspan font-size="13" dx="1" dy="-4">%</tspan></text>
              </svg>
              {#if sysStatus?.memory}
                <div class="mt-1 text-[12px] text-foreground/55">
                  {formatBytes(sysStatus.memory.usedBytes)} / {formatBytes(sysStatus.memory.totalBytes)}
                </div>
              {/if}
            </div>
            {/if}

            <!-- Network I/O -->
            <div class="rounded-xl border border-foreground/[0.12] bg-surface/20 p-4">
              <div class="mb-3 font-serif-display text-[15px] tracking-tight text-foreground/75">Network I/O</div>
              <div class="space-y-3.5">
                <div>
                  <div class="mb-1 text-[11px] font-medium tracking-wide text-emerald-300/75">RECV ↓</div>
                  <div class="font-serif-display text-2xl font-medium leading-none tabular-nums tracking-tight"
                    style="color: oklch(0.78 0.22 145); text-shadow: 0 0 10px oklch(0.78 0.22 145 / 0.22);">
                    {sysStatus?.network ? formatBps(sysStatus.network.receiveBps) : '—'}
                  </div>
                </div>
                <div class="h-px bg-foreground/[0.08]"></div>
                <div>
                  <div class="mb-1 text-[11px] font-medium tracking-wide text-accent/85">XMIT ↑</div>
                  <div class="font-serif-display text-2xl font-medium leading-none tabular-nums tracking-tight"
                    style="color: oklch(0.72 0.16 255); text-shadow: 0 0 10px oklch(0.72 0.16 255 / 0.22);">
                    {sysStatus?.network ? formatBps(sysStatus.network.transmitBps) : '—'}
                  </div>
                </div>
              </div>
            </div>

            <!-- Server -->
            <div class="rounded-xl border border-foreground/[0.12] bg-surface/20 p-4">
              <div class="mb-3 font-serif-display text-[15px] tracking-tight text-foreground/75">Server</div>
              <div class="space-y-3.5">
                <div>
                  <div class="mb-1 text-[11px] font-medium tracking-wide text-foreground/55">UPTIME</div>
                  <div class="font-serif-display text-2xl font-medium leading-none tabular-nums tracking-tight"
                    style="color: oklch(0.78 0.22 145); text-shadow: 0 0 10px oklch(0.78 0.22 145 / 0.2);">
                    {dashUptimeStr}
                  </div>
                </div>
                <div class="h-px bg-foreground/[0.08]"></div>
                <div>
                  <div class="mb-1 text-[11px] font-medium tracking-wide text-foreground/55">NOW PLAYING</div>
                  <div class="font-serif-display text-2xl font-medium leading-none tabular-nums tracking-tight {activeSessions.length > 0 ? '' : 'text-foreground/30'}"
                    style={activeSessions.length > 0 ? 'color: oklch(0.74 0.2 280); text-shadow: 0 0 10px oklch(0.74 0.2 280 / 0.3);' : ''}>
                    {activeSessions.length}<span class="ml-1.5 text-[12px] font-normal text-foreground/55 tracking-normal">{activeSessions.length === 1 ? 'stream' : 'streams'}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- ── Library Manifest ────────────────────────────────────────────── -->
          <div class="mb-4 rounded-xl border border-foreground/[0.12] bg-surface/20 px-5 py-4">
            <div class="mb-4 font-serif-display text-[15px] tracking-tight text-foreground/75">Library Manifest</div>
            <div class="grid grid-cols-3 gap-4 sm:grid-cols-6">
              {#each ([
                { id: 'MOV', label: 'Movies',   value: catalogSummary?.movies   ?? 0, warn: false, live: false },
                { id: 'SER', label: 'TV Shows', value: catalogSummary?.series   ?? 0, warn: false, live: false },
                { id: 'EPS', label: 'Episodes', value: catalogSummary?.episodes ?? 0, warn: false, live: false },
                { id: 'FIL', label: 'Files',    value: dashTotalFiles,               warn: false, live: false },
                { id: 'UNP', label: 'Unprobed', value: dashUnprobed,                 warn: dashUnprobed > 0, live: false },
                { id: 'ACT', label: 'Sessions', value: activeSessions.length,        warn: false, live: activeSessions.length > 0 },
              ] as const) as t (t.id)}
                <div class="relative text-center">
                  <div class="text-[11px] font-medium tracking-wide text-foreground/55 mb-2">{t.label}</div>
                  <div class="font-serif-display text-4xl font-medium leading-none tabular-nums tracking-tight"
                    style={t.warn ? 'color: oklch(0.85 0.22 75); text-shadow: 0 0 12px oklch(0.85 0.22 75 / 0.28);'
                         : t.value === 0 ? 'color: rgba(255,255,255,0.2);'
                         : 'color: oklch(0.74 0.2 280); text-shadow: 0 0 12px oklch(0.74 0.2 280 / 0.22);'}>
                    {t.value.toLocaleString()}
                  </div>
                  {#if t.live}
                    <span class="absolute -right-0.5 top-0 flex h-1.5 w-1.5">
                      <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-50"></span>
                      <span class="relative inline-flex h-1.5 w-1.5 rounded-full bg-emerald-400"></span>
                    </span>
                  {/if}
                </div>
              {/each}
            </div>
          </div>

          <!-- ── Ops Board ───────────────────────────────────────────────────── -->
          <div class="mb-4 grid gap-4 lg:grid-cols-2">

            <!-- File Analysis Intelligence -->
            <div class="rounded-xl border border-foreground/[0.12] bg-surface/20 p-5">
              <div class="mb-4 font-serif-display text-[15px] tracking-tight text-foreground/75">File Analysis Intel</div>
              {#if dashTotalFiles > 0}
                {#if true}
                {@const _pct = dashUnprobed > 0 ? Math.min(99, Math.floor(dashProbed / dashTotalFiles * 100)) : 100}
                <!-- RAG semantics: 100% analysed = green (everything ready
                     to play), partial = amber (work still to do), 0% = red
                     (nothing analysed yet). -->
                {@const _pctColor = _pct >= 100 ? 'oklch(0.78 0.22 145)'
                                  : _pct >= 50  ? 'oklch(0.85 0.22 75)'
                                  :               'oklch(0.68 0.26 22)'}
                <div class="flex items-center gap-5 mb-4">
                  <div class="shrink-0">
                    <ActivityRing probed={dashProbed} total={dashTotalFiles} size={72} />
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-baseline gap-1.5 mb-2">
                      <span class="font-serif-display text-6xl font-medium leading-none tabular-nums tracking-tight"
                        style={`color: ${_pctColor}; text-shadow: 0 0 14px ${_pctColor.replace(')', ' / 0.25)')};`}>{_pct}</span>
                      <span class="font-serif-display text-xl text-foreground/55">%</span>
                      <span class="ml-2 text-[12px] italic text-foreground/60">analysed</span>
                    </div>
                    <div class="h-1 w-full overflow-hidden rounded-full bg-foreground/[0.06] mb-2">
                      <div class="h-full rounded-full transition-all duration-700"
                        style={`width: ${_pct}%; background: ${_pctColor.replace(')', ' / 0.65)')};`}></div>
                    </div>
                    <div class="text-[12px] tabular-nums text-foreground/55">
                      {dashProbed.toLocaleString()} of {dashTotalFiles.toLocaleString()} files
                    </div>
                  </div>
                </div>
                {/if}

                <!-- Health flags -->
                <div class="space-y-1.5">
                  {#if dashUnprobed > 0}
                    <div class="flex items-center justify-between rounded-lg border border-amber-400/15 bg-amber-400/[0.05] px-3 py-2">
                      <div class="flex items-center gap-2">
                        <span class="text-[14px] text-amber-400">⚠</span>
                        <span class="text-[12.5px] text-amber-300/85">
                          {dashUnprobed.toLocaleString()} files awaiting analysis
                        </span>
                      </div>
                      {#if !dashProbeRunning}
                        <button type="button" onclick={handleDashProbeNow} disabled={dashProbeBusy}
                          class="rounded px-2.5 py-1 text-[11px] font-medium text-amber-300 transition-colors hover:bg-amber-400/10 disabled:opacity-50">
                          {dashProbeBusy ? '…' : 'Analyse now'}
                        </button>
                      {:else}
                        <span class="text-[11px] italic text-amber-300/70 animate-pulse">analysing…</span>
                      {/if}
                    </div>
                  {:else}
                    <div class="flex items-center gap-2 rounded-lg border border-emerald-400/12 bg-emerald-400/[0.04] px-3 py-2">
                      <span class="text-[14px] text-emerald-400">✓</span>
                      <span class="text-[12.5px] text-emerald-300/75">All files analysed</span>
                    </div>
                  {/if}
                  <!-- The "files with unsupported codecs" flag used to live here.
                       It's been promoted to a full Library Codecs panel below,
                       which explains what each codec means for playback instead
                       of just shouting a scary number. -->
                  {#if (dashHealth?.highBitrate ?? 0) > 0}
                    <div class="flex items-center gap-2 rounded-lg border border-amber-400/10 bg-amber-400/[0.03] px-3 py-2">
                      <span class="text-[13px] text-amber-400/65">▲</span>
                      <span class="text-[12.5px] text-muted-foreground/65">
                        {dashHealth!.highBitrate!.toLocaleString()} files at high bitrate
                      </span>
                    </div>
                  {/if}
                  {#if (dashHealth?.needsReview ?? 0) > 0}
                    <div class="flex items-center gap-2 rounded-lg border border-foreground/[0.06] bg-foreground/[0.02] px-3 py-2">
                      <span class="text-[14px] text-muted-foreground/55">!</span>
                      <span class="text-[12.5px] text-muted-foreground/55">
                        {dashHealth!.needsReview!.toLocaleString()} items need review
                      </span>
                    </div>
                  {/if}
                </div>

                <!-- ─── Library at a glance ────────────────────────────────
                     Fills the vertical real estate below the health flags
                     with useful summary stats from data we already have:
                     total library footprint on disk + subtitle coverage.
                     Neither needs RAG (these are info, not warnings) so
                     they get the theme accent treatment. -->
                {#if true}
                {@const totalBytes = dashHealth?.totalSizeBytes ?? 0}
                {@const withSubs = dashHealth?.withSubtitles ?? 0}
                {@const subsPct = dashTotalFiles > 0 ? Math.round((withSubs / dashTotalFiles) * 100) : 0}
                <div class="mt-4 grid grid-cols-2 gap-3 border-t border-foreground/[0.06] pt-4">
                  <div class="rounded-lg bg-foreground/[0.02] px-3 py-2.5">
                    <div class="text-[10px] uppercase tracking-[0.18em] text-muted-foreground/55">Library footprint</div>
                    <div class="mt-1 font-serif-display text-xl tracking-tight tabular-nums"
                      style="color: oklch(0.74 0.2 280);">
                      {totalBytes > 0 ? formatBytes(totalBytes) : '—'}
                    </div>
                    <div class="text-[11px] text-muted-foreground/55">
                      across {dashTotalFiles.toLocaleString()} files
                    </div>
                  </div>
                  <div class="rounded-lg bg-foreground/[0.02] px-3 py-2.5">
                    <div class="text-[10px] uppercase tracking-[0.18em] text-muted-foreground/55">Subtitle coverage</div>
                    <div class="mt-1 font-serif-display text-xl tracking-tight tabular-nums"
                      style="color: oklch(0.74 0.2 280);">
                      {subsPct}<span class="text-sm text-muted-foreground/55">%</span>
                    </div>
                    <div class="text-[11px] text-muted-foreground/55">
                      {withSubs.toLocaleString()} have subtitles
                    </div>
                  </div>
                </div>
                {/if}
              {:else}
                <div class="flex items-center justify-center py-12 text-[13px] italic text-muted-foreground/40">
                  No files indexed yet
                </div>
              {/if}
            </div>

            <!-- Automation Control -->
            <div class="rounded-xl border border-foreground/[0.12] bg-surface/20 p-5">
              <div class="mb-4 flex items-center justify-between">
                <div class="font-serif-display text-[15px] tracking-tight text-foreground/75">Automation Control</div>
                <a href="/settings/activity" class="text-[11px] italic text-muted-foreground/55 transition-colors hover:text-muted-foreground/85">
                  Full view →
                </a>
              </div>

              <div class="space-y-2">
                <!-- Library Scan — has an editable schedule (syncIntervalMins) -->
                {#if true}
                {@const scanSt = dashJobs?.scan?.status ?? 'idle'}
                {@const scanInterval = dashJobs?.scan?.intervalMins ?? settingsData?.config?.syncIntervalMins ?? 0}
                <div class="rounded-lg border border-foreground/[0.06] bg-surface/15 p-3">
                  <div class="flex items-start gap-3">
                    <div class="shrink-0 w-2 flex justify-center pt-1">
                      {#if scanSt === 'running'}
                        <span class="relative flex h-2 w-2">
                          <span class="absolute inline-flex h-full w-full animate-ping rounded-full opacity-55" style="background: oklch(0.74 0.2 280);"></span>
                          <span class="relative inline-flex h-2 w-2 rounded-full" style="background: oklch(0.74 0.2 280);"></span>
                        </span>
                      {:else if scanSt === 'paused'}
                        <span class="relative inline-flex h-2 w-2 rounded-full bg-amber-400"></span>
                      {:else}
                        <span class="relative inline-flex h-2 w-2 rounded-full bg-foreground/15"></span>
                      {/if}
                    </div>
                    <div class="flex-1 min-w-0">
                      <div class="font-serif-display text-[16px] tracking-tight text-foreground/90">Library Scan</div>
                      <div class="mt-0.5 text-[12px] text-foreground/55 tracking-tight">
                        {#if scanSt === 'running'}<span style="color: oklch(0.74 0.2 280);">RUNNING NOW</span>
                        {:else if scanSt === 'paused'}<span class="text-amber-300">PAUSED — SESSION ACTIVE</span>
                        {:else if dashJobs?.scan?.lastRunAt}LAST RUN {new Date(dashJobs.scan.lastRunAt).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
                        {:else}NEVER RUN{/if}
                      </div>
                    </div>
                    <button type="button" onclick={handleDashScanNow} disabled={dashScanBusy || scanSt === 'running'}
                      class="shrink-0 inline-flex items-center gap-1.5 rounded-lg border px-3.5 py-2 text-[13px] font-medium tracking-tight transition-all hover:brightness-110 disabled:opacity-40 disabled:cursor-not-allowed"
                      style="border-color: oklch(0.74 0.2 280 / 0.45); color: oklch(0.80 0.2 280); background: oklch(0.74 0.2 280 / 0.12); text-shadow: 0 0 8px oklch(0.74 0.2 280 / 0.5);">
                      {#if dashScanBusy}<span class="h-3 w-3 animate-spin rounded-full border border-current border-t-transparent"></span>{:else}▶{/if}
                      Run Now
                    </button>
                  </div>
                  <!-- Schedule row — click to edit -->
                  <div class="mt-2.5 ml-5 flex items-center gap-2 border-t border-foreground/[0.05] pt-2.5">
                    {#if scanIntervalEditing}
                      <span class="text-[12px] font-medium tracking-tight text-foreground/55">Every</span>
                      <input type="number" min="5" max="1440" step="5" bind:value={scanIntervalDraft}
                        class="w-16 rounded border border-foreground/15 bg-surface/40 px-2 py-1 font-mono text-[11px] tabular-nums text-foreground outline-none focus:border-primary/60" />
                      <span class="text-[12px] text-foreground/55">min</span>
                      <button type="button" onclick={saveScanInterval} disabled={scanIntervalSaving}
                        class="ml-1 rounded px-2.5 py-1 text-[12px] font-medium tracking-tight transition-colors disabled:opacity-50"
                        style="color: oklch(0.80 0.2 280); background: oklch(0.74 0.2 280 / 0.18);">
                        {scanIntervalSaving ? '…' : 'Save'}
                      </button>
                      <button type="button" onclick={() => { scanIntervalEditing = false; }}
                        class="rounded px-2 py-1 text-[12px] font-medium tracking-tight text-foreground/55 transition-colors hover:bg-foreground/[0.06] hover:text-foreground/80">
                        Cancel
                      </button>
                    {:else}
                      <button type="button" onclick={openScanIntervalEditor}
                        class="group inline-flex items-center gap-1.5 rounded text-[12px] font-medium tracking-tight text-foreground/55 transition-colors hover:text-foreground/85"
                        title="Click to change how often the library scans">
                        <span>Schedule</span>
                        <span class="text-foreground/75">·</span>
                        <span style="color: oklch(0.80 0.2 280);">Every {dashIntervalLabel(scanInterval)}</span>
                        {#if dashJobs?.scan?.nextRunAt && scanSt !== 'running'}
                          <span class="text-foreground/75">·</span>
                          <span class="text-foreground/70">Next in {dashCountdown(dashJobs.scan.nextRunAt)}</span>
                        {/if}
                        <span class="ml-1 text-foreground/35 transition-colors group-hover:text-foreground/75">✎</span>
                      </button>
                    {/if}
                  </div>
                </div>
                {/if}

                <!-- Metadata Refresh — schedule visible, "Edit" deep-links to Metadata page -->
                {#if true}
                {@const metaSt = dashJobs?.metadata?.status ?? 'idle'}
                {@const metaInterval = dashJobs?.metadata?.intervalMins ?? 0}
                {@const bf = dashJobs?.metadata?.backfill}
                <div class="rounded-lg border border-foreground/[0.06] bg-surface/15 p-3">
                  <div class="flex items-start gap-3">
                    <div class="shrink-0 w-2 flex justify-center pt-1">
                      {#if dashMetaRunning}
                        <span class="relative flex h-2 w-2">
                          <span class="absolute inline-flex h-full w-full animate-ping rounded-full opacity-55" style="background: oklch(0.74 0.2 280);"></span>
                          <span class="relative inline-flex h-2 w-2 rounded-full" style="background: oklch(0.74 0.2 280);"></span>
                        </span>
                      {:else if metaSt === 'paused'}
                        <span class="relative inline-flex h-2 w-2 rounded-full bg-amber-400"></span>
                      {:else}
                        <span class="relative inline-flex h-2 w-2 rounded-full bg-foreground/15"></span>
                      {/if}
                    </div>
                    <div class="flex-1 min-w-0">
                      <div class="font-serif-display text-[16px] tracking-tight text-foreground/90">Metadata Refresh</div>
                      {#if dashMetaRunning && bf && bf.total > 0}
                        <div class="mt-1 space-y-1">
                          <div class="flex items-center justify-between text-[12px] tracking-tight">
                            <span style="color: oklch(0.74 0.2 280);" class="truncate pr-2">REFRESHING{#if bf.lastTitle} — {bf.lastTitle}{/if}</span>
                            <span class="shrink-0 text-foreground/55 tabular-nums">{(bf.refreshed + bf.failed).toLocaleString()}/{bf.total.toLocaleString()}</span>
                          </div>
                          <div class="h-0.5 w-full overflow-hidden rounded-full bg-foreground/[0.06]">
                            <div class="h-full rounded-full transition-all"
                              style="width: {Math.round((bf.refreshed + bf.failed) / bf.total * 100)}%; background: oklch(0.74 0.2 280 / 0.7);"></div>
                          </div>
                        </div>
                      {:else}
                        <div class="mt-0.5 text-[12px] text-foreground/55 tracking-tight">
                          {#if metaSt === 'paused'}<span class="text-amber-300">PAUSED — SESSION ACTIVE</span>
                          {:else if dashJobs?.metadata?.lastRunAt}LAST RUN {new Date(dashJobs.metadata.lastRunAt).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
                          {:else}NEVER RUN{/if}
                        </div>
                      {/if}
                    </div>
                    <button type="button" onclick={() => { triggerBackfill(); refreshDashJobs(); }} disabled={dashMetaRunning}
                      class="shrink-0 inline-flex items-center gap-1.5 rounded-lg border px-3.5 py-2 text-[13px] font-medium tracking-tight transition-all hover:brightness-110 disabled:opacity-40 disabled:cursor-not-allowed"
                      style="border-color: oklch(0.74 0.2 280 / 0.45); color: oklch(0.80 0.2 280); background: oklch(0.74 0.2 280 / 0.12); text-shadow: 0 0 8px oklch(0.74 0.2 280 / 0.5);">
                      {#if dashMetaRunning}<span class="h-3 w-3 animate-spin rounded-full border border-current border-t-transparent"></span>{:else}▶{/if}
                      Run Now
                    </button>
                  </div>
                  <div class="mt-2.5 ml-5 flex items-center gap-1.5 border-t border-foreground/[0.05] pt-2.5">
                    <button type="button" onclick={() => select('metadata')}
                      class="group inline-flex items-center gap-1.5 rounded text-[12px] font-medium tracking-tight text-foreground/55 transition-colors hover:text-foreground/85"
                      title="Open Metadata settings">
                      <span>Schedule</span>
                      <span class="text-foreground/75">·</span>
                      <span style="color: oklch(0.80 0.2 280);">{metaInterval > 0 ? `Every ${dashIntervalLabel(metaInterval)}` : 'Auto in background'}</span>
                      {#if dashJobs?.metadata?.nextRunAt && !dashMetaRunning}
                        <span class="text-foreground/75">·</span>
                        <span class="text-foreground/70">Next in {dashCountdown(dashJobs.metadata.nextRunAt)}</span>
                      {/if}
                      <span class="ml-1 text-foreground/35 transition-colors group-hover:text-foreground/75">↗</span>
                    </button>
                  </div>
                </div>
                {/if}

                <!-- File Analysis — event-driven, runs automatically after scans find new files -->
                {#if true}
                {@const probeJob = dashJobs?.probe?.activeJobs?.find(j => j.status === 'running')}
                {@const probeSt = dashJobs?.probe?.status ?? 'idle'}
                <div class="rounded-lg border border-foreground/[0.06] bg-surface/15 p-3">
                  <div class="flex items-start gap-3">
                    <div class="shrink-0 w-2 flex justify-center pt-1">
                      {#if dashProbeRunning}
                        <span class="relative flex h-2 w-2">
                          <span class="absolute inline-flex h-full w-full animate-ping rounded-full opacity-55" style="background: oklch(0.74 0.2 280);"></span>
                          <span class="relative inline-flex h-2 w-2 rounded-full" style="background: oklch(0.74 0.2 280);"></span>
                        </span>
                      {:else if probeSt === 'paused'}
                        <span class="relative inline-flex h-2 w-2 rounded-full bg-amber-400"></span>
                      {:else}
                        <span class="relative inline-flex h-2 w-2 rounded-full bg-foreground/15"></span>
                      {/if}
                    </div>
                    <div class="flex-1 min-w-0">
                      <div class="font-serif-display text-[16px] tracking-tight text-foreground/90">File Analysis</div>
                      {#if dashProbeRunning && probeJob && (probeJob.total ?? 0) > 0}
                        {#if true}
                        {@const done = probeJob.completed ?? 0}
                        {@const tot = probeJob.total ?? 1}
                        <div class="mt-1 space-y-1">
                          <div class="flex items-center justify-between text-[12px] tracking-tight">
                            <span style="color: oklch(0.74 0.2 280);">ANALYSING FILES</span>
                            <span class="text-foreground/55 tabular-nums">{done.toLocaleString()}/{tot.toLocaleString()} · {Math.round(done/tot*100)}%</span>
                          </div>
                          <div class="h-0.5 w-full overflow-hidden rounded-full bg-foreground/[0.06]">
                            <div class="h-full rounded-full transition-all"
                              style="width: {Math.round(done/tot*100)}%; background: oklch(0.74 0.2 280 / 0.7);"></div>
                          </div>
                        </div>
                        {/if}
                      {:else}
                        <div class="mt-0.5 text-[12px] text-foreground/55 tracking-tight">
                          {#if probeSt === 'paused'}<span class="text-amber-300">PAUSED — SESSION ACTIVE</span>
                          {:else if dashUnprobed > 0}<span style="color: oklch(0.85 0.22 75);">{dashUnprobed.toLocaleString()} FILES AWAITING ANALYSIS</span>
                          {:else}ALL FILES ANALYSED{/if}
                        </div>
                      {/if}
                    </div>
                    <button type="button" onclick={handleDashProbeNow} disabled={dashProbeBusy || dashProbeRunning || dashUnprobed === 0}
                      class="shrink-0 inline-flex items-center gap-1.5 rounded-lg border px-3.5 py-2 text-[13px] font-medium tracking-tight transition-all hover:brightness-110 disabled:opacity-40 disabled:cursor-not-allowed"
                      style="border-color: oklch(0.74 0.2 280 / 0.45); color: oklch(0.80 0.2 280); background: oklch(0.74 0.2 280 / 0.12); text-shadow: 0 0 8px oklch(0.74 0.2 280 / 0.5);">
                      {#if dashProbeBusy || dashProbeRunning}<span class="h-3 w-3 animate-spin rounded-full border border-current border-t-transparent"></span>{:else}▶{/if}
                      Run Now
                    </button>
                  </div>
                  <div class="mt-2.5 ml-5 flex items-center gap-1.5 border-t border-foreground/[0.05] pt-2.5 text-[12px] font-medium tracking-tight text-foreground/55">
                    <span>Schedule</span>
                    <span class="text-foreground/75">·</span>
                    <span style="color: oklch(0.80 0.2 280);">Auto after each scan</span>
                    <span class="ml-2 normal-case tracking-normal text-foreground/40">— runs when new files are found</span>
                  </div>
                </div>
                {/if}
              </div>
            </div>
          </div>

          <!-- ── Library Codecs ──────────────────────────────────────────────── -->
          <!-- Replaces the old "1,714 files with unsupported codecs" mystery
               number with an actual breakdown of what's in the library and
               what each codec means for playback. -->
          {#if dashCodecs && dashCodecs.videoCodecs.length > 0}
            {#if true}
            {@const groups = [
              // RAG semantics: green = nothing the server has to do, amber =
              // light server work (container repack), red = full re-encode.
              { cls: 'direct'    as const, label: 'Plays instantly',  blurb: 'No server work needed. Browsers decode these natively.', color: 'oklch(0.78 0.22 145)' },
              { cls: 'remux'     as const, label: 'Fast repackage',    blurb: 'Video stream is kept; only the container is rewrapped. Starts in seconds.', color: 'oklch(0.85 0.22 75)' },
              { cls: 'transcode' as const, label: 'Needs transcoding', blurb: "No browser can decode these directly — the server has to re-encode the video, which is slow.", color: 'oklch(0.68 0.26 22)' },
            ]}
            {@const bucketed = groups.map(g => {
              const items = dashCodecs!.videoCodecs.filter(v => classifyCodec(v.codec).cls === g.cls);
              const count = items.reduce((s, v) => s + v.count, 0);
              return { ...g, items, count };
            })}
            {@const total = dashCodecs.total || 1}
            <div class="mb-4 rounded-xl border border-foreground/[0.12] bg-surface/20 p-5">
              <div class="mb-1 flex items-baseline justify-between gap-3">
                <div class="font-serif-display text-[15px] tracking-tight text-foreground/75">Library Codecs</div>
                <span class="text-[12px] tabular-nums text-foreground/50">{dashCodecs.total.toLocaleString()} files analysed</span>
              </div>
              <p class="mb-4 text-[12.5px] text-muted-foreground">
                What your library is made of, and what each codec means when you press Play.
              </p>

              <!-- Stacked horizontal bar — proportional widths of the three buckets -->
              <div class="mb-4 flex h-2 w-full overflow-hidden rounded-full bg-foreground/[0.06]">
                {#each bucketed as g (g.cls)}
                  {#if g.count > 0}
                    <div class="h-full transition-all duration-700" title={`${g.label}: ${g.count.toLocaleString()}`}
                      style={`width: ${(g.count / total) * 100}%; background: ${g.color};`}></div>
                  {/if}
                {/each}
              </div>

              <!-- Three bucket sections -->
              <div class="grid gap-3 md:grid-cols-3">
                {#each bucketed as g (g.cls)}
                  <div class="rounded-lg border border-foreground/[0.06] bg-foreground/[0.015] p-3"
                    style={g.count > 0 ? `border-color: ${g.color} / 0.18; background: ${g.color.replace(')', ' / 0.04)')};` : ''}>
                    <div class="flex items-baseline justify-between gap-2">
                      <span class="font-serif-display text-[14px] tracking-tight" style={`color: ${g.color};`}>{g.label}</span>
                      <span class="text-[12px] tabular-nums" style={`color: ${g.color};`}>
                        {g.count.toLocaleString()}<span class="ml-1 text-foreground/45">· {Math.round((g.count / total) * 100)}%</span>
                      </span>
                    </div>
                    <p class="mt-1 text-[11.5px] leading-relaxed text-muted-foreground">{g.blurb}</p>
                    {#if g.items.length > 0}
                      <div class="mt-2.5 flex flex-wrap gap-1.5">
                        {#each g.items as item (item.codec)}
                          <span class="hairline rounded-full bg-surface/40 px-2.5 py-0.5 text-[11px] text-foreground/75 tabular-nums">
                            {dashCodecLabel(item.codec)} · {item.count.toLocaleString()}
                          </span>
                        {/each}
                      </div>
                    {/if}
                  </div>
                {/each}
              </div>

              <!-- Helpful "what to do" footer when there are files in the transcode bucket -->
              {#if bucketed.find(g => g.cls === 'transcode' && g.count > 0)}
                {@const tCount = bucketed.find(g => g.cls === 'transcode')!.count}
                <div class="mt-4 rounded-lg border border-red-400/15 bg-red-400/[0.04] px-4 py-3">
                  <p class="text-[12.5px] leading-relaxed text-foreground/80">
                    <span class="font-medium text-red-300">{tCount.toLocaleString()} {tCount === 1 ? 'file' : 'files'}</span>
                    use older codecs (MPEG-2, VC-1, WMV, etc.) that no browser plays directly. They'll still
                    play — Xuva will re-encode the video on the fly — but expect a longer wait when you press Play
                    and higher CPU use. If you want instant playback for these, convert them to H.264 or HEVC
                    once with a tool like HandBrake.
                  </p>
                </div>
              {/if}
            </div>
            {/if}
          {/if}

          <!-- ── Storage Systems ─────────────────────────────────────────────── -->
          {#if sysStatus?.disks && sysStatus.disks.length > 0}
            <div class="mb-4 rounded-xl border border-foreground/[0.12] bg-surface/20 p-5">
              <div class="mb-4 font-serif-display text-[15px] tracking-tight text-foreground/75">Storage Systems</div>
              <div class="space-y-3">
                {#each sysStatus.disks as disk (disk.path ?? disk.name)}
                  {#if true}
                  {@const dPct = Math.round(disk.usedPercent ?? 0)}
                  <div class="grid items-center gap-4" style="grid-template-columns: minmax(0,1fr) 160px 3rem">
                    <div class="min-w-0">
                      <div class="truncate text-[13px] text-foreground/75">{disk.path ?? disk.name ?? 'Disk'}</div>
                      <div class="text-[12px] text-foreground/45 tabular-nums">{formatBytes(disk.freeBytes)} free</div>
                    </div>
                    <div class="relative h-1 overflow-hidden rounded-full bg-foreground/[0.06]">
                      <div class="absolute inset-y-0 left-0 rounded-full transition-all duration-700"
                        style="width: {dPct}%; background: {dPct >= 90 ? 'oklch(0.68 0.26 22)' : dPct >= 75 ? 'oklch(0.85 0.22 75)' : 'oklch(0.74 0.2 280 / 0.5)'};"></div>
                    </div>
                    <div class="text-right font-serif-display text-[16px] font-medium tabular-nums tracking-tight {dPct >= 90 ? 'text-red-400' : dPct >= 75 ? 'text-amber-400' : 'text-foreground/55'}">{dPct}%</div>
                  </div>
                  {/if}
                {/each}
              </div>
            </div>
          {/if}

          <!-- ── Now Playing ─────────────────────────────────────────────────── -->
          <!-- Always visible: shows an empty state when idle so people can spot the panel. -->
          {#if true}
          {@const liveCount = activeSessions.length}
          {@const hasLive = liveCount > 0}
          <div class="mb-4 overflow-hidden rounded-xl border"
            style={hasLive ? 'border-color: oklch(0.72 0.55 160 / 0.18); background: oklch(0.72 0.55 160 / 0.025);'
                           : 'border-color: rgba(255,255,255,0.07); background: rgba(255,255,255,0.012);'}>
            <div class="flex items-center justify-between gap-3 border-b px-5 py-3"
              style={hasLive ? 'border-color: oklch(0.72 0.55 160 / 0.12);' : 'border-color: rgba(255,255,255,0.05);'}>
              <div class="flex items-center gap-3">
                <span class="relative flex h-2 w-2">
                  {#if hasLive}
                    <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-50"></span>
                    <span class="relative inline-flex h-2 w-2 rounded-full bg-emerald-400 shadow-[0_0_10px_oklch(0.78_0.2_145)]"></span>
                  {:else}
                    <span class="relative inline-flex h-2 w-2 rounded-full bg-foreground/15"></span>
                  {/if}
                </span>
                <span class="font-serif-display text-[16px] tracking-tight {hasLive ? 'text-emerald-300' : 'text-foreground/75'}"
                  style={hasLive ? 'text-shadow: 0 0 10px oklch(0.78 0.2 145 / 0.3);' : ''}>
                  Now Playing
                </span>
                <span class="text-[12px] tabular-nums italic {hasLive ? 'text-emerald-300/75' : 'text-foreground/45'}">
                  {hasLive ? `${liveCount} ${liveCount === 1 ? 'stream' : 'streams'}` : 'nobody watching'}
                </span>
              </div>
            </div>
            {#if hasLive}
              <div class="divide-y" style="border-color: oklch(0.72 0.55 160 / 0.08);">
                {#each activeSessions as session (session.id)}
                  <div class="flex items-center gap-4 px-5 py-3.5">
                    <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-emerald-400/[0.1] ring-1 ring-emerald-400/20">
                      <Play class="h-4 w-4 fill-emerald-400 text-emerald-400" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <div class="truncate text-[13px] font-semibold text-foreground/90">{session.title ?? session.sourceName ?? 'Unknown'}</div>
                      <div class="mt-0.5 text-[12.5px] text-foreground/55">
                        {session.mode ?? session.route ?? 'Direct Play'}{session.deviceId ? ` · ${session.deviceId}` : ''}
                      </div>
                    </div>
                    <span class="shrink-0 rounded border border-emerald-400/25 bg-emerald-400/[0.1] px-2.5 py-0.5 text-[11px] font-medium tracking-wide text-emerald-300"
                      style="text-shadow: 0 0 8px oklch(0.78 0.2 145 / 0.3);">Live</span>
                  </div>
                {/each}
              </div>
            {:else}
              <div class="flex items-center justify-center px-5 py-6 text-center">
                <div class="text-[12px] text-foreground/45">
                  Nobody's watching right now — when someone starts playback, they'll appear here.
                </div>
              </div>
            {/if}
          </div>
          {/if}

          <!-- ── Recent Activity ─────────────────────────────────────────────── -->
          <div class="overflow-hidden rounded-xl border border-foreground/[0.12] bg-surface/20">
            <div class="border-b border-foreground/[0.06] px-5 py-3">
              <span class="font-serif-display text-[15px] tracking-tight text-foreground/75">Recent Activity</span>
            </div>
            {#if recentScans.length > 0}
              <div class="divide-y divide-foreground/[0.04]">
                {#each recentScans as scan (scan.id)}
                  <div class="flex items-center gap-4 px-5 py-2.5">
                    <div class="min-w-0 flex-1">
                      <div class="flex items-center gap-2">
                        <span class="text-[13px] font-medium text-foreground/75 capitalize">{scan.kind ?? 'Scan'}</span>
                        {#if scan.libraryId}
                          <code class="rounded bg-foreground/[0.04] px-1.5 py-0.5 font-mono text-[10px] text-foreground/45">{scan.libraryId.slice(0, 8)}</code>
                        {/if}
                      </div>
                      {#if scan.updatedAt}
                        <div class="mt-0.5 text-[11px] italic text-muted-foreground/55">
                          {new Date(scan.updatedAt).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
                        </div>
                      {/if}
                    </div>
                    {#if scan.status === 'running'}
                      <div class="text-[12px] italic" style="color: oklch(0.74 0.2 280);">running</div>
                    {:else if scan.status === 'completed'}
                      <div class="text-[12px] italic text-emerald-400/80">done</div>
                    {:else if scan.status}
                      <span class="text-[12px] italic text-muted-foreground/55 capitalize">{scan.status}</span>
                    {/if}
                  </div>
                {/each}
              </div>
            {:else}
              <div class="flex items-center justify-center px-5 py-10 text-[13px] italic text-muted-foreground/50">
                No recent activity
              </div>
            {/if}
          </div>

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

            <!-- Trailers toggle -->
            <div class="hairline rounded-2xl bg-surface/40 p-5">
              <h3 class="mb-4 text-sm font-semibold">Display options</h3>
              <div class="flex items-start gap-4">
                <div class="min-w-0 flex-1">
                  <p class="text-sm font-medium">Show trailers in hero</p>
                  <p class="mt-0.5 text-xs text-muted-foreground">
                    When enabled, the home page hero autoplays a trailer (local MP4 or YouTube) for featured titles.
                    Turn off to always show the static backdrop image instead.
                  </p>
                </div>
                <button
                  type="button"
                  onclick={toggleTrailers}
                  disabled={trailersSaving}
                  class="relative mt-0.5 h-6 w-11 shrink-0 rounded-full transition-colors {trailersOn ? 'bg-primary-glow' : 'bg-border'} disabled:opacity-50"
                  role="switch"
                  aria-checked={trailersOn}
                  aria-label="Show trailers in hero"
                >
                  <span class="absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-all {trailersOn ? 'left-[22px]' : 'left-0.5'}"></span>
                </button>
              </div>
            </div>
          </div>

        {:else if active === "scanning"}
          <div class="space-y-12">
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Sync schedule</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Control how and when Xuva watches for new and changed files in your libraries.
                </p>
              </div>
              <div class="space-y-5">
                <div>
                  <label for="sync-mode" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Sync mode</label>
                  <select id="sync-mode" bind:value={editConfig.librarySyncMode}
                    class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none focus:border-primary/60">
                    <option value="auto">Auto — watch + interval fallback</option>
                    <option value="watch">Watch only — filesystem events</option>
                    <option value="interval">Interval only — timed scans</option>
                    <option value="manual">Manual — scan on demand</option>
                  </select>
                </div>
                {#if editConfig.librarySyncMode !== 'manual' && editConfig.librarySyncMode !== 'watch'}
                  <div>
                    <label for="sync-interval" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Sync interval (minutes)</label>
                    <input id="sync-interval" type="number" min="5" max="10080" bind:value={editConfig.syncIntervalMins}
                      class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none focus:border-primary/60" />
                  </div>
                {/if}
                {#if editConfig.librarySyncMode === 'watch' || editConfig.librarySyncMode === 'auto'}
                  <div>
                    <label for="watch-debounce" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Watch debounce (seconds)</label>
                    <p class="mt-1 text-[11px] text-muted-foreground/70">Delay before processing a filesystem event to let file copies finish.</p>
                    <input id="watch-debounce" type="number" min="1" max="300" bind:value={editConfig.watchDebounceSecs}
                      class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none focus:border-primary/60" />
                  </div>
                {/if}
                <div>
                  <label for="probe-batch" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Probe batch limit</label>
                  <p class="mt-1 text-[11px] text-muted-foreground/70">Max files probed per scan run. Lower values reduce memory spikes.</p>
                  <input id="probe-batch" type="number" min="1" max="500" bind:value={editConfig.probeBatchLimit}
                    class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none focus:border-primary/60" />
                </div>
              </div>
            </section>

            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Manual scan</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Trigger an immediate scan across all libraries.
                </p>
              </div>
              <div class="flex items-start gap-3">
                <button type="button" onclick={handleScanAll} disabled={scanAllRunning}
                  class="inline-flex items-center gap-2 rounded-full bg-gradient-primary px-5 py-2.5 text-sm font-semibold text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110 disabled:opacity-60">
                  {#if scanAllRunning}
                    <span class="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white"></span> Scanning…
                  {:else}
                    <RefreshCw class="h-4 w-4" /> Scan all libraries
                  {/if}
                </button>
              </div>
            </section>

            <!-- Background task scheduler (#58) -->
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Background tasks</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  How Xuva handles library scans, metadata fetches, and media probes while you watch.
                </p>
              </div>
              <div class="space-y-3">
                <div class="hairline flex items-center justify-between rounded-xl bg-surface/40 px-4 py-3">
                  <div>
                    <div class="text-sm font-medium">Pause when streaming</div>
                    <div class="mt-0.5 text-[11px] text-muted-foreground">
                      Library scans and media probes pause automatically whenever a playback session is active, so streaming always takes priority.
                    </div>
                  </div>
                  <span class="shrink-0 rounded-full bg-emerald-500/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.15em] text-emerald-300">On</span>
                </div>
                {#if perfSettings?.queues}
                  {#each perfSettings.queues.filter(q => q.name !== 'transcode') as q (q.name)}
                    <div class="hairline flex items-center gap-4 rounded-xl bg-surface/40 px-4 py-3">
                      <div class="min-w-0 flex-1">
                        <div class="text-sm font-medium capitalize">{q.name} queue</div>
                        <div class="text-[11px] text-muted-foreground">
                          {q.workers ?? 1} worker{(q.workers ?? 1) !== 1 ? 's' : ''} &middot; {q.active ?? 0} active &middot; {q.queued ?? 0} queued
                        </div>
                      </div>
                      {#if (q.active ?? 0) > 0}
                        <div class="h-3 w-3 animate-pulse rounded-full bg-primary-glow" title="Processing"></div>
                      {/if}
                    </div>
                  {/each}
                {/if}
              </div>
            </section>

            {#if recentScans.length > 0}
              <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
                <div>
                  <h3 class="font-serif-display text-lg tracking-tight">Recent scans</h3>
                  <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">Last scan jobs across all libraries.</p>
                </div>
                <div class="space-y-2">
                  {#each recentScans as scan (scan.id)}
                    <div class="hairline flex items-center gap-3 rounded-xl bg-surface/40 px-4 py-3">
                      <div class="min-w-0 flex-1">
                        <div class="flex items-center gap-2 text-sm">
                          <span class="font-medium capitalize">{scan.kind ?? 'Scan'}</span>
                          {#if scan.libraryId}<span class="font-mono text-[11px] text-muted-foreground">{scan.libraryId.slice(0,8)}</span>{/if}
                        </div>
                        <div class="mt-0.5 text-[11px] text-muted-foreground">{scan.status ?? ''}
                          {#if scan.updatedAt} · {new Date(scan.updatedAt).toLocaleString(undefined, { month:'short', day:'numeric', hour:'2-digit', minute:'2-digit' })}{/if}
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
              </section>
            {/if}
          </div>

        {:else if active === "transcoding"}
          <div class="space-y-12">

            {#if perfLoading && !perfSettings}
              <div class="flex items-center justify-center py-12"><div class="h-6 w-6 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div></div>
            {:else}
              <!-- Hardware acceleration -->
              <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
                <div>
                  <h3 class="font-serif-display text-lg tracking-tight">Hardware acceleration</h3>
                  <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                    GPU-accelerated transcoding reduces CPU load and allows more simultaneous streams.
                  </p>
                </div>
                <div class="space-y-5">
                  <div class="hairline rounded-2xl bg-surface/40 p-5">
                    <div class="flex items-center justify-between gap-4">
                      <div>
                        <div class="text-sm font-semibold">GPU Transcoding</div>
                        <p class="mt-0.5 text-xs text-muted-foreground">
                          {#if perfSettings?.hardwareAcceleration?.available}{@const hwStatus = perfSettings.hardwareAcceleration.status ?? ''}{hwStatus && hwStatus.toLowerCase() !== 'available' ? `Available — ${hwStatus}` : 'Available — auto-detected'}{:else}Not available on this hardware{/if}
                        </p>
                      </div>
                      <label class="relative inline-flex cursor-pointer items-center gap-2">
                        <input type="checkbox" bind:checked={editConfig.hardwareUnlocked} class="sr-only peer" />
                        <div class="h-5 w-9 rounded-full border border-border bg-surface-elevated/60 peer-checked:bg-primary-glow/70 transition-colors after:absolute after:left-0.5 after:top-0.5 after:h-4 after:w-4 after:rounded-full after:bg-white after:transition-all peer-checked:after:translate-x-4"></div>
                        <span class="text-xs text-muted-foreground">{editConfig.hardwareUnlocked ? 'Enabled' : 'Disabled'}</span>
                      </label>
                    </div>
                    {#if perfSettings?.hardwareAcceleration?.unlockState && perfSettings.hardwareAcceleration.unlockState !== 'unlocked'}
                      <p class="mt-3 text-xs text-amber-300">{perfSettings.hardwareAcceleration.unlockState}</p>
                    {/if}
                  </div>

                  <!-- Detected accelerators panel (from ffmpeg -encoders scan) -->
                  {#if perfSettings?.hardwareAcceleration?.encoders && perfSettings.hardwareAcceleration.encoders.length > 0}
                    <div class="space-y-1">
                      <p class="text-[10px] font-semibold uppercase tracking-[0.22em] text-muted-foreground">Detected accelerators</p>
                      {#each perfSettings.hardwareAcceleration.encoders as enc (enc.id)}
                        <div class="hairline flex items-center gap-3 rounded-xl bg-surface/40 px-3 py-2 text-xs">
                          <span class="font-mono text-primary-glow/80">{enc.id}</span>
                          <span class="text-muted-foreground">{enc.vendor ?? ''}</span>
                          <span class="ml-auto font-medium text-muted-foreground/60">{enc.codec ?? ''}</span>
                        </div>
                      {/each}
                    </div>
                  {:else if perfSettings && !perfSettings?.hardwareAcceleration?.available}
                    <p class="text-xs text-muted-foreground/60">No hardware encoders detected in FFmpeg.</p>
                  {/if}

                  <button type="button" onclick={runHwTest} disabled={hwTestRunning}
                    class="hairline inline-flex items-center gap-2 rounded-full bg-foreground/[0.04] px-4 py-2 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground disabled:opacity-40">
                    {#if hwTestRunning}
                      <span class="h-3 w-3 animate-spin rounded-full border border-current border-t-transparent"></span> Testing…
                    {:else}
                      Run hardware test
                    {/if}
                  </button>

                  <!-- Active test results -->
                  {#if hwTestResult}
                    <div class="hairline rounded-xl bg-surface/40 p-4 space-y-2">
                      <div class="flex items-center gap-2 text-sm font-semibold">
                        {#if hwTestResult.status === 'ok' || (hwTestResult.working ?? 0) > 0}
                          <Check class="h-4 w-4 text-emerald-300" />
                          <span class="text-emerald-300">{hwTestResult.working}/{hwTestResult.tested} codecs working</span>
                        {:else}
                          <span class="text-red-300">{hwTestResult.error ?? 'No hardware codecs available'}</span>
                        {/if}
                      </div>
                      {#each hwTestResult.tests ?? [] as t (t.id)}
                        <div class="flex items-center gap-3 text-xs text-muted-foreground">
                          {#if t.ok}<Check class="h-3 w-3 shrink-0 text-emerald-300" />{:else}<span class="h-3 w-3 shrink-0 rounded-full bg-red-400/40"></span>{/if}
                          <span class="font-medium">{t.label ?? t.id}</span>
                          <span class="text-muted-foreground/60">{t.codec ?? ''}</span>
                          {#if t.durationMs}<span class="ml-auto">{t.durationMs}ms</span>{/if}
                          {#if t.error}<span class="text-red-300">{t.error}</span>{/if}
                        </div>
                      {/each}
                    </div>
                  <!-- Last persisted test results (shown before any test is run in this session) -->
                  {:else if perfSettings?.hardwareAcceleration?.lastTest}
                    {@const last = perfSettings.hardwareAcceleration.lastTest}
                    <div class="hairline rounded-xl bg-surface/40 p-4 space-y-2">
                      <div class="flex items-center gap-2 text-sm font-semibold">
                        {#if (last.working ?? 0) > 0}
                          <Check class="h-4 w-4 text-emerald-300" />
                          <span class="text-emerald-300">{last.working}/{last.tested} codecs working</span>
                        {:else}
                          <span class="text-amber-300">Last test: no working codecs</span>
                        {/if}
                        {#if last.testedAt}
                          <span class="ml-auto text-[10px] text-muted-foreground/50">
                            {new Date(last.testedAt).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
                          </span>
                        {/if}
                      </div>
                      {#each last.tests ?? [] as t (t.id)}
                        <div class="flex items-center gap-3 text-xs text-muted-foreground">
                          {#if t.ok}<Check class="h-3 w-3 shrink-0 text-emerald-300" />{:else}<span class="h-3 w-3 shrink-0 rounded-full bg-red-400/40"></span>{/if}
                          <span class="font-medium">{t.label ?? t.id}</span>
                          <span class="text-muted-foreground/60">{t.codec ?? ''}</span>
                          {#if t.durationMs}<span class="ml-auto">{t.durationMs}ms</span>{/if}
                          {#if t.error}<span class="text-red-300">{t.error}</span>{/if}
                        </div>
                      {/each}
                    </div>
                  {/if}
                </div>
              </section>

              <!-- Worker limits -->
              {#if perfSettings?.limits}
                <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
                  <div>
                    <h3 class="font-serif-display text-lg tracking-tight">Worker limits</h3>
                    <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                      Current concurrency limits set by the server based on your hardware profile.
                    </p>
                  </div>
                  <div class="grid gap-3 sm:grid-cols-2">
                    {#each ([
                      { label: 'Scan workers', val: perfSettings.limits.scanWorkers },
                      { label: 'Probe workers', val: perfSettings.limits.probeWorkers },
                      { label: 'Transcode workers', val: perfSettings.limits.transcodeWorkers },
                      { label: 'GPU workers', val: perfSettings.limits.gpuWorkers },
                    ] as const) as row (row.label)}
                      <div class="hairline rounded-xl bg-surface/40 px-4 py-3">
                        <div class="text-[10px] font-semibold uppercase tracking-[0.2em] text-muted-foreground">{row.label}</div>
                        <div class="mt-1 font-serif-display text-2xl text-foreground/80">{row.val ?? '—'}</div>
                      </div>
                    {/each}
                  </div>
                </section>
              {/if}

              <!-- Queue status -->
              {#if perfSettings?.queues && perfSettings.queues.length > 0}
                <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
                  <div>
                    <h3 class="font-serif-display text-lg tracking-tight">Queue status</h3>
                    <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">Live queue activity.</p>
                  </div>
                  <div class="space-y-2">
                    {#each perfSettings.queues as q (q.name)}
                      <div class="hairline flex items-center gap-4 rounded-xl bg-surface/40 px-4 py-3">
                        <div class="min-w-0 flex-1">
                          <div class="text-sm font-medium capitalize">{q.name}</div>
                          <div class="text-[11px] text-muted-foreground">{q.class ?? ''}</div>
                        </div>
                        <div class="flex gap-4 text-xs text-muted-foreground">
                          <span>{q.active ?? 0} active</span>
                          <span>{q.queued ?? 0} queued</span>
                          {#if q.workerUtilization != null}
                            <span>{(q.workerUtilization * 100).toFixed(0)}% util</span>
                          {/if}
                        </div>
                      </div>
                    {/each}
                  </div>
                </section>
              {/if}
            {/if}
          </div>

        {:else if active === "storage"}
          <div class="space-y-12">
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Directories</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Paths where Xuva stores transcoded files, downloads, metadata, and caches. Leave blank to use defaults.
                </p>
              </div>
              <div class="space-y-4">
                {#if settingsData?.config?.dataDir}
                  <div class="hairline rounded-xl bg-surface/30 px-4 py-3">
                    <div class="text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">Data directory (read-only)</div>
                    <div class="mt-1 truncate font-mono text-sm text-foreground/70">{settingsData.config.dataDir}</div>
                    <p class="mt-0.5 text-[11px] text-muted-foreground/60">Set via environment variable at startup.</p>
                  </div>
                {/if}
                {#each ([
                  { field: 'transcodeDir', label: 'Transcode directory', hint: 'Where in-progress transcode segments are written.' },
                  { field: 'downloadsDir', label: 'Downloads directory', hint: 'Where downloaded files are saved.' },
                  { field: 'metadataDir', label: 'Metadata directory', hint: 'Artwork, NFO files, and metadata cache.' },
                  { field: 'cacheDir', label: 'Cache directory', hint: 'Thumbnails and temporary cache files.' },
                  { field: 'tempDir', label: 'Temp directory', hint: 'Short-lived working files.' },
                ] as const) as row (row.field)}
                  <div>
                    <label for={`dir-${row.field}`} class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">{row.label}</label>
                    {#if row.hint}<p class="mt-0.5 text-[11px] text-muted-foreground/70">{row.hint}</p>{/if}
                    <div class="mt-2 flex gap-2">
                      <input id={`dir-${row.field}`} type="text" bind:value={editConfig[row.field]}
                        placeholder="Default"
                        class="h-11 flex-1 rounded-xl border border-border bg-background/40 px-4 font-mono text-sm outline-none placeholder:font-sans placeholder:text-muted-foreground/50 focus:border-primary/60 focus:bg-background/70" />
                      <button type="button" onclick={() => openStorageBrowser(row.field)}
                        class="hairline flex h-11 items-center gap-1.5 rounded-xl bg-foreground/[0.06] px-4 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.10] hover:text-foreground">
                        <Folder class="h-3.5 w-3.5" /> Browse
                      </button>
                    </div>
                    <!-- Inline folder browser when this field is active -->
                    {#if showFolderBrowser && browserContext === 'storage' && storageBrowserField === row.field}
                      <div class="mt-2 hairline overflow-hidden rounded-xl bg-background/40">
                        <div class="flex items-center gap-2 border-b border-border px-3 py-2">
                          <FolderOpen class="h-3.5 w-3.5 shrink-0 text-primary-glow" />
                          <span class="min-w-0 flex-1 truncate font-mono text-xs text-foreground/80">{browserCurrentPath || '/'}</span>
                          <button type="button"
                            onclick={() => { editConfig[row.field as keyof typeof editConfig] = browserCurrentPath as never; showFolderBrowser = false; }}
                            class="shrink-0 rounded-full bg-primary-glow/10 px-3 py-1 text-[11px] font-semibold text-primary-glow transition-colors hover:bg-primary-glow/20">Select</button>
                          <button type="button" onclick={() => (showFolderBrowser = false)}
                            class="shrink-0 rounded-full bg-foreground/[0.06] px-2 py-1 text-[11px] text-muted-foreground transition-colors hover:text-foreground">✕</button>
                        </div>
                        {#if browserLoading}
                          <div class="flex items-center justify-center py-5"><div class="h-5 w-5 animate-spin rounded-full border border-border border-t-primary-glow"></div></div>
                        {:else}
                          <ul class="max-h-48 overflow-y-auto py-1">
                            {#if browserParentPath !== undefined}
                              <li><button type="button" onclick={() => navigateFolder(browserParentPath!)}
                                class="flex w-full items-center gap-3 px-3 py-2 text-sm text-muted-foreground hover:bg-foreground/[0.04] hover:text-foreground">
                                <ChevronLeft class="h-3.5 w-3.5 shrink-0"/><span class="font-mono text-xs">..</span></button></li>
                            {/if}
                            {#each browserEntries as entry (entry.path ?? entry.name)}
                              <li><button type="button" onclick={() => navigateFolder(entry.path ?? '')}
                                class="flex w-full items-center gap-3 px-3 py-2 text-sm text-foreground/80 hover:bg-foreground/[0.04] hover:text-foreground">
                                <Folder class="h-3.5 w-3.5 shrink-0 text-muted-foreground/60"/>
                                <span class="flex-1 truncate font-mono text-xs">{entry.name ?? entry.path}</span>
                                <ChevronRight class="h-3 w-3 shrink-0 text-muted-foreground/40"/></button></li>
                            {:else}
                              <li class="px-4 py-3 text-xs text-muted-foreground/60">No sub-folders found.</li>
                            {/each}
                          </ul>
                        {/if}
                      </div>
                    {/if}
                  </div>
                {/each}
              </div>
            </section>

            <!-- Disk usage from system status -->
            {#if sysStatus?.disks && sysStatus.disks.length > 0}
              <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
                <div>
                  <h3 class="font-serif-display text-lg tracking-tight">Disk usage</h3>
                  <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">Current usage across mounted volumes visible to the server.</p>
                </div>
                <div class="space-y-4">
                  {#each sysStatus.disks as disk (disk.path ?? disk.name)}
                    {@const pct = disk.usedPercent ?? 0}
                    <div>
                      <div class="flex items-center justify-between">
                        <span class="max-w-[60%] truncate font-mono text-xs text-muted-foreground">{disk.path ?? disk.name ?? 'Disk'}</span>
                        <span class="text-xs {pct > 90 ? 'text-amber-300' : 'text-foreground/60'}">{formatBytes(disk.usedBytes)} / {formatBytes(disk.totalBytes)}</span>
                      </div>
                      <div class="mt-1.5 h-2 w-full overflow-hidden rounded-full bg-surface-elevated/60">
                        <div class="h-full rounded-full {pct > 90 ? 'bg-amber-400' : 'bg-foreground/25'}" style="width: {Math.min(pct, 100)}%"></div>
                      </div>
                      {#if disk.error}<p class="mt-0.5 text-[11px] text-red-300">{disk.error}</p>{/if}
                    </div>
                  {/each}
                </div>
              </section>
            {/if}
          </div>

        {:else if active === "backup"}
          <div class="space-y-12">
            <!-- Export -->
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Export backup</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Download a <code class="rounded bg-surface-elevated px-1 py-0.5 font-mono text-[11px]">.tar.gz</code> archive containing your catalog database, settings, and a manifest. Keep this file safe — it contains your watch history, metadata matches, and all curation work.
                </p>
              </div>
              <div class="space-y-4">
                <button
                  type="button"
                  onclick={handleExport}
                  disabled={backupExporting}
                  class="inline-flex items-center gap-2 rounded-xl bg-foreground/[0.08] hairline px-5 py-3 text-sm font-medium text-foreground transition-colors hover:bg-foreground/[0.14] disabled:opacity-50"
                >
                  <ArchiveRestore class="h-4 w-4 shrink-0" />
                  {backupExporting ? 'Preparing archive…' : 'Download backup'}
                </button>
                {#if backupExportError}
                  <p class="text-sm text-red-400">{backupExportError}</p>
                {/if}
              </div>
            </section>

            <!-- Import -->
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Restore from backup</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Upload a previously exported <code class="rounded bg-surface-elevated px-1 py-0.5 font-mono text-[11px]">.tar.gz</code> archive. The restore is staged and applied the next time the server starts — your current database is preserved as <code class="rounded bg-surface-elevated px-1 py-0.5 font-mono text-[11px]">xuva.db.bak</code> until the new one proves stable.
                </p>
              </div>
              <div class="space-y-4">
                {#if backupImportResult}
                  <div class="hairline rounded-xl bg-surface/40 px-4 py-4 space-y-2">
                    <div class="flex items-center gap-2 text-sm font-medium text-foreground">
                      <Check class="h-4 w-4 text-emerald-400" /> Restore staged successfully
                    </div>
                    <p class="text-[13px] text-muted-foreground">
                      Backup created <span class="font-mono">{backupImportResult.manifest?.createdAt ?? '—'}</span>.
                      Restart the server to apply the restore.
                    </p>
                    {#if backupImportResult.manifest?.mediaPaths?.movies || backupImportResult.manifest?.mediaPaths?.tv}
                      <p class="text-[12px] text-muted-foreground/70">
                        Original media paths — Movies: <span class="font-mono">{backupImportResult.manifest?.mediaPaths?.movies || '—'}</span>,
                        TV: <span class="font-mono">{backupImportResult.manifest?.mediaPaths?.tv || '—'}</span>.
                        Verify these still match your current library locations after restart.
                      </p>
                    {/if}
                  </div>
                {:else}
                  <label class="inline-flex cursor-pointer items-center gap-2 rounded-xl bg-foreground/[0.08] hairline px-5 py-3 text-sm font-medium text-foreground transition-colors hover:bg-foreground/[0.14]" class:opacity-50={backupImporting}>
                    <Upload class="h-4 w-4 shrink-0" />
                    {backupImporting ? 'Uploading…' : 'Choose backup file'}
                    <input type="file" accept=".tar.gz,.tgz" onchange={handleImport} disabled={backupImporting} class="sr-only" />
                  </label>
                {/if}
                {#if backupImportError}
                  <p class="text-sm text-red-400">{backupImportError}</p>
                {/if}
              </div>
            </section>
          </div>

        {:else if active === "network"}
          <div class="space-y-12">
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">mDNS discovery</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Xuva advertises a friendly instance name plus the network address clients should use to connect.
                </p>
              </div>
              <div class="space-y-4">
                {#if discoveryLoading && !discoveryStatus}
                  <div class="flex items-center justify-center py-8"><div class="h-6 w-6 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div></div>
                {:else if discoveryStatus}
                  <div class="hairline rounded-2xl bg-surface/40 p-5 space-y-3">
                    <div class="flex items-center justify-between">
                      <span class="text-sm font-semibold">Status</span>
                      {#if discoveryStatus.running}
                        <span class="hairline inline-flex items-center gap-1.5 rounded-full bg-emerald-400/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.15em] text-emerald-300"><span class="h-1.5 w-1.5 rounded-full bg-emerald-400"></span>Running</span>
                      {:else if discoveryStatus.enabled}
                        <span class="hairline inline-flex items-center gap-1.5 rounded-full bg-amber-400/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.15em] text-amber-300">Enabled, not running</span>
                      {:else}
                        <span class="hairline rounded-full bg-foreground/[0.06] px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.15em] text-muted-foreground">Disabled</span>
                      {/if}
                    </div>
                    {#if discoveryStatus.serviceName}
                      <div class="flex items-center justify-between text-sm">
                        <span class="text-muted-foreground">Service name</span>
                        <span class="font-mono text-foreground/80">{discoveryStatus.serviceName}</span>
                      </div>
                    {/if}
                    {#if discoveryStatus.hostName}
                      <div class="flex items-center justify-between text-sm">
                        <span class="text-muted-foreground">Network host name</span>
                        <span class="font-mono text-foreground/80">{discoveryStatus.hostName}</span>
                      </div>
                    {/if}
                    {#if discoveryStatus.webUrl}
                      <div class="flex items-center justify-between text-sm">
                        <span class="text-muted-foreground">Web address</span>
                        <span class="font-mono text-foreground/80">{discoveryStatus.webUrl}</span>
                      </div>
                    {/if}
                    {#if discoveryStatus.serviceType}
                      <div class="flex items-center justify-between text-sm">
                        <span class="text-muted-foreground">Service type</span>
                        <span class="font-mono text-foreground/80">{discoveryStatus.serviceType}</span>
                      </div>
                    {/if}
                    {#if discoveryStatus.port}
                      <div class="flex items-center justify-between text-sm">
                        <span class="text-muted-foreground">Port</span>
                        <span class="font-mono text-foreground/80">{discoveryStatus.port}</span>
                      </div>
                    {/if}
                    {#if discoveryStatus.lastError}
                      <p class="text-xs text-red-300">{discoveryStatus.lastError}</p>
                    {/if}
                    {#if discoveryStatus.note}
                      <p class="text-xs text-muted-foreground/70">{discoveryStatus.note}</p>
                    {/if}
                  </div>
                  <button type="button" onclick={loadDiscovery} class="hairline inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground">
                    <RefreshCw class="h-3 w-3" /> Refresh
                  </button>
                {:else}
                  <p class="text-sm text-muted-foreground">Unable to load discovery status.</p>
                  <button type="button" onclick={loadDiscovery} class="hairline mt-2 inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground"><RefreshCw class="h-3 w-3" /> Retry</button>
                {/if}
              </div>
            </section>

            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">HTTP port</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  The port Xuva listens on is configured at startup via the <code class="font-mono text-xs">XUVA_HTTP_ADDR</code> environment variable and cannot be changed here.
                </p>
              </div>
              <div class="hairline rounded-2xl bg-surface/30 p-5">
                <p class="text-sm text-muted-foreground">To change the HTTP port, restart the server with a different <code class="font-mono text-xs">XUVA_HTTP_ADDR</code> value (e.g. <code class="font-mono text-xs">0.0.0.0:8097</code>).</p>
              </div>
            </section>
          </div>

        {:else if active === "playback"}
          <div class="space-y-12">
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Playback policy</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Controls how the server decides between direct streaming and transcoding for each session.
                </p>
              </div>
              <div class="space-y-3">
                {#each ([
                  { id: 'original_only', label: 'Original files only', desc: 'Stream originals as-is. Files that need conversion will show fallback options instead of converting automatically.' },
                  { id: 'light', label: 'Light compatibility', desc: 'Allow audio conversion and container remux. Video stays untouched so quality is always preserved.' },
                  { id: 'full', label: 'Full compatibility', desc: 'Allow video and audio conversion for maximum device compatibility. Recommended for most setups.' },
                  { id: 'cinema', label: 'Cinema server', desc: 'Full conversion plus future automated optimisation controls for power users with heavy workloads.' },
                ] as const) as opt (opt.id)}
                  <button type="button" onclick={() => (editConfig.playbackPolicy = opt.id)}
                    class="hairline w-full rounded-2xl p-4 text-left transition-all {editConfig.playbackPolicy === opt.id ? 'bg-surface-elevated/80 shadow-elev' : 'bg-surface/40 hover:bg-surface/70'}">
                    <div class="flex items-center gap-3">
                      <span class="flex h-4 w-4 items-center justify-center rounded-full border {editConfig.playbackPolicy === opt.id ? 'border-primary-glow bg-primary-glow' : 'border-border'}">
                        {#if editConfig.playbackPolicy === opt.id}<Check class="h-2.5 w-2.5 text-black" />{/if}
                      </span>
                      <span class="text-sm font-semibold">{opt.label}</span>
                    </div>
                    <p class="mt-2 pl-7 text-xs text-muted-foreground">{opt.desc}</p>
                  </button>
                {/each}
              </div>
            </section>

            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Subtitle preferences</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Control how the server handles subtitles and video quality during playback.
                </p>
              </div>
              <div class="space-y-4">
                <!-- Prefer text subtitles toggle -->
                <div class="hairline flex items-start gap-4 rounded-2xl bg-surface/40 p-4">
                  <div class="min-w-0 flex-1">
                    <p class="text-sm font-semibold">Prefer text subtitles</p>
                    <p class="mt-0.5 text-xs text-muted-foreground">
                      Always use SRT/ASS text tracks over bitmap (PGS/DVB) subtitles when both are available. Reduces CPU/GPU usage during transcode.
                    </p>
                  </div>
                  <button
                    type="button"
                    onclick={() => (editConfig.preferTextSubtitles = !editConfig.preferTextSubtitles)}
                    class="relative mt-0.5 h-6 w-11 shrink-0 rounded-full transition-colors {editConfig.preferTextSubtitles ? 'bg-primary-glow' : 'bg-border'}"
                    role="switch"
                    aria-checked={editConfig.preferTextSubtitles}
                    aria-label="Prefer text subtitles"
                  >
                    <span class="absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-all {editConfig.preferTextSubtitles ? 'left-[22px]' : 'left-0.5'}"></span>
                  </button>
                </div>
                <!-- Original quality only toggle -->
                <div class="hairline flex items-start gap-4 rounded-2xl bg-surface/40 p-4">
                  <div class="min-w-0 flex-1">
                    <p class="text-sm font-semibold">Original quality only</p>
                    <p class="mt-0.5 text-xs text-muted-foreground">
                      Refuse to transcode video. Files that can't be direct-played will return an error instead of being transcoded. Overrides the playback policy above.
                    </p>
                  </div>
                  <button
                    type="button"
                    onclick={() => (editConfig.originalQualityOnly = !editConfig.originalQualityOnly)}
                    class="relative mt-0.5 h-6 w-11 shrink-0 rounded-full transition-colors {editConfig.originalQualityOnly ? 'bg-primary-glow' : 'bg-border'}"
                    role="switch"
                    aria-checked={editConfig.originalQualityOnly}
                    aria-label="Original quality only"
                  >
                    <span class="absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-all {editConfig.originalQualityOnly ? 'left-[22px]' : 'left-0.5'}"></span>
                  </button>
                </div>
                <!-- Default subtitles on for movies -->
                <div class="hairline flex items-start gap-4 rounded-2xl bg-surface/40 p-4">
                  <div class="min-w-0 flex-1">
                    <p class="text-sm font-semibold">Subtitles on by default — Movies</p>
                    <p class="mt-0.5 text-xs text-muted-foreground">
                      When you start a movie, automatically enable the first available subtitle track. You can still toggle subtitles off in the player.
                    </p>
                  </div>
                  <button
                    type="button"
                    onclick={() => (editConfig.defaultSubtitlesMovies = !editConfig.defaultSubtitlesMovies)}
                    class="relative mt-0.5 h-6 w-11 shrink-0 rounded-full transition-colors {editConfig.defaultSubtitlesMovies ? 'bg-primary-glow' : 'bg-border'}"
                    role="switch"
                    aria-checked={editConfig.defaultSubtitlesMovies}
                    aria-label="Subtitles on by default — Movies"
                  >
                    <span class="absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-all {editConfig.defaultSubtitlesMovies ? 'left-[22px]' : 'left-0.5'}"></span>
                  </button>
                </div>
                <!-- Default subtitles on for TV -->
                <div class="hairline flex items-start gap-4 rounded-2xl bg-surface/40 p-4">
                  <div class="min-w-0 flex-1">
                    <p class="text-sm font-semibold">Subtitles on by default — TV</p>
                    <p class="mt-0.5 text-xs text-muted-foreground">
                      When you start a TV episode, automatically enable the first available subtitle track.
                    </p>
                  </div>
                  <button
                    type="button"
                    onclick={() => (editConfig.defaultSubtitlesTV = !editConfig.defaultSubtitlesTV)}
                    class="relative mt-0.5 h-6 w-11 shrink-0 rounded-full transition-colors {editConfig.defaultSubtitlesTV ? 'bg-primary-glow' : 'bg-border'}"
                    role="switch"
                    aria-checked={editConfig.defaultSubtitlesTV}
                    aria-label="Subtitles on by default — TV"
                  >
                    <span class="absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-all {editConfig.defaultSubtitlesTV ? 'left-[22px]' : 'left-0.5'}"></span>
                  </button>
                </div>
              </div>
            </section>

            {#if deviceProfiles.length > 0}
              <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
                <div>
                  <h3 class="font-serif-display text-lg tracking-tight">Client profiles</h3>
                  <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                    Known device capability profiles. The server uses these to pick the best playback route.
                  </p>
                </div>
                <div class="space-y-2">
                  {#each deviceProfiles as p (p.id)}
                    <div class="hairline rounded-xl bg-surface/40 px-4 py-3">
                      <div class="flex flex-wrap items-center gap-2">
                        <span class="text-sm font-medium">{p.name ?? p.id}</span>
                        {#if p.supportsHevc}<span class="rounded-full bg-foreground/[0.06] px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground">HEVC</span>{/if}
                        {#if p.supportsAv1}<span class="rounded-full bg-foreground/[0.06] px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground">AV1</span>{/if}
                        {#if p.supportsHdr}<span class="rounded-full bg-amber-400/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-amber-300">HDR</span>{/if}
                        {#if p.supportsDolbyVision}<span class="rounded-full bg-blue-400/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-blue-300">DV</span>{/if}
                        {#if p.preferDirectPlay}<span class="rounded-full bg-emerald-400/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-emerald-300">Direct play</span>{/if}
                        {#if p.maxBitrate}<span class="ml-auto font-mono text-[11px] text-muted-foreground/60">{(p.maxBitrate / 1000).toFixed(0)} Mbps</span>{/if}
                      </div>
                      {#if p.description}<p class="mt-0.5 text-xs text-muted-foreground">{p.description}</p>{/if}
                    </div>
                  {/each}
                </div>
              </section>
            {/if}
          </div>

        {:else if active === "users"}
          <div class="space-y-8">
            <div class="flex items-center justify-between">
              <p class="text-sm text-muted-foreground">
                {usersList.length === 0 && !usersLoading ? 'No users yet.' : `${usersList.length} user${usersList.length !== 1 ? 's' : ''}`}
              </p>
              <button type="button" onclick={() => { showAddUser = !showAddUser; usersError = null; }}
                class="inline-flex items-center gap-2 rounded-full bg-gradient-primary px-4 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110">
                <Plus class="h-3.5 w-3.5" /> {showAddUser ? 'Cancel' : 'Add user'}
              </button>
            </div>

            {#if usersError}
              <div class="rounded-xl bg-red-400/10 px-4 py-3 text-sm text-red-300">{usersError}</div>
            {/if}

            {#if showAddUser}
              <div class="hairline rounded-2xl bg-surface/50 p-6 space-y-5">
                <h3 class="font-semibold">New user</h3>
                <div class="grid gap-5 sm:grid-cols-2">
                  <div>
                    <label for="new-username" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Username</label>
                    <input id="new-username" type="text" bind:value={newUserName} placeholder="username"
                      class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60" />
                  </div>
                  <div>
                    <label for="new-displayname" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Display name</label>
                    <input id="new-displayname" type="text" bind:value={newUserDisplay} placeholder="Full name"
                      class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60" />
                  </div>
                  <div>
                    <label for="new-pass" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Password</label>
                    <input id="new-pass" type="password" autocomplete="new-password" bind:value={newUserPass} placeholder="••••••••"
                      class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60" />
                  </div>
                  <div>
                    <div class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Role</div>
                    <div class="mt-2 flex gap-2">
                      {#each (['viewer', 'admin'] as const) as role (role)}
                        <button type="button" onclick={() => (newUserRole = role)}
                          class="flex-1 rounded-xl border py-2.5 text-sm font-medium capitalize transition-colors {newUserRole === role ? 'border-primary/60 bg-primary-glow/10 text-foreground' : 'border-border bg-surface/40 text-muted-foreground hover:text-foreground'}">
                          {role}
                        </button>
                      {/each}
                    </div>
                  </div>
                </div>
                <div class="flex items-center gap-3 border-t border-border pt-4">
                  <button type="button" onclick={handleCreateUser} disabled={userSaving || !newUserName.trim() || !newUserPass.trim()}
                    class="inline-flex items-center gap-2 rounded-full bg-gradient-primary px-5 py-2.5 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110 disabled:opacity-60">
                    {#if userSaving}<span class="h-3.5 w-3.5 animate-spin rounded-full border border-white/30 border-t-white"></span> Creating…{:else}Create user{/if}
                  </button>
                  <button type="button" onclick={() => { showAddUser = false; usersError = null; }}
                    class="hairline rounded-full bg-foreground/[0.04] px-5 py-2.5 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground">
                    Cancel
                  </button>
                </div>
              </div>
            {/if}

            {#if usersLoading}
              <div class="flex items-center justify-center py-8"><div class="h-6 w-6 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div></div>
            {:else if usersList.length > 0}
              <div class="space-y-3">
                {#each usersList as user (user.id ?? user.username)}
                  {@const uid = user.id ?? ''}
                  {@const isEditing = editingProfileUserId === uid}
                  {@const isPinEditing = pinEditUserId === uid}
                  <div class="hairline rounded-2xl bg-surface/40 overflow-hidden">
                    <!-- Header row -->
                    <div class="flex flex-wrap items-center gap-4 p-4">
                      <!-- Avatar preview -->
                      <div class="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-xl bg-surface-elevated/60">
                        {#if user.avatarPreset}
                          <img src="/avatars/{user.avatarPreset}.svg" alt="" class="h-full w-full object-cover" />
                        {:else}
                          <span class="text-xs font-semibold uppercase text-muted-foreground">
                            {(user.displayName ?? user.username ?? '?').slice(0, 2)}
                          </span>
                        {/if}
                      </div>
                      <div class="min-w-0 flex-1">
                        <div class="flex flex-wrap items-center gap-2">
                          <span class="font-medium">{user.displayName ?? user.username}</span>
                          {#if user.displayName && user.username}
                            <span class="font-mono text-xs text-muted-foreground/60">@{user.username}</span>
                          {/if}
                          {#if user.role}
                            <span class="rounded-full bg-foreground/[0.06] px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-muted-foreground">{user.role}</span>
                          {/if}
                          {#if user.isRestricted}
                            <span class="rounded-full bg-amber-400/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-amber-400">Kids</span>
                          {/if}
                          {#if user.maxRating}
                            <span class="rounded-full bg-blue-400/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-blue-400">{user.maxRating}</span>
                          {/if}
                          {#if user.hasPin}
                            <span class="rounded-full bg-green-400/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-green-400">PIN set</span>
                          {/if}
                        </div>
                      </div>
                      <div class="flex shrink-0 flex-wrap gap-2">
                        <button type="button" onclick={() => uid && (isEditing ? cancelEditProfile() : startEditProfile(user))}
                          class="hairline inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground">
                          {isEditing ? 'Cancel' : 'Edit profile'}
                        </button>
                        <button type="button" onclick={() => uid && (isPinEditing ? cancelEditPin() : startEditPin(uid))}
                          class="hairline inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground">
                          {isPinEditing ? 'Cancel' : (user.hasPin ? 'Change PIN' : 'Set PIN')}
                        </button>
                        {#if userDeletingId === uid}
                          <button type="button" onclick={() => uid && handleDeleteUser(uid)}
                            class="inline-flex items-center gap-1.5 rounded-full bg-red-400/10 px-3 py-1.5 text-xs font-semibold text-red-300 transition-colors hover:bg-red-400/20">Confirm?</button>
                          <button type="button" onclick={() => (userDeletingId = null)}
                            class="hairline rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08]">Cancel</button>
                        {:else}
                          <button type="button" onclick={() => uid && handleDeleteUser(uid)} disabled={!uid}
                            class="hairline inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-red-400/10 hover:text-red-300 disabled:opacity-40">
                            <Trash2 class="h-3 w-3" /> Delete
                          </button>
                        {/if}
                      </div>
                    </div>

                    <!-- Profile edit panel -->
                    {#if isEditing}
                      <div class="border-t border-border bg-surface/30 px-5 py-5 space-y-5">
                        {#if profileSaveError}<p class="text-sm text-red-400">{profileSaveError}</p>{/if}

                        <div class="grid gap-4 sm:grid-cols-2">
                          <!-- Display name -->
                          <div>
                            <label for={`profile-name-${uid}`} class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Display name</label>
                            <input id={`profile-name-${uid}`} type="text" bind:value={profileEditName} placeholder="Name shown on profile picker"
                              class="mt-2 h-10 w-full rounded-xl border border-border bg-background/40 px-3 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60" />
                          </div>

                          <!-- Rating ceiling -->
                          <div>
                            <label for={`profile-rating-${uid}`} class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Max content rating</label>
                            <select id={`profile-rating-${uid}`} bind:value={profileEditMaxRating}
                              class="mt-2 h-10 w-full rounded-xl border border-border bg-background/40 px-3 text-sm text-foreground outline-none focus:border-primary/60">
                              {#each RATING_OPTIONS as opt (opt.value)}
                                <option value={opt.value}>{opt.label}</option>
                              {/each}
                            </select>
                          </div>
                        </div>

                        <!-- Kids toggle -->
                        <div class="flex items-center gap-3">
                          <button type="button" role="switch" aria-checked={profileEditRestricted}
                            aria-label="Kids profile"
                            onclick={() => (profileEditRestricted = !profileEditRestricted)}
                            class={`relative inline-flex h-6 w-11 shrink-0 items-center rounded-full transition-colors ${profileEditRestricted ? 'bg-primary' : 'bg-surface-elevated'}`}>
                            <span class={`h-4 w-4 rounded-full bg-white shadow transition-transform ${profileEditRestricted ? 'translate-x-6' : 'translate-x-1'}`}></span>
                          </button>
                          <div>
                            <p class="text-sm font-medium">Kids profile</p>
                            <p class="text-xs text-muted-foreground">PIN guards exit; entry is open to all</p>
                          </div>
                        </div>

                        <!-- Avatar preset -->
                        <div>
                          <p class="mb-2 text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">Avatar</p>
                          <div class="flex flex-wrap gap-2">
                            {#each AVATAR_PRESETS as preset (preset)}
                              <button type="button" aria-label={preset} onclick={() => (profileEditPreset = preset === profileEditPreset ? '' : preset)}
                                class={`h-12 w-12 overflow-hidden rounded-xl ring-2 transition-all ${profileEditPreset === preset ? 'ring-primary scale-110' : 'ring-transparent opacity-70 hover:opacity-100'}`}>
                                <img src="/avatars/{preset}.svg" alt="" class="h-full w-full object-cover" />
                              </button>
                            {/each}
                            <!-- Clear -->
                            {#if profileEditPreset}
                              <button type="button" onclick={() => (profileEditPreset = '')}
                                class="flex h-12 w-12 items-center justify-center rounded-xl border border-border text-xs text-muted-foreground transition-colors hover:bg-surface-elevated hover:text-foreground">
                                Clear
                              </button>
                            {/if}
                          </div>
                        </div>

                        <div class="flex gap-3 border-t border-border pt-4">
                          <button type="button" onclick={() => uid && handleSaveProfile(uid)} disabled={profileSaving || !profileEditName.trim()}
                            class="inline-flex items-center gap-2 rounded-full bg-gradient-primary px-5 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110 disabled:opacity-60">
                            {#if profileSaving}<span class="h-3 w-3 animate-spin rounded-full border border-white/30 border-t-white"></span> Saving…{:else}Save{/if}
                          </button>
                          <button type="button" onclick={cancelEditProfile}
                            class="hairline rounded-full bg-foreground/[0.04] px-5 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground">Cancel</button>
                        </div>
                      </div>
                    {/if}

                    <!-- PIN edit panel -->
                    {#if isPinEditing}
                      <div class="border-t border-border bg-surface/30 px-5 py-5 space-y-4">
                        {#if pinSaveError}<p class="text-sm text-red-400">{pinSaveError}</p>{/if}
                        <div>
                          <label for={`pin-input-${uid}`} class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                            {user.isRestricted ? 'Exit PIN (required to leave this profile)' : 'Entry PIN (required to enter this profile)'}
                          </label>
                          <input id={`pin-input-${uid}`} type="password" inputmode="numeric" bind:value={pinEditValue} placeholder="4-digit PIN (blank to remove)"
                            maxlength={8}
                            class="mt-2 h-10 w-full rounded-xl border border-border bg-background/40 px-3 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60" />
                          <p class="mt-1 text-xs text-muted-foreground">Leave blank to remove the PIN from this profile.</p>
                        </div>
                        <div class="flex gap-3">
                          <button type="button" onclick={() => uid && handleSavePin(uid)} disabled={pinSaving}
                            class="inline-flex items-center gap-2 rounded-full bg-gradient-primary px-5 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110 disabled:opacity-60">
                            {#if pinSaving}<span class="h-3 w-3 animate-spin rounded-full border border-white/30 border-t-white"></span> Saving…{:else}Save PIN{/if}
                          </button>
                          <button type="button" onclick={cancelEditPin}
                            class="hairline rounded-full bg-foreground/[0.04] px-5 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground">Cancel</button>
                        </div>
                      </div>
                    {/if}
                  </div>
                {/each}
              </div>
            {:else if !showAddUser}
              <div class="hairline flex flex-col items-center justify-center rounded-3xl bg-surface/30 px-8 py-20 text-center">
                <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-surface-elevated/70 text-primary-glow shadow-elev"><Users class="h-6 w-6" /></div>
                <div class="font-serif-display mt-5 text-xl tracking-tight">No users yet</div>
                <p class="mt-2 max-w-sm text-sm text-muted-foreground">Add the first user to enable authentication.</p>
                <button type="button" onclick={() => (showAddUser = true)}
                  class="mt-6 inline-flex items-center gap-2 rounded-full bg-gradient-primary px-5 py-2.5 text-sm font-semibold text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110">
                  <Plus class="h-4 w-4" /> Add first user
                </button>
              </div>
            {/if}
          </div>

        {:else if active === "pending-approvals"}
          <div class="space-y-6">

            <!-- QR pair token -->
            {#if qrToken}
              <div class="hairline rounded-2xl bg-surface/50 p-6">
                <div class="flex flex-wrap items-start gap-6">
                  <img src={qrToken.imageUrl} alt="Pairing QR code" class="h-40 w-40 shrink-0 rounded-xl bg-white object-contain p-1" />
                  <div class="flex min-w-0 flex-1 flex-col gap-3">
                    <div>
                      <p class="text-sm font-semibold">Scan to pair instantly</p>
                      <p class="mt-1 text-xs text-muted-foreground">Point your phone camera at this QR code. On Apple TV, type the code below.</p>
                    </div>
                    <div class="font-mono text-2xl font-black tracking-[0.25em] text-primary-glow">{qrToken.token}</div>
                    <div class="flex items-center gap-2 text-xs text-muted-foreground">
                      <span class="inline-block h-2 w-2 rounded-full {qrSecondsLeft > 15 ? 'bg-emerald-400' : 'bg-amber-400'} animate-pulse"></span>
                      Expires in {qrSecondsLeft}s
                    </div>
                    <button type="button" onclick={dismissQR}
                      class="hairline mt-auto inline-flex w-fit items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground">
                      Dismiss
                    </button>
                  </div>
                </div>
              </div>
            {:else}
              <div class="hairline rounded-2xl bg-surface/30 p-5">
                <div class="flex items-center justify-between gap-4">
                  <div>
                    <p class="text-sm font-semibold">QR code pairing</p>
                    <p class="mt-0.5 text-xs text-muted-foreground">Generate a one-time QR code. Scan with iPhone or enter the code on Apple TV — no approval step needed.</p>
                  </div>
                  <button type="button" onclick={generateQR} disabled={qrGenerating}
                    class="inline-flex shrink-0 items-center gap-2 rounded-full bg-primary-glow/10 px-4 py-2 text-xs font-semibold text-primary-glow transition-colors hover:bg-primary-glow/20 disabled:opacity-40">
                    {#if qrGenerating}<span class="h-3 w-3 animate-spin rounded-full border border-current border-t-transparent"></span>{:else}<QrCode class="h-3.5 w-3.5" />{/if}
                    Generate QR
                  </button>
                </div>
              </div>
            {/if}

            <div class="flex items-center justify-between">
              <p class="text-sm text-muted-foreground">
                {pairingRequests.length === 0 && !pairingLoading ? 'No pending requests.' : `${pairingRequests.length} pending`}
              </p>
              <button type="button" onclick={loadPairingRequests} disabled={pairingLoading}
                class="hairline inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground disabled:opacity-40">
                <RefreshCw class="h-3 w-3 {pairingLoading ? 'animate-spin' : ''}" /> Refresh
              </button>
            </div>

            {#if pairingLoading && pairingRequests.length === 0}
              <div class="flex items-center justify-center py-8"><div class="h-6 w-6 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div></div>
            {:else if pairingRequests.length > 0}
              <div class="space-y-3">
                {#each pairingRequests as req (req.id)}
                  <div class="hairline rounded-2xl bg-surface/40 p-5">
                    <div class="flex flex-wrap items-start justify-between gap-4">
                      <div class="min-w-0">
                        <div class="flex flex-wrap items-center gap-2">
                          <span class="font-semibold">{req.deviceName ?? 'Unknown device'}</span>
                          {#if req.clientProfile}
                            <span class="rounded-full bg-foreground/[0.06] px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-muted-foreground">{req.clientProfile}</span>
                          {/if}
                        </div>
                        {#if req.code}
                          <div class="mt-1 font-mono text-sm text-primary-glow tracking-widest">{req.code}</div>
                        {/if}
                        <div class="mt-1 text-xs text-muted-foreground">
                          {#if req.expiresAt}Expires {new Date(req.expiresAt).toLocaleString(undefined, { month:'short', day:'numeric', hour:'2-digit', minute:'2-digit' })}{/if}
                        </div>
                      </div>
                      <div class="flex shrink-0 gap-2">
                        <button type="button" onclick={() => req.id && handleApprove(req.id)} disabled={pairingActionId === req.id}
                          class="inline-flex items-center gap-1.5 rounded-full bg-emerald-400/10 px-4 py-2 text-xs font-semibold text-emerald-300 transition-colors hover:bg-emerald-400/20 disabled:opacity-40">
                          {#if pairingActionId === req.id}<span class="h-3 w-3 animate-spin rounded-full border border-current border-t-transparent"></span>{:else}<Check class="h-3.5 w-3.5" />{/if}
                          Approve
                        </button>
                        <button type="button" onclick={() => req.id && handleDeny(req.id)} disabled={pairingActionId === req.id}
                          class="hairline inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-4 py-2 text-xs font-medium text-muted-foreground transition-colors hover:bg-red-400/10 hover:text-red-300 disabled:opacity-40">
                          Deny
                        </button>
                      </div>
                    </div>
                  </div>
                {/each}
              </div>
            {:else}
              <div class="hairline flex flex-col items-center justify-center rounded-3xl bg-surface/30 px-8 py-20 text-center">
                <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-surface-elevated/70 text-primary-glow shadow-elev"><Link2 class="h-6 w-6" /></div>
                <div class="font-serif-display mt-5 text-xl tracking-tight">No pending requests</div>
                <p class="mt-2 max-w-sm text-sm text-muted-foreground">Pairing requests from new devices appear here for you to approve or deny.</p>
              </div>
            {/if}
          </div>

        {:else if active === "approved-devices"}
          <div class="space-y-6">
            <div class="flex items-center justify-between">
              <p class="text-sm text-muted-foreground">
                {approvedDevices.length === 0 && !devicesLoading ? 'No approved devices.' : `${approvedDevices.length} device${approvedDevices.length !== 1 ? 's' : ''}`}
              </p>
              <button type="button" onclick={loadApprovedDevices} disabled={devicesLoading}
                class="hairline inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground disabled:opacity-40">
                <RefreshCw class="h-3 w-3 {devicesLoading ? 'animate-spin' : ''}" /> Refresh
              </button>
            </div>

            {#if devicesLoading && approvedDevices.length === 0}
              <div class="flex items-center justify-center py-8"><div class="h-6 w-6 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div></div>
            {:else if approvedDevices.length > 0}
              <div class="space-y-2">
                {#each approvedDevices as dev (dev.id)}
                  <div class="hairline flex flex-wrap items-center gap-4 rounded-2xl bg-surface/40 p-4">
                    <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-surface-elevated/60 text-primary-glow">
                      <ShieldCheck class="h-4 w-4" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <div class="flex flex-wrap items-center gap-2">
                        <span class="font-medium">{dev.displayName ?? dev.deviceName ?? 'Unknown'}</span>
                        {#if dev.clientProfile}<span class="rounded-full bg-foreground/[0.06] px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-muted-foreground">{dev.clientProfile}</span>{/if}
                        {#if dev.status === 'active'}
                          <span class="hairline inline-flex items-center gap-1.5 rounded-full bg-emerald-400/10 px-2.5 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-emerald-300">Active</span>
                        {:else if dev.status}
                          <span class="rounded-full bg-foreground/[0.06] px-2 py-0.5 text-[10px] uppercase tracking-[0.15em] text-muted-foreground">{dev.status}</span>
                        {/if}
                      </div>
                      {#if dev.approvedAt}
                        <div class="mt-0.5 text-[11px] text-muted-foreground/60">Approved {new Date(dev.approvedAt).toLocaleDateString()}{dev.approvedBy ? ` by ${dev.approvedBy}` : ''}</div>
                      {/if}
                    </div>
                    <div class="flex shrink-0 gap-2">
                      {#if deviceRevokingId === dev.id}
                        <button type="button" onclick={() => dev.id && handleRevoke(dev.id)}
                          class="inline-flex items-center gap-1.5 rounded-full bg-red-400/10 px-3 py-1.5 text-xs font-semibold text-red-300 transition-colors hover:bg-red-400/20">Confirm revoke?</button>
                        <button type="button" onclick={() => (deviceRevokingId = null)}
                          class="hairline rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08]">Cancel</button>
                      {:else}
                        <button type="button" onclick={() => dev.id && handleRevoke(dev.id)} disabled={!dev.id}
                          class="hairline inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-red-400/10 hover:text-red-300 disabled:opacity-40">
                          <Unlink class="h-3 w-3" /> Revoke
                        </button>
                      {/if}
                    </div>
                  </div>
                {/each}
              </div>
            {:else}
              <div class="hairline flex flex-col items-center justify-center rounded-3xl bg-surface/30 px-8 py-20 text-center">
                <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-surface-elevated/70 text-primary-glow shadow-elev"><ShieldCheck class="h-6 w-6" /></div>
                <div class="font-serif-display mt-5 text-xl tracking-tight">No approved devices</div>
                <p class="mt-2 max-w-sm text-sm text-muted-foreground">Devices you approve appear here. You can revoke access at any time.</p>
              </div>
            {/if}
          </div>

        {:else if active === "general"}
          <div class="space-y-12">
            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Server identity</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Separate the friendly Xuva name from the network address browsers and native clients use to connect.
                </p>
              </div>
              <div class="space-y-5">
                <div>
                  <label for="server-name" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                    Xuva display name
                  </label>
                  <input
                    id="server-name"
                    type="text"
                    bind:value={editConfig.serverName}
                    placeholder="Xuva"
                    class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70"
                  />
                  <p class="mt-2 text-xs leading-relaxed text-muted-foreground">
                    This is a human-readable instance name for titles, setup, and discovery lists. It is not a DNS name and does not need to resolve on your network.
                  </p>
                </div>
                <div>
                  <label for="canonical-web-origin" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                    Canonical network web address
                  </label>
                  <input
                    id="canonical-web-origin"
                    type="url"
                    bind:value={editConfig.canonicalWebOrigin}
                    placeholder="http://DESKTOP-7UV0925:8097"
                    class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70"
                  />
                  <p class="mt-2 text-xs leading-relaxed text-muted-foreground">
                    This must be a real address for the machine, DNS name, mDNS name, reverse proxy, or container host that clients can reach. Browser visits through IP addresses redirect here so cookies, profile choice, and sign-in state stay consistent. Leave blank to use the operating-system hostname automatically when the server listens on the LAN.
                  </p>
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
                <h3 class="font-serif-display text-lg tracking-tight">Region</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Used to show trending titles from your country. Timezone is stored for scheduling and display purposes.
                </p>
              </div>
              <div class="space-y-5">
                <div>
                  <label for="st-country" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                    Country
                  </label>
                  <select
                    id="st-country"
                    bind:value={editConfig.country}
                    class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none focus:border-primary/60 focus:bg-background/70"
                  >
                    <option value="">— Not set —</option>
                    {#each [
                      { code: 'AU', label: 'Australia' },
                      { code: 'BR', label: 'Brazil' },
                      { code: 'CA', label: 'Canada' },
                      { code: 'FR', label: 'France' },
                      { code: 'DE', label: 'Germany' },
                      { code: 'IN', label: 'India' },
                      { code: 'IT', label: 'Italy' },
                      { code: 'JP', label: 'Japan' },
                      { code: 'MX', label: 'Mexico' },
                      { code: 'NL', label: 'Netherlands' },
                      { code: 'NZ', label: 'New Zealand' },
                      { code: 'PL', label: 'Poland' },
                      { code: 'PT', label: 'Portugal' },
                      { code: 'ES', label: 'Spain' },
                      { code: 'SE', label: 'Sweden' },
                      { code: 'CH', label: 'Switzerland' },
                      { code: 'GB', label: 'United Kingdom' },
                      { code: 'US', label: 'United States' },
                    ] as c (c.code)}
                      <option value={c.code}>{c.label}</option>
                    {/each}
                  </select>
                </div>
                <div>
                  <label for="st-tz" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                    Timezone
                  </label>
                  <select
                    id="st-tz"
                    bind:value={editConfig.timezone}
                    class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none focus:border-primary/60 focus:bg-background/70"
                  >
                    <option value="">— Not set —</option>
                    {#each [
                      { id: 'Pacific/Honolulu',    label: 'Hawaii (UTC−10)' },
                      { id: 'America/Los_Angeles', label: 'Pacific Time (UTC−8/−7)' },
                      { id: 'America/Denver',      label: 'Mountain Time (UTC−7/−6)' },
                      { id: 'America/Chicago',     label: 'Central Time (UTC−6/−5)' },
                      { id: 'America/New_York',    label: 'Eastern Time (UTC−5/−4)' },
                      { id: 'America/Sao_Paulo',   label: 'Brasília (UTC−3)' },
                      { id: 'Atlantic/Reykjavik',  label: 'Reykjavik (UTC+0)' },
                      { id: 'Europe/London',       label: 'London (UTC+0/+1)' },
                      { id: 'Europe/Paris',        label: 'Paris / Berlin (UTC+1/+2)' },
                      { id: 'Europe/Helsinki',     label: 'Helsinki (UTC+2/+3)' },
                      { id: 'Europe/Moscow',       label: 'Moscow (UTC+3)' },
                      { id: 'Asia/Dubai',          label: 'Dubai (UTC+4)' },
                      { id: 'Asia/Kolkata',        label: 'India (UTC+5:30)' },
                      { id: 'Asia/Bangkok',        label: 'Bangkok (UTC+7)' },
                      { id: 'Asia/Singapore',      label: 'Singapore (UTC+8)' },
                      { id: 'Asia/Tokyo',          label: 'Tokyo (UTC+9)' },
                      { id: 'Australia/Sydney',    label: 'Sydney (UTC+10/+11)' },
                      { id: 'Pacific/Auckland',    label: 'Auckland (UTC+12/+13)' },
                    ] as tz (tz.id)}
                      <option value={tz.id}>{tz.label}</option>
                    {/each}
                  </select>
                </div>
              </div>
            </section>

            <section class="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
              <div>
                <h3 class="font-serif-display text-lg tracking-tight">Metadata language</h3>
                <p class="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                  Language used when fetching titles, overviews, and artwork from TMDB. Changes apply to future metadata refreshes.
                </p>
              </div>
              <div>
                <label for="st-metalang" class="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                  Language
                </label>
                <select
                  id="st-metalang"
                  bind:value={editConfig.metadataLanguage}
                  class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none focus:border-primary/60 focus:bg-background/70"
                >
                  {#each [
                    { code: 'en-US', label: 'English (US)' },
                    { code: 'en-GB', label: 'English (UK)' },
                    { code: 'fr-FR', label: 'French' },
                    { code: 'de-DE', label: 'German' },
                    { code: 'es-ES', label: 'Spanish (Spain)' },
                    { code: 'es-MX', label: 'Spanish (Mexico)' },
                    { code: 'it-IT', label: 'Italian' },
                    { code: 'pt-BR', label: 'Portuguese (Brazil)' },
                    { code: 'pt-PT', label: 'Portuguese (Portugal)' },
                    { code: 'nl-NL', label: 'Dutch' },
                    { code: 'pl-PL', label: 'Polish' },
                    { code: 'sv-SE', label: 'Swedish' },
                    { code: 'nb-NO', label: 'Norwegian' },
                    { code: 'da-DK', label: 'Danish' },
                    { code: 'fi-FI', label: 'Finnish' },
                    { code: 'ru-RU', label: 'Russian' },
                    { code: 'ja-JP', label: 'Japanese' },
                    { code: 'ko-KR', label: 'Korean' },
                    { code: 'zh-CN', label: 'Chinese (Simplified)' },
                    { code: 'zh-TW', label: 'Chinese (Traditional)' },
                  ] as lang (lang.code)}
                    <option value={lang.code}>{lang.label}</option>
                  {/each}
                </select>
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
                    bind:value={acctDisplayName}
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
                    value={currentUser?.username ?? ''}
                    placeholder="username"
                    class="mt-2 h-11 w-full cursor-not-allowed rounded-xl border border-border bg-background/20 px-4 text-sm text-muted-foreground/60 outline-none"
                  />
                  <p class="mt-1.5 text-[11px] text-muted-foreground/60">Username cannot be changed after setup.</p>
                </div>
                {#if acctProfileError}
                  <p class="text-xs text-red-400">{acctProfileError}</p>
                {/if}
                {#if acctProfileSuccess}
                  <p class="text-xs text-green-400">Profile saved.</p>
                {/if}
                <button
                  type="button"
                  onclick={handleAccountSaveProfile}
                  disabled={acctProfileSaving}
                  class="rounded-full bg-gradient-primary px-5 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110 disabled:opacity-50"
                >
                  {acctProfileSaving ? 'Saving…' : 'Save profile'}
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
                    bind:value={acctCurrentPw}
                    autocomplete="current-password"
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
                    bind:value={acctNewPw}
                    autocomplete="new-password"
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
                    bind:value={acctConfirmPw}
                    autocomplete="new-password"
                    placeholder="••••••••"
                    class="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70"
                  />
                </div>
                {#if acctPwError}
                  <p class="text-xs text-red-400">{acctPwError}</p>
                {/if}
                {#if acctPwSuccess}
                  <p class="text-xs text-green-400">Password updated successfully.</p>
                {/if}
                <button
                  type="button"
                  onclick={handleAccountUpdatePassword}
                  disabled={acctPwSaving}
                  class="rounded-full bg-gradient-primary px-5 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110 disabled:opacity-50"
                >
                  {acctPwSaving ? 'Updating…' : 'Update password'}
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
                  onclick={handleSignOut}
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
