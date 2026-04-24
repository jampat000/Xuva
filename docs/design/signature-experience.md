# Vyrden Signature Experience

## Problem

Most movie and TV apps converge on the same visual pattern:

- Large hero.
- Horizontal poster rows.
- Detail page.
- Hidden settings.
- Playback diagnostics only after something goes wrong.

That pattern is familiar, but it is also overused. Vyrden should not look like a slightly darker Plex, a cleaner Emby, or a paid Jellyfin skin.

## Design Goal

Create a browsing system that feels native to Vyrden:

- Movie and TV first.
- Remote-friendly.
- Rich enough for a premium TV product.
- Practical enough for daily use.
- Playback intelligence visible by default.
- Distinct from standard streaming-service layouts.

## Signature Pattern: Signal Deck

The Signal Deck treats each movie or episode as both a piece of entertainment and a playback route.

Instead of separating browsing from diagnostics, Vyrden shows:

- The focused title.
- The selected version.
- The selected audio track.
- The selected subtitle track.
- The route to the client.
- The expected server cost.

This makes Vyrden's technical advantage feel like part of the core product, not an admin-only feature.

## Interaction Model

The user navigates through a deck of media slates rather than generic poster rows.

Focus changes:

- Backdrop color field.
- Title and synopsis.
- Playback route.
- Codec path.
- Server cost.
- Download suggestion.
- Primary actions for the focused item.

The design should feel like tuning into a private media signal, not browsing a commercial streaming storefront.

## Action Rules

Signal Deck must never hide basic watch controls.

Every focused movie or episode needs visible controls for:

- Resume or play.
- Play from start.
- Mark watched or mark episode played.
- Versions.
- Audio.
- Subtitles.
- Download.

These controls should live near the focused title and also remain reachable from the side control stack for TV remote navigation.

## Differentiation Rules

- Avoid endless horizontal rows as the only browsing model.
- Avoid oversized marketing hero treatment in the app itself.
- Avoid hiding audio, subtitles, and versions behind secondary modals.
- Avoid making diagnostics look like developer tools.
- Avoid generic purple/blue SaaS gradients.
- Avoid decorative backgrounds that do not explain the media or playback state.

## What Must Stay Familiar

Fresh does not mean confusing.

Keep:

- Clear focus states.
- Predictable arrow-key movement.
- Posters and titles.
- Play/resume as the primary action.
- Back behavior.
- Search.
- Movie, TV, collection, and download entry points.

## What Should Feel New

- Playback route as a first-class visual object.
- Version/audio/subtitle choices visible before play.
- A browsing layout with depth and hierarchy, not only rows.
- Cinematic visual language that belongs to personal libraries rather than subscription storefronts.
- Admin-grade clarity without admin-grade ugliness.
