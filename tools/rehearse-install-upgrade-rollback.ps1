param(
	[string]$DataDir = "data",
	[string]$OutputRoot = "artifacts/rehearsals",
	[switch]$SkipDatabase
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

$repoRoot = Split-Path -Parent $PSScriptRoot
$dataRoot = if ([System.IO.Path]::IsPathRooted($DataDir)) { $DataDir } else { Join-Path $repoRoot $DataDir }
$settingsPath = Join-Path $dataRoot "settings.json"
$dbPath = Join-Path $dataRoot "xuva.db"

if (-not (Test-Path -LiteralPath $settingsPath)) { throw "settings.json not found at $settingsPath" }
if (-not $SkipDatabase -and -not (Test-Path -LiteralPath $dbPath)) { throw "xuva.db not found at $dbPath" }

$timestamp = (Get-Date).ToUniversalTime().ToString("yyyyMMddTHHmmssZ")
$outputBase = if ([System.IO.Path]::IsPathRooted($OutputRoot)) { $OutputRoot } else { Join-Path $repoRoot $OutputRoot }
$runDir = Join-Path $outputBase ("rehearsal_" + $timestamp)
$backupDir = Join-Path $runDir "backup"
$stagingDir = Join-Path $runDir "staging"
New-Item -ItemType Directory -Path $backupDir -Force | Out-Null
New-Item -ItemType Directory -Path $stagingDir -Force | Out-Null

$beforeSettingsHash = Get-NormalizedHash -Path $settingsPath
$beforeDBHash = if ($SkipDatabase) { "" } else { Get-NormalizedHash -Path $dbPath }
if (-not $SkipDatabase -and [string]::IsNullOrWhiteSpace($beforeDBHash)) {
	throw "xuva.db is locked or unreadable at $dbPath. Stop Xuva for a full DB rehearsal or rerun with -SkipDatabase."
}

Copy-Item -LiteralPath $settingsPath -Destination (Join-Path $backupDir "settings.json") -Force
if (-not $SkipDatabase) {
	Copy-Item -LiteralPath $dbPath -Destination (Join-Path $backupDir "xuva.db") -Force
}
Copy-Item -LiteralPath $settingsPath -Destination (Join-Path $stagingDir "settings.json") -Force
if (-not $SkipDatabase) {
	Copy-Item -LiteralPath $dbPath -Destination (Join-Path $stagingDir "xuva.db") -Force
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
	Copy-Item -LiteralPath (Join-Path $backupDir "xuva.db") -Destination (Join-Path $stagingDir "xuva.db") -Force
}

$afterSettingsHash = Get-NormalizedHash -Path $stagingSettingsPath
$afterDBHash = if ($SkipDatabase) { "" } else { Get-NormalizedHash -Path (Join-Path $stagingDir "xuva.db") }
$settingsRestored = $beforeSettingsHash -eq $afterSettingsHash
$dbRestored = if ($SkipDatabase) { $true } else { $beforeDBHash -eq $afterDBHash }

$report = [ordered]@{
	createdAtUtc = (Get-Date).ToUniversalTime().ToString("o")
	dataRoot = $dataRoot
	runDir = $runDir
	checks = [ordered]@{
		backupCreated = $true
		settingsRestored = $settingsRestored
		dbRestored = $dbRestored
		dbCheckSkipped = [bool]$SkipDatabase
	}
	hashes = [ordered]@{
		before = [ordered]@{ settings = $beforeSettingsHash; db = $beforeDBHash }
		afterRollback = [ordered]@{ settings = $afterSettingsHash; db = $afterDBHash }
	}
}

$reportPath = Join-Path $runDir "report.json"
$report | ConvertTo-Json -Depth 12 | Set-Content -LiteralPath $reportPath -Encoding UTF8

if (-not $settingsRestored -or -not $dbRestored) {
	throw "Rollback rehearsal failed. See $reportPath"
}

Write-Host "Rollback rehearsal passed. Report: $reportPath"
