# nas-sync.ps1 — one-way sync from local working copy to NAS backup
# Runs as a Windows scheduled task every 5 minutes.
# Source:  C:\Projects\Xuva
# Dest:    \\storage-city\data\Projects\Xuva
#
# Excluded directories:
#   .git\          — git objects live on GitHub, not the NAS backup
#   node_modules\  — 100k+ files, fully reproducible via npm install
#   .air\          — compiled Go binaries, rebuilt per machine
#   .svelte-kit\   — vite/SvelteKit cache, rebuilds automatically
#
# NOTE: /MIR is NOT used here so that directories existing only on the NAS
# (legacy native app projects: apps/android-tv, apps/apple-core, apps/desktop,
# apps/ios, apps/tvos) are preserved until explicitly removed by the user.
# Switch to /MIR once you've confirmed nothing valuable lives on the NAS only.

$src  = "C:\Projects\Xuva"
$dest = "\\storage-city\data\Projects\Xuva"
$log  = "C:\Projects\Xuva\tools\nas-sync.log"

# Keep last 500 lines of log
if (Test-Path $log) {
    $lines = Get-Content $log
    if ($lines.Count -gt 500) {
        $lines | Select-Object -Last 500 | Set-Content $log -Encoding UTF8
    }
}

robocopy $src $dest `
    /E /R:2 /W:5 /MT:4 `
    /XD .git node_modules .air .svelte-kit `
    /XF "*.log" "*.tmp" `
    /NP /NFL /NDL /NJH /NJS

$rc        = $LASTEXITCODE
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

# 0 = nothing to do, 1 = files copied — only log interesting results
if ($rc -ge 1) {
    $verb = if ($rc -ge 8) { "ERROR" } elseif ($rc -ge 4) { "WARN" } else { "SYNC" }
    Add-Content -Path $log -Value "$timestamp [$verb] robocopy exit $rc" -Encoding UTF8
}
