param(
	[Parameter(Mandatory = $true)][string]$PackagePath,
	[string]$ChecksumPath = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Assert-Exists {
	param([Parameter(Mandatory = $true)][string]$Path)
	if (-not (Test-Path -LiteralPath $Path)) {
		throw "Expected package entry missing: $Path"
	}
}

function Assert-Absent {
	param([Parameter(Mandatory = $true)][string]$Path)
	if (Test-Path -LiteralPath $Path) {
		throw "Package contains forbidden entry: $Path"
	}
}

function Assert-NoLegacyDesktopIconConfig {
	$repoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot "..\..")).Path
	$packageJsonPath = Join-Path $repoRoot "apps\desktop\package.json"
	$configText = Get-Content -LiteralPath $packageJsonPath -Raw

	$legacyIcon = "assets/xuva.ico"
	if ($configText.Contains($legacyIcon)) {
		throw "Desktop package config still references removed legacy icon $legacyIcon."
	}
	Assert-Absent -Path (Join-Path (Split-Path -Parent $packageJsonPath) $legacyIcon)
}

Assert-NoLegacyDesktopIconConfig

$resolvedPackage = (Resolve-Path -LiteralPath $PackagePath).Path
if ([string]::IsNullOrWhiteSpace($ChecksumPath)) {
	$ChecksumPath = "$resolvedPackage.sha256"
}
$resolvedChecksum = (Resolve-Path -LiteralPath $ChecksumPath).Path

$actualHash = (Get-FileHash -LiteralPath $resolvedPackage -Algorithm SHA256).Hash.ToLowerInvariant()
$checksumText = Get-Content -LiteralPath $resolvedChecksum -Raw
if ($checksumText -notmatch "(?i)\b$actualHash\b") {
	throw "Package checksum mismatch. Expected checksum file to contain $actualHash."
}

$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("xuva-package-verify-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tempRoot -Force | Out-Null
try {
	Expand-Archive -LiteralPath $resolvedPackage -DestinationPath $tempRoot -Force

	Assert-Exists -Path (Join-Path $tempRoot "Xuva.exe")
	Assert-Exists -Path (Join-Path $tempRoot "resources\app.asar")
	Assert-Exists -Path (Join-Path $tempRoot "resources\runtime\xuva-server.exe")
	Assert-Exists -Path (Join-Path $tempRoot "resources\runtime\README.txt")
	Assert-Exists -Path (Join-Path $tempRoot "resources\runtime\PACKAGE-NOTES.txt")
	Assert-Exists -Path (Join-Path $tempRoot "resources\runtime\THIRD_PARTY_FFMPEG.txt")
	Assert-Exists -Path (Join-Path $tempRoot "resources\runtime\bin\ffmpeg.exe")
	Assert-Exists -Path (Join-Path $tempRoot "resources\runtime\bin\ffprobe.exe")

	Assert-Absent -Path (Join-Path $tempRoot "apps")
	Assert-Absent -Path (Join-Path $tempRoot "server\data")
	Assert-Absent -Path (Join-Path $tempRoot ".git")
	Assert-Absent -Path (Join-Path $tempRoot "app\Start-Xuva.ps1")

	$forbiddenSwift = Get-ChildItem -LiteralPath $tempRoot -Recurse -Filter "*.swift" -ErrorAction SilentlyContinue | Select-Object -First 1
	if ($forbiddenSwift) {
		throw "Package contains forbidden Swift file: $($forbiddenSwift.FullName)"
	}
} finally {
	Remove-Item -LiteralPath $tempRoot -Recurse -Force -ErrorAction SilentlyContinue
}

Write-Host "Package verification passed: $resolvedPackage"
