Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Test-UrlUp {
	param([string]$Url)
	try {
		$response = Invoke-WebRequest -UseBasicParsing -Uri $Url -TimeoutSec 2
		return $response.StatusCode -ge 200 -and $response.StatusCode -lt 500
	} catch {
		return $false
	}
}

$goUp = Test-UrlUp "http://127.0.0.1:8097/"
$viteUp = Test-UrlUp "http://127.0.0.1:5173/"

$goState = if ($goUp) { "UP" } else { "DOWN" }
$viteState = if ($viteUp) { "UP" } else { "DOWN" }

Write-Host "Go server (8097): $goState"
Write-Host "Vite dev (5173): $viteState"

if ($goUp -and $viteUp) {
	Write-Host "Dev mode status: HEALTHY (WebDev live proxy likely active)"
	exit 0
}

Write-Host "Dev mode status: DEGRADED"
exit 1
