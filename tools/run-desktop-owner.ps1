param(
	[string]$HttpAddr = "127.0.0.1:8097",
	[switch]$SkipWebBuild
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot

function Resolve-CommandPath {
	param(
		[Parameter(Mandatory = $true)]
		[string[]]$Candidates,
		[string[]]$FallbackPaths = @(),
		[Parameter(Mandatory = $true)]
		[string]$ErrorMessage
	)
	foreach ($candidate in $Candidates) {
		$command = Get-Command $candidate -ErrorAction SilentlyContinue
		if ($null -ne $command -and -not [string]::IsNullOrWhiteSpace($command.Source)) {
			return $command.Source
		}
	}
	foreach ($path in $FallbackPaths) {
		if (-not [string]::IsNullOrWhiteSpace($path) -and (Test-Path -LiteralPath $path)) {
			return $path
		}
	}
	throw $ErrorMessage
}

$npmFallbacks = @(
	"$(Join-Path $env:ProgramFiles 'nodejs\npm.cmd')",
	"$(Join-Path $env:ProgramFiles 'nodejs\npm.exe')",
	"$(Join-Path ${env:ProgramFiles(x86)} 'nodejs\npm.cmd')"
)
$npmCommand = Resolve-CommandPath -Candidates @("npm.cmd", "npm") -FallbackPaths $npmFallbacks -ErrorMessage "npm is required. Install Node.js and ensure npm is available on PATH."
$goFallbacks = @(
	"$(Join-Path $env:ProgramFiles 'Go\bin\go.exe')",
	"C:\Go\bin\go.exe"
)
$goCommand = Resolve-CommandPath -Candidates @("go.exe", "go") -FallbackPaths $goFallbacks -ErrorMessage "go is required. Install Go and ensure go is available on PATH."

if (-not $SkipWebBuild) {
	Write-Host "Building and publishing web assets for embedded server..."
	Push-Location (Join-Path $repoRoot "apps/web/svelte")
	try {
		& $npmCommand run build
		& $npmCommand run publish:go-static
	} finally {
		Pop-Location
	}
}

$env:XUVA_HTTP_ADDR = $HttpAddr
$env:XUVA_AUTH_DISABLED = "false"
$env:XUVA_HARDWARE_UNLOCKED = "true"
$env:XUVA_WEB_DISABLE_ASSET_CACHE = "true"

Write-Host "Starting Xuva desktop owner mode on $HttpAddr (bootstrap/sign-in flow enabled)..."
Push-Location (Join-Path $repoRoot "server")
try {
	& $goCommand run ./cmd/Xuva
} finally {
	Pop-Location
}
