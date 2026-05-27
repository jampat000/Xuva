# Xuva Icon Pack Surfaces

Status: no approved icon pack is currently wired.

The previous MediaMop and legacy Xuva chevron icon files were removed because they are not applicable to the current product branding.

Removed surfaces:

- Windows desktop package icon: `apps/desktop/assets/xuva.ico`
- Browser favicons under `apps/web/svelte/static/`
- Generated server web copies under `server/internal/webapp/static-next/`
- Temporary MediaMop icons under `.tmp/MediaMop/`

Before the next Windows release, add a new approved icon pack and wire it deliberately:

- Windows executable, installer, uninstaller, tray, and shortcuts.
- Browser favicon SVG/ICO/PNG.
- Apple touch icon if web install support is desired.
- Generated server web static copy.

Guardrail:

- Do not reintroduce MediaMop icons or the removed legacy Xuva chevron icon.
- Do not modify Apple client assets from this server/package task.
