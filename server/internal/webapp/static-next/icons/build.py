"""Generate all Xuva icon assets from SVG masters."""
import os, io, struct, zipfile, json, shutil
from pathlib import Path
from resvg_py import svg_to_bytes
from PIL import Image

class _C:
    @staticmethod
    def svg2png(bytestring=None, output_width=None, output_height=None, write_to=None):
        data = bytes(svg_to_bytes(svg_string=bytestring.decode(), width=output_width, height=output_height))
        if write_to:
            with open(write_to, "wb") as f: f.write(data)
        return data
cairosvg = _C()

OUT = Path("/tmp/xuva-icons/dist")
if OUT.exists(): shutil.rmtree(OUT)
OUT.mkdir(parents=True)

STOPS = ("#A78BFA", "#7C3AED", "#DB2777")

# --- Master SVG: bare mark (matches XuvaLogo.tsx geometry) ---
MARK_SVG = f"""<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">
  <defs>
    <linearGradient id="g" x1="4" y1="32" x2="60" y2="32" gradientUnits="userSpaceOnUse">
      <stop offset="0%" stop-color="{STOPS[0]}"/>
      <stop offset="55%" stop-color="{STOPS[1]}"/>
      <stop offset="100%" stop-color="{STOPS[2]}"/>
    </linearGradient>
    <linearGradient id="sh" x1="0" y1="0" x2="0" y2="64" gradientUnits="userSpaceOnUse">
      <stop offset="0%" stop-color="#000" stop-opacity="0.35"/>
      <stop offset="100%" stop-color="#000" stop-opacity="0"/>
    </linearGradient>
  </defs>
  <g transform="translate(0.6 1.2)" opacity="0.5">
    <path d="M 4 4 L 16 4 L 60 32 L 30 32 Z" fill="#000"/>
    <path d="M 4 60 L 16 60 L 60 32 L 30 32 Z" fill="#000"/>
  </g>
  <path d="M 4 4 L 16 4 L 60 32 L 30 32 Z" fill="url(#g)"/>
  <path d="M 4 60 L 16 60 L 60 32 L 30 32 Z" fill="url(#g)"/>
  <path d="M 6 4.6 L 14 4.6 L 58.5 31 L 56 32 Z" fill="#FFF" opacity="0.32"/>
  <path d="M 30 32 L 33 30.2 L 33 33.8 Z" fill="#FFF" opacity="0.55"/>
  <path d="M 4 60 L 16 60 L 60 32 L 30 32 Z" fill="url(#sh)"/>
</svg>"""

# --- Master SVG: app-icon tile (rounded square + radial bg + mark inset 60%) ---
def tile_svg(size=1024, rounded=True):
    r = size * 0.235 if rounded else 0
    inset = size * 0.20  # mark = 60% of tile, centered
    mark_size = size * 0.60
    return f"""<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {size} {size}">
  <defs>
    <radialGradient id="bg" cx="32%" cy="38%" r="80%">
      <stop offset="0%" stop-color="{STOPS[2]}" stop-opacity="0.55"/>
      <stop offset="60%" stop-color="#0a0a0d"/>
      <stop offset="100%" stop-color="#050507"/>
    </radialGradient>
    <linearGradient id="g" x1="4" y1="32" x2="60" y2="32" gradientUnits="userSpaceOnUse">
      <stop offset="0%" stop-color="{STOPS[0]}"/>
      <stop offset="55%" stop-color="{STOPS[1]}"/>
      <stop offset="100%" stop-color="{STOPS[2]}"/>
    </linearGradient>
  </defs>
  <rect width="{size}" height="{size}" rx="{r}" ry="{r}" fill="#0a0a0d"/>
  <rect width="{size}" height="{size}" rx="{r}" ry="{r}" fill="url(#bg)"/>
  <g transform="translate({inset} {inset}) scale({mark_size/64})">
    <path d="M 4 4 L 16 4 L 60 32 L 30 32 Z" fill="url(#g)"/>
    <path d="M 4 60 L 16 60 L 60 32 L 30 32 Z" fill="url(#g)"/>
    <path d="M 6 4.6 L 14 4.6 L 58.5 31 L 56 32 Z" fill="#FFF" opacity="0.32"/>
    <path d="M 30 32 L 33 30.2 L 33 33.8 Z" fill="#FFF" opacity="0.55"/>
  </g>
</svg>"""

# --- Master SVG: monochrome tray icon (white silhouette, no tile, no gradient) ---
TRAY_SVG = """<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">
  <path d="M 4 4 L 16 4 L 60 32 L 30 32 Z" fill="#FFF"/>
  <path d="M 4 60 L 16 60 L 60 32 L 30 32 Z" fill="#FFF"/>
</svg>"""

def render(svg, size, path):
    cairosvg.svg2png(bytestring=svg.encode(), output_width=size, output_height=size,
                     write_to=str(path))

# === Outputs ===

# 1. Browser favicons
(OUT / "favicon").mkdir()
with open(OUT / "favicon" / "favicon.svg", "w") as f:
    f.write(tile_svg(64, rounded=False))  # square favicon, browser may round
for s in [16, 32, 48, 64]:
    render(tile_svg(s*4, rounded=False), s, OUT / "favicon" / f"favicon-{s}.png")

# Build favicon.ico (16,32,48)
ico_imgs = [Image.open(OUT / "favicon" / f"favicon-{s}.png") for s in [16,32,48]]
ico_imgs[0].save(OUT / "favicon" / "favicon.ico", format="ICO",
                 sizes=[(16,16),(32,32),(48,48)])

# 2. Apple touch icon
(OUT / "apple").mkdir()
render(tile_svg(720), 180, OUT / "apple" / "apple-touch-icon.png")

# 3. PWA icons + manifest
(OUT / "pwa").mkdir()
for s in [192, 512]:
    render(tile_svg(s*2), s, OUT / "pwa" / f"icon-{s}.png")
    # maskable variant: same art with extra safe-zone padding
    pad_svg = f"""<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <rect width="100" height="100" fill="#0a0a0d"/>
      <g transform="translate(10 10) scale(0.8)">
        <svg viewBox="0 0 100 100" width="100" height="100">
          <foreignObject width="100" height="100"/>
        </svg>
      </g>
    </svg>"""
    # simpler: render larger tile and inset
    big = Image.open(io.BytesIO(cairosvg.svg2png(bytestring=tile_svg(s*2, rounded=False).encode(),
                                                  output_width=int(s*0.8), output_height=int(s*0.8))))
    canvas = Image.new("RGBA", (s, s), (10,10,13,255))
    canvas.paste(big, (int(s*0.1), int(s*0.1)), big)
    canvas.save(OUT / "pwa" / f"icon-{s}-maskable.png")

manifest = {
    "name": "Xuva", "short_name": "Xuva",
    "description": "Your media, beautifully played.",
    "start_url": "/", "display": "standalone",
    "background_color": "#08080c", "theme_color": "#7C3AED",
    "icons": [
        {"src": "/icons/pwa/icon-192.png", "sizes": "192x192", "type": "image/png"},
        {"src": "/icons/pwa/icon-512.png", "sizes": "512x512", "type": "image/png"},
        {"src": "/icons/pwa/icon-192-maskable.png", "sizes": "192x192", "type": "image/png", "purpose": "maskable"},
        {"src": "/icons/pwa/icon-512-maskable.png", "sizes": "512x512", "type": "image/png", "purpose": "maskable"},
    ],
}
with open(OUT / "pwa" / "manifest.json", "w") as f:
    json.dump(manifest, f, indent=2)

# 4. iOS App Store: 1024x1024, no transparency, no rounded corners (Apple rounds)
(OUT / "ios-app-store").mkdir()
render(tile_svg(1024, rounded=False), 1024, OUT / "ios-app-store" / "AppIcon-1024.png")
# flatten alpha to opaque
img = Image.open(OUT / "ios-app-store" / "AppIcon-1024.png").convert("RGBA")
bg = Image.new("RGB", img.size, (10,10,13))
bg.paste(img, mask=img.split()[3])
bg.save(OUT / "ios-app-store" / "AppIcon-1024.png", "PNG")

# 5. macOS .icns (multi-size)
(OUT / "macos").mkdir()
icns_sizes = [16, 32, 64, 128, 256, 512, 1024]
icns_dir = OUT / "macos" / "Xuva.iconset"
icns_dir.mkdir()
for s in icns_sizes:
    render(tile_svg(max(s*2, 256)), s, icns_dir / f"icon_{s}x{s}.png")
    if s <= 512:
        render(tile_svg(max(s*4, 256)), s*2, icns_dir / f"icon_{s}x{s}@2x.png")

# 6. Windows .ico (multi-size)
(OUT / "windows").mkdir()
win_sizes = [(s,s) for s in [16,24,32,48,64,128,256]]
win_imgs = []
for s,_ in win_sizes:
    render(tile_svg(max(s*2,256), rounded=False), s, OUT / "windows" / f"icon-{s}.png")
    win_imgs.append(Image.open(OUT / "windows" / f"icon-{s}.png"))
win_imgs[-1].save(OUT / "windows" / "Xuva.ico", format="ICO", sizes=win_sizes)

# Windows Store tiles
for s in [44, 150, 310]:
    render(tile_svg(max(s*2,256), rounded=False), s, OUT / "windows" / f"StoreLogo-{s}.png")

# 7. System tray (monochrome, multiple sizes incl. @2x)
(OUT / "tray").mkdir()
for s in [16, 22, 24, 32, 44]:
    render(TRAY_SVG, s, OUT / "tray" / f"tray-{s}.png")
    render(TRAY_SVG, s*2, OUT / "tray" / f"tray-{s}@2x.png")
# macOS template version (must end in Template.png)
render(TRAY_SVG, 22, OUT / "tray" / "trayTemplate.png")
render(TRAY_SVG, 44, OUT / "tray" / "trayTemplate@2x.png")

# 8. Source SVGs
src = OUT / "_source-svg"
src.mkdir()
with open(src / "mark.svg", "w") as f: f.write(MARK_SVG)
with open(src / "tile.svg", "w") as f: f.write(tile_svg(1024))
with open(src / "tile-square.svg", "w") as f: f.write(tile_svg(1024, rounded=False))
with open(src / "tray-mono.svg", "w") as f: f.write(TRAY_SVG)

print("Done. Files:")
for p in sorted(OUT.rglob("*")):
    if p.is_file():
        print(f"  {p.relative_to(OUT)}  ({p.stat().st_size} bytes)")
