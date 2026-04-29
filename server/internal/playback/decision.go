package playback

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
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
	Policy              string `json:"policy,omitempty"`
	RouteType           string `json:"routeType,omitempty"`
	MaxNetworkBitrate   int64  `json:"maxNetworkBitrate,omitempty"`
	AudioTrackIndex     int    `json:"audioTrackIndex,omitempty"`
	AudioCodec          string `json:"audioCodec,omitempty"`
	AudioChannels       int    `json:"audioChannels,omitempty"`
	SubtitleTrackIndex  int    `json:"subtitleTrackIndex,omitempty"`
	SubtitleCodec       string `json:"subtitleCodec,omitempty"`
	SubtitleMode        string `json:"subtitleMode,omitempty"`
	SubtitleTrackActive bool   `json:"subtitleTrackActive,omitempty"`
	Containers          []string
	VideoCodecs         []string
	AudioCodecs         []string
	SubtitleCodecs      []string
}

type SourceFacts struct {
	MediaSourceID    string
	Container        string
	VideoCodec       string
	VideoProfile     string
	VideoLevel       string
	VideoBitDepth    int
	HDR              string
	Width            int
	Height           int
	FrameRate        float64
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
	ReasonCode              string            `json:"reasonCode"`
	ReasonText              string            `json:"reasonText"`
	DecisionTraceID         string            `json:"decisionTraceId"`
	MediaSourceID           string            `json:"mediaSourceId,omitempty"`
	ClientProfile           string            `json:"clientProfile,omitempty"`
	ContainerAction         string            `json:"containerAction"`
	VideoAction             string            `json:"videoAction"`
	AudioAction             string            `json:"audioAction"`
	SubtitleAction          string            `json:"subtitleAction"`
	SubtitleClass           string            `json:"subtitleClass,omitempty"`
	SubtitleImpact          map[string]string `json:"subtitleImpact,omitempty"`
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
	if request.ClientProfile == "" {
		request.ClientProfile = "web"
	}
	decision := Decision{
		Mode:             DecisionDeferred,
		ReasonCode:       "source_required",
		ReasonText:       "Choose a media source before Vyrden can check playback compatibility.",
		MediaSourceID:    request.MediaSourceID,
		ClientProfile:    request.ClientProfile,
		ContainerAction:  "pending",
		VideoAction:      "pending",
		AudioAction:      "pending",
		SubtitleAction:   "pending",
		EstimatedCPUCost: "unknown",
		EstimatedGPUCost: "unknown",
		Selected:         map[string]string{},
		SuggestedFixes:   []string{"Select a file version", "Run the media check if file details are missing"},
	}
	return finalizeDecision(request, SourceFacts{}, decision)
}

func (s *Service) DecideSource(_ context.Context, request Request, source SourceFacts) Decision {
	if request.ClientProfile == "" {
		request.ClientProfile = "web"
	}
	request.Policy = strings.TrimSpace(request.Policy)
	request.RouteType = strings.TrimSpace(request.RouteType)
	decision := Decision{
		MediaSourceID:           request.MediaSourceID,
		ClientProfile:           request.ClientProfile,
		EstimatedNetworkBitrate: source.Bitrate,
		Selected:                map[string]string{},
		SuggestedFixes:          []string{},
	}
	if !source.Probed {
		decision.Mode = DecisionDeferred
		decision.ReasonCode = "probe_required"
		decision.ReasonText = "Vyrden needs to inspect this file once before it can confirm the best playback path."
		decision.ContainerAction = "probe_required"
		decision.VideoAction = "probe_required"
		decision.AudioAction = "probe_required"
		decision.SubtitleAction = "probe_required"
		decision.EstimatedCPUCost = "unknown"
		decision.EstimatedGPUCost = "unknown"
		decision.SuggestedFixes = []string{"Run media check for this file", "Let the library scheduler inspect it in the background"}
		return finalizeDecision(request, source, decision)
	}
	if strings.TrimSpace(source.Container) == "" || strings.TrimSpace(source.VideoCodec) == "" {
		decision.Mode = DecisionDeferred
		decision.ReasonCode = "media_facts_incomplete"
		decision.ReasonText = "The media check completed without enough container or video information to make a safe decision."
		decision.ContainerAction = "unknown"
		decision.VideoAction = "unknown"
		decision.AudioAction = "unknown"
		decision.SubtitleAction = "unknown"
		decision.EstimatedCPUCost = "unknown"
		decision.EstimatedGPUCost = "unknown"
		decision.SuggestedFixes = []string{"Recheck this file", "Inspect the source details", "Try a different file version"}
		return finalizeDecision(request, source, decision)
	}

	decision.Selected["container"] = source.Container
	decision.Selected["videoCodec"] = source.VideoCodec
	if source.Width > 0 && source.Height > 0 {
		decision.Selected["resolution"] = fmt.Sprintf("%dx%d", source.Width, source.Height)
	}
	if source.AudioCodec != "" {
		decision.Selected["audioCodec"] = source.AudioCodec
		decision.Selected["audioTrack"] = fmt.Sprintf("%d", request.AudioTrackIndex)
	}
	if source.SubtitleActive {
		decision.Selected["subtitleCodec"] = source.SubtitleCodec
		decision.Selected["subtitleTrack"] = fmt.Sprintf("%d", request.SubtitleTrackIndex)
	}
	audioAction := audioAction(request, source.AudioCodec, source.AudioChannels)
	subtitles := selectedSubtitleAction(request, source)
	decision.SubtitleClass = subtitleClass(source.SubtitleCodec)
	decision.SubtitleImpact = subtitleImpact(request, source, subtitles)
	if subtitles == "burn_in" {
		decision.Mode = SubtitleBurn
		decision.ReasonCode = "subtitle_burn_required"
		decision.ReasonText = "The selected subtitle track is image-based for this player, so subtitles require video conversion."
		decision.ContainerAction = "direct_or_remux"
		decision.VideoAction = "transcode_for_subtitles"
		decision.AudioAction = audioAction
		decision.SubtitleAction = subtitles
		decision.EstimatedCPUCost = "high"
		decision.EstimatedGPUCost = "optional"
		decision.SuggestedFixes = []string{"Choose a text subtitle track", "Disable subtitles", "Add an SRT sidecar"}
		return finalizeDecision(request, source, decision)
	}
	if networkConstrained(request, source) {
		decision.Mode = VideoTranscode
		decision.ReasonCode = "network_bitrate_limit"
		decision.ReasonText = "The file bitrate is above the selected network limit, so Vyrden should create a lower-bitrate stream for smoother playback."
		decision.ContainerAction = "transcode"
		decision.VideoAction = "transcode"
		decision.AudioAction = audioAction
		decision.SubtitleAction = subtitles
		decision.EstimatedCPUCost = "high"
		decision.EstimatedGPUCost = "optional"
		decision.SuggestedFixes = []string{"Use original quality on LAN", "Raise the network bitrate limit", "Use hardware acceleration for lower server impact"}
		return finalizeDecision(request, source, decision)
	}
	if isDirectPlayable(request, source.Container, source.VideoCodec) {
		decision.Mode = DirectPlay
		decision.ReasonCode = "direct_play_supported"
		decision.ReasonText = "This file matches the selected player profile, so Vyrden can stream it without conversion."
		decision.ContainerAction = "direct"
		decision.VideoAction = "direct"
		decision.AudioAction = audioAction
		decision.SubtitleAction = subtitles
		decision.EstimatedCPUCost = "none"
		decision.EstimatedGPUCost = "none"
		if audioAction == "transcode" {
			decision.Mode = AudioTranscode
			decision.ReasonCode = "audio_conversion_required"
			decision.ReasonText = "The video can play as-is, but the selected audio track needs conversion for this player."
			decision.EstimatedCPUCost = "medium"
			decision.SuggestedFixes = []string{"Choose a compatible audio track", "Allow temporary audio conversion", "Create an optimized version for this device"}
		} else if subtitles == "convert" {
			decision.ReasonCode = "subtitle_conversion_available"
			decision.ReasonText = "The video can play as-is, and Vyrden can convert this text subtitle to a compatible sidecar format."
			decision.EstimatedCPUCost = "low"
			decision.SuggestedFixes = []string{"Use the converted WebVTT subtitle", "Choose a subtitle format this player supports", "Turn subtitles off for pure direct play"}
		}
		return finalizeDecision(request, source, decision)
	}

	if canRemux(request, source.VideoCodec) {
		decision.Mode = Remux
		decision.ReasonCode = "container_remux_required"
		decision.ReasonText = "The video stream is compatible, but the file container needs repackaging for this player."
		decision.ContainerAction = "remux"
		decision.VideoAction = "copy"
		decision.AudioAction = audioAction
		decision.SubtitleAction = subtitles
		decision.EstimatedCPUCost = "low"
		decision.EstimatedGPUCost = "none"
		decision.SuggestedFixes = []string{"Allow temporary repackaging", "Create an optimized version only if you want a stored copy"}
		if audioAction == "transcode" {
			decision.Mode = AudioTranscode
			decision.ReasonCode = "audio_and_container_conversion_required"
			decision.ReasonText = "The video can be kept, but this player needs audio conversion and a streamable container."
			decision.EstimatedCPUCost = "medium"
			decision.SuggestedFixes = []string{"Choose a compatible audio track", "Allow temporary conversion", "Use hardware acceleration if available"}
		}
		return finalizeDecision(request, source, decision)
	}

	decision.Mode = VideoTranscode
	decision.ReasonCode = "video_conversion_required"
	decision.ReasonText = "This player profile needs video conversion for the selected file."
	decision.ContainerAction = "transcode_or_remux"
	decision.VideoAction = "transcode"
	decision.AudioAction = audioAction
	decision.SubtitleAction = subtitles
	decision.EstimatedCPUCost = "high"
	decision.EstimatedGPUCost = "optional"
	decision.SuggestedFixes = []string{"Try a player that supports the original file", "Allow temporary video conversion", "Unlock hardware acceleration to reduce CPU load"}
	return finalizeDecision(request, source, decision)
}

func finalizeDecision(request Request, source SourceFacts, decision Decision) Decision {
	if decision.Reason == "" {
		decision.Reason = decision.ReasonText
	}
	if decision.ReasonText == "" {
		decision.ReasonText = decision.Reason
	}
	if decision.ReasonCode == "" {
		decision.ReasonCode = "decision_ready"
	}
	if decision.Selected == nil {
		decision.Selected = map[string]string{}
	}
	if decision.SuggestedFixes == nil {
		decision.SuggestedFixes = []string{}
	}
	decision.DecisionTraceID = decisionTraceID(request, source, decision)
	return decision
}

func decisionTraceID(request Request, source SourceFacts, decision Decision) string {
	selectedKeys := make([]string, 0, len(decision.Selected))
	for key := range decision.Selected {
		selectedKeys = append(selectedKeys, key)
	}
	sort.Strings(selectedKeys)
	selected := make([]string, 0, len(selectedKeys))
	for _, key := range selectedKeys {
		selected = append(selected, key+"="+decision.Selected[key])
	}
	payload := strings.Join([]string{
		normalize(request.MediaSourceID),
		normalize(request.ClientProfile),
		normalize(request.Policy),
		normalize(request.RouteType),
		fmt.Sprintf("%d", request.MaxNetworkBitrate),
		fmt.Sprintf("%d", request.AudioTrackIndex),
		normalize(request.AudioCodec),
		fmt.Sprintf("%d", request.AudioChannels),
		fmt.Sprintf("%d", request.SubtitleTrackIndex),
		normalize(request.SubtitleCodec),
		fmt.Sprintf("%t", request.SubtitleTrackActive),
		normalize(source.Container),
		normalize(source.VideoCodec),
		fmt.Sprintf("%d", source.VideoBitDepth),
		normalize(source.HDR),
		fmt.Sprintf("%d", source.Width),
		fmt.Sprintf("%d", source.Height),
		fmt.Sprintf("%d", source.Bitrate),
		string(decision.Mode),
		decision.ReasonCode,
		strings.Join(selected, ","),
	}, "|")
	sum := sha256.Sum256([]byte(payload))
	return "dec_" + hex.EncodeToString(sum[:])[:16]
}

func networkConstrained(request Request, source SourceFacts) bool {
	return request.MaxNetworkBitrate > 0 && source.Bitrate > request.MaxNetworkBitrate
}

func isDirectPlayable(request Request, container string, videoCodec string) bool {
	container = strings.ToLower(container)
	videoCodec = strings.ToLower(videoCodec)
	if len(request.Containers) > 0 || len(request.VideoCodecs) > 0 {
		return codecIn(videoCodec, request.VideoCodecs...) && containerIn(container, request.Containers...)
	}
	switch request.ClientProfile {
	case "android-tv":
		return codecIn(videoCodec, "h264", "hevc", "av1", "vp9") && containerIn(container, "mp4", "matroska", "webm", "mpegts")
	case "apple-tv":
		return codecIn(videoCodec, "h264", "hevc") && containerIn(container, "mp4", "mov")
	default:
		return codecIn(videoCodec, "h264", "av1", "vp9") && containerIn(container, "mp4", "mov", "webm")
	}
}

func canRemux(request Request, videoCodec string) bool {
	videoCodec = strings.ToLower(videoCodec)
	if len(request.VideoCodecs) > 0 {
		return codecIn(videoCodec, request.VideoCodecs...)
	}
	switch request.ClientProfile {
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

func selectedSubtitleAction(request Request, source SourceFacts) string {
	if source.SubtitleActive {
		codec := strings.ToLower(source.SubtitleCodec)
		if codec == "" {
			return "select_or_convert"
		}
		if isImageSubtitle(codec) {
			if codecIn(codec, request.SubtitleCodecs...) || request.ClientProfile == "android-tv" || request.ClientProfile == "apple-tv" {
				return "direct_or_burn"
			}
			return "burn_in"
		}
		if len(request.SubtitleCodecs) > 0 && !codecIn(codec, request.SubtitleCodecs...) {
			return "convert"
		}
		if isTextSubtitle(codec) {
			return "direct"
		}
		return "direct_or_convert"
	}
	return "none"
}

func subtitleImpact(request Request, source SourceFacts, action string) map[string]string {
	if !source.SubtitleActive {
		return map[string]string{
			"class":       "off",
			"action":      "none",
			"serverLoad":  "none",
			"userMessage": "Subtitles are off, so they do not affect playback.",
		}
	}
	class := subtitleClass(source.SubtitleCodec)
	impact := map[string]string{
		"class":  class,
		"action": action,
		"codec":  normalize(source.SubtitleCodec),
	}
	switch action {
	case "direct", "direct_or_convert", "direct_or_burn":
		impact["serverLoad"] = "none"
		impact["userMessage"] = "This subtitle can be used without changing the video."
	case "convert":
		impact["serverLoad"] = "low"
		impact["output"] = "webvtt sidecar"
		impact["userMessage"] = "Vyrden can convert this text subtitle to WebVTT without video conversion."
	case "burn_in":
		impact["serverLoad"] = "high"
		impact["output"] = "video with subtitles burned in"
		impact["userMessage"] = "This image subtitle requires video conversion for the selected player."
	default:
		impact["serverLoad"] = "unknown"
		impact["userMessage"] = "Subtitle impact needs a media check."
	}
	if request.ClientProfile != "" {
		impact["profile"] = request.ClientProfile
	}
	return impact
}

func audioAction(request Request, codec string, channels int) string {
	codec = strings.ToLower(codec)
	if codec == "" {
		return "copy_or_transcode"
	}
	if len(request.AudioCodecs) > 0 {
		if codecIn(codec, request.AudioCodecs...) && channels <= 8 {
			return "direct"
		}
		return "transcode"
	}
	switch request.ClientProfile {
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

func isTextSubtitle(codec string) bool {
	return codecIn(codec, "srt", "subrip", "webvtt", "vtt", "ass", "ssa", "mov_text")
}

func subtitleClass(codec string) string {
	codec = normalize(codec)
	switch {
	case codec == "":
		return "none"
	case isImageSubtitle(codec):
		return "image"
	case isTextSubtitle(codec):
		return "text"
	default:
		return "unknown"
	}
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

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
