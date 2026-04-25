package playback

import (
	"context"
	"fmt"
	"strings"
)

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
	MediaSourceID       string `json:"mediaSourceId"`
	ClientProfile       string `json:"clientProfile"`
	AudioTrackIndex     int    `json:"audioTrackIndex,omitempty"`
	AudioCodec          string `json:"audioCodec,omitempty"`
	AudioChannels       int    `json:"audioChannels,omitempty"`
	SubtitleTrackIndex  int    `json:"subtitleTrackIndex,omitempty"`
	SubtitleCodec       string `json:"subtitleCodec,omitempty"`
	SubtitleMode        string `json:"subtitleMode,omitempty"`
	SubtitleTrackActive bool   `json:"subtitleTrackActive,omitempty"`
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
	AudioCodec       string
	AudioChannels    int
	SubtitleCodec    string
	SubtitleActive   bool
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
	if request.ClientProfile == "" {
		request.ClientProfile = "web"
	}
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
	if source.AudioCodec != "" {
		decision.Selected["audioCodec"] = source.AudioCodec
		decision.Selected["audioTrack"] = fmt.Sprintf("%d", request.AudioTrackIndex)
	}
	if source.SubtitleActive {
		decision.Selected["subtitleCodec"] = source.SubtitleCodec
		decision.Selected["subtitleTrack"] = fmt.Sprintf("%d", request.SubtitleTrackIndex)
	}
	audioAction := audioAction(request.ClientProfile, source.AudioCodec, source.AudioChannels)
	subtitles := selectedSubtitleAction(request.ClientProfile, source)
	if subtitles == "burn_in" {
		decision.Mode = SubtitleBurn
		decision.Reason = "The selected subtitle track is image-based for this client profile, so Vyrden must burn it into the video."
		decision.ContainerAction = "direct_or_remux"
		decision.VideoAction = "transcode_for_subtitles"
		decision.AudioAction = audioAction
		decision.SubtitleAction = subtitles
		decision.EstimatedCPUCost = "high"
		decision.EstimatedGPUCost = "optional"
		decision.SuggestedFixes = []string{"Choose a text subtitle track", "Disable subtitles", "Prepare an SRT sidecar"}
		return decision
	}
	if isDirectPlayable(request.ClientProfile, source.Container, source.VideoCodec) {
		decision.Mode = DirectPlay
		decision.Reason = "Container and video codec match the selected client profile, so the server can stream the source directly."
		decision.ContainerAction = "direct"
		decision.VideoAction = "direct"
		decision.AudioAction = audioAction
		decision.SubtitleAction = subtitles
		decision.EstimatedCPUCost = "none"
		decision.EstimatedGPUCost = "none"
		if audioAction == "transcode" {
			decision.Mode = AudioTranscode
			decision.Reason = "Video can direct play, but the selected audio track is not safe for this client profile."
			decision.EstimatedCPUCost = "medium"
		}
		return decision
	}

	if canRemux(request.ClientProfile, source.VideoCodec) {
		decision.Mode = Remux
		decision.Reason = "The video codec is compatible, but the container is not ideal for the selected client profile."
		decision.ContainerAction = "remux"
		decision.VideoAction = "copy"
		decision.AudioAction = audioAction
		decision.SubtitleAction = subtitles
		decision.EstimatedCPUCost = "low"
		decision.EstimatedGPUCost = "none"
		decision.SuggestedFixes = []string{"Run an ffmpeg remux job for streamable output"}
		return decision
	}

	decision.Mode = VideoTranscode
	decision.Reason = "The probed container or video codec is not safely direct-playable for the requested client profile."
	decision.ContainerAction = "transcode_or_remux"
	decision.VideoAction = "transcode"
	decision.AudioAction = audioAction
	decision.SubtitleAction = subtitles
	decision.EstimatedCPUCost = "high"
	decision.EstimatedGPUCost = "optional"
	decision.SuggestedFixes = []string{"Add client capability profiles", "Implement remux/transcode pipeline"}
	return decision
}

func isDirectPlayable(profile string, container string, videoCodec string) bool {
	container = strings.ToLower(container)
	videoCodec = strings.ToLower(videoCodec)
	switch profile {
	case "android-tv":
		return codecIn(videoCodec, "h264", "hevc", "av1", "vp9") && containerIn(container, "mp4", "matroska", "webm", "mpegts")
	case "apple-tv":
		return codecIn(videoCodec, "h264", "hevc") && containerIn(container, "mp4", "mov")
	default:
		return codecIn(videoCodec, "h264", "av1", "vp9") && containerIn(container, "mp4", "mov", "webm")
	}
}

func canRemux(profile string, videoCodec string) bool {
	videoCodec = strings.ToLower(videoCodec)
	switch profile {
	case "android-tv":
		return codecIn(videoCodec, "h264", "hevc", "av1", "vp9")
	case "apple-tv":
		return codecIn(videoCodec, "h264", "hevc")
	default:
		return codecIn(videoCodec, "h264", "av1", "vp9")
	}
}

func codecIn(value string, codecs ...string) bool {
	value = strings.ToLower(value)
	for _, codec := range codecs {
		if value == codec {
			return true
		}
	}
	return false
}

func containerIn(container string, values ...string) bool {
	for _, value := range values {
		if contains(container, value) {
			return true
		}
	}
	return false
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

func selectedSubtitleAction(profile string, source SourceFacts) string {
	if source.SubtitleActive {
		codec := strings.ToLower(source.SubtitleCodec)
		if codec == "" {
			return "select_or_convert"
		}
		if isImageSubtitle(codec) {
			if profile == "android-tv" || profile == "apple-tv" {
				return "direct_or_burn"
			}
			return "burn_in"
		}
		return "direct_or_convert"
	}
	return "none"
}

func audioAction(profile string, codec string, channels int) string {
	codec = strings.ToLower(codec)
	if codec == "" {
		return "copy_or_transcode"
	}
	switch profile {
	case "android-tv":
		if codecIn(codec, "aac", "ac3", "eac3", "opus", "flac", "dts", "truehd") {
			return "direct"
		}
	case "apple-tv":
		if codecIn(codec, "aac", "ac3", "eac3", "alac", "flac") {
			return "direct"
		}
	default:
		if codecIn(codec, "aac", "mp3", "opus", "vorbis") && channels <= 6 {
			return "direct"
		}
	}
	return "transcode"
}

func isImageSubtitle(codec string) bool {
	return codecIn(codec, "hdmv_pgs_subtitle", "dvd_subtitle", "dvb_subtitle", "pgs")
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
