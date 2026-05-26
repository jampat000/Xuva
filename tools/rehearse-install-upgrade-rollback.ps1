param(
	[string]$DataDir = "data",
	[string]$OutputRoot = "artifacts/rehearsals",
	[switch]$SkipDatabase,
	[switch]$AllowLiveWalCopy
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Get-NormalizedHash {
	param([Parameter(Mandatory = $true)][string]$Path)
	if (-not (Test-Path -LiteralPath $Path)) { return "" }
	try {
		return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash.ToLowerInvariant()
	} catch {
		return ""
	}
}

function Get-DatabaseFileSet {
	param([Parameter(Mandatory = $true)][string]$DatabasePath)
	$files = @($DatabasePath)
	foreach ($suffix in @("-wal", "-shm")) {
		$sidecar = "$DatabasePath$suffix"
		if (Test-Path -LiteralPath $sidecar) {
			$files += $sidecar
		}
	}
	return $files
}

function Copy-DatabaseFileSet {
	param(
		[Parameter(Mandatory = $true)][string[]]$Files,
		[Parameter(Mandatory = $true)][string]$DestinationDir
	)
	foreach ($file in $Files) {
		Copy-Item -LiteralPath $file -Destination (Join-Path $DestinationDir (Split-Path -Leaf $file)) -Force
	}
}

function Get-HashSet {
	param([Parameter(Mandatory = $true)][string[]]$Files)
	$result = [ordered]@{}
	foreach ($file in $Files) {
		$result[(Split-Path -Leaf $file)] = Get-NormalizedHash -Path $file
	}
	return $result
}

function Test-ExclusiveOpen {
	param([Parameter(Mandatory = $true)][string]$Path)
	if (-not (Test-Path -LiteralPath $Path)) { return $true }
	$stream = $null
	try {
		$stream = [System.IO.File]::Open($Path, [System.IO.FileMode]::Open, [System.IO.FileAccess]::ReadWrite, [System.IO.FileShare]::None)
		return $true
	} catch {
		return $false
	} finally {
		if ($stream) { $stream.Dispose() }
	}
}

$repoRoot = Split-Path -Parent $PSScriptRoot
$dataRoot = if ([System.IO.Path]::IsPathRooted($DataDir)) { $DataDir } else { Join-Path $repoRoot $DataDir }
$settingsPath = Join-Path $dataRoot "settings.json"
$dbPath = Join-Path $dataRoot "xuva.db"

if (-not (Test-Path -LiteralPath $settingsPath)) { throw "settings.json not found at $settingsPath" }
if (-not $SkipDatabase -and -not (Test-Path -LiteralPath $dbPath)) { throw "xuva.db not found at $dbPath" }
if (-not $SkipDatabase -and -not $AllowLiveWalCopy) {
	foreach ($candidate in (Get-DatabaseFileSet -DatabasePath $dbPath)) {
		if (-not (Test-ExclusiveOpen -Path $candidate)) {
			throw "$candidate appears to be in use. Stop Xuva for a stable DB rehearsal, rerun with -SkipDatabase, or explicitly pass -AllowLiveWalCopy."
		}
	}
}

$timestamp = (Get-Date).ToUniversalTime().ToString("yyyyMMddTHHmmssZ")
$outputBase = if ([System.IO.Path]::IsPathRooted($OutputRoot)) { $OutputRoot } else { Join-Path $repoRoot $OutputRoot }
$runDir = Join-Path $outputBase ("rehearsal_" + $timestamp)
$backupDir = Join-Path $runDir "backup"
$stagingDir = Join-Path $runDir "staging"
New-Item -ItemType Directory -Path $backupDir -Force | Out-Null
New-Item -ItemType Directory -Path $stagingDir -Force | Out-Null

$beforeSettingsHash = Get-NormalizedHash -Path $settingsPath
$databaseFiles = if ($SkipDatabase) { @() } else { Get-DatabaseFileSet -DatabasePath $dbPath }
$beforeDBHashes = if ($SkipDatabase) { [ordered]@{} } else { Get-HashSet -Files $databaseFiles }
if (-not $SkipDatabase -and [string]::IsNullOrWhiteSpace($beforeDBHashes["xuva.db"])) {
	throw "xuva.db is locked or unreadable at $dbPath. Stop Xuva for a full DB rehearsal or rerun with -SkipDatabase."
}

Copy-Item -LiteralPath $settingsPath -Destination (Join-Path $backupDir "settings.json") -Force
if (-not $SkipDatabase) {
	Copy-DatabaseFileSet -Files $databaseFiles -DestinationDir $backupDir
}
Copy-Item -LiteralPath $settingsPath -Destination (Join-Path $stagingDir "settings.json") -Force
if (-not $SkipDatabase) {
	Copy-DatabaseFileSet -Files $databaseFiles -DestinationDir $stagingDir
}

# Simulate upgrade mutation in staging only.
$stagingSettingsPath = Join-Path $stagingDir "settings.json"
$config = Get-Content -LiteralPath $stagingSettingsPath -Raw | ConvertFrom-Json
$existingName = [string]($config.serverName)
if ([string]::IsNullOrWhiteSpace($existingName)) { $existingName = "Xuva" }
$config.serverName = ($existingName + " (upgrade-rehearsal)")
$config | ConvertTo-Json -Depth 16 | Set-Content -LiteralPath $stagingSettingsPath -Encoding UTF8

# Rollback rehearsal: restore backup artifacts to staging.
Copy-Item -LiteralPath (Join-Path $backupDir "settings.json") -Destination $stagingSettingsPath -Force
if (-not $SkipDatabase) {
	$backupDatabaseFiles = Get-DatabaseFileSet -DatabasePath (Join-Path $backupDir "xuva.db")
	Copy-DatabaseFileSet -Files $backupDatabaseFiles -DestinationDir $stagingDir
}

$afterSettingsHash = Get-NormalizedHash -Path $stagingSettingsPath
$stagedDatabaseFiles = if ($SkipDatabase) { @() } else { Get-DatabaseFileSet -DatabasePath (Join-Path $stagingDir "xuva.db") }
$afterDBHashes = if ($SkipDatabase) { [ordered]@{} } else { Get-HashSet -Files $stagedDatabaseFiles }
$settingsRestored = $beforeSettingsHash -eq $afterSettingsHash
$dbRestored = $true
if (-not $SkipDatabase) {
	$beforeJson = $beforeDBHashes | ConvertTo-Json -Compress
	$afterJson = $afterDBHashes | ConvertTo-Json -Compress
	$dbRestored = $beforeJson -eq $afterJson
}

$report = [ordered]@{
	createdAtUtc = (Get-Date).ToUniversalTime().ToString("o")
	dataRoot = $dataRoot
	runDir = $runDir
	checks = [ordered]@{
		backupCreated = $true
		settingsRestored = $settingsRestored
		dbRestored = $dbRestored
		dbCheckSkipped = [bool]$SkipDatabase
		liveWalCopyAllowed = [bool]$AllowLiveWalCopy
	}
	hashes = [ordered]@{
		before = [ordered]@{ settings = $beforeSettingsHash; database = $beforeDBHashes }
		afterRollback = [ordered]@{ settings = $afterSettingsHash; database = $afterDBHashes }
	}
}

$reportPath = Join-Path $runDir "report.json"
$report | ConvertTo-Json -Depth 12 | Set-Content -LiteralPath $reportPath -Encoding UTF8

if (-not $settingsRestored -or -not $dbRestored) {
	throw "Rollback rehearsal failed. See $reportPath"
}

Write-Host "Rollback rehearsal passed. Report: $reportPath"
