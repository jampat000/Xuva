param(
	[string]$HttpAddr = "127.0.0.1:8097",
	[switch]$SkipWebBuild,
	[switch]$WebDev
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot

function Get-HttpPort {
	param(
		[Parameter(Mandatory = $true)]
		[string]$Address
	)
	$parts = $Address.Split(":")
	if ($parts.Length -lt 2) {
		throw "Unable to parse port from HttpAddr '$Address'. Expected host:port."
	}
	$portText = $parts[$parts.Length - 1]
	$port = 0
	if (-not [int]::TryParse($portText, [ref]$port)) {
		throw "Unable to parse numeric port from HttpAddr '$Address'."
	}
	return $port
}

function Stop-ProcessOnPort {
	param(
		[Parameter(Mandatory = $true)]
		[int]$Port
	)
	$listeners = @(Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue)
	foreach ($listener in $listeners) {
		$owningPid = $listener.OwningProcess
		if ($null -eq $owningPid -or $owningPid -le 0) {
			continue
		}
		try {
			$proc = Get-Process -Id $owningPid -ErrorAction Stop
			Write-Host "Stopping existing process on port ${Port}: $($proc.ProcessName) (PID $owningPid)"
			Stop-Process -Id $owningPid -Force -ErrorAction Stop
		} catch {
			Write-Warning "Unable to stop process $owningPid on port ${Port}: $($_.Exception.Message)"
		}
	}
}

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

if (-not $WebDev -and -not $SkipWebBuild) {
	Write-Host "Building and publishing web assets for embedded server..."
	Push-Location (Join-Path $repoRoot "apps/web/svelte")
	try {
		& $npmCommand run build
		& $npmCommand run publish:go-static
	} finally {
		Pop-Location
	}
}

$webDevProcess = $null
if ($WebDev) {
	Write-Host "Starting web dev server (Vite) for live UI updates..."
	$webDevPort = if ([string]::IsNullOrWhiteSpace($env:XUVA_WEB_DEV_PORT)) { "5174" } else { $env:XUVA_WEB_DEV_PORT }
	$httpPort = Get-HttpPort -Address $HttpAddr
	$apiOrigin = if ([string]::IsNullOrWhiteSpace($env:XUVA_API_ORIGIN)) { "http://127.0.0.1:$httpPort" } else { $env:XUVA_API_ORIGIN }
	$env:XUVA_WEB_DEV_ORIGIN = "http://127.0.0.1:$webDevPort"
	$env:XUVA_API_ORIGIN = $apiOrigin
	$webDevProcess = Start-Process -FilePath $npmCommand -ArgumentList "run","dev","--","--host","127.0.0.1","--port",$webDevPort,"--strictPort" -WorkingDirectory (Join-Path $repoRoot "apps/web/svelte") -PassThru -WindowStyle Hidden
} else {
	$env:XUVA_WEB_DEV_ORIGIN = ""
}

$env:XUVA_HTTP_ADDR = $HttpAddr
$env:XUVA_AUTH_DISABLED = "false"
$env:XUVA_HARDWARE_UNLOCKED = "true"
$env:XUVA_WEB_DISABLE_ASSET_CACHE = "true"

Stop-ProcessOnPort -Port (Get-HttpPort -Address $HttpAddr)

Write-Host "Starting Xuva desktop owner mode on $HttpAddr (bootstrap/sign-in flow enabled)..."
Push-Location (Join-Path $repoRoot "server")
try {
	& $goCommand run ./cmd/Xuva
} finally {
	if ($null -ne $webDevProcess) {
		try {
			Stop-Process -Id $webDevProcess.Id -Force -ErrorAction SilentlyContinue
		} catch {}
	}
	Pop-Location
}
