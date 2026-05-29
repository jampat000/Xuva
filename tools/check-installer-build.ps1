# Pre-push installer compile check.
#
# Purpose: catch NSIS / electron-builder failures BEFORE pushing to CI.
# Three back-to-back v0.0.34 release runs failed because the previous
# checks (svelte-check, vitest, agent-check) covered everything EXCEPT
# the Windows installer's NSIS compile - the most fragile part of the
# build because NSIS macros expand at compile time and electron-builder
# bundles plugin DLLs out of a cache that can drift from local to CI.
#
# What this does:
#   1. Builds the Go server binary into apps/desktop/runtime/xuva-server.exe
#      so electron-builder has something to package.
#   2. Stages ffmpeg.exe + ffprobe.exe into apps/desktop/runtime/bin/.
#      For local checks we accept any ffmpeg on PATH or in a known
#      WinGet location; the release build uses the LGPL Windows archive
#      via Save-VerifiedFFmpeg in packaging/windows/build-package.ps1.
#   3. Runs the same `npx electron-builder --win nsis --x64` command CI
#      uses. This is the actual NSIS compile + signing pipeline.
#
# When to run:
#   - Before pushing any change to apps/desktop/installer.nsh
#   - Before pushing any change to apps/desktop/package.json's nsis block
#   - Before tagging a release
#
# Required tools:
#   - go (on PATH)
#   - npm + node 20+ (on PATH)
#   - electron-builder dependencies (installed via apps/desktop/npm ci once)
#   - ffmpeg.exe + ffprobe.exe somewhere findable (WinGet location auto-detected)
#
# Exits 0 on success, non-zero with the failing tool's exit code otherwise.

$ErrorActionPreference = 'Stop'

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$desktopRoot = Join-Path $repoRoot 'apps\desktop'
$runtimeRoot = Join-Path $desktopRoot 'runtime'
$binRoot = Join-Path $runtimeRoot 'bin'
$distRoot = Join-Path $desktopRoot 'dist'

Write-Host "Pre-push installer build check" -ForegroundColor Cyan
Write-Host "  repo:    $repoRoot"
Write-Host "  desktop: $desktopRoot"

# ── 1. Build Go server ───────────────────────────────────────────────────────
Write-Host "[1/4] Building xuva-server.exe..." -ForegroundColor Cyan
if (Test-Path $runtimeRoot) { Remove-Item $runtimeRoot -Recurse -Force }
New-Item -ItemType Directory -Path $binRoot -Force | Out-Null

$gitCommit = (& git -C $repoRoot rev-parse --short HEAD).Trim()
$buildDate = (Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')
$ldflags = "-s -w " +
    "-X github.com/jampat000/Xuva/server/internal/buildinfo.Version=v0.0.0-pretest " +
    "-X github.com/jampat000/Xuva/server/internal/buildinfo.Commit=$gitCommit " +
    "-X github.com/jampat000/Xuva/server/internal/buildinfo.Date=$buildDate"

Push-Location (Join-Path $repoRoot 'server')
try {
    $env:CGO_ENABLED = '0'
    $env:GOOS = 'windows'
    $env:GOARCH = 'amd64'
    & go build -trimpath -ldflags="$ldflags" -o (Join-Path $runtimeRoot 'xuva-server.exe') ./cmd/Xuva
    if ($LASTEXITCODE -ne 0) { throw "go build failed with exit $LASTEXITCODE" }
} finally {
    Pop-Location
    Remove-Item Env:CGO_ENABLED, Env:GOOS, Env:GOARCH -ErrorAction SilentlyContinue
}
Write-Host "  ok: $((Get-Item (Join-Path $runtimeRoot 'xuva-server.exe')).Length) bytes"

# ── 2. Stage ffmpeg ──────────────────────────────────────────────────────────
Write-Host "[2/4] Locating ffmpeg.exe + ffprobe.exe..." -ForegroundColor Cyan
$ffmpegCandidates = @(
    "$env:LOCALAPPDATA\Microsoft\WinGet\Packages\Gyan.FFmpeg.Essentials_Microsoft.Winget.Source_8wekyb3d8bbwe\ffmpeg-*-essentials_build\bin\ffmpeg.exe",
    "$env:LOCALAPPDATA\Microsoft\WinGet\Links\ffmpeg.exe"
) | ForEach-Object { Get-ChildItem -Path $_ -ErrorAction SilentlyContinue | Select-Object -First 1 -ExpandProperty FullName }
$ffmpegPath = $ffmpegCandidates | Where-Object { $_ } | Select-Object -First 1
if (-not $ffmpegPath) {
    $cmd = Get-Command ffmpeg.exe -ErrorAction SilentlyContinue
    if ($cmd) { $ffmpegPath = $cmd.Source }
}
if (-not $ffmpegPath) {
    throw "ffmpeg.exe not found. Install via winget: winget install Gyan.FFmpeg.Essentials"
}
$ffprobePath = $ffmpegPath -replace 'ffmpeg\.exe$', 'ffprobe.exe'
if (-not (Test-Path $ffprobePath)) {
    throw "ffprobe.exe not found next to $ffmpegPath"
}
Copy-Item $ffmpegPath  (Join-Path $binRoot 'ffmpeg.exe')  -Force
Copy-Item $ffprobePath (Join-Path $binRoot 'ffprobe.exe') -Force
Write-Host "  ok: ffmpeg+ffprobe staged at $binRoot"

# ── 3. Ensure desktop deps are installed ─────────────────────────────────────
Write-Host "[3/4] Verifying apps/desktop/node_modules..." -ForegroundColor Cyan
if (-not (Test-Path (Join-Path $desktopRoot 'node_modules\electron-builder'))) {
    Write-Host "  node_modules missing - running npm ci (one-time, ~2 min)..."
    Push-Location $desktopRoot
    try {
        & npm ci
        if ($LASTEXITCODE -ne 0) { throw "npm ci failed with exit $LASTEXITCODE" }
    } finally {
        Pop-Location
    }
}
Write-Host "  ok"

# ── 4. Run electron-builder NSIS compile ─────────────────────────────────────
Write-Host "[4/4] Running electron-builder --win nsis --x64..." -ForegroundColor Cyan
if (Test-Path $distRoot) { Remove-Item $distRoot -Recurse -Force }
Push-Location $desktopRoot
try {
    # Don't use $LASTEXITCODE here - electron-builder writes Node.js
    # deprecation warnings to stderr, which PowerShell wraps in
    # NativeCommandError and surfaces as $LASTEXITCODE=1 even when the
    # underlying build succeeds. Check the actual artifact instead: if the
    # installer .exe exists in $distRoot, the build worked regardless of
    # what PowerShell thinks the exit code was.
    try { & npx electron-builder --win nsis --x64 } catch { Write-Host "  (stderr noise from electron-builder ignored - checking artifacts below)" -ForegroundColor DarkGray }
} finally {
    Pop-Location
}

$installer = Get-ChildItem -Path $distRoot -Filter '*.exe' -ErrorAction SilentlyContinue | Where-Object { $_.Name -notlike '*uninstaller*' } | Select-Object -First 1
if (-not $installer) {
    throw "Installer .exe not produced - electron-builder failed at the NSIS compile step. Inspect the output above."
}
Write-Host ""
Write-Host "PASS - installer built: $($installer.FullName) ($([math]::Round($installer.Length / 1MB, 1)) MB)" -ForegroundColor Green
Write-Host "Safe to push installer-related changes."
