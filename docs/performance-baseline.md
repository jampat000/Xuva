# Performance Baseline

This baseline is produced by the Go benchmark harness added for large-library work.

## Benchmarks

Run from `server`:

```powershell
go test ./internal/scanner ./internal/catalog -bench Benchmark -benchmem
```

Covered paths:

- `BenchmarkScanFullLibrary`: full filesystem walk over 1,000 media files.
- `BenchmarkScanIncrementalLowChangeLibrary`: filesystem walk with known file signatures and one changed file.
- `BenchmarkCatalogMovieBrowse`: browse query over a 1,000 movie catalog.
- `BenchmarkCatalogMovieScanPersistFull`: persist a 1,000 file full movie scan.
- `BenchmarkCatalogMovieScanPersistIncrementalLowChange`: persist a low-change incremental movie scan with one changed file.

## Baseline From Local Run

Environment: Windows amd64, AMD Ryzen 7 9800X3D.

| Benchmark | Result |
| --- | ---: |
| Scan full library | 20.47 ms/op |
| Scan incremental low-change library | 20.51 ms/op, lower memory than full scan |
| Catalog movie browse | 5.06 ms/op |
| Persist full movie scan | 94.03 ms/op |
| Persist incremental low-change scan | 6.82 ms/op |

The filesystem walk remains bounded by the storage device, so the major win is in persistence and downstream work: low-change incremental persistence is about 13x faster than rewriting every catalog row.

## Expected Signals

- Incremental scan reports `changedFiles` and `unchangedFiles`, and changed paths are ordered first for downstream probe/metadata priority.
- Catalog browse should stay bounded by `LIMIT` and indexed sort paths.
- Queue metrics from `/api/metrics` should show playback-critical transcode work isolated from scan/probe background work.

## Current Safeguards

- Per-file scan state is persisted in `scan_file_state`.
- Runtime paths and queue saturation are visible through `/api/ready`.
- Scan/probe/transcode queues remain separate, with worker utilization exposed in `/api/metrics`.
- Storage defaults recommend conservative workers for network/removable/mounted libraries.

## Regression Risk

Incremental state depends on `rel_path`, `size_bytes`, and `modified_at`. Filesystems with coarse timestamp resolution may occasionally mark a file unchanged until the next scan. A manual full scan remains available by clearing scan state or moving the library path.
