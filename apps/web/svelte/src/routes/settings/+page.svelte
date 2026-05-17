<script lang="ts">
  import {
    ArrowRightLeft,
    Check,
    ClipboardList,
    Compass,
    Database,
    HardDrive,
    Info,
    KeyRound,
    LayoutDashboard,
    Library,
    Link2,
    Play,
    ScanSearch,
    Search,
    ShieldCheck,
    Sliders,
    User,
    Users,
    Wifi
  } from "lucide-svelte";
  import Header from "$lib/components/Header.svelte";

  type Group = "Media-Server" | "Devices" | "Advanced";
  type Mode = "balanced" | "automatic" | "advanced";

  type Section = {
    id: string;
    label: string;
    icon: typeof User;
    group: Group;
    hint?: string;
  };

  const sections: Section[] = [
    { id: "dashboard", label: "Dashboard", icon: LayoutDashboard, group: "Media-Server", hint: "Overview & health" },
    { id: "general", label: "General", icon: User, group: "Media-Server", hint: "Server name, locale, theme" },
    { id: "libraries", label: "Libraries", icon: Library, group: "Media-Server", hint: "Folders, scanners, art" },
    { id: "scanning", label: "Scanning", icon: ScanSearch, group: "Media-Server", hint: "Cadence & rules" },
    { id: "metadata", label: "Metadata", icon: Database, group: "Media-Server", hint: "Providers, keys, matches" },
    { id: "playback", label: "Playback", icon: Play, group: "Media-Server", hint: "Streaming defaults" },
    { id: "transcoding", label: "Transcoding", icon: Sliders, group: "Media-Server", hint: "Hardware & quality" },
    { id: "storage", label: "Storage", icon: HardDrive, group: "Media-Server", hint: "Disks & cache" },
    { id: "network", label: "Network", icon: Wifi, group: "Media-Server", hint: "Ports, remote access" },
    { id: "migration", label: "Migration", icon: ArrowRightLeft, group: "Media-Server", hint: "Import from elsewhere" },
    { id: "discovery", label: "Discovery", icon: Compass, group: "Media-Server", hint: "Recommendations" },
    { id: "users", label: "Users", icon: Users, group: "Media-Server", hint: "Accounts & roles" },
    { id: "pairing", label: "Pairing", icon: Link2, group: "Devices", hint: "Link a new device" },
    { id: "approved", label: "Approved Devices", icon: ShieldCheck, group: "Devices", hint: "Trusted clients" },
    { id: "planned", label: "Planned", icon: ClipboardList, group: "Advanced", hint: "Coming soon" },
    { id: "about", label: "About", icon: Info, group: "Advanced", hint: "Build & licenses" }
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
    { id: "fanart", name: "Fanart.tv", blurb: "Logos, banners, thumbs, and extra artwork" },
    { id: "musicbrainz", name: "MusicBrainz", blurb: "Track, album, and artist metadata for music libraries" }
  ];

  const groups: Group[] = ["Media-Server", "Devices", "Advanced"];

  let active = $state("metadata");
  let mode = $state<Mode>("advanced");
  let keys = $state<Record<string, string>>({});
  let q = $state("");
  let headerScrolled = $state(false);
  let mainRef = $state<HTMLDivElement | null>(null);

  let current = $derived(sections.find((section) => section.id === active) ?? sections[4]);

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
            Media-<em>Server</em>
          </h1>
          <p class="mt-3 max-w-xl text-sm text-muted-foreground">
            Choose metadata sources, review matches, and tune how Xuva pulls artwork and information for your library.
          </p>
        </div>
        <div class="flex items-center gap-3">
          <span class="hairline inline-flex items-center gap-2 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-[10px] font-semibold uppercase tracking-[0.25em] text-emerald-300">
            <span class="h-1.5 w-1.5 rounded-full bg-emerald-400 shadow-[0_0_10px_currentColor]"></span>
            Healthy
          </span>
          <div class="hidden text-right text-[11px] uppercase tracking-[0.22em] text-muted-foreground md:block">
            Updated <span class="text-foreground/90">10:39:30 am</span>
          </div>
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
          <div class="flex items-center gap-2">
            <button class="hairline rounded-full bg-foreground/[0.04] px-4 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground">
              Discard
            </button>
            <button class="rounded-full bg-gradient-primary px-5 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110">
              Save changes
            </button>
          </div>
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
