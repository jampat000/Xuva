package adaptive

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Request struct {
	MediaSourceID     string `json:"mediaSourceId"`
	ClientProfile     string `json:"clientProfile"`
	RouteType         string `json:"routeType"`
	SourceBitrate     int64  `json:"sourceBitrate"`
	MaxNetworkBitrate int64  `json:"maxNetworkBitrate"`
	Width             int    `json:"width"`
	Height            int    `json:"height"`
	SupportsHLS       bool   `json:"supportsHls"`
}

type Plan struct {
	Protocol       string    `json:"protocol"`
	MediaSourceID  string    `json:"mediaSourceId"`
	Enabled        bool      `json:"enabled"`
	ReasonCode     string    `json:"reasonCode"`
	Reason         string    `json:"reason"`
	SegmentSeconds int       `json:"segmentSeconds"`
	Variants       []Variant `json:"variants"`
}

type Variant struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Bitrate     int64  `json:"bitrate"`
	VideoCodec  string `json:"videoCodec"`
	AudioCodec  string `json:"audioCodec"`
	PlaylistURL string `json:"playlistUrl"`
}

type Telemetry struct {
	SessionID       string  `json:"sessionId"`
	MediaSourceID   string  `json:"mediaSourceId"`
	ClientProfile   string  `json:"clientProfile"`
	Event           string  `json:"event"`
	VariantID       string  `json:"variantId"`
	PreviousVariant string  `json:"previousVariantId"`
	BufferSeconds   float64 `json:"bufferSeconds"`
	StallMS         int64   `json:"stallMs"`
	ObservedBitrate int64   `json:"observedBitrate"`
	CorrelationID   string  `json:"correlationId"`
	CreatedAt       string  `json:"createdAt"`
}

func BuildPlan(request Request) Plan {
	plan := Plan{
		Protocol:       "hls",
		MediaSourceID:  request.MediaSourceID,
		SegmentSeconds: 4,
	}
	if !request.SupportsHLS {
		plan.ReasonCode = "client_no_hls"
		plan.Reason = "This player does not advertise HLS support, so Xuva should use direct, repackage, or transcode routes instead."
		return plan
	}
	if request.MediaSourceID == "" {
		plan.ReasonCode = "source_required"
		plan.Reason = "Choose a media source before Xuva can build an adaptive route."
		return plan
	}
	if request.SourceBitrate <= 0 || request.Width <= 0 || request.Height <= 0 {
		plan.ReasonCode = "media_facts_required"
		plan.Reason = "Xuva needs bitrate and resolution from the media check before creating an adaptive ladder."
		return plan
	}
	if !isRemoteLike(request.RouteType) && request.MaxNetworkBitrate <= 0 {
		plan.ReasonCode = "local_route"
		plan.Reason = "Adaptive streaming is reserved for remote or constrained routes; local playback should use the cheapest direct path."
		return plan
	}
	if request.MaxNetworkBitrate > 0 && request.SourceBitrate <= request.MaxNetworkBitrate {
		plan.ReasonCode = "source_within_limit"
		plan.Reason = "The source bitrate is already within the selected network limit."
		return plan
	}
	plan.Variants = selectVariants(request)
	if len(plan.Variants) == 0 {
		plan.ReasonCode = "no_variant_fit"
		plan.Reason = "No adaptive variant fits this source and network target."
		return plan
	}
	plan.Enabled = true
	plan.ReasonCode = "adaptive_remote_route"
	plan.Reason = "Xuva can use an adaptive HLS ladder so remote playback can step down before buffering stalls."
	return plan
}

func MasterPlaylist(plan Plan) string {
	var builder strings.Builder
	builder.WriteString("#EXTM3U\n#EXT-X-VERSION:7\n#EXT-X-INDEPENDENT-SEGMENTS\n")
	for _, variant := range plan.Variants {
		builder.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,AVERAGE-BANDWIDTH=%d,RESOLUTION=%dx%d,CODECS=\"%s,%s\"\n",
			variant.Bitrate, int64(float64(variant.Bitrate)*0.88), variant.Width, variant.Height, variant.VideoCodec, variant.AudioCodec))
		builder.WriteString(variant.PlaylistURL + "\n")
	}
	return builder.String()
}

func MediaPlaylist(plan Plan, variantID string) (string, bool) {
	var selected Variant
	for _, variant := range plan.Variants {
		if variant.ID == variantID {
			selected = variant
			break
		}
	}
	if selected.ID == "" {
		return "", false
	}
	return strings.Join([]string{
		"#EXTM3U",
		"#EXT-X-VERSION:7",
		fmt.Sprintf("#EXT-X-TARGETDURATION:%d", plan.SegmentSeconds),
		"#EXT-X-PLAYLIST-TYPE:EVENT",
		"#EXT-X-MEDIA-SEQUENCE:0",
		"#EXT-X-DISCONTINUITY",
		fmt.Sprintf("# Xuva adaptive variant %s, %dp, %d bps", selected.ID, selected.Height, selected.Bitrate),
		"",
	}, "\n"), true
}

func NormalizeTelemetry(input Telemetry, correlationID string) Telemetry {
	input.Event = normalizeEvent(input.Event)
	input.CorrelationID = correlationID
	input.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if input.Event == "" {
		input.Event = "adaptation.sample"
	}
	return input
}

func selectVariants(request Request) []Variant {
	limit := request.SourceBitrate
	if request.MaxNetworkBitrate > 0 && request.MaxNetworkBitrate < limit {
		limit = request.MaxNetworkBitrate
	}
	profiles := []Variant{
		{ID: "2160p", Name: "4K Adaptive", Width: 3840, Height: 2160, Bitrate: 35_000_000, VideoCodec: "avc1.640033", AudioCodec: "mp4a.40.2"},
		{ID: "1440p", Name: "1440p Adaptive", Width: 2560, Height: 1440, Bitrate: 20_000_000, VideoCodec: "avc1.640032", AudioCodec: "mp4a.40.2"},
		{ID: "1080p", Name: "1080p Adaptive", Width: 1920, Height: 1080, Bitrate: 8_000_000, VideoCodec: "avc1.640028", AudioCodec: "mp4a.40.2"},
		{ID: "720p", Name: "720p Adaptive", Width: 1280, Height: 720, Bitrate: 4_000_000, VideoCodec: "avc1.4d401f", AudioCodec: "mp4a.40.2"},
		{ID: "480p", Name: "480p Adaptive", Width: 854, Height: 480, Bitrate: 1_500_000, VideoCodec: "avc1.4d401e", AudioCodec: "mp4a.40.2"},
	}
	var variants []Variant
	for _, profile := range profiles {
		if profile.Height > request.Height || profile.Width > request.Width || profile.Bitrate > limit {
			continue
		}
		profile.PlaylistURL = "variant-" + profile.ID + ".m3u8"
		variants = append(variants, profile)
	}
	if len(variants) == 0 {
		for _, profile := range profiles {
			if profile.Height <= request.Height && profile.Width <= request.Width {
				profile.PlaylistURL = "variant-" + profile.ID + ".m3u8"
				variants = append(variants, profile)
				break
			}
		}
	}
	sort.Slice(variants, func(i, j int) bool {
		return variants[i].Bitrate > variants[j].Bitrate
	})
	return variants
}

func isRemoteLike(route string) bool {
	route = strings.ToLower(strings.TrimSpace(route))
	return route == "remote" || route == "wan" || route == "internet" || route == "constrained"
}

func normalizeEvent(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "startup", "variant_selected", "variant_changed", "stall", "recovered", "ended":
		return "adaptive." + value
	case "adaptive.startup", "adaptive.variant_selected", "adaptive.variant_changed", "adaptive.stall", "adaptive.recovered", "adaptive.ended":
		return value
	default:
		return ""
	}
}
