# Xuva Icon Pack Surfaces

Source pack location used for this import:

- `C:\Users\User\Downloads\xuva-icons.zip`

Imported asset root:

- `apps/web/svelte/static/icons/`

Root lookup files promoted for browser and iOS:

- `apps/web/svelte/static/favicon.ico`
- `apps/web/svelte/static/favicon.svg`
- `apps/web/svelte/static/favicon-16.png`
- `apps/web/svelte/static/favicon-32.png`
- `apps/web/svelte/static/favicon-48.png`
- `apps/web/svelte/static/favicon-64.png`
- `apps/web/svelte/static/apple-touch-icon.png`

Current wired surfaces:

- Browser favicon via `<link rel="icon"...>` in `apps/web/svelte/src/routes/+layout.svelte`
- iOS home screen icon via `<link rel="apple-touch-icon"...>` in `apps/web/svelte/src/routes/+layout.svelte`
- Theme color via `<meta name="theme-color"...>` in `apps/web/svelte/src/routes/+layout.svelte`

Future surfaces kept for later integration:

- `apps/web/svelte/static/icons/pwa/`: web app install metadata and maskable icons
- `apps/web/svelte/static/icons/ios-app-store/`: App Store and TestFlight app icon
- `apps/web/svelte/static/icons/macos/`: dock and bundle icon assets
- `apps/web/svelte/static/icons/windows/`: desktop/taskbar/store assets
- `apps/web/svelte/static/icons/tray/`: tray/menu bar icons
- `apps/web/svelte/static/icons/_source-svg/`: editable source vectors

Guardrail:

- Do not modify the in-app brand components for this icon pack task:
  `apps/web/svelte/src/lib/components/brand/XuvaLogo.svelte` and
  `apps/web/svelte/src/lib/components/brand/XuvaWordmark.svelte`
