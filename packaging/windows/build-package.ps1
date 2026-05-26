param(
	[string]$Version = "dev",
	[string]$OutputRoot = "dist/windows",
	[switch]$SkipWebInstall,
	[switch]$SkipDesktopInstall,
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

function Get-DesktopVersion {
	param([Parameter(Mandatory = $true)][string]$ReleaseVersion)
	$normalized = $ReleaseVersion.Trim()
	if ($normalized.StartsWith("v")) {
		$normalized = $normalized.Substring(1)
	}
	if ($normalized -match "^\d+\.\d+\.\d+([-.+][0-9A-Za-z.-]+)?$") {
		return $normalized
	}
	return "0.0.0-dev"
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
$desktopRoot = Join-Path $repoRoot "apps\desktop"
$desktopRuntimeRoot = Join-Path $desktopRoot "runtime"
$runtimeBinRoot = Join-Path $desktopRuntimeRoot "bin"
$electronOutput = Join-Path $outputBase "electron"
$desktopVersion = Get-DesktopVersion -ReleaseVersion $Version

New-Item -ItemType Directory -Path $outputBase -Force | Out-Null
Remove-Item -LiteralPath $desktopRuntimeRoot -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item -LiteralPath $electronOutput -Recurse -Force -ErrorAction SilentlyContinue
Get-ChildItem -LiteralPath $outputBase -File -ErrorAction SilentlyContinue |
	Where-Object { $_.Name -like "xuva-v*-win-x64.*" } |
	Remove-Item -Force
New-Item -ItemType Directory -Path $runtimeBinRoot -Force | Out-Null

if (-not $SkipWebInstall) {
	Invoke-Native -FilePath $npm -ArgumentList @("ci") -WorkingDirectory (Join-Path $repoRoot "apps\web\svelte")
}
if (-not $SkipDesktopInstall) {
	Invoke-Native -FilePath $npm -ArgumentList @("ci") -WorkingDirectory $desktopRoot
}
if (-not $LeavePublishedStatic) {
	Assert-StaticNextClean -RepoRoot $repoRoot
}
$publishedStatic = $false
try {
	Invoke-Native -FilePath $npm -ArgumentList @("run", "publish:go-static") -WorkingDirectory (Join-Path $repoRoot "apps\web\svelte")
	$publishedStatic = $true

	$ldflags = "-s -w -X github.com/jampat000/Xuva/server/internal/buildinfo.Version=$Version -X github.com/jampat000/Xuva/server/internal/buildinfo.Commit=$gitCommit -X github.com/jampat000/Xuva/server/internal/buildinfo.Date=$buildDate"
	Invoke-Native -FilePath $go -ArgumentList @("build", "-trimpath", "-ldflags=$ldflags", "-o", (Join-Path $desktopRuntimeRoot "xuva-server.exe"), ".\cmd\Xuva") -WorkingDirectory (Join-Path $repoRoot "server")
} finally {
	if (-not $LeavePublishedStatic -and $publishedStatic) {
		Restore-StaticNext -RepoRoot $repoRoot
	}
}

Save-VerifiedFFmpeg -DestinationDir $runtimeBinRoot

@"
Xuva Runtime

This directory is managed by the packaged Xuva desktop app.

Normal users should launch Xuva.exe from the Start Menu, desktop shortcut,
or extracted portable package root. Do not launch xuva-server.exe directly
unless you are debugging the server runtime.
"@ | Set-Content -LiteralPath (Join-Path $desktopRuntimeRoot "README.txt") -Encoding UTF8

@"
Xuva Windows Package

This build produces unsigned Windows desktop artifacts:

  - xuva-v$desktopVersion-win-x64.exe: per-user installer
  - xuva-v$desktopVersion-win-x64.zip: portable desktop package

Run Xuva.exe from the installed app or extracted zip. The desktop app starts
and supervises the bundled server runtime.

Included runtime dependencies:
  - Xuva.exe desktop shell
  - xuva-server.exe
  - embedded web UI
  - ffmpeg.exe
  - ffprobe.exe

Default data directory:
  %LOCALAPPDATA%\Xuva\data

Default runtime directories:
  %LOCALAPPDATA%\Xuva\transcode
  %LOCALAPPDATA%\Xuva\downloads
  %LOCALAPPDATA%\Xuva\metadata
  %LOCALAPPDATA%\Xuva\cache
  %LOCALAPPDATA%\Xuva\temp
  %LOCALAPPDATA%\Xuva\trailers

Override with environment variables before launch:
  XUVA_DATA_DIR
  XUVA_TRANSCODE_DIR
  XUVA_DOWNLOADS_DIR
  XUVA_METADATA_DIR
  XUVA_CACHE_DIR
  XUVA_TEMP_DIR
  XUVA_TRAILERS_DIR
  XUVA_HTTP_ADDR
  XUVA_FFMPEG_PATH
  XUVA_FFPROBE_PATH

The package is unsigned. Verify the SHA256 checksum published with the GitHub Release.
"@ | Set-Content -LiteralPath (Join-Path $desktopRuntimeRoot "PACKAGE-NOTES.txt") -Encoding UTF8

@"
FFmpeg and FFprobe are bundled from BtbN FFmpeg Builds:
https://github.com/BtbN/FFmpeg-Builds

The build script downloads the LGPL Windows archive and verifies it against
the upstream checksums.sha256 manifest before packaging.
"@ | Set-Content -LiteralPath (Join-Path $desktopRuntimeRoot "THIRD_PARTY_FFMPEG.txt") -Encoding UTF8

try {
	Invoke-Native -FilePath $npm -ArgumentList @(
		"run",
		"dist:win",
		"--",
		"--config.directories.output=$electronOutput",
		"--config.extraMetadata.version=$desktopVersion"
	) -WorkingDirectory $desktopRoot

	$artifacts = Get-ChildItem -LiteralPath $electronOutput -File -ErrorAction Stop |
		Where-Object { $_.Extension -in @(".exe", ".zip") }
	if (-not $artifacts) {
		throw "Electron build did not produce a Windows installer or portable zip."
	}

	foreach ($artifact in $artifacts) {
		$dest = Join-Path $outputBase $artifact.Name
		Copy-Item -LiteralPath $artifact.FullName -Destination $dest -Force
		$sha = Get-Sha256 -Path $dest
		"$sha  $($artifact.Name)" | Set-Content -LiteralPath "$dest.sha256" -Encoding ASCII
		Write-Host "Package: $dest"
		Write-Host "SHA256:  $dest.sha256"
	}
} finally {
	Remove-Item -LiteralPath $desktopRuntimeRoot -Recurse -Force -ErrorAction SilentlyContinue
}
