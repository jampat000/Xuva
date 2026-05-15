# Xuva Icon Pack

Generated from the live `XuvaAppIcon` / `XuvaLogo` components. Drop these into
your project (or hand the whole pack to Codex) and wire up the use cases below
as needed — now or later.

## File map → use case

### `favicon/` — Browser tab
| File | Where it goes | HTML |
|---|---|---|
| `favicon.ico` | `/public/favicon.ico` | served automatically at `/favicon.ico` |
| `favicon.svg` | `/public/favicon.svg` | `<link rel="icon" type="image/svg+xml" href="/favicon.svg">` |
| `favicon-16.png` `-32` `-48` `-64` | `/public/` | `<link rel="icon" type="image/png" sizes="32x32" href="/favicon-32.png">` |

### `apple/` — iOS home-screen bookmark
| File | HTML |
|---|---|
| `apple-touch-icon.png` (180×180) | `<link rel="apple-touch-icon" href="/apple-touch-icon.png">` |

### `pwa/` — Progressive Web App / Add to Home Screen
| File | Notes |
|---|---|
| `icon-192.png`, `icon-512.png` | standard PWA icons |
| `icon-192-maskable.png`, `icon-512-maskable.png` | Android adaptive (safe-zone padded) |
| `manifest.json` | reference manifest, edit paths to match your `/public` layout |

Wire-up: `<link rel="manifest" href="/manifest.json">`

### `ios-app-store/` — Apple App Store / TestFlight
| File | Notes |
|---|---|
| `AppIcon-1024.png` | 1024×1024, opaque (no alpha), no rounded corners — Apple rounds it. Drop into `Assets.xcassets/AppIcon.appiconset/`. |

### `macos/` — macOS app bundle / dock icon
| File | Notes |
|---|---|
| `Xuva.icns` | multi-resolution icns ready to use as `CFBundleIconFile` |
| `Xuva.iconset/` | source PNGs (16→1024 + @2x) if you want to rebuild via `iconutil` |

### `windows/` — Windows desktop / taskbar / Store
| File | Notes |
|---|---|
| `Xuva.ico` | multi-resolution `.ico` (16/24/32/48/64/128/256). Use as Electron `icon:` and Windows shortcut icon. |
| `icon-{16..256}.png` | individual PNGs if needed |
| `StoreLogo-44.png` `-150` `-310` | Microsoft Store / MSIX tiles |

### `tray/` — System tray / menu bar
Monochrome white silhouette (no tile, no gradient — those collapse to mush at 16–22px).

| File | Notes |
|---|---|
| `tray-16.png` `-22` `-24` `-32` `-44` (+ `@2x`) | platform-agnostic tray PNGs |
| `trayTemplate.png` + `trayTemplate@2x.png` | **macOS template image** — filename suffix `Template` is required so macOS auto-inverts for dark/light menu bar |

For Electron tray: `new Tray(path.join(__dirname, 'trayTemplate.png'))`.

### `_source-svg/` — Editable masters
| File | Notes |
|---|---|
| `mark.svg` | bare mark, gradient, no tile |
| `tile.svg` | full app-icon tile, rounded corners |
| `tile-square.svg` | tile without rounded corners (for OS that rounds itself: iOS, favicon) |
| `tray-mono.svg` | monochrome silhouette |

Re-run `build.py` (in this pack's parent dir) to regenerate everything.

---

## Codex prompt

> The Xuva project has a live React component `XuvaAppIcon` (in
> `src/components/brand/XuvaLogo.tsx`) that renders the brand tile.
> I have an attached icon pack (`xuva-icons.zip`) with the same artwork
> exported as real image files (PNG/ICO/ICNS/SVG) plus a monochrome tray
> variant.
>
> Do the following:
>
> 1. Unzip the pack into `public/icons/` (preserve the subfolder layout).
> 2. Move `favicon/favicon.ico`, `favicon/favicon.svg`, `favicon/favicon-*.png`,
>    and `apple/apple-touch-icon.png` to `public/` (root) for default
>    browser/iOS lookup.
> 3. Update `src/routes/__root.tsx` `head()` to include:
>    - `<link rel="icon" type="image/svg+xml" href="/favicon.svg">`
>    - `<link rel="icon" type="image/png" sizes="32x32" href="/favicon-32.png">`
>    - `<link rel="icon" type="image/png" sizes="16x16" href="/favicon-16.png">`
>    - `<link rel="apple-touch-icon" href="/apple-touch-icon.png">`
>    - `<meta name="theme-color" content="#7C3AED">`
> 4. Leave `pwa/`, `ios-app-store/`, `macos/`, `windows/`, `tray/`, and
>    `_source-svg/` in `public/icons/` untouched — they are not wired up
>    yet. They're for future use (PWA install, Electron desktop builds,
>    App Store submission, system tray). Do NOT add a service worker or
>    `vite-plugin-pwa` unless I explicitly ask.
> 5. Do NOT replace or modify `XuvaLogo.tsx` or `XuvaWordmark.tsx` — the
>    React components stay as the in-app brand surface; these files are
>    the rasterized counterparts for OS-level surfaces only.
