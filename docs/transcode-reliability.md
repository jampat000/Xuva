# Transcode Reliability

Vyrden normalizes FFmpeg failures so playback and support surfaces can explain what happened and what to do next.

## Failure Taxonomy

| Class | Reason Code | Retryable | User Meaning |
| --- | --- | --- | --- |
| `input_missing` | `source_file_missing` | no | Source file moved, NAS/USB offline, or path is unavailable. |
| `permission_denied` | `transcode_permission_denied` | no | FFmpeg cannot read the source or write temp output. |
| `disk_full` | `transcode_disk_full` | no | Transcode folder has no available space. |
| `retryable_io` | `transient_storage_io` | yes | Storage/network I/O had a transient failure. |
| `unsupported_codec` | `unsupported_codec` | no | FFmpeg build cannot decode one selected stream. |
| `timeout` | `transcode_timeout` | no | Job exceeded its configured timeout. |
| `cancelled` | `transcode_cancelled` | no | User/system cancelled the job. |
| `ffmpeg_failed` | `ffmpeg_failed` | no | FFmpeg failed without a more specific known signature. |

## Job Diagnostics

Transcode jobs now include:

- `attempts`
- `maxAttempts`
- `timeout`
- `failureClass`
- `reasonCode`
- `remediation`
- terminal status, including `failed-timeout` and `cancelled`

## Retry And Cleanup

Retryable storage I/O failures retry up to `maxAttempts` with short backoff. Timeout, cancellation, and failure paths remove partial output files so stale temp artifacts do not look playable.

## Cancel Endpoint

`DELETE /api/work/{id}` cancels queued/running work and returns the terminal job payload.

## Example Failure Payload

```json
{
  "status": "failed-timeout",
  "attempts": 1,
  "maxAttempts": 1,
  "failureClass": "timeout",
  "reasonCode": "transcode_timeout",
  "remediation": "Try a lower quality target, enable hardware acceleration, or increase the timeout."
}
```
