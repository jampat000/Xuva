param(
	[string]$HttpAddr = "127.0.0.1:8097",
	[switch]$SkipWebBuild
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot

if (-not $SkipWebBuild) {
	Write-Host "Building and publishing web assets for embedded server..."
	Push-Location (Join-Path $repoRoot "apps/web/svelte")
	try {
		npm run build
		npm run publish:go-static
	} finally {
		Pop-Location
	}
}

$env:LORIVO_HTTP_ADDR = $HttpAddr
$env:LORIVO_DEV_AUTH_BYPASS = "true"
$env:LORIVO_AUTH_DISABLED = "false"

Write-Host "Starting Lorivo desktop owner mode on $HttpAddr ..."
Push-Location (Join-Path $repoRoot "server")
try {
	go run ./cmd/lorivo
} finally {
	Pop-Location
}
