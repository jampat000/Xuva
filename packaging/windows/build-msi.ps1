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

# 2) Build the MSI. Run wix from the staging dir so File/@Source bare filenames
#    (xuva-server.exe) resolve, with the .wxs copied alongside.
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
