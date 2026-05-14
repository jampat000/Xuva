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

$env:XUVA_HTTP_ADDR = $HttpAddr
$env:XUVA_AUTH_DISABLED = "false"
$env:XUVA_HARDWARE_UNLOCKED = "true"

Write-Host "Starting Xuva desktop owner mode on $HttpAddr (bootstrap/sign-in flow enabled)..."
Push-Location (Join-Path $repoRoot "server")
try {
	go run ./cmd/xuva
} finally {
	Pop-Location
}
