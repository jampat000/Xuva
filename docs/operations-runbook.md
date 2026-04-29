# Operations Runbook

Vyrden exposes local-first diagnostics with no vendor APM dependency.

## Request Tracing

Every request receives an `X-Request-ID` response header. If the client sends a safe `X-Request-ID`, Vyrden keeps it; otherwise it generates one. Structured request logs include:

- `correlation_id`
- `method`
- `path`
- `status`
- `duration_ms`
- `remote_addr`

Operational events created by API actions also include `correlationId`, so an operator can follow one user action from request log to queued work.

## Endpoints

- `GET /api/health`: liveness and subsystem checks. Returns `200` even when degraded so monitors can distinguish process liveness from readiness.
- `GET /api/ready`: readiness. Returns `503` when a required runtime path is unavailable or a queue is saturated.
- `GET /api/metrics`: API request latency/errors, queue depth, active workers, event counts, playback/transcode/download/probe outcome counts, and generated alerts.

## Baseline Alerts

| Alert | Trigger | Operator action |
| --- | --- | --- |
| `queue_saturated` | Active workers are at capacity and queued work is waiting. | Pause background scans/checks, wait for active jobs, or increase worker limits. |
| `api_error_rate` | A route has at least 5 requests and 20% or more are server errors. | Search logs by the last `correlation_id`, inspect `/api/ready`, then check the related subsystem. |
| `path.*` degraded | Runtime folder is missing, not a directory, or not writable. | Fix permissions or move the folder in Settings before starting heavy work. |

## First Response Flow

1. Open `/api/ready`. If degraded, fix the listed subsystem first.
2. Open `/api/metrics`. Check `alerts`, then `queues`.
3. Use the latest `correlation_id` from the affected request metric to search structured logs.
4. If the issue involves playback, compare the session inspector and transcode outcome counts.
5. If the issue involves library sync, compare scan/probe queues and probe outcomes.
