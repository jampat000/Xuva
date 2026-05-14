package subtitles

import "strings"

type Sidecar struct {
	Path              string `json:"path"`
	RelPath           string `json:"relPath,omitempty"`
	Format            string `json:"format"`
	Language          string `json:"language,omitempty"`
	Forced            bool   `json:"forced"`
	HearingImpaired   bool   `json:"hearingImpaired"`
	RequiresVideoBurn bool   `json:"requiresVideoBurn"`
}

type ConversionPlan struct {
	Status         string `json:"status"`
	SourceFormat   string `json:"sourceFormat"`
	OutputFormat   string `json:"outputFormat,omitempty"`
	OutputBehavior string `json:"outputBehavior"`
	ServerImpact   string `json:"serverImpact"`
	ReasonCode     string `json:"reasonCode"`
	ReasonText     string `json:"reasonText"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) PlanConversion(sidecar Sidecar, clientProfile string) ConversionPlan {
	_ = clientProfile
	format := strings.ToLower(strings.TrimSpace(sidecar.Format))
	switch format {
	case "vtt", "webvtt":
		return ConversionPlan{
			Status:         "not_required",
			SourceFormat:   format,
			OutputFormat:   "webvtt",
			OutputBehavior: "Use the existing subtitle file directly.",
			ServerImpact:   "none",
			ReasonCode:     "subtitle_direct_supported",
			ReasonText:     "This subtitle is already compatible with the selected player.",
		}
	case "srt", "ass", "ssa":
		return ConversionPlan{
			Status:         "available",
			SourceFormat:   format,
			OutputFormat:   "webvtt",
			OutputBehavior: "Generate a temporary WebVTT sidecar for playback; do not modify the original media file.",
			ServerImpact:   "low",
			ReasonCode:     "subtitle_text_conversion_available",
			ReasonText:     "This text subtitle can be converted without video conversion.",
		}
	case "sub":
		return ConversionPlan{
			Status:         "unsupported",
			SourceFormat:   format,
			OutputBehavior: "Do not convert automatically until the subtitle type is known.",
			ServerImpact:   "unknown",
			ReasonCode:     "subtitle_type_unknown",
			ReasonText:     "This subtitle extension can contain different formats, so Xuva needs inspection before conversion.",
		}
	default:
		return ConversionPlan{
			Status:         "unsupported",
			SourceFormat:   format,
			OutputBehavior: "No automatic conversion path is available.",
			ServerImpact:   "unknown",
			ReasonCode:     "subtitle_conversion_unsupported",
			ReasonText:     "Xuva does not have a safe conversion path for this subtitle format yet.",
		}
	}
}
