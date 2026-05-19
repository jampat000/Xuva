# Xuva dev — starts the Go server (Air hot-reload) and the Vite frontend together.
# Run from the repo root:  .\dev.ps1
#
# Both processes stream their output here. Ctrl+C stops everything.
#
# Local API keys: drop a `.env.local` next to this script (gitignored) with
# lines like:
#   XUVA_DEFAULT_TMDB_API_KEY=your_tmdb_key
#   XUVA_DEFAULT_FANARTTV_API_KEY=your_fanart_key
#   XUVA_DEFAULT_OMDB_API_KEY=your_omdb_key
# The Go server reads these as embedded-default fallbacks (see
# server/internal/config/keys.go for the four-tier resolution).

$ErrorActionPreference = "Stop"

$repoRoot  = $PSScriptRoot
$serverDir = Join-Path $repoRoot "server"
$webDir    = Join-Path $repoRoot "apps\web\svelte"
$envLocal  = Join-Path $repoRoot ".env.local"

# Check Air is installed
if (-not (Get-Command air -ErrorAction SilentlyContinue)) {
    Write-Host "Air not found. Installing..." -ForegroundColor Yellow
    go install github.com/air-verse/air@latest
}

# Load .env.local into the current shell so spawned processes inherit it.
# Lines starting with # are comments. Values may be quoted or unquoted.
$envLoaded = @()
if (Test-Path $envLocal) {
    Get-Content $envLocal | ForEach-Object {
        $line = $_.Trim()
        if ($line.Length -eq 0 -or $line.StartsWith("#")) { return }
        $eq = $line.IndexOf("=")
        if ($eq -lt 1) { return }
        $name = $line.Substring(0, $eq).Trim()
        $value = $line.Substring($eq + 1).Trim().Trim('"').Trim("'")
        [System.Environment]::SetEnvironmentVariable($name, $value, "Process")
        $envLoaded += $name
    }
}

Write-Host ""
Write-Host "  Xuva dev" -ForegroundColor Magenta
Write-Host "  Go  -> http://127.0.0.1:8097  (Air hot-reload)" -ForegroundColor Cyan
Write-Host "  Web -> http://localhost:5173   (Vite HMR)" -ForegroundColor Green
if ($envLoaded.Count -gt 0) {
    Write-Host "  Env -> loaded from .env.local: $($envLoaded -join ', ')" -ForegroundColor DarkGray
} else {
    Write-Host "  Env -> no .env.local found (drop one in repo root to inject API keys)" -ForegroundColor DarkGray
}
Write-Host ""
Write-Host "  Ctrl+C to stop both." -ForegroundColor DarkGray
Write-Host ""

# Start Go server with Air in its own window so logs are separate and readable.
# Use cmd /c set to splat the loaded env vars into the child shell.
$envExports = ""
foreach ($name in $envLoaded) {
    $val = [System.Environment]::GetEnvironmentVariable($name, "Process")
    # PowerShell-quote the value (single quotes don't expand).
    $escaped = $val -replace "'", "''"
    $envExports += "`$env:$name = '$escaped'; "
}

$airJob = Start-Process powershell `
    -ArgumentList "-NoExit", "-Command", "$envExports Set-Location '$serverDir'; air" `
    -PassThru

# Start Vite in its own window
$viteJob = Start-Process powershell `
    -ArgumentList "-NoExit", "-Command", "Set-Location '$webDir'; npm run dev" `
    -PassThru

Write-Host "Started Go server (PID $($airJob.Id)) and Vite (PID $($viteJob.Id))." -ForegroundColor DarkGray
Write-Host "Close those windows or press Ctrl+C here to stop." -ForegroundColor DarkGray

# Wait until the user presses Ctrl+C
try {
    while ($true) { Start-Sleep -Seconds 60 }
} finally {
    Write-Host "`nStopping..." -ForegroundColor Yellow
    Stop-Process -Id $airJob.Id  -ErrorAction SilentlyContinue
    Stop-Process -Id $viteJob.Id -ErrorAction SilentlyContinue
}
