<script lang="ts">
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import { Bell, Bookmark, Clock, Menu, Search, Settings, Film, Tv, LogOut, User, Layers, X, AlertTriangle, CheckCircle, Users } from "lucide-svelte";
  import Logo from "./Logo.svelte";
  import { primeSearchCatalogue, runSearch, getSearchResults, isSearchLoading } from "$lib/stores/searchStore.svelte";
  import { getPlaybackRecent } from "$lib/api/home";
  import type { PlaybackRecentItem } from "$lib/api/home";
  import type { SearchHit } from "$lib/api/browse";
  import { getNotifications, dismissNotification, dismissAllNotifications } from "$lib/api/operator";
  import type { NotificationItem } from "$lib/api/operator";
  import { profileStore } from "$lib/stores/profileStore.svelte";

  const nav = [
    { label: "Home", href: "/" },
    { label: "Movies", href: "/movies" },
    { label: "TV", href: "/tv" },
    { label: "Collections", href: "/collections" },
    { label: "Watchlist", href: "/watchlist" },
  ];

  let scrolled = $state(false);
  const currentPath = $derived(page.url.pathname);

  // ── Search ─────────────────────────────────────────────────────────────────
  let searchQuery = $state("");
  let searchFocused = $state(false);
  let searchInputEl = $state<HTMLInputElement | null>(null);

  // Kick off backend search whenever the input changes; results are
  // available via `getSearchResults()` once loading finishes.
  $effect(() => {
    runSearch(searchQuery, 8);
  });

  const searchResp = $derived(getSearchResults());
  const searchResults = $derived<SearchHit[]>(
    !searchResp || searchResp.query !== searchQuery.trim()
      ? []
      : [
          ...searchResp.movies.slice(0, 4),
          ...searchResp.series.slice(0, 4),
          ...searchResp.people.slice(0, 3),
          ...searchResp.collections.slice(0, 3),
        ]
  );
  const showSearchDropdown = $derived(searchFocused && searchQuery.trim().length > 0);

  function handleSearchKeydown(e: KeyboardEvent) {
    if (e.key === "Enter" && searchQuery.trim()) {
      goto(`/search?q=${encodeURIComponent(searchQuery.trim())}`);
      searchFocused = false;
      searchInputEl?.blur();
    }
    if (e.key === "Escape") {
      searchFocused = false;
      searchInputEl?.blur();
    }
  }

  function handleResultClick(hit: SearchHit) {
    searchQuery = "";
    searchFocused = false;
    switch (hit.kind) {
      case "movie":
        goto(`/movies/${hit.id}`);
        return;
      case "series":
        goto(`/tv/${hit.id}`);
        return;
      case "person":
        goto(`/people/${encodeURIComponent(hit.name)}`);
        return;
      case "collection":
        goto(`/collections/${hit.id}`);
        return;
    }
  }

  function hitKey(h: SearchHit): string {
    switch (h.kind) {
      case "movie":
        return `m:${h.id}`;
      case "series":
        return `s:${h.id}`;
      case "person":
        return `p:${h.name}`;
      case "collection":
        return `c:${h.id}`;
    }
  }

  function handleSearchFocus() {
    searchFocused = true;
    primeSearchCatalogue();
  }

  // ── Notifications ──────────────────────────────────────────────────────────
  let showNotifications = $state(false);
  let notifications = $state<NotificationItem[]>([]);
  let notificationsLoaded = $state(false);

  async function loadNotifications() {
    try {
      const resp = await getNotifications();
      notifications = resp.notifications ?? [];
    } catch {
      notifications = [];
    } finally {
      notificationsLoaded = true;
    }
  }

  async function openNotifications() {
    showNotifications = !showNotifications;
    showProfile = false;
    showRecentPlayed = false;
    if (showNotifications) {
      await loadNotifications();
    }
  }

  async function handleDismiss(id: string) {
    await dismissNotification(id).catch(() => {});
    notifications = notifications.filter(n => n.id !== id);
  }

  async function handleDismissAll() {
    await dismissAllNotifications().catch(() => {});
    notifications = [];
  }

  // ── Recently Played ────────────────────────────────────────────────────────
  let showRecentPlayed = $state(false);
  let recentItems = $state<PlaybackRecentItem[]>([]);
  let recentLoaded = $state(false);

  async function openRecentPlayed() {
    showRecentPlayed = !showRecentPlayed;
    showProfile = false;
    showNotifications = false;
    if (showRecentPlayed && !recentLoaded) {
      try {
        const resp = await getPlaybackRecent(undefined, 8);
        recentItems = resp.recent ?? [];
      } catch {
        recentItems = [];
      } finally {
        recentLoaded = true;
      }
    }
  }

  function formatProgress(item: PlaybackRecentItem): string {
    const pct = item.percent ?? 0;
    if (item.watched) return 'Watched';
    if (pct > 0) return `${Math.round(pct)}%`;
    return '';
  }

  function friendlyName(item: PlaybackRecentItem): string {
    if (item.name) return item.name;
    if (item.relPath) {
      return item.relPath
        .replace(/\.[a-z0-9]{2,4}$/i, '')
        .replace(/\s*\([^)]*(?:remux|bluray|1080p|2160p|720p|hdtv)[^)]*\)/gi, '')
        .split('/').pop() ?? item.relPath;
    }
    return 'Unknown';
  }

  // ── Profile ────────────────────────────────────────────────────────────────
  let showProfile = $state(false);
  const activeProfile = $derived(profileStore.activeProfile);

  // ── Mobile menu ────────────────────────────────────────────────────────────
  let showMobileMenu = $state(false);

  function toggleProfile() {
    showProfile = !showProfile;
    showNotifications = false;
  }

  function openProfileSwitcher() {
    showProfile = false;
    profileStore.openPicker();
  }

  function profileInitials(name: string): string {
    return name.split(/\s+/).slice(0, 2).map(w => w[0]?.toUpperCase() ?? '').join('');
  }

  function profileAvatarSrc(p: typeof activeProfile): string | null {
    if (!p) return null;
    if (p.avatarUrl) return p.avatarUrl;
    if (p.avatarPreset) return `/avatars/${p.avatarPreset}.svg`;
    return null;
  }

  // ── Scroll ─────────────────────────────────────────────────────────────────
  $effect(() => {
    if (typeof window === "undefined") return;
    const onScroll = () => { scrolled = window.scrollY > 12; };
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  });

  // ── Click-outside to close any overlay ────────────────────────────────────
  $effect(() => {
    if (typeof document === "undefined") return;
    const handler = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      if (!target.closest("[data-search-container]")) searchFocused = false;
      if (!target.closest("[data-notif-container]")) { showNotifications = false; showRecentPlayed = false; }
      if (!target.closest("[data-profile-container]")) showProfile = false;
      if (!target.closest("[data-mobile-menu]") && !target.closest("[data-hamburger]")) showMobileMenu = false;
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  });

  // ── ⌘K / Ctrl+K shortcut ─────────────────────────────────────────────────
  let isMac = $state(false);

  $effect(() => {
    if (typeof navigator === "undefined") return;
    isMac = /Macintosh|MacIntel|MacPPC|Mac68K/.test(navigator.userAgent);
  });

  $effect(() => {
    if (typeof document === "undefined") return;
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        searchInputEl?.focus();
        searchFocused = true;
        primeSearchCatalogue();
      }
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  });
</script>

<header
  class={`fixed inset-x-0 top-0 z-50 transition-all duration-300 ${
    scrolled
      ? "border-b border-border bg-background/70 backdrop-blur-xl"
      : "bg-gradient-to-b from-background/80 to-transparent"
  }`}
>
  <div class="flex h-16 items-center gap-6 px-6 md:h-18 md:px-12 lg:px-20">
    <button
      type="button"
      data-hamburger
      onclick={() => (showMobileMenu = !showMobileMenu)}
      aria-label="Menu"
      aria-expanded={showMobileMenu}
      class={`flex h-11 w-11 lg:h-9 lg:w-9 items-center justify-center rounded-lg transition-colors hover:bg-surface lg:hidden ${showMobileMenu ? 'bg-surface text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
    >
      <Menu class="h-5 w-5" />
    </button>

    <a href="/" class="shrink-0"><Logo /></a>

    <nav
      class="hidden items-center gap-1 lg:flex"
      data-sveltekit-preload-code="eager"
      data-sveltekit-preload-data="hover"
    >
      {#each nav as item (item.href)}
        <a
          href={item.href}
          class={`relative rounded-full px-4 py-1.5 text-sm font-medium transition-colors ${
            currentPath === item.href
              ? "text-foreground bg-primary-glow/10 ring-1 ring-primary-glow/30"
              : "text-muted-foreground hover:text-foreground"
          }`}
        >
          {item.label}
        </a>
      {/each}
    </nav>

    <div class="flex flex-1 items-center justify-end gap-2">

      <!-- Desktop search -->
      <div class="relative hidden w-full max-w-md md:block" data-search-container>
        <Search class="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          bind:this={searchInputEl}
          bind:value={searchQuery}
          type="search"
          autocomplete="off"
          placeholder="Search movies, shows..."
          class="h-10 w-full rounded-full border border-border bg-surface/60 pl-11 pr-10 text-sm text-foreground outline-none ring-0 transition-all placeholder:text-muted-foreground/70 focus:border-primary/60 focus:bg-surface focus:shadow-glow"
          onfocus={handleSearchFocus}
          onkeydown={handleSearchKeydown}
        />
        <kbd class="pointer-events-none absolute right-3 top-1/2 hidden -translate-y-1/2 rounded border border-border bg-background/60 px-1.5 py-0.5 text-[10px] text-muted-foreground lg:inline-block">{isMac ? '⌘K' : 'Ctrl K'}</kbd>

        {#if showSearchDropdown}
          <div class="absolute left-0 right-0 top-[calc(100%+6px)] z-50 overflow-hidden rounded-2xl border border-border bg-surface-elevated shadow-2xl backdrop-blur-xl">
            {#if isSearchLoading()}
              <div class="px-4 py-3 text-sm text-muted-foreground">Loading…</div>
            {:else if searchResults.length === 0}
              <div class="px-4 py-3 text-sm text-muted-foreground">No results for "{searchQuery}"</div>
            {:else}
              <ul>
                {#each searchResults as hit (hitKey(hit))}
                  <li>
                    <button
                      type="button"
                      class="flex w-full items-center gap-3 px-4 py-2.5 text-left transition-colors hover:bg-surface/60"
                      onmousedown={(e) => { e.preventDefault(); handleResultClick(hit); }}
                    >
                      {#if hit.kind === 'movie' || hit.kind === 'series'}
                        {#if hit.posterUrl}
                          <img src={hit.posterUrl} alt="" class="h-10 w-7 shrink-0 rounded object-cover" onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')} />
                        {:else}
                          <div class="flex h-10 w-7 shrink-0 items-center justify-center rounded bg-surface-elevated text-muted-foreground">
                            {#if hit.kind === 'series'}<Tv class="h-3.5 w-3.5" />{:else}<Film class="h-3.5 w-3.5" />{/if}
                          </div>
                        {/if}
                        <div class="min-w-0 flex-1">
                          <div class="truncate text-sm font-medium">{hit.title}</div>
                          <div class="text-xs text-muted-foreground">
                            {hit.year && hit.year > 0 ? `${hit.year} · ` : ''}{hit.kind === 'series' ? 'TV Series' : 'Movie'}
                          </div>
                        </div>
                      {:else if hit.kind === 'person'}
                        {#if hit.profileUrl}
                          <img src={hit.profileUrl} alt="" class="h-10 w-10 shrink-0 rounded-full object-cover" onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')} />
                        {:else}
                          <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-surface-elevated text-muted-foreground">
                            <User class="h-4 w-4" />
                          </div>
                        {/if}
                        <div class="min-w-0 flex-1">
                          <div class="truncate text-sm font-medium">{hit.name}</div>
                          <div class="text-xs text-muted-foreground">
                            Person · {hit.creditCount} credit{hit.creditCount === 1 ? '' : 's'}
                          </div>
                        </div>
                      {:else}
                        {#if hit.posterUrl}
                          <img src={hit.posterUrl} alt="" class="h-10 w-7 shrink-0 rounded object-cover" onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')} />
                        {:else}
                          <div class="flex h-10 w-7 shrink-0 items-center justify-center rounded bg-surface-elevated text-muted-foreground">
                            <Layers class="h-3.5 w-3.5" />
                          </div>
                        {/if}
                        <div class="min-w-0 flex-1">
                          <div class="truncate text-sm font-medium">{hit.name}</div>
                          <div class="text-xs text-muted-foreground">
                            Collection · {hit.movieCount} movie{hit.movieCount === 1 ? '' : 's'}
                          </div>
                        </div>
                      {/if}
                    </button>
                  </li>
                {/each}
              </ul>
              <a
                href={`/search?q=${encodeURIComponent(searchQuery.trim())}`}
                class="flex items-center gap-2 border-t border-border px-4 py-2.5 text-xs text-muted-foreground transition-colors hover:bg-surface/40 hover:text-foreground"
                onmousedown={(e) => { e.preventDefault(); searchFocused = false; goto(`/search?q=${encodeURIComponent(searchQuery.trim())}`); }}
              >
                <Search class="h-3.5 w-3.5" /> See all results for "{searchQuery}"
              </a>
            {/if}
          </div>
        {/if}
      </div>

      <!-- Mobile search icon -->
      <button
        type="button"
        aria-label="Search"
        class="flex h-11 w-11 lg:h-9 lg:w-9 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-surface hover:text-foreground md:hidden"
        onclick={() => goto('/search')}
      >
        <Search class="h-5 w-5" />
      </button>

      <!-- Notifications + Recently Played -->
      <div class="relative hidden items-center gap-1 sm:flex" data-notif-container>
        <!-- Bell: notifications -->
        <button
          type="button"
          aria-label="Notifications"
          title="Notifications"
          onclick={openNotifications}
          class={`relative flex h-11 w-11 lg:h-9 lg:w-9 items-center justify-center rounded-full transition-colors ${showNotifications ? 'bg-surface text-foreground' : 'text-muted-foreground hover:bg-surface hover:text-foreground'}`}
        >
          <Bell class="h-5 w-5" />
          {#if notificationsLoaded && notifications.filter(n => !n.dismissed).length > 0}
            <span class="absolute right-2 top-2 h-1.5 w-1.5 rounded-full bg-accent shadow-glow"></span>
          {/if}
        </button>

        <!-- Clock: recently played -->
        <button
          type="button"
          aria-label="Recently Played"
          title="Recently Played"
          onclick={openRecentPlayed}
          class={`relative flex h-11 w-11 lg:h-9 lg:w-9 items-center justify-center rounded-full transition-colors ${showRecentPlayed ? 'bg-surface text-foreground' : 'text-muted-foreground hover:bg-surface hover:text-foreground'}`}
        >
          <Clock class="h-5 w-5" />
          {#if recentLoaded && recentItems.length > 0}
            <span class="absolute right-2 top-2 h-1.5 w-1.5 rounded-full bg-accent shadow-glow"></span>
          {/if}
        </button>

        <!-- Notifications dropdown (under Bell) -->
        {#if showNotifications}
          <div class="absolute right-9 top-[calc(100%+8px)] z-50 w-80 overflow-hidden rounded-2xl border border-border bg-surface-elevated shadow-2xl backdrop-blur-xl">
            <div class="flex items-center justify-between border-b border-border px-4 py-3">
              <h3 class="text-sm font-semibold">Notifications</h3>
              {#if notifications.length > 0}
                <button onclick={handleDismissAll} class="text-xs text-muted-foreground transition-colors hover:text-foreground">
                  Dismiss all
                </button>
              {/if}
            </div>
            {#if !notificationsLoaded}
              <div class="px-4 py-3 text-sm text-muted-foreground">Loading…</div>
            {:else if notifications.length === 0}
              <div class="flex flex-col items-center gap-2 px-4 py-8 text-center">
                <Bell class="h-8 w-8 text-muted-foreground/30" />
                <p class="text-sm text-muted-foreground">No notifications</p>
              </div>
            {:else}
              <ul class="max-h-80 overflow-y-auto">
                {#each notifications as notif (notif.id)}
                  <li class="flex items-start gap-3 px-4 py-3 transition-colors hover:bg-surface/40">
                    <div class="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-surface text-muted-foreground">
                      {#if notif.kind?.includes('failed')}
                        <AlertTriangle class="h-3.5 w-3.5" />
                      {:else}
                        <CheckCircle class="h-3.5 w-3.5" />
                      {/if}
                    </div>
                    <div class="min-w-0 flex-1">
                      <p class="text-sm font-medium leading-tight">{notif.title}</p>
                      {#if notif.message}
                        <p class="mt-0.5 text-xs text-muted-foreground">{notif.message}</p>
                      {/if}
                    </div>
                    <button
                      onclick={() => handleDismiss(notif.id!)}
                      aria-label="Dismiss"
                      class="mt-0.5 shrink-0 text-muted-foreground transition-colors hover:text-foreground"
                    >
                      <X class="h-3.5 w-3.5" />
                    </button>
                  </li>
                {/each}
              </ul>
            {/if}
          </div>
        {/if}

        <!-- Recently Played dropdown (under Clock) -->
        {#if showRecentPlayed}
          <div class="absolute right-0 top-[calc(100%+8px)] z-50 w-80 overflow-hidden rounded-2xl border border-border bg-surface-elevated shadow-2xl backdrop-blur-xl">
            <div class="border-b border-border px-4 py-3">
              <h3 class="text-sm font-semibold">Recently Played</h3>
            </div>
            {#if !recentLoaded}
              <div class="px-4 py-3 text-sm text-muted-foreground">Loading…</div>
            {:else if recentItems.length === 0}
              <div class="flex flex-col items-center gap-2 px-4 py-8 text-center">
                <Clock class="h-8 w-8 text-muted-foreground/30" />
                <p class="text-sm text-muted-foreground">Nothing played yet</p>
              </div>
            {:else}
              <ul class="max-h-80 overflow-y-auto">
                {#each recentItems as item (item.mediaSourceId)}
                  <li class="flex items-center gap-3 px-4 py-3 transition-colors hover:bg-surface/40">
                    <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-surface-elevated text-muted-foreground">
                      {#if item.kind === 'series'}<Tv class="h-4 w-4" />{:else}<Film class="h-4 w-4" />{/if}
                    </div>
                    <div class="min-w-0 flex-1">
                      <p class="truncate text-sm font-medium">{friendlyName(item)}</p>
                      {#if formatProgress(item)}
                        <p class="text-xs text-muted-foreground">{formatProgress(item)}</p>
                      {/if}
                    </div>
                    {#if item.mediaSourceId}
                      <a
                        href={`/play/${item.mediaSourceId}?title=${encodeURIComponent(friendlyName(item))}&back=/`}
                        class="shrink-0 rounded-full bg-foreground/[0.06] px-2.5 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.12] hover:text-foreground"
                      >
                        Resume
                      </a>
                    {/if}
                  </li>
                {/each}
              </ul>
              <a
                href="/continue-watching"
                class="flex items-center justify-center border-t border-border px-4 py-3 text-xs text-muted-foreground transition-colors hover:bg-surface/40 hover:text-foreground"
              >
                View all
              </a>
            {/if}
          </div>
        {/if}
      </div>

      <!-- Settings link -->
      <a
        href="/settings"
        aria-label="Settings"
        title="Settings"
        class={`hidden h-9 w-9 items-center justify-center rounded-full transition-colors sm:flex ${
          currentPath === "/settings"
            ? "text-foreground bg-surface"
            : "text-muted-foreground hover:bg-surface hover:text-foreground"
        }`}
      >
        <Settings class="h-5 w-5" />
      </a>

      <!-- Profile avatar -->
      <div class="relative" data-profile-container>
        <button
          type="button"
          onclick={toggleProfile}
          class={`flex h-11 w-11 lg:h-9 lg:w-9 items-center justify-center overflow-hidden rounded-full ring-1 ring-white/20 transition-shadow ${showProfile ? 'shadow-glow' : ''}`}
          aria-label="Profile"
          title={activeProfile ? activeProfile.displayName : 'Profile'}
        >
          {#if activeProfile}
            {@const src = profileAvatarSrc(activeProfile)}
            {#if src}
              <img src={src} alt={activeProfile.displayName} class="h-full w-full object-cover" />
            {:else}
              <div class="flex h-full w-full items-center justify-center bg-gradient-primary text-sm font-semibold text-white">
                {profileInitials(activeProfile.displayName)}
              </div>
            {/if}
          {:else}
            <div class="flex h-full w-full items-center justify-center bg-gradient-primary text-sm font-semibold text-white">
              <User class="h-4 w-4" />
            </div>
          {/if}
        </button>

        {#if showProfile}
          <div class="absolute right-0 top-[calc(100%+8px)] z-50 w-56 overflow-hidden rounded-2xl border border-border bg-surface-elevated shadow-2xl backdrop-blur-xl">
            <div class="border-b border-border px-4 py-3">
              {#if activeProfile}
                <p class="text-sm font-semibold">{activeProfile.displayName}</p>
                <p class="text-xs text-muted-foreground">{activeProfile.isRestricted ? 'Kids profile' : 'Profile'}</p>
              {:else}
                <p class="text-sm font-semibold">Account</p>
                <p class="text-xs text-muted-foreground">Local account</p>
              {/if}
            </div>
            <ul class="py-1">
              <li>
                <button
                  type="button"
                  onclick={openProfileSwitcher}
                  class="flex w-full items-center gap-3 px-4 py-2.5 text-sm text-muted-foreground transition-colors hover:bg-surface/60 hover:text-foreground"
                >
                  <Users class="h-4 w-4" /> Switch Profile
                </button>
              </li>
              <li>
                <a
                  href="/settings"
                  onclick={() => (showProfile = false)}
                  class="flex items-center gap-3 px-4 py-2.5 text-sm text-muted-foreground transition-colors hover:bg-surface/60 hover:text-foreground"
                >
                  <Settings class="h-4 w-4" /> Settings
                </a>
              </li>
              <li>
                <a
                  href="/watchlist"
                  onclick={() => (showProfile = false)}
                  class="flex items-center gap-3 px-4 py-2.5 text-sm text-muted-foreground transition-colors hover:bg-surface/60 hover:text-foreground"
                >
                  <Bookmark class="h-4 w-4" /> Watchlist
                </a>
              </li>
              <li class="border-t border-border">
                <a
                  href="/signin"
                  onclick={() => (showProfile = false)}
                  class="flex items-center gap-3 px-4 py-2.5 text-sm text-muted-foreground transition-colors hover:bg-surface/60 hover:text-foreground"
                >
                  <LogOut class="h-4 w-4" /> Sign out
                </a>
              </li>
            </ul>
          </div>
        {/if}
      </div>
    </div>
  </div>
</header>

{#if showMobileMenu}
  <div
    data-mobile-menu
    class="fixed inset-x-0 top-16 z-40 border-b border-border bg-surface-elevated shadow-xl backdrop-blur-xl lg:hidden"
  >
    <nav
      class="px-5 py-3"
      data-sveltekit-preload-code="eager"
      data-sveltekit-preload-data="tap"
    >
      {#each nav as item (item.href)}
        <a
          href={item.href}
          onclick={() => (showMobileMenu = false)}
          class={`flex items-center rounded-xl px-4 py-3 text-sm font-medium transition-colors ${
            currentPath === item.href
              ? 'bg-primary-glow/10 text-foreground ring-1 ring-primary-glow/20'
              : 'text-muted-foreground hover:bg-foreground/[0.04] hover:text-foreground'
          }`}
        >
          {item.label}
        </a>
      {/each}
    </nav>
  </div>
{/if}
