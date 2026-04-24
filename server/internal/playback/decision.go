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

type SourceFacts struct {
	MediaSourceID    string
	Container        string
	VideoCodec       string
	AudioStreams     int
	SubtitleStreams  int
	SidecarSubtitles int
	Bitrate          int64
	Probed           bool
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

func (s *Service) DecideSource(_ context.Context, request Request, source SourceFacts) Decision {
	decision := Decision{
		MediaSourceID:           request.MediaSourceID,
		ClientProfile:           request.ClientProfile,
		EstimatedNetworkBitrate: source.Bitrate,
		Selected:                map[string]string{},
		SuggestedFixes:          []string{},
	}
	if !source.Probed {
		decision.Mode = DecisionDeferred
		decision.Reason = "Media source has not been probed yet, so Vyrden cannot safely choose direct play or transcoding."
		decision.ContainerAction = "probe_required"
		decision.VideoAction = "probe_required"
		decision.AudioAction = "probe_required"
		decision.SubtitleAction = "probe_required"
		decision.EstimatedCPUCost = "unknown"
		decision.EstimatedGPUCost = "unknown"
		decision.SuggestedFixes = []string{"Run ffprobe for this media source"}
		return decision
	}

	decision.Selected["container"] = source.Container
	decision.Selected["videoCodec"] = source.VideoCodec
	if isBrowserDirectPlayable(source.Container, source.VideoCodec) {
		decision.Mode = DirectPlay
		decision.Reason = "Container and video codec are browser-compatible, so the server can stream the source directly."
		decision.ContainerAction = "direct"
		decision.VideoAction = "direct"
		decision.AudioAction = "direct"
		decision.SubtitleAction = subtitleAction(source.SubtitleStreams + source.SidecarSubtitles)
		decision.EstimatedCPUCost = "none"
		decision.EstimatedGPUCost = "none"
		return decision
	}

	decision.Mode = VideoTranscode
	decision.Reason = "The probed container or video codec is not safely direct-playable for the requested client profile."
	decision.ContainerAction = "transcode_or_remux"
	decision.VideoAction = "transcode"
	decision.AudioAction = "copy_or_transcode"
	decision.SubtitleAction = subtitleAction(source.SubtitleStreams + source.SidecarSubtitles)
	decision.EstimatedCPUCost = "high"
	decision.EstimatedGPUCost = "optional"
	decision.SuggestedFixes = []string{"Add client capability profiles", "Implement remux/transcode pipeline"}
	return decision
}

func isBrowserDirectPlayable(container string, videoCodec string) bool {
	switch videoCodec {
	case "h264", "av1", "vp9":
	default:
		return false
	}
	return contains(container, "mp4") || contains(container, "mov") || contains(container, "webm")
}

func subtitleAction(count int) string {
	if count > 0 {
		return "select_or_convert"
	}
	return "none"
}

func contains(value string, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
