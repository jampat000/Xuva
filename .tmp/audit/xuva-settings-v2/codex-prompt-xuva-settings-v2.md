# Codex Prompt — Xuva Settings Page (v2, Vercel/Linear style)

## Goal
Port the React/TanStack Start `/settings` route in this bundle to **SvelteKit + Svelte 5 runes** and integrate it into the existing Xuva plan. This replaces any previous settings draft. Same Header, same tokens, same Tailwind v4 setup as `/movies` and `/home`. **Do not invent new design tokens.**

## Files in this bundle
- `settings.tsx` — React source of truth for layout, state, copy, classes
- `Header.tsx` — shared header (already ported in the plan; use the existing Svelte version)
- `styles.css` — Tailwind v4 + token reference (already in the plan; do not duplicate)

## Layout contract (do NOT deviate)
1. Page shell:
   - `min-h-screen bg-background`
   - `<Header />` then `<main class="px-6 pb-32 pt-24 md:px-12 md:pt-28 lg:px-20">`
2. **Compact editorial header** (single line, NO giant radial hero):
   - Eyebrow `Settings` in `text-primary-glow`, uppercase, tracking `0.35em`
   - `font-serif-display` H1: `Media-<em>Server</em>`, clamp(2rem,4vw,3.25rem)
   - Sub copy, then right-side: emerald "Healthy" pill + "Updated 10:39:30 am"
   - Subtle radial gradient as `aria-hidden` absolute, height ~220px, opacity 60
3. **2-column grid**: `grid gap-10 lg:grid-cols-[260px_minmax(0,1fr)] lg:gap-14`
   - **Left = in-content sticky TOC** (`lg:sticky lg:top-24 lg:self-start`). This is NOT a global sidebar. It lives inside `main`'s horizontal padding and scrolls with the page on mobile.
     - Search input (rounded-full, `bg-surface/60`, focus `border-primary/60`)
     - Three groups, in order: **Media-Server**, **Devices**, **Advanced**
     - Group label: `text-[10px] uppercase tracking-[0.28em] text-muted-foreground/70`
     - Item button: icon + label, active state = `bg-foreground/[0.06]` + 2px `primary-glow` left rail with `shadow-glow`, icon recolors to `text-primary-glow`
   - **Right = content panel**:
     - Sticky sub-header bar: group eyebrow + section title (`font-serif-display text-2xl`) + hint + `Discard` / `Save changes` buttons. When `window.scrollY > 80`, add `border-border bg-background/80 backdrop-blur-xl`.
     - `Save changes` = `bg-gradient-primary text-white shadow-glow ring-1 ring-white/20`
     - Body uses `SettingBlock` (`md:grid-cols-[280px_minmax(0,1fr)]`): description column left, controls right

## Worked example: Metadata section
Render fully (other 15 sections are scaffold placeholders):
- **Mode selector** — 3 cards (`Balanced`, `Prefer local artwork`, `Advanced provider settings`). Active card = `bg-surface-elevated/80 shadow-elev`, radio dot fills with `primary-glow` + `shadow-glow`, "Selected" eyebrow appears. Default active: `advanced`.
- **Provider keys** — cards for `TMDB`, `TheTVDB`, `Fanart.tv`, `MusicBrainz`. Each:
  - `font-serif-display` name + status pill
  - Empty key → amber `Key required` pill (KeyRound icon)
  - Non-empty key → emerald `Active` pill (Check icon)
  - `type="password"` input, `Clear key` button disabled when empty

## Other sections
Render a placeholder card with the section icon, label, hint, and "Coming soon" note. Keep the same `SettingBlock` spacing so swapping in real content later is trivial.

## React → Svelte 5 translation rules
- `useState` → `let x = $state(...)`
- `useRef` → `let mainRef: HTMLDivElement | null = $state(null)` + `bind:this`
- `useEffect` for scroll listener → `$effect(() => { ... return () => ... })`
- `onClick` → `onclick`, `onChange` → `oninput` for inputs, value bound via `bind:value`
- `className` → `class`; conditional classes via template strings or `clsx`-equivalent
- `lucide-react` → `lucide-svelte` (same icon names)
- `@tanstack/react-router` `createFileRoute` → SvelteKit route file `+page.svelte` at `src/routes/settings/+page.svelte` and `+page.ts` for `<title>`/meta via `<svelte:head>`
- React `<em>` inside H1 → keep `<em>` in Svelte
- `React.ReactNode` children → `{@render children()}` with `let { children }: { children: Snippet } = $props()`

## Forbidden
- No left **app-shell** sidebar (the TOC lives inside `main`'s padding — that's the whole point of this revision)
- No giant editorial hero on `/settings` (movies/home keep theirs; settings is intentionally compact)
- No new color tokens, no new fonts, no new radii — reuse `bg-surface`, `bg-surface-elevated`, `hairline`, `shadow-glow`, `shadow-elev`, `font-serif-display`, `bg-gradient-primary`, `text-primary-glow`
- No backend, no API calls, no persistence — pure client state
- No Svelte stores; use runes only
- Do not edit `Header.tsx` semantics — Svelte Header already in plan stays as-is

## Acceptance checklist (verify in browser before reporting done)
1. `/settings` renders with the shared Header; nav highlights Settings.
2. Compact header shows `Settings` eyebrow, `Media-Server` title with italic `Server`, Healthy pill, Updated timestamp.
3. Left TOC is sticky on `lg+`, scrolls inline on mobile. Search filters items live across all three groups.
4. Clicking a TOC item updates the active state, the section header strip, and scrolls the content panel to top. Active item shows the `primary-glow` left rail.
5. Sticky sub-header gains `border-border bg-background/80 backdrop-blur-xl` after scrolling 80px.
6. Metadata → Mode selector: exactly one selected at a time, defaults to `Advanced provider settings`.
7. Metadata → Provider cards: pasting a key flips amber `Key required` → emerald `Active`; `Clear key` resets it and disables when empty.
8. All other 15 sections render the placeholder without console errors.
9. No layout shift between sections (sticky sub-header stays put).
10. Lighthouse: no contrast regressions vs `/movies`.
