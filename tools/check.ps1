param(
	[switch]$Release,
	[switch]$SkipFrontendInstall,
	[switch]$SkipFrontendTypeCheck,
	[switch]$SkipFrontendTests,
	[switch]$SkipGoTest,
	[switch]$SkipVulnerabilityScan
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Invoke-Step {
	param(
		[Parameter(Mandatory = $true)][string]$Name,
		[Parameter(Mandatory = $true)][scriptblock]$Script
	)
	Write-Host ""
	Write-Host "==> $Name" -ForegroundColor Cyan
	& $Script
}

function Invoke-Native {
	param(
		[Parameter(Mandatory = $true)][string]$FilePath,
		[string[]]$ArgumentList = @(),
		[string]$WorkingDirectory = (Get-Location).Path
	)
	Push-Location $WorkingDirectory
	try {
		& $FilePath @ArgumentList
		if ($LASTEXITCODE -ne 0) {
			throw "$FilePath $($ArgumentList -join ' ') failed with exit code $LASTEXITCODE"
		}
	} finally {
		Pop-Location
	}
}

function Assert-Command {
	param([Parameter(Mandatory = $true)][string]$Name)
	if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
		throw "Required command '$Name' was not found on PATH."
	}
}

$repoRoot = Split-Path -Parent $PSScriptRoot
$serverRoot = Join-Path $repoRoot "server"
$webRoot = Join-Path $repoRoot "apps\web\svelte"

Assert-Command -Name "git"
Assert-Command -Name "node"
Assert-Command -Name "npm"
Assert-Command -Name "go"

Invoke-Step -Name "Agent route/policy harness" -Script {
	Invoke-Native -FilePath "node" -ArgumentList @("tools/agent-check.cjs") -WorkingDirectory $repoRoot
}

if (-not $SkipFrontendTests) {
	Invoke-Step -Name "Frontend contract tests" -Script {
		Invoke-Native -FilePath "node" -ArgumentList @("--test", "server/internal/webapp/frontend_tests/*.test.cjs") -WorkingDirectory $repoRoot
	}
}

if (-not $SkipFrontendTypeCheck) {
	if (-not $SkipFrontendInstall) {
		Invoke-Step -Name "Frontend dependency install" -Script {
			Invoke-Native -FilePath "npm" -ArgumentList @("ci") -WorkingDirectory $webRoot
		}
	}
	Invoke-Step -Name "Frontend type check" -Script {
		Invoke-Native -FilePath "npm" -ArgumentList @("run", "check") -WorkingDirectory $webRoot
	}
}

if (-not $SkipGoTest) {
	Invoke-Step -Name "Server Go tests" -Script {
		Invoke-Native -FilePath "go" -ArgumentList @("test", "./...") -WorkingDirectory $serverRoot
	}
}

if ($Release -and -not $SkipVulnerabilityScan) {
	Invoke-Step -Name "Go vulnerability scan" -Script {
		Invoke-Native -FilePath "go" -ArgumentList @("run", "golang.org/x/vuln/cmd/govulncheck@latest", "./...") -WorkingDirectory $serverRoot
	}
}

Invoke-Step -Name "Whitespace diff check" -Script {
	Invoke-Native -FilePath "git" -ArgumentList @("diff", "--check") -WorkingDirectory $repoRoot
}

Write-Host ""
Write-Host "All requested checks passed." -ForegroundColor Green
