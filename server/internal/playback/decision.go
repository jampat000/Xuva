package playback

import "context"

type Mode string

const (
	DirectPlay       Mode = "Direct Play"
	Remux            Mode = "Remux"
	AudioTranscode   Mode = "Audio Transcode"
	SubtitleBurn     Mode = "Subtitle Burn"
	VideoTranscode   Mode = "Video Transcode"
	DecisionDeferred Mode = "Decision Deferred"
)

type Request struct {
	MediaSourceID string `json:"mediaSourceId"`
	ClientProfile string `json:"clientProfile"`
}

type Decision struct {
	Mode                    Mode              `json:"mode"`
	Reason                  string            `json:"reason"`
	MediaSourceID           string            `json:"mediaSourceId,omitempty"`
	ClientProfile           string            `json:"clientProfile,omitempty"`
	ContainerAction         string            `json:"containerAction"`
	VideoAction             string            `json:"videoAction"`
	AudioAction             string            `json:"audioAction"`
	SubtitleAction          string            `json:"subtitleAction"`
	EstimatedCPUCost        string            `json:"estimatedCpuCost"`
	EstimatedGPUCost        string            `json:"estimatedGpuCost"`
	EstimatedNetworkBitrate int64             `json:"estimatedNetworkBitrate"`
	Selected                map[string]string `json:"selected"`
	SuggestedFixes          []string          `json:"suggestedFixes"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Decide(_ context.Context, request Request) Decision {
	return Decision{
		Mode:             DecisionDeferred,
		Reason:           "Playback decision engine is scaffolded; media-source probing and client capability matching are the next implementation step.",
		MediaSourceID:    request.MediaSourceID,
		ClientProfile:    request.ClientProfile,
		ContainerAction:  "pending",
		VideoAction:      "pending",
		AudioAction:      "pending",
		SubtitleAction:   "pending",
		EstimatedCPUCost: "unknown",
		EstimatedGPUCost: "unknown",
		Selected:         map[string]string{},
		SuggestedFixes:   []string{"Import probed media sources into SQLite", "Attach a client capability profile", "Evaluate selected version/audio/subtitle tracks"},
	}
}
