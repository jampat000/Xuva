<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { Check, ChevronDown, Play, Plus, RotateCcw, Search, SlidersHorizontal, X } from "lucide-svelte";
  import type { Media } from "$lib/mock-data";
  import { toggleWatchlist, isInWatchlist } from '$lib/stores/watchlistStore.svelte';
  import { getAuthSession } from '$lib/api/auth';
  import { updateUserPreferences } from '$lib/api/operator';
  import { artworkSrc, artworkSrcset } from '$lib/api/artwork-url';

  let { eyebrow, title, tagline, items, kind, loading = false, baseHref = "", showHero = true } = $props<{
    eyebrow: string;
    title: string;
    tagline: string;
    items: Media[];
    kind: "Movies" | "TV";
    loading?: boolean;
    baseHref?: string;
    /** When false, renders a compact text header instead of the featured-item hero. */
    showHero?: boolean;
  }>();

  // ── Types ──────────────────────────────────────────────────────────────────
  type Density = "S" | "M" | "L";
  type Sort =
    | "az" | "za"
    | "year-desc" | "year-asc"
    | "added-desc" | "added-asc"
    | "rating-desc" | "rating-asc"
    | "runtime-asc" | "runtime-desc"
    | "parental-asc"
    | "unwatched-first" | "watched-first"
    | "versions-desc" | "versions-asc"
    | "random";

  // ── Density grid ──────────────────────────────────────────────────────────
  // Bumped +2 columns at every breakpoint vs the previous values. Users wanted
  // smaller posters across S / M / L so more titles fit on screen per row —
  // the previous L (5 cols at xl) was reading more like a marketing grid than
  // a browse view. Card heights scale automatically with column width since
  // they use aspect-[2/3], so no other tuning needed.
  const densityGrid: Record<Density, string> = {
    S: "grid-cols-4 sm:grid-cols-6 md:grid-cols-8 lg:grid-cols-10 xl:grid-cols-12",
    M: "grid-cols-3 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-8 xl:grid-cols-9",
    L: "grid-cols-2 sm:grid-cols-3 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7"
  };

  // ── Virtual window state ──────────────────────────────────────────────────
  // Renders ONLY ~70 cards (visible + overscan) instead of all 4000+. Without
  // this, Svelte's keyed teardown of the full grid blocks the main thread for
  // 4-5 seconds when navigating away from /movies — confirmed via Chrome
  // PerformanceObserver: a single longtask of 4607 ms on a 4008-item library.
  // content-visibility:auto avoids PAINT for off-screen cards but doesn't help
  // with Svelte's destroy phase or with the DOM-node count the browser has to
  // walk. True virtualization (this) drops the in-DOM count to ~70, making
  // route transitions sub-100ms.
  //
  // The window is recomputed on scroll (throttled via requestAnimationFrame)
  // and on resize / density-change (which alter row height and items-per-row).
  // Spacer divs above/below the rendered slice preserve scroll height so the
  // scrollbar, anchor jumps, and Find-in-Page all still work on items that
  // haven't been rendered.
  let gridEl: HTMLElement | undefined = $state();
  let scrollY     = $state(typeof window !== 'undefined' ? window.scrollY : 0);
  let viewportH   = $state(typeof window !== 'undefined' ? window.innerHeight : 800);
  let gridTop     = $state(0);
  let rowHeight   = $state(360); // measured on mount/resize; fallback covers cold paint
  let itemsPerRow = $state(7);
  const OVERSCAN_ROWS = 3;       // rows beyond viewport edges to keep mounted

  // ── Sort options ───────────────────────────────────────────────────────────
  const sortOptions: { value: Sort; label: string }[] = [
    { value: "az",              label: "Title A → Z" },
    { value: "za",              label: "Title Z → A" },
    { value: "year-desc",       label: "Year — Newest" },
    { value: "year-asc",        label: "Year — Oldest" },
    { value: "added-desc",      label: "Date Added — Newest" },
    { value: "added-asc",       label: "Date Added — Oldest" },
    { value: "rating-desc",     label: "Rating — Highest" },
    { value: "rating-asc",      label: "Rating — Lowest" },
    { value: "runtime-asc",     label: "Runtime — Shortest" },
    { value: "runtime-desc",    label: "Runtime — Longest" },
    { value: "unwatched-first", label: "Unwatched First" },
    { value: "watched-first",   label: "Watched First" },
    { value: "versions-desc",   label: "Most Files" },
    { value: "versions-asc",    label: "Fewest Files" },
    { value: "parental-asc",    label: "Parental Rating" },
    { value: "random",          label: "Random" },
  ];

  // Parental-rating sort order (G → most restricted)
  const RATING_ORDER = ['G','PG','PG-13','R','NC-17','TV-Y','TV-Y7','TV-Y7-FV','TV-G','TV-PG','TV-14','TV-MA'];

  // Header description sentence: "{count} films across {N} genres." or
  // "{count} TV shows." Building this in JS instead of in the template avoids
  // Svelte 5's whitespace handling around {#if} blocks, which collapsed the
  // space between expression and conditional on one line ("filmsacross") and
  // emitted a stray space before the period on another line ("TV shows .").
  // A $derived expression is the only stable way to compose this sentence.

  // ── Filter & UI state ─────────────────────────────────────────────────────
  let q            = $state("");
  let sort         = $state<Sort>("az");
  // Read cached density from localStorage synchronously so the correct size is
  // applied on the very first render — no flash of the wrong grid size.
  const DENSITY_KEY = 'xuva:poster-size';
  function readCachedDensity(): Density {
    try {
      const v = typeof localStorage !== 'undefined' ? localStorage.getItem(DENSITY_KEY) : null;
      if (v === 'S' || v === 'M' || v === 'L') return v;
    } catch { /* SSR or privacy mode */ }
    return 'M';
  }
  let density      = $state<Density>(readCachedDensity());
  let sortOpen     = $state(false);
  let filterOpen   = $state(false);
  let randomSeed   = $state(Date.now());

  // Multi-select filter sets
  let selectedGenres   = $state(new Set<string>());
  let selectedDecades  = $state(new Set<string>());
  let selectedRatings  = $state(new Set<string>());
  let selectedStudios  = $state(new Set<string>());
  let watchFilter      = $state<"all" | "watched" | "unwatched">("all");
  let onlyNeedsReview  = $state(false);
  let onlyMultiVersion = $state(false);
  let onlyMissingMeta  = $state(false);
  let onlyUnprobed     = $state(false);

  // ── Pseudo-random helper ───────────────────────────────────────────────────
  function pseudoRandom(id: string, seed: number): number {
    let h = seed & 0x7fffffff;
    for (let i = 0; i < id.length; i++) h = (Math.imul(h, 31) + id.charCodeAt(i)) & 0x7fffffff;
    return h / 0x7fffffff;
  }

  // ── Decade helper ─────────────────────────────────────────────────────────
  function getDecade(year: number): string {
    if (!year || year <= 0) return '';
    if (year < 1970) return 'Older';
    return `${Math.floor(year / 10) * 10}s`;
  }

  // ── Derived filter options ─────────────────────────────────────────────────
  type GenreChip = { name: string; count: number };

  const genreChips = $derived.by<GenreChip[]>(() => {
    const counts = new Map<string, number>();
    for (const item of items) {
      for (const g of item.genres ?? []) {
        counts.set(g, (counts.get(g) ?? 0) + 1);
      }
    }
    const arr: GenreChip[] = [];
    for (const [name, count] of counts) arr.push({ name, count });
    arr.sort((a, b) => (b.count - a.count) || a.name.localeCompare(b.name));
    return arr;
  });

  // See the comment above RATING_ORDER for why this is composed in JS.
  // No trailing period — matches Watchlist's title styling and stops the page
  // reading "title. description." which felt over-punctuated to the user.
  const headerSentence = $derived.by(() => {
    const noun = kind === 'TV' ? 'TV shows' : 'films';
    const count = items.length.toLocaleString();
    if (genreChips.length > 0) {
      return `${count} ${noun} across ${genreChips.length} genres`;
    }
    return `${count} ${noun}`;
  });

  // ── Alphabet jump rail (#403) ──────────────────────────────────────────────
  // Letter → index of the first item in the *filtered + sorted* list whose
  // sort-title starts with that letter. Computed against `filtered` so the
  // rail tracks the visible result set: filter to a single genre and the
  // index map shrinks accordingly, with letters that have no items rendered
  // dim and non-interactive.
  //
  // Only renders when sort is alphabetical (az / za) and on lg+ breakpoint;
  // the filter panel covers small-screen navigation. Click handler computes
  // the y-coordinate from gridTop + (index/itemsPerRow)*rowHeight and
  // window.scrollTo — works with the virtualization spacer because the
  // outer page scroll position is what drives the virtual window.
  function jumpLetter(title: string): string {
    // Strip common articles so "The Godfather" indexes under G, not T.
    const stripped = title.replace(/^(the|a|an)\s+/i, '');
    const ch = stripped.charAt(0).toUpperCase();
    if (!ch) return '#';
    if (ch >= '0' && ch <= '9') return '#';
    if (ch >= 'A' && ch <= 'Z') return ch;
    return '#';
  }
  const ALPHABET = ['#', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'];
  const isAlphabeticalSort = $derived(sort === 'az' || sort === 'za');
  const letterIndex = $derived.by(() => {
    if (!isAlphabeticalSort) return new Map<string, number>();
    const map = new Map<string, number>();
    for (let i = 0; i < filtered.length; i++) {
      const letter = jumpLetter(filtered[i].title);
      if (!map.has(letter)) map.set(letter, i);
    }
    return map;
  });
  function jumpToLetter(letter: string) {
    const idx = letterIndex.get(letter);
    if (idx === undefined) return;
    const rowH = rowHeight > 0 ? rowHeight : 360;
    const perRow = itemsPerRow > 0 ? itemsPerRow : 7;
    const targetRow = Math.floor(idx / perRow);
    const targetY = Math.max(0, gridTop + targetRow * rowH - 90); // 90px = header offset
    if (typeof window !== 'undefined') {
      window.scrollTo({ top: targetY, behavior: 'smooth' });
    }
  }

  const availableDecades = $derived.by<string[]>(() => {
    const set = new Set<string>();
    for (const item of items) {
      const d = getDecade(item.year);
      if (d) set.add(d);
    }
    return [...set].sort((a, b) => {
      if (a === 'Older') return 1;
      if (b === 'Older') return -1;
      return parseInt(b) - parseInt(a);
    });
  });

  const availableRatings = $derived.by<string[]>(() => {
    const set = new Set<string>();
    for (const item of items) {
      const r = item.contentRating?.trim().toUpperCase() || 'NR';
      set.add(r);
    }
    return [...set].sort((a, b) => {
      const ai = RATING_ORDER.indexOf(a);
      const bi = RATING_ORDER.indexOf(b);
      if (ai >= 0 && bi >= 0) return ai - bi;
      if (ai >= 0) return -1;
      if (bi >= 0) return 1;
      return a.localeCompare(b);
    });
  });

  const hasNeedsReview     = $derived(items.some((i: Media) => i.needsReview));
  const hasMultipleVersion = $derived(items.some((i: Media) => (i.versionCount ?? 1) > 1));
  const hasWatchData       = $derived(items.some((i: Media) => i.watched !== undefined));
  const hasMissingMeta     = $derived(items.some((i: Media) => !i.poster && !i.synopsis));
  const hasUnprobedItems   = $derived(items.some((i: Media) => i.probed === false));

  type StudioChip = { name: string; count: number };
  const studioChips = $derived.by<StudioChip[]>(() => {
    const counts = new Map<string, number>();
    for (const item of items) {
      for (const s of item.studio ?? []) {
        counts.set(s, (counts.get(s) ?? 0) + 1);
      }
    }
    const arr: StudioChip[] = [];
    for (const [name, count] of counts) arr.push({ name, count });
    arr.sort((a, b) => (b.count - a.count) || a.name.localeCompare(b.name));
    return arr.slice(0, 20); // top 20 per spec
  });

  // Snap invalid genre selections to empty if genres disappear (library reload).
  $effect(() => {
    if (selectedGenres.size === 0) return;
    const valid = new Set(genreChips.map(g => g.name));
    let changed = false;
    const next = new Set<string>();
    for (const g of selectedGenres) {
      if (valid.has(g)) next.add(g);
      else changed = true;
    }
    if (changed) selectedGenres = next;
  });

  // ── Active filter count (for badge) ───────────────────────────────────────
  const activeFilterCount = $derived(
    selectedGenres.size + selectedDecades.size + selectedRatings.size + selectedStudios.size +
    (watchFilter !== 'all' ? 1 : 0) +
    (onlyNeedsReview ? 1 : 0) + (onlyMultiVersion ? 1 : 0) + (onlyMissingMeta ? 1 : 0) +
    (onlyUnprobed ? 1 : 0)
  );

  // ── Sort label ─────────────────────────────────────────────────────────────
  const sortLabel = $derived(sortOptions.find(o => o.value === sort)?.label ?? "Sort");

  // ── Filtered + sorted list ─────────────────────────────────────────────────
  const filtered = $derived.by(() => {
    let list = items.filter((item: Media) => {
      if (q && !item.title.toLowerCase().includes(q.toLowerCase())) return false;
      if (selectedGenres.size > 0 && !(item.genres ?? []).some(g => selectedGenres.has(g))) return false;
      if (selectedDecades.size > 0 && !selectedDecades.has(getDecade(item.year))) return false;
      if (selectedRatings.size > 0) {
        const r = item.contentRating?.trim().toUpperCase() || 'NR';
        if (!selectedRatings.has(r)) return false;
      }
      if (selectedStudios.size > 0 && !(item.studio ?? []).some((s: string) => selectedStudios.has(s))) return false;
      if (watchFilter === 'watched' && !item.watched) return false;
      if (watchFilter === 'unwatched' && item.watched) return false;
      if (onlyNeedsReview && !item.needsReview) return false;
      if (onlyMultiVersion && (item.versionCount ?? 1) <= 1) return false;
      if (onlyMissingMeta && (item.poster || item.synopsis)) return false;
      if (onlyUnprobed && item.probed !== false) return false;
      return true;
    });

    switch (sort) {
      case 'za':           list = [...list].sort((a, b) => b.title.localeCompare(a.title)); break;
      case 'year-desc':    list = [...list].sort((a, b) => (b.year || 0) - (a.year || 0)); break;
      case 'year-asc':     list = [...list].sort((a, b) => (a.year || 0) - (b.year || 0)); break;
      case 'added-desc':   list = [...list].sort((a, b) => (b.addedAt || '').localeCompare(a.addedAt || '')); break;
      case 'added-asc':    list = [...list].sort((a, b) => (a.addedAt || '').localeCompare(b.addedAt || '')); break;
      case 'rating-desc':  list = [...list].sort((a, b) => (b.rating || 0) - (a.rating || 0)); break;
      case 'runtime-asc':  list = [...list].sort((a, b) => (a.runtimeMins || 9999) - (b.runtimeMins || 9999)); break;
      case 'runtime-desc': list = [...list].sort((a, b) => (b.runtimeMins || 0) - (a.runtimeMins || 0)); break;
      case 'rating-asc':   list = [...list].sort((a, b) => (a.rating || 99) - (b.rating || 99)); break;
      case 'parental-asc': list = [...list].sort((a, b) => {
        const ra = RATING_ORDER.indexOf(a.contentRating?.trim().toUpperCase() || 'NR');
        const rb = RATING_ORDER.indexOf(b.contentRating?.trim().toUpperCase() || 'NR');
        return (ra < 0 ? 999 : ra) - (rb < 0 ? 999 : rb);
      }); break;
      // Unwatched first: cards with no watched flag AND no/low progress come
      // before in-progress, then watched. Tie-break by title so the order is
      // stable across renders.
      case 'unwatched-first': list = [...list].sort((a, b) => {
        const score = (m: typeof a) => (m.watched ? 2 : (m.progress && m.progress >= 0.05) ? 1 : 0);
        const d = score(a) - score(b);
        return d !== 0 ? d : a.title.localeCompare(b.title);
      }); break;
      case 'watched-first': list = [...list].sort((a, b) => {
        const score = (m: typeof a) => (m.watched ? 2 : (m.progress && m.progress >= 0.05) ? 1 : 0);
        const d = score(b) - score(a);
        return d !== 0 ? d : a.title.localeCompare(b.title);
      }); break;
      // File-count sorts: useful for storage triage ("which titles have the
      // most duplicate versions?") and finding under-represented matches.
      case 'versions-desc': list = [...list].sort((a, b) => (b.versionCount || 0) - (a.versionCount || 0)); break;
      case 'versions-asc':  list = [...list].sort((a, b) => (a.versionCount || 99) - (b.versionCount || 99)); break;
      case 'random': {
        const seed = randomSeed;
        list = [...list].sort((a, b) => pseudoRandom(a.id, seed) - pseudoRandom(b.id, seed));
        break;
      }
      case 'az':
      default:
        list = [...list].sort((a, b) => a.title.localeCompare(b.title));
    }
    return list;
  });

  const featured = $derived(filtered[0] ?? items[0]);

  // ── Windowed slice of `filtered` actually rendered to the DOM ───────────
  // This is the entire payoff of the virtual-grid work: visibleSlice contains
  // ~50-100 items max (visible rows + OVERSCAN_ROWS above and below). Without
  // virtualization a 4000-item library shipped 4000 keyed Svelte components
  // and a 4.6 s teardown longtask on every route-out — confirmed against the
  // live media-server via PerformanceObserver.
  const totalRows       = $derived(Math.max(1, Math.ceil(filtered.length / itemsPerRow)));
  const firstVisibleRow = $derived(Math.max(0, Math.floor((scrollY - gridTop) / rowHeight) - OVERSCAN_ROWS));
  const lastVisibleRow  = $derived(Math.min(totalRows, Math.ceil((scrollY + viewportH - gridTop) / rowHeight) + OVERSCAN_ROWS));
  const visibleStart    = $derived(firstVisibleRow * itemsPerRow);
  const visibleEnd      = $derived(Math.min(filtered.length, lastVisibleRow * itemsPerRow));
  const visibleSlice    = $derived(filtered.slice(visibleStart, visibleEnd));
  // Spacer heights preserve scrollbar accuracy + native Find-in-Page anchor
  // jumps for not-yet-mounted items. Without these, scroll height would
  // collapse to whatever the rendered slice occupies and the scrollbar
  // would jump erratically.
  const topSpacer       = $derived(firstVisibleRow * rowHeight);
  const bottomSpacer    = $derived(Math.max(0, (totalRows - lastVisibleRow) * rowHeight));

  // ── Toggle helpers ─────────────────────────────────────────────────────────
  function toggleGenre(name: string) {
    const next = new Set(selectedGenres);
    if (next.has(name)) next.delete(name); else next.add(name);
    selectedGenres = next;
  }
  function toggleDecade(d: string) {
    const next = new Set(selectedDecades);
    if (next.has(d)) next.delete(d); else next.add(d);
    selectedDecades = next;
  }
  function toggleRating(r: string) {
    const next = new Set(selectedRatings);
    if (next.has(r)) next.delete(r); else next.add(r);
    selectedRatings = next;
  }
  function toggleStudio(s: string) {
    const next = new Set(selectedStudios);
    if (next.has(s)) next.delete(s); else next.add(s);
    selectedStudios = next;
  }
  function clearFilters() {
    selectedGenres   = new Set();
    selectedDecades  = new Set();
    selectedRatings  = new Set();
    selectedStudios  = new Set();
    watchFilter      = 'all';
    onlyNeedsReview  = false;
    onlyMultiVersion = false;
    onlyMissingMeta  = false;
    onlyUnprobed     = false;
  }

  // Re-measure the rendered grid: actual row height (poster + caption + gap)
  // and items-per-row (varies with viewport breakpoint + density). We read
  // both from the live DOM via getComputedStyle / offsetHeight so we don't
  // have to keep a fragile map of tailwind breakpoints in JS.
  function remeasureGrid() {
    if (!gridEl) return;
    const cs = window.getComputedStyle(gridEl);
    const cols = cs.gridTemplateColumns.split(/\s+/).filter(s => s && s !== 'none').length;
    if (cols > 0) itemsPerRow = cols;
    // First child is one rendered card. offsetHeight + row-gap = stride.
    const first = gridEl.querySelector(':scope > a') as HTMLElement | null;
    if (first) {
      const gap = parseFloat(cs.rowGap || cs.gap || '0') || 0;
      rowHeight = first.offsetHeight + gap;
    }
    const rect = gridEl.getBoundingClientRect();
    gridTop = rect.top + window.scrollY;
  }

  // ── URL state ──────────────────────────────────────────────────────────────
  // Read initial state from URL params on mount, then sync changes back.
  onMount(() => {
    const p = page.url.searchParams;
    const sortParam = p.get('sort') as Sort | null;
    if (sortParam && sortOptions.some(o => o.value === sortParam)) sort = sortParam;
    const qParam = p.get('q');
    if (qParam) q = qParam;
    const g = p.get('genres');
    if (g) selectedGenres = new Set(g.split(',').filter(Boolean));
    const dec = p.get('decades');
    if (dec) selectedDecades = new Set(dec.split(',').filter(Boolean));
    const rat = p.get('ratings');
    if (rat) selectedRatings = new Set(rat.split(',').filter(Boolean));
    const stu = p.get('studios');
    if (stu) selectedStudios = new Set(stu.split(',').filter(Boolean));
    const wf = p.get('watch') as "watched" | "unwatched" | null;
    if (wf === 'watched' || wf === 'unwatched') watchFilter = wf;
    if (p.get('review') === '1') onlyNeedsReview = true;
    if (p.get('multi') === '1') onlyMultiVersion = true;
    if (p.get('missing') === '1') onlyMissingMeta = true;
    if (p.get('unprobed') === '1') onlyUnprobed = true;
    const seed = parseInt(p.get('seed') ?? '', 10);
    if (!isNaN(seed) && seed > 0) randomSeed = seed;

    // Sync density with server preference (runs after first render; localStorage
    // already applied the cached value above so there is no visible flash).
    getAuthSession().then(s => {
      const size = s?.preferences?.posterSize;
      if (size === 'S' || size === 'M' || size === 'L') {
        density = size;
        try { localStorage.setItem(DENSITY_KEY, size); } catch { /* privacy mode */ }
      }
    }).catch(() => {});

    // Virtual-grid bookkeeping. Throttle scroll via rAF — we only need the
    // updated scrollY once per frame; firing on every wheel tick would burn
    // CPU and cause derived recomputes to thrash. Resize/orientationchange
    // re-measure both rowHeight and itemsPerRow since both depend on
    // viewport width via the responsive grid-cols-* breakpoints.
    let scrollTick = false;
    const onScroll = () => {
      if (scrollTick) return;
      scrollTick = true;
      requestAnimationFrame(() => {
        scrollY = window.scrollY;
        scrollTick = false;
      });
    };
    const onResize = () => {
      viewportH = window.innerHeight;
      remeasureGrid();
    };
    window.addEventListener('scroll', onScroll, { passive: true });
    window.addEventListener('resize', onResize);
    // Initial measurement happens after the first paint when at least one
    // card has been rendered. A microtask + rAF gets us there without
    // hardcoding a setTimeout.
    queueMicrotask(() => requestAnimationFrame(remeasureGrid));
    return () => {
      window.removeEventListener('scroll', onScroll);
      window.removeEventListener('resize', onResize);
    };
  });

  // Re-measure whenever density flips — that changes both items-per-row
  // (different grid-cols-*) and card height (no caption on density S).
  $effect(() => {
    // touch density so the effect re-runs
    void density;
    if (gridEl) requestAnimationFrame(remeasureGrid);
  });

  function setDensity(d: Density) {
    density = d;
    try { localStorage.setItem(DENSITY_KEY, d); } catch { /* privacy mode */ }
    updateUserPreferences({ posterSize: d }).catch(() => {});
  }

  // Sync filter state → URL (replaceState, no scroll, no focus loss)
  $effect(() => {
    const p = new URLSearchParams();
    if (q) p.set('q', q);
    if (sort !== 'az') p.set('sort', sort);
    if (sort === 'random') p.set('seed', String(randomSeed));
    if (selectedGenres.size)  p.set('genres',  [...selectedGenres].sort().join(','));
    if (selectedDecades.size) p.set('decades', [...selectedDecades].sort().join(','));
    if (selectedRatings.size) p.set('ratings', [...selectedRatings].sort().join(','));
    if (selectedStudios.size) p.set('studios', [...selectedStudios].sort().join(','));
    if (watchFilter !== 'all') p.set('watch', watchFilter);
    if (onlyNeedsReview)  p.set('review', '1');
    if (onlyMultiVersion) p.set('multi', '1');
    if (onlyMissingMeta)  p.set('missing', '1');
    if (onlyUnprobed)     p.set('unprobed', '1');
    const qs = p.toString();
    const cur = page.url.searchParams.toString();
    if (qs !== cur) {
      goto(`?${qs}`, { replaceState: true, noScroll: true, keepFocus: true });
    }
  });
</script>

<main class="pb-32">
  <!-- Alphabet jump rail (#403) ────────────────────────────────────────────
       Sticky vertical strip of A–Z (+ # for digits / non-Latin titles), shown
       only on alphabetical sort and only on lg+ where there's horizontal room.
       Letters that have no items in the current filter render dim and
       non-interactive. Click → smooth-scrolls the page to the row of the
       first matching item via the existing virtualization metrics. -->
  {#if isAlphabeticalSort}
    <nav
      class="pointer-events-none fixed right-2 top-1/2 z-30 hidden -translate-y-1/2 lg:block"
      aria-label="Jump to letter"
    >
      <ul class="pointer-events-auto flex flex-col items-center gap-px rounded-full bg-background/70 px-1.5 py-2 text-[10px] font-semibold tracking-widest backdrop-blur-sm">
        {#each ALPHABET as letter (letter)}
          {@const has = letterIndex.has(letter)}
          <li>
            <button
              type="button"
              disabled={!has}
              onclick={() => jumpToLetter(letter)}
              aria-label={`Jump to ${letter}`}
              class={`flex h-4 w-5 items-center justify-center rounded transition-colors ${
                has
                  ? 'text-muted-foreground hover:bg-foreground/[0.08] hover:text-foreground'
                  : 'text-muted-foreground/25 cursor-default'
              }`}
            >
              {letter}
            </button>
          </li>
        {/each}
      </ul>
    </nav>
  {/if}

  {#if !showHero}
    <!-- ── Compact header (matches /collections and /watchlist) ───────────────
         Sentence-style description in `text-sm muted-foreground` rather than
         the previous cramped uppercase stat strip, so Movies / TV /
         Collections / Watchlist all read the same. The library count is moved
         into the sticky toolbar below (as "{filtered} of {total}") where it
         lives alongside the active filter chips. -->
    <div class="px-6 pb-0 pt-28 md:px-12 lg:px-20">
      <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">
        {eyebrow}
      </div>
      <h1 class="font-serif-display text-[clamp(2rem,5vw,3.5rem)] leading-[0.95] tracking-tight">
        {title}
      </h1>
      <p class="mt-1.5 max-w-xl text-sm text-muted-foreground">{headerSentence}</p>
    </div>
  {:else}
  <!-- ── Hero header ──────────────────────────────────────────────────────── -->
  <section class="relative isolate overflow-hidden px-6 pb-10 pt-32 md:px-12 md:pb-14 md:pt-40 lg:px-20">
    {#if featured?.backdrop}
      {#key featured.id}
        <img
          src={artworkSrc(featured, 'backdrop', 720, featured.backdrop)}
          srcset={artworkSrcset(featured, 'backdrop', 720)}
          alt=""
          aria-hidden="true"
          decoding="async"
          class="pointer-events-none absolute inset-x-0 top-0 -z-20 h-full w-full object-cover object-top"
          style="opacity: 0.35;"
          onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
        />
      {/key}
    {/if}
    <div class="pointer-events-none absolute inset-x-0 bottom-0 -z-10 h-2/3 bg-gradient-to-t from-background to-transparent"></div>
    <div
      aria-hidden="true"
      class="pointer-events-none absolute inset-x-0 top-0 -z-10 h-[520px] opacity-70"
      style={`background: ${featured
        ? `radial-gradient(60% 70% at 20% 0%, ${featured.palette[0]}55, transparent 60%), radial-gradient(50% 60% at 90% 10%, ${featured.palette[1]}40, transparent 70%)`
        : ""}`}
    ></div>
    <div class="grain pointer-events-none absolute inset-0 -z-10"></div>

    <div class="grid items-end gap-10 md:grid-cols-[1.6fr_1fr] lg:grid-cols-[2fr_1fr]">
      <div>
        <div class="mb-3 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">{eyebrow}</div>
        <h1 class="font-serif-display text-[clamp(2.6rem,6vw,5rem)] leading-[0.95] tracking-tight">{title}</h1>
        <p class="mt-5 max-w-xl text-sm leading-relaxed text-muted-foreground md:text-base">{tagline}</p>
        <div class="mt-6 flex flex-wrap gap-x-6 gap-y-2 text-[11px] uppercase tracking-[0.22em] text-muted-foreground">
          <span><span class="text-foreground/90">{items.length}</span> in library</span>
          <span class="opacity-30">·</span>
          <span><span class="text-foreground/90">{genreChips.length}</span> genres</span>
        </div>
      </div>

      {#if featured}
        <article
          class="shadow-poster relative ml-auto aspect-[2/3] w-full max-w-[260px] overflow-hidden rounded-2xl md:max-w-[280px] lg:max-w-[320px]"
          style={`background: linear-gradient(135deg, ${featured.palette[0]}, ${featured.palette[1]})`}
        >
          {#if featured.poster}
            <img
              src={artworkSrc(featured, 'poster', 320, featured.poster)}
              srcset={artworkSrcset(featured, 'poster', 320)}
              alt={featured.title}
              decoding="async"
              class="absolute inset-0 h-full w-full object-cover"
              onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
            />
          {:else}
            <div class="absolute inset-0 opacity-40 mix-blend-overlay" style="background-image: radial-gradient(circle at 25% 15%, rgba(255,255,255,0.45), transparent 50%), radial-gradient(circle at 80% 85%, rgba(0,0,0,0.6), transparent 60%);"></div>
          {/if}
          <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/10 to-transparent"></div>
          <div class="absolute inset-0 flex flex-col justify-between p-5">
            <div class="flex items-center gap-2 text-[9px] font-semibold uppercase tracking-[0.3em] text-white/85">
              <span class="h-1.5 w-1.5 rounded-full" style={`background: ${featured.accent}; box-shadow: 0 0 12px ${featured.accent};`}></span>
              Editor's pick
            </div>
            <div>
              <div class="text-[10px] uppercase tracking-[0.28em] text-white/65">
                {featured.year || ""}
                {#if featured.type === "Series" && featured.seasons}
                  {featured.year ? " · " : ""}{featured.seasons} Season{featured.seasons !== 1 ? "s" : ""}
                {:else if featured.runtime}
                  {featured.year ? " · " : ""}{featured.runtime}
                {/if}
              </div>
              <h2 class="font-serif-display mt-1.5 text-2xl leading-[0.95] text-white md:text-3xl" style="text-shadow: 0 4px 24px rgba(0,0,0,0.5);">
                {featured.title}
              </h2>
              <div class="mt-4 flex gap-2">
                <a
                  href={baseHref ? `${baseHref}/${featured.id}` : (featured.type === 'Series' ? `/tv/${featured.id}` : `/movies/${featured.id}`)}
                  class="inline-flex items-center gap-1.5 rounded-full bg-white px-4 py-2 text-xs font-semibold text-black hover:bg-white/90"
                >
                  <Play class="h-3.5 w-3.5 fill-black" /> Play
                </a>
                <button
                  aria-label={isInWatchlist(featured.id, featured.type === 'Series' ? 'series' : 'movie') ? "Remove from watchlist" : "Add to watchlist"}
                  onclick={() => toggleWatchlist({
                    id: featured.id,
                    kind: featured.type === 'Series' ? 'series' : 'movie',
                    title: featured.title,
                    year: featured.year,
                    posterUrl: featured.poster,
                    backdropUrl: featured.backdrop,
                    genres: featured.genres,
                  })}
                  class={`hairline flex h-9 w-9 items-center justify-center rounded-full backdrop-blur-md transition-colors ${
                    isInWatchlist(featured.id, featured.type === 'Series' ? 'series' : 'movie')
                      ? 'bg-primary-glow/30 text-white ring-1 ring-primary-glow/60 hover:bg-primary-glow/40'
                      : 'bg-white/10 text-white hover:bg-white/20'
                  }`}
                >
                  {#if isInWatchlist(featured.id, featured.type === 'Series' ? 'series' : 'movie')}
                    <Check class="h-4 w-4" />
                  {:else}
                    <Plus class="h-4 w-4" />
                  {/if}
                </button>
              </div>
            </div>
          </div>
        </article>
      {/if}
    </div>
  </section>
  {/if}

  <!-- ── Toolbar (sticky) ───────────────────────────────────────────────────
       3-column flex layout matching /collections: a left flex-1 holds the
       Filters chip + (when filtering) "{filtered} of {total}" result count;
       the search input is centered (max-w-sm, w-full); the right flex-1
       holds sort + density. Previously this was "search hard-left, sort
       ml-auto" which left search visually anchored to the page gutter rather
       than centred on the viewport. -->
  <div class="sticky top-16 z-30 -mb-px border-y border-border bg-background/75 backdrop-blur-xl md:top-18">
    <div class="flex items-center gap-3 px-6 py-3 md:px-12 lg:px-20">

      <!-- Left: Filters chip + result count -->
      <div class="flex flex-1 items-center gap-3">
        <div class="relative">
          <button
            type="button"
            onclick={() => { filterOpen = !filterOpen; sortOpen = false; }}
            class={`hairline flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-medium transition-colors ${
              filterOpen || activeFilterCount > 0
                ? 'bg-primary-glow/15 text-foreground ring-1 ring-primary-glow/30'
                : 'bg-foreground/[0.04] text-muted-foreground hover:bg-foreground/[0.08] hover:text-foreground'
            }`}
            aria-expanded={filterOpen}
          >
            <SlidersHorizontal class="h-3.5 w-3.5" />
            Filters
            {#if activeFilterCount > 0}
              <span class="flex h-4 min-w-4 items-center justify-center rounded-full bg-primary-glow px-1 text-[10px] font-bold text-white">
                {activeFilterCount}
              </span>
            {/if}
            <ChevronDown class={`h-3.5 w-3.5 transition-transform ${filterOpen ? 'rotate-180' : ''}`} />
          </button>
        </div>
        {#if q || activeFilterCount > 0}
          <span class="hidden text-xs text-muted-foreground tabular-nums sm:inline">
            {filtered.length.toLocaleString()} of {items.length.toLocaleString()}
          </span>
        {/if}
      </div>

      <!-- Center: search -->
      <div class="relative w-full max-w-sm">
        <Search class="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          bind:value={q}
          placeholder={`Search ${kind === "TV" ? "TV shows" : "movies"}...`}
          class="h-9 w-full rounded-full border border-border bg-surface/60 pl-10 pr-4 text-sm outline-none placeholder:text-muted-foreground/70 focus:border-primary/60 focus:bg-surface"
        />
        {#if q}
          <button type="button" onclick={() => (q = '')} class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground" aria-label="Clear search">
            <X class="h-3.5 w-3.5" />
          </button>
        {/if}
      </div>

      <!-- Right: sort + density -->
      <div class="flex flex-1 items-center justify-end gap-2">
        <!-- Sort picker -->
        <div class="flex items-center gap-1">
          <!-- Shuffle button shown alongside when random is active -->
          {#if sort === 'random'}
            <button
              type="button"
              title="Shuffle"
              onclick={() => { randomSeed = Date.now(); }}
              class="flex h-7 w-7 items-center justify-center rounded-full bg-foreground/[0.04] text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground"
              aria-label="Shuffle"
            >
              <RotateCcw class="h-3 w-3" />
            </button>
          {/if}
          <div class="relative">
            <button
              type="button"
              onclick={() => { sortOpen = !sortOpen; filterOpen = false; }}
              class={`hairline flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs transition-colors hover:bg-foreground/[0.08] ${sortOpen ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
            >
              {sortLabel}
              <ChevronDown class={`h-3.5 w-3.5 transition-transform ${sortOpen ? 'rotate-180' : ''}`} />
            </button>
            {#if sortOpen}
              <button type="button" class="fixed inset-0 z-40 cursor-default" aria-hidden="true" tabindex="-1" onclick={() => (sortOpen = false)}></button>
              <div class="absolute right-0 top-full z-50 mt-1.5 min-w-[172px] overflow-hidden rounded-xl border border-border bg-surface/95 py-1 shadow-xl backdrop-blur-xl">
                {#each sortOptions as opt (opt.value)}
                  <button
                    type="button"
                    onclick={() => { sort = opt.value; sortOpen = false; if (opt.value === 'random') randomSeed = Date.now(); }}
                    class={`flex w-full items-center px-3.5 py-2 text-left text-xs transition-colors ${
                      sort === opt.value
                        ? 'bg-foreground/[0.07] font-medium text-foreground'
                        : 'text-muted-foreground hover:bg-foreground/[0.05] hover:text-foreground'
                    }`}
                  >
                    {opt.label}
                  </button>
                {/each}
              </div>
            {/if}
          </div>
        </div>

        <!-- Density switcher -->
        <div class="hairline hidden items-center rounded-full bg-foreground/[0.04] p-1 md:flex">
          {#each (["S", "M", "L"] as const) as option (option)}
            <button
              type="button"
              aria-label={option === "S" ? "Small cards" : option === "M" ? "Medium cards" : "Large cards"}
              onclick={() => setDensity(option)}
              class={`flex h-7 min-w-7 items-center justify-center rounded-full px-2.5 text-[11px] font-semibold tracking-wider transition-colors ${
                density === option ? 'bg-foreground text-background' : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              {option}
            </button>
          {/each}
        </div>
      </div>
    </div>

    <!-- ── Filter panel ──────────────────────────────────────────────────── -->
    {#if filterOpen}
      <div class="border-t border-border bg-surface/30 px-6 py-5 md:px-12 lg:px-20">
        <div class="space-y-5">

          <!-- Genre -->
          {#if genreChips.length > 0}
            <div>
              <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">Genre</div>
              <div class="flex flex-wrap gap-1.5">
                {#each genreChips as chip (chip.name)}
                  {@const active = selectedGenres.has(chip.name)}
                  <button
                    type="button"
                    onclick={() => toggleGenre(chip.name)}
                    class={`inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                      active
                        ? 'bg-foreground text-background'
                        : 'bg-foreground/[0.05] text-foreground/70 hover:bg-foreground/[0.10] hover:text-foreground'
                    }`}
                    aria-pressed={active}
                  >
                    {chip.name}
                    <span class={`tabular-nums text-[10px] ${active ? 'text-background/60' : 'text-foreground/40'}`}>{chip.count}</span>
                  </button>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Year / decade -->
          {#if availableDecades.length > 0}
            <div>
              <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">Year</div>
              <div class="flex flex-wrap gap-1.5">
                {#each availableDecades as decade (decade)}
                  {@const active = selectedDecades.has(decade)}
                  <button
                    type="button"
                    onclick={() => toggleDecade(decade)}
                    class={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                      active
                        ? 'bg-foreground text-background'
                        : 'bg-foreground/[0.05] text-foreground/70 hover:bg-foreground/[0.10] hover:text-foreground'
                    }`}
                    aria-pressed={active}
                  >
                    {decade}
                  </button>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Parental rating -->
          {#if availableRatings.length > 0}
            <div>
              <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">Rating</div>
              <div class="flex flex-wrap gap-1.5">
                {#each availableRatings as r (r)}
                  {@const active = selectedRatings.has(r)}
                  <button
                    type="button"
                    onclick={() => toggleRating(r)}
                    class={`rounded-full border px-3 py-1 text-xs font-medium transition-colors ${
                      active
                        ? 'border-foreground/60 bg-foreground text-background'
                        : 'border-border bg-transparent text-foreground/70 hover:border-foreground/30 hover:text-foreground'
                    }`}
                    aria-pressed={active}
                  >
                    {r}
                  </button>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Studio / Network -->
          {#if studioChips.length > 0}
            <div>
              <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">Studio</div>
              <div class="flex flex-wrap gap-1.5">
                {#each studioChips as chip (chip.name)}
                  {@const active = selectedStudios.has(chip.name)}
                  <button
                    type="button"
                    onclick={() => toggleStudio(chip.name)}
                    class={`inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                      active
                        ? 'bg-foreground text-background'
                        : 'bg-foreground/[0.05] text-foreground/70 hover:bg-foreground/[0.10] hover:text-foreground'
                    }`}
                    aria-pressed={active}
                  >
                    {chip.name}
                    <span class={`tabular-nums text-[10px] ${active ? 'text-background/60' : 'text-foreground/40'}`}>{chip.count}</span>
                  </button>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Watched / Unwatched -->
          {#if hasWatchData}
            <div>
              <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">Watch State</div>
              <div class="hairline inline-flex items-center overflow-hidden rounded-full bg-foreground/[0.04] p-0.5">
                {#each ([['all', 'All'], ['unwatched', 'Unwatched'], ['watched', 'Watched']] as const) as [val, label] (val)}
                  <button
                    type="button"
                    onclick={() => (watchFilter = val)}
                    class={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                      watchFilter === val
                        ? 'bg-foreground text-background'
                        : 'text-muted-foreground hover:text-foreground'
                    }`}
                    aria-pressed={watchFilter === val}
                  >
                    {label}
                  </button>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Flags -->
          {#if hasNeedsReview || hasMultipleVersion || hasMissingMeta}
            <div>
              <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">Show only</div>
              <div class="flex flex-wrap gap-1.5">
                {#if hasNeedsReview}
                  <button
                    type="button"
                    onclick={() => (onlyNeedsReview = !onlyNeedsReview)}
                    class={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                      onlyNeedsReview
                        ? 'bg-amber-500/20 text-amber-300 ring-1 ring-amber-500/30'
                        : 'bg-foreground/[0.05] text-foreground/70 hover:bg-foreground/[0.10] hover:text-foreground'
                    }`}
                    aria-pressed={onlyNeedsReview}
                  >
                    Needs Review
                  </button>
                {/if}
                {#if hasMultipleVersion}
                  <button
                    type="button"
                    onclick={() => (onlyMultiVersion = !onlyMultiVersion)}
                    class={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                      onlyMultiVersion
                        ? 'bg-primary-glow/15 text-foreground ring-1 ring-primary-glow/30'
                        : 'bg-foreground/[0.05] text-foreground/70 hover:bg-foreground/[0.10] hover:text-foreground'
                    }`}
                    aria-pressed={onlyMultiVersion}
                  >
                    Multiple Versions
                  </button>
                {/if}
                {#if hasMissingMeta}
                  <button
                    type="button"
                    onclick={() => (onlyMissingMeta = !onlyMissingMeta)}
                    class={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                      onlyMissingMeta
                        ? 'bg-orange-500/15 text-orange-300 ring-1 ring-orange-500/25'
                        : 'bg-foreground/[0.05] text-foreground/70 hover:bg-foreground/[0.10] hover:text-foreground'
                    }`}
                    aria-pressed={onlyMissingMeta}
                  >
                    Missing Metadata
                  </button>
                {/if}
                {#if hasUnprobedItems}
                  <button
                    type="button"
                    onclick={() => (onlyUnprobed = !onlyUnprobed)}
                    class={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                      onlyUnprobed
                        ? 'bg-amber-400/15 text-amber-300 ring-1 ring-amber-400/30'
                        : 'bg-foreground/[0.05] text-foreground/70 hover:bg-foreground/[0.10] hover:text-foreground'
                    }`}
                    aria-pressed={onlyUnprobed}
                  >
                    Not Analysed
                  </button>
                {/if}
              </div>
            </div>
          {/if}

          <!-- Clear / status row -->
          <div class="flex items-center justify-between pt-1">
            <p class="text-xs text-muted-foreground">
              {filtered.length === items.length
                ? `${items.length} ${kind === 'TV' ? 'shows' : 'movies'}`
                : `${filtered.length} of ${items.length} ${kind === 'TV' ? 'shows' : 'movies'}`}
            </p>
            {#if activeFilterCount > 0}
              <button
                type="button"
                onclick={clearFilters}
                class="flex items-center gap-1.5 rounded-full bg-foreground/[0.06] px-3 py-1 text-xs text-muted-foreground transition-colors hover:bg-foreground/[0.10] hover:text-foreground"
              >
                <X class="h-3 w-3" /> Clear {activeFilterCount} filter{activeFilterCount !== 1 ? 's' : ''}
              </button>
            {/if}
          </div>
        </div>
      </div>
    {/if}
  </div>

  <!-- ── Grid ─────────────────────────────────────────────────────────────── -->
  <section class="px-6 pt-10 md:px-12 lg:px-20">
    {#if loading}
      <div class={`grid gap-x-5 gap-y-10 ${densityGrid[density]}`}>
        {#each { length: 18 } as _, i (i)}
          <div class="animate-pulse">
            <div class="aspect-[2/3] rounded-xl bg-foreground/[0.07]"></div>
            {#if density !== "S"}
              <div class="mt-3 h-4 w-3/4 rounded bg-foreground/[0.07]"></div>
              <div class="mt-1.5 h-3 w-1/2 rounded bg-foreground/[0.05]"></div>
            {/if}
          </div>
        {/each}
      </div>
    {:else if filtered.length === 0}
      <div class="hairline flex flex-col items-center justify-center rounded-3xl bg-surface/30 py-24 text-center">
        <div class="font-serif-display text-3xl">Nothing matches that</div>
        <p class="mt-2 max-w-sm text-sm text-muted-foreground">
          Try different filters or clear them to see your full library.
        </p>
        {#if activeFilterCount > 0 || q}
          <button
            type="button"
            onclick={() => { clearFilters(); q = ''; }}
            class="mt-6 inline-flex items-center gap-2 rounded-full bg-foreground/[0.06] px-5 py-2.5 text-sm text-muted-foreground transition-colors hover:bg-foreground/[0.10] hover:text-foreground"
          >
            <X class="h-4 w-4" /> Clear all filters
          </button>
        {/if}
      </div>
    {:else}
      <!-- Top spacer reserves scroll height for rows above the rendered window. -->
      <div aria-hidden="true" style="height: {topSpacer}px"></div>
      <div bind:this={gridEl} class={`grid gap-x-5 gap-y-10 library-grid ${densityGrid[density]}`}>
        {#each visibleSlice as media (media.id)}
          <a
            href={baseHref ? `${baseHref}/${media.id}` : (media.type === "Series" ? `/tv/${media.id}` : `/movies/${media.id}`)}
            class="group block library-grid-item"
          >
            <div
              class="shadow-poster relative aspect-[2/3] overflow-hidden rounded-xl transition-all duration-300 group-hover:-translate-y-2 group-hover:scale-[1.04] group-hover:ring-[3px] group-hover:ring-white/85 group-hover:shadow-[0_28px_60px_-12px_oklch(0_0_0/0.85)]"
              style={`background: linear-gradient(135deg, ${media.palette[0]}, ${media.palette[1]})`}
            >
              {#if media.poster}
                {@const cardWidth = density === 'S' ? 160 : density === 'M' ? 220 : 300}
                <img
                  src={artworkSrc(media, 'poster', cardWidth, media.poster)}
                  srcset={artworkSrcset(media, 'poster', cardWidth)}
                  alt={media.title}
                  loading="lazy"
                  decoding="async"
                  class="absolute inset-0 h-full w-full object-cover"
                  onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                />
              {:else}
                <div class="absolute inset-0 opacity-40 mix-blend-overlay" style="background-image: radial-gradient(circle at 30% 20%, rgba(255,255,255,0.4), transparent 50%), radial-gradient(circle at 80% 80%, rgba(0,0,0,0.55), transparent 60%);"></div>
              {/if}
              <div class="absolute inset-x-0 bottom-0 h-1/2 bg-gradient-to-t from-black/80 to-transparent"></div>
              <div class="absolute inset-x-3 bottom-3">
                <h3 class="font-display text-[13px] font-semibold leading-tight text-white md:text-sm" style="text-shadow: 0 2px 12px rgba(0,0,0,0.7);">
                  {media.title}
                </h3>
              </div>
              <!-- Needs review badge -->
              {#if media.needsReview}
                <div class="absolute right-2 top-2 rounded-md bg-amber-500/80 px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wide text-black backdrop-blur-sm">
                  Review
                </div>
              {/if}
              <!-- Watched / unwatched badge -->
              {#if media.type === 'Series' && (media.unwatchedCount ?? 0) > 0}
                <div class="absolute left-2 top-2 min-w-[22px] rounded-full bg-primary/80 px-1.5 py-0.5 text-center text-[10px] font-bold leading-tight text-white backdrop-blur-sm">
                  {media.unwatchedCount}
                </div>
              {:else if media.type !== 'Series' && media.watched === true}
                <div class="absolute left-2 top-2 flex h-5 w-5 items-center justify-center rounded-full bg-green-500/80 backdrop-blur-sm">
                  <Check class="h-3 w-3 text-white" />
                </div>
              {/if}
              <div class="absolute inset-0 flex items-center justify-center bg-black/35 opacity-0 transition-opacity duration-300 group-hover:opacity-100">
                <button aria-label={`Play ${media.title}`} class="flex h-12 w-12 items-center justify-center rounded-full bg-white text-black ring-1 ring-white/40">
                  <Play class="h-5 w-5 translate-x-0.5 fill-black" />
                </button>
              </div>
            </div>
            {#if density !== "S"}
              <div class="mt-3 flex items-baseline justify-between gap-2 px-0.5">
                <h4 class="truncate text-sm font-medium text-foreground">{media.title}</h4>
                <span class="shrink-0 text-[11px] tabular-nums text-muted-foreground">{media.year || ''}</span>
              </div>
              <p class="mt-0.5 truncate text-[11px] uppercase tracking-[0.18em] text-muted-foreground">
                {media.genres.slice(0, 2).join(" · ")}
                {#if typeof media.rating === "number" && media.rating > 0}
                  <span class="ml-2 normal-case tracking-normal text-foreground/70">★ {media.rating.toFixed(1)}</span>
                {/if}
              </p>
            {/if}
          </a>
        {/each}
      </div>
      <!-- Bottom spacer reserves scroll height for rows below the window. -->
      <div aria-hidden="true" style="height: {bottomSpacer}px"></div>
    {/if}
  </section>
</main>

<style>
  /*
   * Off-screen grid items use content-visibility:auto so the browser skips
   * layout, paint, and style recalc until they scroll into view. For a
   * 5000-poster library this drops initial paint time from "the browser
   * laid out every card" to "the browser laid out the visible 12-60 cards".
   *
   * The contain-intrinsic-size hint tells the browser the size each card
   * would have if rendered — so the scrollbar is accurate, page height is
   * stable, and anchor links / Find In Page still work even on items the
   * browser hasn't laid out yet. The aspect ratio is 2:3 (poster) plus
   * ~64px for the title/year text below the poster at M/L densities.
   *
   * Supported in Chromium 85+, Edge 85+, Safari 18+, Firefox 125+. Older
   * browsers ignore the property and render normally — no regression.
   */
  .library-grid-item {
    content-visibility: auto;
    contain-intrinsic-size: 180px 334px;
  }
</style>
