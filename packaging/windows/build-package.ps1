param(
	[string]$Version = "dev",
	[string]$OutputRoot = "dist/windows",
	[switch]$SkipWebInstall,
	[switch]$LeavePublishedStatic
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Invoke-Native {
	param(
		[Parameter(Mandatory = $true)][string]$FilePath,
		[string[]]$ArgumentList = @(),
		[string]$WorkingDirectory = (Get-Location).Path
	)
	Write-Host "Running: $FilePath $($ArgumentList -join ' ')" -ForegroundColor DarkGray
	Push-Location $WorkingDirectory
	try {
		& $FilePath @ArgumentList
		if ($LASTEXITCODE -ne 0) {
			throw "$FilePath failed with exit code $LASTEXITCODE"
		}
	} finally {
		Pop-Location
	}
}

function Get-RequiredCommand {
	param([Parameter(Mandatory = $true)][string]$Name)
	$cmd = Get-Command $Name -ErrorAction SilentlyContinue
	if (-not $cmd) {
		throw "Required build command '$Name' was not found on PATH."
	}
	return $cmd.Source
}

function Get-Sha256 {
	param([Parameter(Mandatory = $true)][string]$Path)
	return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash.ToLowerInvariant()
}

function Get-GitValue {
	param([Parameter(Mandatory = $true)][string[]]$Args)
	try {
		$value = (& git @Args 2>$null)
		if ($LASTEXITCODE -eq 0) { return ([string]$value).Trim() }
	} catch {}
	return ""
}

function Test-GitWorkTree {
	param([Parameter(Mandatory = $true)][string]$RepoRoot)
	& git -C $RepoRoot rev-parse --is-inside-work-tree *> $null
	return $LASTEXITCODE -eq 0
}

function Assert-StaticNextClean {
	param([Parameter(Mandatory = $true)][string]$RepoRoot)
	if (-not (Test-GitWorkTree -RepoRoot $RepoRoot)) { return }
	& git -C $RepoRoot diff --quiet -- server/internal/webapp/static-next
	if ($LASTEXITCODE -ne 0) {
		throw "server/internal/webapp/static-next has tracked changes. Commit/stash them before packaging, or rerun with -LeavePublishedStatic."
	}
	$untracked = & git -C $RepoRoot ls-files --others --exclude-standard -- server/internal/webapp/static-next
	if ($LASTEXITCODE -ne 0) {
		throw "Could not inspect static-next untracked files."
	}
	if ($untracked) {
		throw "server/internal/webapp/static-next has untracked files. Clean them before packaging, or rerun with -LeavePublishedStatic."
	}
}

function Restore-StaticNext {
	param([Parameter(Mandatory = $true)][string]$RepoRoot)
	if (-not (Test-GitWorkTree -RepoRoot $RepoRoot)) { return }
	Write-Host "Restoring generated static web assets in working tree..."
	& git -C $RepoRoot restore --worktree --staged -- server/internal/webapp/static-next
	if ($LASTEXITCODE -ne 0) {
		throw "Failed to restore tracked static-next files after packaging."
	}
	& git -C $RepoRoot clean -fdx -- server/internal/webapp/static-next
	if ($LASTEXITCODE -ne 0) {
		throw "Failed to clean generated static-next files after packaging."
	}
}

function Save-VerifiedFFmpeg {
	param([Parameter(Mandatory = $true)][string]$DestinationDir)

	$archiveName = "ffmpeg-master-latest-win64-lgpl.zip"
	$archiveUrl = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/$archiveName"
	$checksumsUrl = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/checksums.sha256"
	$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("xuva-ffmpeg-" + [System.Guid]::NewGuid().ToString("N"))
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
		$actual = Get-Sha256 -Path $archivePath
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
$go = Get-RequiredCommand -Name "go"
$npm = Get-RequiredCommand -Name "npm"
$gitCommit = Get-GitValue -Args @("rev-parse", "HEAD")
if ([string]::IsNullOrWhiteSpace($gitCommit)) { $gitCommit = "unknown" }
$buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

$outputBase = if ([System.IO.Path]::IsPathRooted($OutputRoot)) { $OutputRoot } else { Join-Path $repoRoot $OutputRoot }
$stageRoot = Join-Path $outputBase "stage"
$packageName = "xuva-$Version-win-x64"
$packageRoot = Join-Path $stageRoot $packageName
$appRoot = Join-Path $packageRoot "app"
$binRoot = Join-Path $appRoot "bin"
Remove-Item -LiteralPath $packageRoot -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Path $binRoot -Force | Out-Null

if (-not $SkipWebInstall) {
	Invoke-Native -FilePath $npm -ArgumentList @("ci") -WorkingDirectory (Join-Path $repoRoot "apps\web\svelte")
}
if (-not $LeavePublishedStatic) {
	Assert-StaticNextClean -RepoRoot $repoRoot
}
$publishedStatic = $false
try {
	Invoke-Native -FilePath $npm -ArgumentList @("run", "publish:go-static") -WorkingDirectory (Join-Path $repoRoot "apps\web\svelte")
	$publishedStatic = $true

	$ldflags = "-s -w -X github.com/jampat000/Xuva/server/internal/buildinfo.Version=$Version -X github.com/jampat000/Xuva/server/internal/buildinfo.Commit=$gitCommit -X github.com/jampat000/Xuva/server/internal/buildinfo.Date=$buildDate"
	Invoke-Native -FilePath $go -ArgumentList @("build", "-trimpath", "-ldflags=$ldflags", "-o", (Join-Path $appRoot "xuva.exe"), ".\cmd\Xuva") -WorkingDirectory (Join-Path $repoRoot "server")
} finally {
	if (-not $LeavePublishedStatic -and $publishedStatic) {
		Restore-StaticNext -RepoRoot $repoRoot
	}
}

Save-VerifiedFFmpeg -DestinationDir $binRoot

@"
param()

`$root = Split-Path -Parent `$MyInvocation.MyCommand.Path
if (-not `$env:XUVA_DATA_DIR) {
	`$env:XUVA_DATA_DIR = Join-Path `$env:LOCALAPPDATA "Xuva\data"
}
if (-not `$env:XUVA_HTTP_ADDR) {
	`$env:XUVA_HTTP_ADDR = "0.0.0.0:8097"
}
`$env:XUVA_FFMPEG_PATH = Join-Path `$root "bin\ffmpeg.exe"
`$env:XUVA_FFPROBE_PATH = Join-Path `$root "bin\ffprobe.exe"

& (Join-Path `$root "xuva.exe")
"@ | Set-Content -LiteralPath (Join-Path $appRoot "Start-Xuva.ps1") -Encoding UTF8

@"
Xuva Windows Package

This is an unsigned portable package.

Run:
  powershell -ExecutionPolicy Bypass -File .\Start-Xuva.ps1

Then open:
  http://localhost:8097/

Included runtime dependencies:
  - xuva.exe
  - embedded web UI
  - ffmpeg.exe
  - ffprobe.exe

Default data directory:
  %LOCALAPPDATA%\Xuva\data

Override with environment variables before launch:
  XUVA_DATA_DIR
  XUVA_HTTP_ADDR
  XUVA_FFMPEG_PATH
  XUVA_FFPROBE_PATH

The package is unsigned. Verify the SHA256 checksum published with the GitHub Release.
"@ | Set-Content -LiteralPath (Join-Path $packageRoot "README.txt") -Encoding UTF8

@"
FFmpeg and FFprobe are bundled from BtbN FFmpeg Builds:
https://github.com/BtbN/FFmpeg-Builds

The build script downloads the LGPL Windows archive and verifies it against
the upstream checksums.sha256 manifest before packaging.
"@ | Set-Content -LiteralPath (Join-Path $packageRoot "THIRD_PARTY_FFMPEG.txt") -Encoding UTF8

$zipPath = Join-Path $outputBase "$packageName.zip"
Remove-Item -LiteralPath $zipPath -Force -ErrorAction SilentlyContinue
Compress-Archive -Path (Join-Path $packageRoot "*") -DestinationPath $zipPath -Force

$shaPath = "$zipPath.sha256"
$sha = Get-Sha256 -Path $zipPath
"$sha  $(Split-Path -Leaf $zipPath)" | Set-Content -LiteralPath $shaPath -Encoding ASCII

Write-Host "Package: $zipPath"
Write-Host "SHA256:  $shaPath"
