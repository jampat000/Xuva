# nas-sync.ps1 — one-way sync from local working copy to NAS backup
# Runs as a Windows scheduled task every 5 minutes.
# Source:  C:\Projects\Xuva
# Dest:    \\storage-city\data\Projects\Xuva
#
# Excluded:
#   node_modules\  — 100k+ files, fully reproducible via npm install
#   .air\          — compiled binaries, rebuilt per machine
#   .svelte-kit\   — vite cache, rebuilds automatically
#   *.log          — transient log files

$src  = "C:\Projects\Xuva"
$dest = "\\storage-city\data\Projects\Xuva"
$log  = "C:\Projects\Xuva\tools\nas-sync.log"

$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

# Keep last 500 lines of log (trim on each run)
if (Test-Path $log) {
    $lines = Get-Content $log
    if ($lines.Count -gt 500) {
        $lines | Select-Object -Last 500 | Set-Content $log -Encoding UTF8
    }
}

robocopy $src $dest `
    /MIR /R:2 /W:5 /MT:4 `
    /XD node_modules .air .svelte-kit `
    /XF "*.log" "*.tmp" `
    /NP /NFL /NDL /NJH /NJS

$rc = $LASTEXITCODE
# Robocopy exit codes: 0=nothing to do, 1=files copied, 2=extras deleted,
# 3=both, 4+=errors. Only log when something happened or errored.
if ($rc -ge 1) {
    $verb = if ($rc -ge 8) { "ERROR" } elseif ($rc -ge 4) { "WARN" } else { "SYNC" }
    Add-Content -Path $log -Value "$timestamp [$verb] robocopy exit $rc" -Encoding UTF8
}
