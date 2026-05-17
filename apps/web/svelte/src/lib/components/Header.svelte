<script lang="ts">
  import { page } from "$app/state";
  import { Bell, Menu, Search, Settings } from "lucide-svelte";
  import Logo from "./Logo.svelte";

  const nav = [
    { label: "Home", href: "/" },
    { label: "Movies", href: "/movies" },
    { label: "TV", href: "/tv" }
  ];

  let scrolled = $state(false);
  const currentPath = $derived(page.url.pathname);

  $effect(() => {
    if (typeof window === "undefined") return;
    const onScroll = () => {
      scrolled = window.scrollY > 12;
    };
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
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
      class="flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-surface hover:text-foreground lg:hidden"
      aria-label="Menu"
    >
      <Menu class="h-5 w-5" />
    </button>

    <a href="/" class="shrink-0">
      <Logo />
    </a>

    <nav class="hidden items-center gap-1 lg:flex">
      {#each nav as item (item.href)}
        <a
          href={item.href}
          class={`relative rounded-full px-4 py-1.5 text-sm font-medium transition-colors ${
            currentPath === item.href
              ? "text-foreground bg-surface"
              : "text-muted-foreground hover:text-foreground"
          }`}
        >
          {item.label}
        </a>
      {/each}
    </nav>

    <div class="flex flex-1 items-center justify-end gap-2">
      <div class="relative hidden w-full max-w-md md:block">
        <Search class="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          type="search"
          placeholder="Search movies, shows, people..."
          class="h-10 w-full rounded-full border border-border bg-surface/60 pl-11 pr-4 text-sm text-foreground outline-none ring-0 transition-all placeholder:text-muted-foreground/70 focus:border-primary/60 focus:bg-surface focus:shadow-glow"
        />
        <kbd class="pointer-events-none absolute right-3 top-1/2 hidden -translate-y-1/2 rounded border border-border bg-background/60 px-1.5 py-0.5 text-[10px] text-muted-foreground lg:inline-block">
          ⌘K
        </kbd>
      </div>

      <button
        type="button"
        aria-label="Search"
        class="flex h-9 w-9 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-surface hover:text-foreground md:hidden"
      >
        <Search class="h-5 w-5" />
      </button>
      <button
        type="button"
        aria-label="Notifications"
        class="relative hidden h-9 w-9 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-surface hover:text-foreground sm:flex"
      >
        <Bell class="h-5 w-5" />
        <span class="absolute right-2 top-2 h-1.5 w-1.5 rounded-full bg-accent shadow-glow"></span>
      </button>
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
      <button
        type="button"
        class="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-primary text-sm font-semibold text-white shadow-glow ring-1 ring-white/20"
        aria-label="Profile"
      >
        A
      </button>
    </div>
  </div>
</header>

