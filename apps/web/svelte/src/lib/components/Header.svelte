<script lang="ts">
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import { Bell, Menu, Search, Settings, Film, Tv, LogOut, User } from "lucide-svelte";
  import Logo from "./Logo.svelte";
  import { primeSearchCatalogue, searchCatalogue, isSearchLoading } from "$lib/stores/searchStore.svelte";
  import { getPlaybackRecent } from "$lib/api/home";
  import type { Media } from "$lib/mock-data";
  import type { PlaybackRecentItem } from "$lib/api/home";

  const nav = [
    { label: "Home", href: "/" },
    { label: "Movies", href: "/movies" },
    { label: "TV", href: "/tv" }
  ];

  let scrolled = $state(false);
  const currentPath = $derived(page.url.pathname);

  // ── Search ─────────────────────────────────────────────────────────────────
  let searchQuery = $state("");
  let searchFocused = $state(false);
  let searchInputEl = $state<HTMLInputElement | null>(null);

  const searchResults = $derived(searchCatalogue(searchQuery, 8));
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

  function handleResultClick(m: Media) {
    searchQuery = "";
    searchFocused = false;
    goto(m.type === "Series" ? `/tv/${m.id}` : `/movies/${m.id}`);
  }

  function handleSearchFocus() {
    searchFocused = true;
    primeSearchCatalogue();
  }

  // ── Notifications ──────────────────────────────────────────────────────────
  let showNotifications = $state(false);
  let recentItems = $state<PlaybackRecentItem[]>([]);
  let notificationsLoaded = $state(false);

  async function openNotifications() {
    showNotifications = !showNotifications;
    showProfile = false;
    if (!notificationsLoaded) {
      try {
        const resp = await getPlaybackRecent(undefined, 8);
        recentItems = resp.recent ?? [];
      } catch {
        recentItems = [];
      } finally {
        notificationsLoaded = true;
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
      // strip extension and quality tags
      return item.relPath
        .replace(/\.[a-z0-9]{2,4}$/i, '')
        .replace(/\s*\([^)]*(?:remux|bluray|1080p|2160p|720p|hdtv)[^)]*\)/gi, '')
        .split('/').pop() ?? item.relPath;
    }
    return 'Unknown';
  }

  // ── Profile ────────────────────────────────────────────────────────────────
  let showProfile = $state(false);

  // ── Mobile menu ────────────────────────────────────────────────────────────
  let showMobileMenu = $state(false);

  function toggleProfile() {
    showProfile = !showProfile;
    showNotifications = false;
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
      if (!target.closest("[data-notif-container]")) showNotifications = false;
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
      class={`flex h-9 w-9 items-center justify-center rounded-lg transition-colors hover:bg-surface lg:hidden ${showMobileMenu ? 'bg-surface text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
    >
      <Menu class="h-5 w-5" />
    </button>

    <a href="/" class="shrink-0"><Logo /></a>

    <nav class="hidden items-center gap-1 lg:flex">
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
                {#each searchResults as m (m.id)}
                  <li>
                    <button
                      type="button"
                      class="flex w-full items-center gap-3 px-4 py-2.5 text-left transition-colors hover:bg-surface/60"
                      onmousedown={(e) => { e.preventDefault(); handleResultClick(m); }}
                    >
                      {#if m.poster}
                        <img src={m.poster} alt="" class="h-10 w-7 shrink-0 rounded object-cover" onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')} />
                      {:else}
                        <div class="flex h-10 w-7 shrink-0 items-center justify-center rounded bg-surface-elevated text-muted-foreground">
                          {#if m.type === "Series"}<Tv class="h-3.5 w-3.5" />{:else}<Film class="h-3.5 w-3.5" />{/if}
                        </div>
                      {/if}
                      <div class="min-w-0 flex-1">
                        <div class="truncate text-sm font-medium">{m.title}</div>
                        <div class="text-xs text-muted-foreground">
                          {m.year > 0 ? m.year : ''}{m.year > 0 ? ' · ' : ''}{m.type === 'Series' ? 'TV Series' : 'Movie'}
                        </div>
                      </div>
                    </button>
                  </li>
                {/each}
              </ul>
              <a
                href={`/search?q=${encodeURIComponent(searchQuery.trim())}`}
                class="flex items-center gap-2 border-t border-border px-4 py-2.5 text-xs text-muted-foreground transition-colors hover:bg-surface/40 hover:text-foreground"
                onmousedown={() => { searchFocused = false; }}
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
        class="flex h-9 w-9 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-surface hover:text-foreground md:hidden"
        onclick={() => goto('/search')}
      >
        <Search class="h-5 w-5" />
      </button>

      <!-- Notifications -->
      <div class="relative hidden sm:block" data-notif-container>
        <button
          type="button"
          aria-label="Notifications"
          onclick={openNotifications}
          class={`relative flex h-9 w-9 items-center justify-center rounded-full transition-colors ${showNotifications ? 'bg-surface text-foreground' : 'text-muted-foreground hover:bg-surface hover:text-foreground'}`}
        >
          <Bell class="h-5 w-5" />
          {#if notificationsLoaded && recentItems.length > 0}
            <span class="absolute right-2 top-2 h-1.5 w-1.5 rounded-full bg-accent shadow-glow"></span>
          {/if}
        </button>

        {#if showNotifications}
          <div class="absolute right-0 top-[calc(100%+8px)] z-50 w-80 overflow-hidden rounded-2xl border border-border bg-surface-elevated shadow-2xl backdrop-blur-xl">
            <div class="border-b border-border px-4 py-3">
              <h3 class="text-sm font-semibold">Recently Played</h3>
            </div>
            {#if !notificationsLoaded}
              <div class="px-4 py-3 text-sm text-muted-foreground">Loading…</div>
            {:else if recentItems.length === 0}
              <div class="flex flex-col items-center gap-2 px-4 py-8 text-center">
                <Bell class="h-8 w-8 text-muted-foreground/30" />
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
                        href={`/play/${item.mediaSourceId}`}
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
          class={`flex h-9 w-9 items-center justify-center rounded-full bg-gradient-primary text-sm font-semibold text-white ring-1 ring-white/20 transition-shadow ${showProfile ? 'shadow-glow' : ''}`}
          aria-label="Profile"
        >
          A
        </button>

        {#if showProfile}
          <div class="absolute right-0 top-[calc(100%+8px)] z-50 w-56 overflow-hidden rounded-2xl border border-border bg-surface-elevated shadow-2xl backdrop-blur-xl">
            <div class="border-b border-border px-4 py-3">
              <p class="text-sm font-semibold">Admin</p>
              <p class="text-xs text-muted-foreground">Local account</p>
            </div>
            <ul class="py-1">
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
                  <User class="h-4 w-4" /> Watchlist
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
    <nav class="px-5 py-3">
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
