<script lang="ts">
  import { page } from '$app/state';
  import { onMount } from 'svelte';
  import { ChevronLeft, Film, Tv, Star, User } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
  import ErrorState from '$lib/components/ErrorState.svelte';
  import { appState } from '$lib/stores/appState.svelte';
  import { getPersonDetail } from '$lib/api/home';
  import type { PersonDetailResponse, PersonCreditItem } from '$lib/api/home';

  const name = $derived(page.params.name ?? '');

  let data = $state<PersonDetailResponse | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let filterKind = $state<'all' | 'movie' | 'series'>('all');

  const person = $derived(data?.person);
  const allCredits = $derived(data?.credits ?? []);
  const credits = $derived(
    filterKind === 'all' ? allCredits : allCredits.filter((c) => c.kind === filterKind)
  );
  const movieCount = $derived(allCredits.filter((c) => c.kind === 'movie').length);
  const tvCount = $derived(allCredits.filter((c) => c.kind === 'series').length);

  // Deterministic palette fallback
  const PALETTES = [
    'linear-gradient(135deg,#0f172a,#1e3a8a)',
    'linear-gradient(135deg,#1c1917,#7c2d12)',
    'linear-gradient(135deg,#0f172a,#4c1d95)',
    'linear-gradient(135deg,#0a0a1a,#1e3a5f)',
    'linear-gradient(135deg,#1a0a2e,#6b21a8)',
  ];
  function fallbackGradient(id: string): string {
    let h = 0;
    for (let i = 0; i < id.length; i++) h = (h + id.charCodeAt(i)) % PALETTES.length;
    return PALETTES[h];
  }

  async function load() {
    try {
      loading = true;
      error = null;
      data = await getPersonDetail(name);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Person not found';
    } finally {
      loading = false;
    }
  }

  onMount(load);
</script>

<svelte:head>
  <title>{person?.name ?? name} — {appState.serverName}</title>
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  {#if loading}
    <div class="flex min-h-[60vh] items-center justify-center">
      <div class="h-10 w-10 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div>
    </div>

  {:else if error || !data}
    <ErrorState
      title="No credits found"
      message="{name} doesn't appear to have any titles in your library."
      actions={[{ label: 'Browse movies', href: '/movies' }]}
      diagnosticInfo={error ?? undefined}
    />

  {:else}
    <div class="px-6 pb-32 pt-28 md:px-12 lg:px-20">
      <!-- Person header -->
      <button
        type="button"
        onclick={() => history.back()}
        class="mb-8 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
      >
        <ChevronLeft class="h-4 w-4" /> Back
      </button>

      <div class="flex items-center gap-6">
        <!-- Profile photo -->
        <div class="h-24 w-24 shrink-0 overflow-hidden rounded-full bg-surface-elevated ring-2 ring-border md:h-32 md:w-32">
          {#if person?.profileUrl}
            <img
              src={person.profileUrl}
              alt={person.name}
              class="h-full w-full object-cover"
              onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
            />
          {:else}
            <div class="flex h-full w-full items-center justify-center text-muted-foreground">
              <User class="h-10 w-10" />
            </div>
          {/if}
        </div>

        <div class="min-w-0">
          <h1 class="font-serif-display text-[clamp(1.8rem,4vw,3.5rem)] leading-tight tracking-tight">
            {person?.name ?? name}
          </h1>
          {#if person?.department}
            <p class="mt-1 text-sm uppercase tracking-[0.2em] text-muted-foreground">{person.department}</p>
          {/if}
          <p class="mt-3 text-sm text-muted-foreground">
            {allCredits.length} title{allCredits.length !== 1 ? 's' : ''} in your library
          </p>
        </div>
      </div>

      <!-- Filter tabs -->
      {#if movieCount > 0 && tvCount > 0}
        <div class="mt-8 flex items-center gap-2">
          {#each ([['all', `All (${allCredits.length})`], ['movie', `Movies (${movieCount})`], ['series', `TV (${tvCount})`]] as const) as [kind, label] (kind)}
            <button
              type="button"
              onclick={() => (filterKind = kind)}
              class={`rounded-full px-4 py-1.5 text-sm font-medium transition-colors ${
                filterKind === kind
                  ? 'bg-foreground text-background'
                  : 'hairline bg-foreground/[0.04] text-muted-foreground hover:bg-foreground/[0.08] hover:text-foreground'
              }`}
            >
              {label}
            </button>
          {/each}
        </div>
      {/if}

      <!-- Credits grid -->
      {#if credits.length === 0}
        <div class="mt-16 flex flex-col items-center gap-3 text-center">
          <p class="text-muted-foreground">No {filterKind === 'movie' ? 'movies' : 'TV shows'} found.</p>
        </div>
      {:else}
        <div class="mt-8 grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
          {#each credits as credit (credit.id)}
            {@const href = credit.id ? (credit.kind === 'series' ? `/tv/${credit.id}` : `/movies/${credit.id}`) : '#'}
            <a {href} class="group block">
              <!-- Poster card -->
              <div
                class="shadow-poster relative aspect-[2/3] w-full overflow-hidden rounded-xl transition-all duration-300 group-hover:-translate-y-1 group-hover:shadow-glow"
                style={credit.posterUrl ? '' : `background: ${fallbackGradient(credit.id ?? '')}`}
              >
                {#if credit.posterUrl}
                  <img
                    src={credit.posterUrl}
                    alt={credit.title}
                    loading="lazy"
                    class="absolute inset-0 h-full w-full object-cover"
                    onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                  />
                {/if}
                <div class="absolute inset-x-0 bottom-0 h-1/2 bg-gradient-to-t from-black/80 to-transparent"></div>

                <!-- Kind badge -->
                <div class="absolute left-2 top-2 rounded-md bg-black/55 px-1.5 py-0.5 backdrop-blur-sm ring-1 ring-white/15">
                  {#if credit.kind === 'series'}
                    <Tv class="h-3 w-3 text-white/80" />
                  {:else}
                    <Film class="h-3 w-3 text-white/80" />
                  {/if}
                </div>

                <!-- Play overlay -->
                <div class="absolute inset-0 flex items-center justify-center bg-black/30 opacity-0 backdrop-blur-[1px] transition-opacity duration-200 group-hover:opacity-100">
                  <span class="flex h-10 w-10 items-center justify-center rounded-full bg-white/90">
                    <Film class="h-4 w-4 fill-black text-black" />
                  </span>
                </div>
              </div>

              <!-- Text below poster -->
              <div class="mt-2.5 px-0.5">
                <h3 class="truncate text-sm font-medium text-foreground">{credit.title ?? 'Unknown'}</h3>
                <div class="mt-0.5 flex items-center gap-1.5 text-[11px] text-muted-foreground">
                  {#if credit.year && credit.year > 0}<span>{credit.year}</span>{/if}
                  {#if credit.voteAverage && credit.voteAverage > 0}
                    <span class="flex items-center gap-0.5 text-amber-300/80">
                      <Star class="h-2.5 w-2.5 fill-current" />{credit.voteAverage.toFixed(1)}
                    </span>
                  {/if}
                </div>
                {#if credit.character}
                  <p class="mt-0.5 truncate text-[10px] uppercase tracking-[0.15em] text-muted-foreground/70">
                    {credit.character}
                  </p>
                {:else if credit.role}
                  <p class="mt-0.5 truncate text-[10px] uppercase tracking-[0.15em] text-muted-foreground/70">
                    {credit.role}
                  </p>
                {/if}
              </div>
            </a>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>
