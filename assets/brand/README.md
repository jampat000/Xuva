# Xuva Brand Pack

Complete icon & branding asset bundle for web, iOS, Android, Windows, macOS, and social.

## Brand

- **Primary purple**: `#7C3AED`
- **Primary glow**: `#A78BFA`
- **Accent blue**: `#60A5FA`
- **Background (cinema black)**: `#0A0A1A`
- **Foreground (white)**: `#FAFAFA`
- **Typography**: Geist (display + body), weight 500–700, tight tracking (`-0.035em`)
- **Tagline**: *Your cinema, everywhere*

## Folder map

```
svg/                   Vector masters (edit these to regenerate everything)
  mark.svg                 Mark only, transparent, on-brand gradient
  mark-padded-dark.svg     Mark + safe-area padding on cinema-black background
  mark-padded-light.svg    Same on white (for light-mode contexts)
  wordmark.svg             Horizontal lockup: mark + "Xuva"
  safari-pinned.svg        Monochrome path for Safari pinned tabs

web/                   Browser favicons & PWA
  favicon.ico              Multi-res 16/32/48
  favicon.svg              Modern SVG favicon (preferred)
  favicon-{16,32,48,64,96,128,192,256,512}.png
  apple-touch-icon.png     180×180, no alpha (iOS Home Screen)
  android-chrome-192.png   PWA manifest icon
  android-chrome-512.png   PWA manifest icon
  maskable-512.png         Android adaptive icon (safe-area baked in)
  safari-pinned-tab.svg    Monochrome SVG for Safari

ios/                   Apple AppIcon.appiconset (drag the folder into Xcode)
  icon-20 / icon-20@2x / icon-20@3x
  icon-29 / icon-29@2x / icon-29@3x
  icon-40 / icon-40@2x / icon-40@3x
  icon-60@2x / icon-60@3x          iPhone home screen
  icon-76 / icon-76@2x             iPad
  icon-83.5@2x                     iPad Pro
  icon-1024                        App Store

android/               Android Studio mipmap-* densities
  mipmap-mdpi/     ic_launcher.png · ic_launcher_round.png · ic_launcher_foreground.png
  mipmap-hdpi/     same set
  mipmap-xhdpi/    same set
  mipmap-xxhdpi/   same set
  mipmap-xxxhdpi/  same set
  play-store-512.png            Play Console listing icon
  feature-graphic-1024x500.png  Play Store feature graphic

windows/               Windows app + installer
  xuva.ico                       Multi-res 16/32/48/64/128/256 (use for .exe / installer)
  installer-256.png              Installer hero
  Square44x44Logo.png            Taskbar / start tile
  Square71x71Logo.png            Small tile
  Square150x150Logo.png          Medium tile
  Square310x310Logo.png          Large tile
  Wide310x150Logo.png            Wide tile

macos/                 macOS .iconset (build .icns with iconutil)
  icon_16x16.png · icon_16x16@2x.png
  icon_32x32.png · icon_32x32@2x.png
  icon_128x128.png · icon_128x128@2x.png
  icon_256x256.png · icon_256x256@2x.png
  icon_512x512.png · icon_512x512@2x.png
  icon_1024x1024.png

social/                Open Graph & social cards
  og-1200x630.png                Facebook / LinkedIn / generic OG
  twitter-card-1200x600.png      Twitter / X large card
  square-1080.png                Instagram / general square

raw/                   High-res masters (for marketing, decks, print)
  icon-master-dark-1024.png      Cinematic dark app icon (AI render)
  icon-master-mark-1024.png      Mark only
  wordmark-dark.png              Lockup on dark background
  og-image-cinematic.png         Premium social hero
```

## How to use

### Web (drop into your app)

Copy `web/` contents into the public folder of your site, then in your HTML head:

```html
<link rel="icon" type="image/svg+xml" href="/favicon.svg" />
<link rel="icon" type="image/png" sizes="32x32" href="/favicon-32.png" />
<link rel="icon" type="image/png" sizes="16x16" href="/favicon-16.png" />
<link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png" />
<link rel="mask-icon" href="/safari-pinned-tab.svg" color="#7C3AED" />
<link rel="manifest" href="/site.webmanifest" />
<meta name="theme-color" content="#0A0A1A" />
<meta name="msapplication-TileColor" content="#0A0A1A" />

<!-- Social -->
<meta property="og:image" content="/og-1200x630.png" />
<meta name="twitter:card" content="summary_large_image" />
<meta name="twitter:image" content="/twitter-card-1200x600.png" />
```

`site.webmanifest` example:

```json
{
  "name": "Xuva",
  "short_name": "Xuva",
  "icons": [
    { "src": "/android-chrome-192.png", "sizes": "192x192", "type": "image/png" },
    { "src": "/android-chrome-512.png", "sizes": "512x512", "type": "image/png" },
    { "src": "/maskable-512.png", "sizes": "512x512", "type": "image/png", "purpose": "maskable" }
  ],
  "theme_color": "#0A0A1A",
  "background_color": "#0A0A1A",
  "display": "standalone"
}
```

### iOS (Xcode)

Open your Xcode project → `Assets.xcassets` → `AppIcon`. Drag each PNG from `ios/` into the matching slot. Or build an `.appiconset` Contents.json if you prefer.

### Android (Android Studio)

Copy each `mipmap-*` folder into `app/src/main/res/`. For an adaptive icon, create `mipmap-anydpi-v26/ic_launcher.xml`:

```xml
<adaptive-icon xmlns:android="http://schemas.android.com/apk/res/android">
  <background android:drawable="@color/ic_launcher_background" />
  <foreground android:drawable="@mipmap/ic_launcher_foreground" />
</adaptive-icon>
```

…with `ic_launcher_background = #0A0A1A`.

### Windows (MSIX / installer)

- Use `xuva.ico` as the `.exe` / Inno Setup / NSIS installer icon.
- Drop `Square*` and `Wide310x150Logo.png` into your MSIX `Assets/` folder and reference them in `Package.appxmanifest`.

### macOS (.icns)

On a Mac:

```bash
mv macos Xuva.iconset
iconutil -c icns Xuva.iconset
# produces Xuva.icns
```

### Regenerating

Edit `svg/mark.svg` (or any sibling) and re-rasterize at the sizes you need. The vector masters are the source of truth — everything else is derived.

---

Built for Xuva. Every screen, one library.
