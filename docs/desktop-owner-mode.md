# Desktop Owner Mode

Desktop Owner Mode is the product operating mode for Xuva Phase 1.

## Product boundary

Media Mode:
- Home
- Movies
- TV
- media details and playback

Settings Mode:
- server-owner control room
- server identity, libraries, scanning, metadata, playback, transcoding, storage
- network discovery, pairing, approved devices, owner access, about

## Development boundary

For UI and settings work:

- run Xuva in local loopback with development owner bypass
- do not use mobile/tablet polish as acceptance criteria
- keep mobile/tablet checks to smoke verification only:
  - app loads
  - drawer opens
  - no catastrophic overflow
  - no stuck loading

## Auth boundary

Phase 1 does not remove auth from production paths.

- development owner bypass is local-only and temporary for desktop owner iteration
- production access flow remains sign-in/bootstrap based
- broader auth/users work is tracked in Phase 2
