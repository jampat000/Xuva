package transcode

import (
	"context"
	"strings"
)

type Failure struct {
	Class       string `json:"class"`
	ReasonCode  string `json:"reasonCode"`
	ReasonText  string `json:"reasonText"`
	Remediation string `json:"remediation"`
	Retryable   bool   `json:"retryable"`
}

func ClassifyFailure(stderr string, contextErr error) Failure {
	text := strings.ToLower(stderr)
	switch contextErr {
	case context.DeadlineExceeded:
		return Failure{Class: "timeout", ReasonCode: "transcode_timeout", ReasonText: "Transcode exceeded its timeout.", Remediation: "Try a lower quality target, enable hardware acceleration, or increase the timeout.", Retryable: false}
	case context.Canceled:
		return Failure{Class: "cancelled", ReasonCode: "transcode_cancelled", ReasonText: "Transcode was cancelled.", Remediation: "Start the job again if playback still needs a prepared stream.", Retryable: false}
	}
	switch {
	case strings.Contains(text, "no such file") || strings.Contains(text, "does not exist") || strings.Contains(text, "could not open"):
		return Failure{Class: "input_missing", ReasonCode: "source_file_missing", ReasonText: "FFmpeg could not read the source file.", Remediation: "Check that the library path is online and the file has not moved.", Retryable: false}
	case strings.Contains(text, "permission denied") || strings.Contains(text, "access is denied"):
		return Failure{Class: "permission_denied", ReasonCode: "transcode_permission_denied", ReasonText: "FFmpeg does not have permission to read or write the file.", Remediation: "Check folder permissions for the media source and transcode temp folder.", Retryable: false}
	case strings.Contains(text, "no space left") || strings.Contains(text, "disk full"):
		return Failure{Class: "disk_full", ReasonCode: "transcode_disk_full", ReasonText: "The transcode folder ran out of disk space.", Remediation: "Free disk space or move the transcode temp folder to a larger drive.", Retryable: false}
	case strings.Contains(text, "connection reset") || strings.Contains(text, "resource temporarily unavailable") || strings.Contains(text, "i/o error") || strings.Contains(text, "input/output error"):
		return Failure{Class: "retryable_io", ReasonCode: "transient_storage_io", ReasonText: "The storage path reported a transient I/O failure.", Remediation: "Lorivo retried the job; check NAS/USB/network stability if it continues.", Retryable: true}
	case strings.Contains(text, "unknown decoder") || strings.Contains(text, "decoder not found") || strings.Contains(text, "unsupported codec"):
		return Failure{Class: "unsupported_codec", ReasonCode: "unsupported_codec", ReasonText: "FFmpeg cannot decode one of the selected streams.", Remediation: "Try a different file version or install an FFmpeg build with the required codec.", Retryable: false}
	default:
		return Failure{Class: "ffmpeg_failed", ReasonCode: "ffmpeg_failed", ReasonText: "FFmpeg failed without a more specific known reason.", Remediation: "Inspect the FFmpeg output and try a different target profile.", Retryable: false}
	}
}

func statusForFailure(failure Failure) Status {
	switch failure.Class {
	case "timeout":
		return StatusTimeout
	case "cancelled":
		return StatusCanceled
	default:
		return StatusFailed
	}
}
