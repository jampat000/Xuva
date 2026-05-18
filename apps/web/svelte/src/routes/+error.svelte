<script lang="ts">
  import { page } from '$app/state';
  import { ChevronLeft, AlertTriangle, SearchX, ShieldOff } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';

  const status = $derived(page.status);

  const title = $derived(
    status === 404 ? 'Page not found'
    : status === 403 ? 'Access denied'
    : status === 401 ? 'Sign in required'
    : 'Something went wrong'
  );

  const subtitle = $derived(
    status === 404
      ? "We couldn't find that page. It may have moved or never existed."
      : status === 403
      ? "You don't have permission to view this page."
      : status === 401
      ? "Please sign in to access this content."
      : "An unexpected error occurred. Please try again or head back home."
  );

  const icon = $derived(
    status === 404 ? SearchX
    : status === 403 || status === 401 ? ShieldOff
    : AlertTriangle
  );
</script>

<svelte:head>
  <title>{status} — Xuva</title>
  <meta name="robots" content="noindex" />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  <div class="relative flex min-h-screen flex-col items-center justify-center px-6 pb-24 pt-16 text-center">
    <!-- Ambient glow -->
    <div
      aria-hidden="true"
      class="pointer-events-none absolute inset-0 -z-10"
      style="background: radial-gradient(ellipse at 50% 35%, oklch(0.62 0.22 285 / 0.10), transparent 55%), radial-gradient(ellipse at 70% 70%, oklch(0.72 0.16 255 / 0.07), transparent 50%);"
    ></div>

    <!-- Giant status code watermark -->
    <div
      aria-hidden="true"
      class="font-serif-display pointer-events-none select-none text-[clamp(8rem,25vw,14rem)] font-bold leading-none tracking-tighter text-foreground/[0.04]"
    >
      {status}
    </div>

    <!-- Icon -->
    <div class="-mt-8 hairline flex h-14 w-14 items-center justify-center rounded-2xl bg-surface-elevated/80 text-muted-foreground shadow-elev">
      <svelte:component this={icon} class="h-6 w-6" />
    </div>

    <!-- Copy -->
    <h1 class="font-serif-display mt-5 text-[clamp(1.5rem,4vw,2.5rem)] tracking-tight">
      {title}
    </h1>
    <p class="mt-3 max-w-sm text-sm leading-relaxed text-muted-foreground">
      {subtitle}
    </p>

    <!-- Actions -->
    <div class="mt-8 flex flex-wrap items-center justify-center gap-3">
      <a
        href="/"
        class="inline-flex items-center gap-2 rounded-full bg-foreground px-6 py-3 text-sm font-semibold text-background transition-all hover:bg-foreground/90"
      >
        <ChevronLeft class="h-4 w-4" /> Back to home
      </a>
      {#if status === 401}
        <a
          href="/signin"
          class="hairline inline-flex items-center gap-2 rounded-full bg-foreground/[0.06] px-6 py-3 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
        >
          Sign in
        </a>
      {/if}
    </div>
  </div>
</div>
