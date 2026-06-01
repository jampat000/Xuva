#requires -version 5
<#
.SYNOPSIS
  Builds the Xuva single-installer MSI (WiX v5).

.DESCRIPTION
  First slice: builds xuva-server.exe (which embeds the web UI) and packages it
  into an MSI that installs the engine + the XuvaMediaServer Windows Service.
  ffmpeg/ffprobe, the tray component, firewall rules and the updater are layered
  in by later slices.

  Requires the WiX v5 .NET tool on PATH (`dotnet tool install --global wix`).
  No Node/SPA build is needed: the SPA is already embedded in xuva-server.exe via
  the committed server/internal/webapp/static-next bundle.

.PARAMETER Version
  Four-part version for the output filename (e.g. 1.2.3.0). The MSI's internal
  ProductVersion is currently fixed in Xuva.wxs; wiring this through is a follow-up.

.PARAMETER OutputDir
  Where the .msi is written. Relative paths are resolved from the repo root.
#>
[CmdletBinding()]
param(
    [string]$Version = "0.0.0.0",
    [string]$OutputDir = "dist/windows-msi"
)

$ErrorActionPreference = "Stop"

# Downloads the BtbN LGPL ffmpeg build, verifies its SHA-256 against the upstream
# checksums manifest, and stages ffmpeg.exe + ffprobe.exe into $DestinationDir.
# (Self-contained mirror of build-package.ps1's Save-VerifiedFFmpeg; the two are
# deduped in a later cleanup.)
function Save-VerifiedFFmpeg {
    param([Parameter(Mandatory = $true)][string]$DestinationDir)

    $archiveName = "ffmpeg-master-latest-win64-lgpl.zip"
    $archiveUrl = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/$archiveName"
    $checksumsUrl = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/checksums.sha256"
    $tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("xuva-msi-ffmpeg-" + [System.Guid]::NewGuid().ToString("N"))
    New-Item -ItemType Directory -Path $tempRoot -Force | Out-Null
    try {
        $archivePath = Join-Path $tempRoot $archiveName
        $checksumsPath = Join-Path $tempRoot "checksums.sha256"
        Write-Host "Downloading FFmpeg checksum manifest..."
        Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath -UseBasicParsing
        $checksumsText = Get-Content -LiteralPath $checksumsPath -Raw
        $checksumPattern = "(?im)^([a-f0-9]{64})\s+\*?$([regex]::Escape($archiveName))\s*$"
        $checksumMatch = [regex]::Match($checksumsText, $checksumPattern)
        if (-not $checksumMatch.Success) {
            throw "FFmpeg checksum entry for '$archiveName' was not found."
        }
        $expected = $checksumMatch.Groups[1].Value.ToLowerInvariant()

        Write-Host "Downloading FFmpeg..."
        Invoke-WebRequest -Uri $archiveUrl -OutFile $archivePath -UseBasicParsing
        $actual = (Get-FileHash -LiteralPath $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()
        if ($actual -ne $expected) {
            throw "FFmpeg checksum mismatch. Expected $expected, got $actual."
        }

        $extractDir = Join-Path $tempRoot "extract"
        Expand-Archive -LiteralPath $archivePath -DestinationPath $extractDir -Force
        $binDir = Get-ChildItem -LiteralPath $extractDir -Recurse -Directory |
            Where-Object {
                (Test-Path -LiteralPath (Join-Path $_.FullName "ffmpeg.exe")) -and
                (Test-Path -LiteralPath (Join-Path $_.FullName "ffprobe.exe"))
            } |
            Select-Object -First 1
        if (-not $binDir) {
            throw "Downloaded FFmpeg archive did not contain ffmpeg.exe and ffprobe.exe."
        }

        New-Item -ItemType Directory -Path $DestinationDir -Force | Out-Null
        Copy-Item -LiteralPath (Join-Path $binDir.FullName "ffmpeg.exe") -Destination (Join-Path $DestinationDir "ffmpeg.exe") -Force
        Copy-Item -LiteralPath (Join-Path $binDir.FullName "ffprobe.exe") -Destination (Join-Path $DestinationDir "ffprobe.exe") -Force
    } finally {
        Remove-Item -LiteralPath $tempRoot -Recurse -Force -ErrorAction SilentlyContinue
    }
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$serverDir = Join-Path $repoRoot "server"
$wxsSource = Join-Path $PSScriptRoot "wix\Xuva.wxs"
$staging = Join-Path $repoRoot "dist\msi-staging"
$outputBase = if ([System.IO.Path]::IsPathRooted($OutputDir)) { $OutputDir } else { Join-Path $repoRoot $OutputDir }

Write-Host "Repo root : $repoRoot"
Write-Host "Staging   : $staging"
Write-Host "Output    : $outputBase"

# Fresh staging + output dirs.
Remove-Item -LiteralPath $staging -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path $staging | Out-Null
New-Item -ItemType Directory -Force -Path $outputBase | Out-Null

# 1) Build the server binary (embeds the SPA). Run from the module root.
Write-Host "`n== go build xuva-server.exe =="
$serverExe = Join-Path $staging "xuva-server.exe"
Push-Location $serverDir
try {
    & go build -trimpath -o $serverExe ".\cmd\Xuva"
    if ($LASTEXITCODE -ne 0) { throw "go build failed (exit $LASTEXITCODE)" }
} finally {
    Pop-Location
}
if (-not (Test-Path -LiteralPath $serverExe)) { throw "xuva-server.exe was not produced" }

# 1b) Stage ffmpeg/ffprobe into bin\ (the wxs packages them; the server
#     auto-detects <exeDir>\bin\ffmpeg.exe at runtime).
Write-Host "`n== ffmpeg/ffprobe =="
$binStaging = Join-Path $staging "bin"
Save-VerifiedFFmpeg -DestinationDir $binStaging
foreach ($tool in @("ffmpeg.exe", "ffprobe.exe")) {
    if (-not (Test-Path -LiteralPath (Join-Path $binStaging $tool))) { throw "$tool was not staged" }
}

# 2) Build the MSI. Run wix from the staging dir so File/@Source paths
#    (xuva-server.exe, bin\ffmpeg.exe) resolve, with the .wxs copied alongside.
Write-Host "`n== wix build =="
Copy-Item -LiteralPath $wxsSource -Destination (Join-Path $staging "Xuva.wxs") -Force
$msiOut = Join-Path $outputBase "xuva-server-v$Version.msi"
Push-Location $staging
try {
    & wix build "Xuva.wxs" -arch x64 -o $msiOut
    if ($LASTEXITCODE -ne 0) { throw "wix build failed (exit $LASTEXITCODE)" }
} finally {
    Pop-Location
}
if (-not (Test-Path -LiteralPath $msiOut)) { throw "MSI was not produced at $msiOut" }

$size = [math]::Round((Get-Item -LiteralPath $msiOut).Length / 1MB, 2)
Write-Host "`nMSI built: $msiOut ($size MB)"
