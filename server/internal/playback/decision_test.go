package playback

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestDecideSourceDirectPlayForBrowserCompatibleMedia(t *testing.T) {
	decision := NewService().DecideSource(context.Background(), Request{
		MediaSourceID: "source-1",
		ClientProfile: "web",
	}, SourceFacts{
		MediaSourceID: "source-1",
		Container:     "mov,mp4,m4a,3gp,3g2,mj2",
		VideoCodec:    "h264",
		Bitrate:       8000000,
		Probed:        true,
	})

	if decision.Mode != DirectPlay {
		t.Fatalf("expected direct play, got %q", decision.Mode)
	}
	if decision.VideoAction != "direct" {
		t.Fatalf("expected direct video action, got %q", decision.VideoAction)
	}
	if decision.ReasonCode != "direct_play_supported" || decision.ReasonText == "" || decision.DecisionTraceID == "" {
		t.Fatalf("expected v2 reason fields, got %#v", decision)
	}
}

func TestDecideSourceDefersUntilProbe(t *testing.T) {
	decision := NewService().DecideSource(context.Background(), Request{
		MediaSourceID: "source-1",
		ClientProfile: "web",
	}, SourceFacts{MediaSourceID: "source-1"})

	if decision.Mode != DecisionDeferred {
		t.Fatalf("expected deferred decision, got %q", decision.Mode)
	}
	if decision.ReasonCode != "probe_required" {
		t.Fatalf("expected probe_required reason, got %q", decision.ReasonCode)
	}
}

func TestDecideSourceTranscodesUnsupportedVideo(t *testing.T) {
	decision := NewService().DecideSource(context.Background(), Request{
		MediaSourceID: "source-1",
		ClientProfile: "web",
	}, SourceFacts{
		MediaSourceID: "source-1",
		Container:     "matroska,webm",
		VideoCodec:    "hevc",
		Probed:        true,
	})

	if decision.Mode != VideoTranscode {
		t.Fatalf("expected video transcode, got %q", decision.Mode)
	}
	if decision.ReasonCode != "video_conversion_required" {
		t.Fatalf("expected video conversion reason, got %q", decision.ReasonCode)
	}
}

func TestDecideSourceBurnsImageSubtitleForBrowser(t *testing.T) {
	decision := NewService().DecideSource(context.Background(), Request{
		MediaSourceID:       "source-1",
		ClientProfile:       "web",
		SubtitleTrackIndex:  3,
		SubtitleCodec:       "hdmv_pgs_subtitle",
		SubtitleTrackActive: true,
	}, SourceFacts{
		MediaSourceID:  "source-1",
		Container:      "mov,mp4,m4a,3gp,3g2,mj2",
		VideoCodec:     "h264",
		SubtitleCodec:  "hdmv_pgs_subtitle",
		SubtitleActive: true,
		Probed:         true,
	})

	if decision.Mode != SubtitleBurn {
		t.Fatalf("expected subtitle burn, got %q", decision.Mode)
	}
	if decision.SubtitleAction != "burn_in" {
		t.Fatalf("expected burn_in subtitle action, got %q", decision.SubtitleAction)
	}
	if decision.ReasonCode != "subtitle_burn_required" {
		t.Fatalf("expected subtitle reason code, got %q", decision.ReasonCode)
	}
}

func TestDecideSourceKeepsDirectPathForCompatibleTextSubtitle(t *testing.T) {
	decision := NewService().DecideSource(context.Background(), Request{
		MediaSourceID:       "source-1",
		ClientProfile:       "web",
		SubtitleTrackIndex:  2,
		SubtitleCodec:       "webvtt",
		SubtitleTrackActive: true,
		SubtitleCodecs:      []string{"webvtt", "srt"},
	}, SourceFacts{
		MediaSourceID:   "source-1",
		Container:       "mov,mp4,m4a,3gp,3g2,mj2",
		VideoCodec:      "h264",
		AudioCodec:      "aac",
		AudioChannels:   2,
		SubtitleCodec:   "webvtt",
		SubtitleActive:  true,
		SubtitleStreams: 1,
		Probed:          true,
	})

	if decision.Mode != DirectPlay {
		t.Fatalf("expected direct play with compatible text subtitle, got %q", decision.Mode)
	}
	if decision.SubtitleAction != "direct" || decision.SubtitleClass != "text" {
		t.Fatalf("expected direct text subtitle impact, got %#v", decision)
	}
	if decision.SubtitleImpact["serverLoad"] != "none" {
		t.Fatalf("expected no subtitle server load, got %#v", decision.SubtitleImpact)
	}
}

func TestDecideSourceReportsTextSubtitleConversionPath(t *testing.T) {
	decision := NewService().DecideSource(context.Background(), Request{
		MediaSourceID:       "source-1",
		ClientProfile:       "web",
		SubtitleTrackIndex:  2,
		SubtitleCodec:       "ass",
		SubtitleTrackActive: true,
		SubtitleCodecs:      []string{"webvtt", "srt"},
	}, SourceFacts{
		MediaSourceID:   "source-1",
		Container:       "mov,mp4,m4a,3gp,3g2,mj2",
		VideoCodec:      "h264",
		AudioCodec:      "aac",
		AudioChannels:   2,
		SubtitleCodec:   "ass",
		SubtitleActive:  true,
		SubtitleStreams: 1,
		Probed:          true,
	})

	if decision.Mode != DirectPlay {
		t.Fatalf("expected direct video path with text subtitle conversion, got %q", decision.Mode)
	}
	if decision.ReasonCode != "subtitle_conversion_available" || decision.SubtitleAction != "convert" {
		t.Fatalf("expected conversion reason/action, got %#v", decision)
	}
	if decision.SubtitleImpact["output"] != "webvtt sidecar" || decision.SubtitleImpact["serverLoad"] != "low" {
		t.Fatalf("expected conversion output behavior, got %#v", decision.SubtitleImpact)
	}
}

func TestDecideSourceSubtitleOffRemovesSubtitleImpact(t *testing.T) {
	off := NewService().DecideSource(context.Background(), Request{
		MediaSourceID: "source-1",
		ClientProfile: "web",
	}, SourceFacts{
		MediaSourceID: "source-1",
		Container:     "mov,mp4,m4a,3gp,3g2,mj2",
		VideoCodec:    "h264",
		AudioCodec:    "aac",
		AudioChannels: 2,
		Probed:        true,
	})
	on := NewService().DecideSource(context.Background(), Request{
		MediaSourceID:       "source-1",
		ClientProfile:       "web",
		SubtitleTrackActive: true,
		SubtitleCodec:       "hdmv_pgs_subtitle",
	}, SourceFacts{
		MediaSourceID:   "source-1",
		Container:       "mov,mp4,m4a,3gp,3g2,mj2",
		VideoCodec:      "h264",
		AudioCodec:      "aac",
		AudioChannels:   2,
		SubtitleCodec:   "hdmv_pgs_subtitle",
		SubtitleActive:  true,
		SubtitleStreams: 1,
		Probed:          true,
	})

	if off.Mode != DirectPlay || off.SubtitleAction != "none" {
		t.Fatalf("expected subtitles off to direct play, got %#v", off)
	}
	if on.Mode != SubtitleBurn || on.ReasonCode != "subtitle_burn_required" {
		t.Fatalf("expected subtitles on to change forecast to burn-in, got %#v", on)
	}
}

func TestDecideSourceAudioTranscodeForBrowserIncompatibleAudio(t *testing.T) {
	decision := NewService().DecideSource(context.Background(), Request{
		MediaSourceID:   "source-1",
		ClientProfile:   "web",
		AudioTrackIndex: 1,
		AudioCodec:      "truehd",
		AudioChannels:   8,
	}, SourceFacts{
		MediaSourceID: "source-1",
		Container:     "mov,mp4,m4a,3gp,3g2,mj2",
		VideoCodec:    "h264",
		AudioCodec:    "truehd",
		AudioChannels: 8,
		Probed:        true,
	})

	if decision.Mode != AudioTranscode {
		t.Fatalf("expected audio transcode, got %q", decision.Mode)
	}
	if decision.VideoAction != "direct" {
		t.Fatalf("expected direct video action, got %q", decision.VideoAction)
	}
	if decision.ReasonCode != "audio_conversion_required" {
		t.Fatalf("expected audio conversion reason, got %q", decision.ReasonCode)
	}
}

func TestDecideSourceRemuxesCompatibleVideoInBadContainer(t *testing.T) {
	decision := NewService().DecideSource(context.Background(), Request{
		MediaSourceID: "source-1",
		ClientProfile: "web",
	}, SourceFacts{
		MediaSourceID: "source-1",
		Container:     "matroska",
		VideoCodec:    "h264",
		AudioCodec:    "aac",
		AudioChannels: 2,
		Probed:        true,
	})

	if decision.Mode != Remux {
		t.Fatalf("expected remux, got %q", decision.Mode)
	}
	if decision.ContainerAction != "remux" || decision.VideoAction != "copy" {
		t.Fatalf("expected remux/copy actions, got %#v", decision)
	}
	if decision.ReasonCode != "container_remux_required" {
		t.Fatalf("expected container reason, got %q", decision.ReasonCode)
	}
}

func TestDecideSourceNetworkLimitChoosesVideoTranscode(t *testing.T) {
	decision := NewService().DecideSource(context.Background(), Request{
		MediaSourceID:     "source-1",
		ClientProfile:     "web",
		MaxNetworkBitrate: 10_000_000,
		RouteType:         "remote",
	}, SourceFacts{
		MediaSourceID: "source-1",
		Container:     "mov,mp4,m4a,3gp,3g2,mj2",
		VideoCodec:    "h264",
		AudioCodec:    "aac",
		AudioChannels: 2,
		Bitrate:       60_000_000,
		Probed:        true,
	})

	if decision.Mode != VideoTranscode {
		t.Fatalf("expected network-limited video transcode, got %q", decision.Mode)
	}
	if decision.ReasonCode != "network_bitrate_limit" {
		t.Fatalf("expected network reason, got %q", decision.ReasonCode)
	}
}

func TestDecideSourceNetworkLimitChoosesAdaptiveWhenSupported(t *testing.T) {
	decision := NewService().DecideSource(context.Background(), Request{
		MediaSourceID:     "source-1",
		ClientProfile:     "web",
		MaxNetworkBitrate: 10_000_000,
		RouteType:         "remote",
		SupportsAdaptive:  true,
		Containers:        []string{"mp4", "webm"},
		VideoCodecs:       []string{"h264", "vp9"},
		AudioCodecs:       []string{"aac"},
		SubtitleCodecs:    []string{"webvtt"},
	}, SourceFacts{
		MediaSourceID: "source-1",
		Container:     "matroska",
		VideoCodec:    "hevc",
		Width:         3840,
		Height:        2160,
		AudioCodec:    "aac",
		AudioChannels: 2,
		Bitrate:       61_000_000,
		Probed:        true,
	})

	if decision.Mode != AdaptiveStream {
		t.Fatalf("expected adaptive stream, got %#v", decision)
	}
	if decision.ReasonCode != "adaptive_remote_route" || decision.VideoAction != "adaptive" {
		t.Fatalf("expected adaptive reason/action, got %#v", decision)
	}
}

func TestDecideSourceIncompleteProbeFailsGracefully(t *testing.T) {
	decision := NewService().DecideSource(context.Background(), Request{
		MediaSourceID: "source-1",
		ClientProfile: "web",
	}, SourceFacts{
		MediaSourceID: "source-1",
		Container:     "mov,mp4,m4a,3gp,3g2,mj2",
		Probed:        true,
	})

	if decision.Mode != DecisionDeferred {
		t.Fatalf("expected deferred incomplete media facts, got %q", decision.Mode)
	}
	if decision.ReasonCode != "media_facts_incomplete" {
		t.Fatalf("expected incomplete reason, got %q", decision.ReasonCode)
	}
	if len(decision.SuggestedFixes) == 0 || !strings.Contains(decision.ReasonText, "container or video") {
		t.Fatalf("expected actionable incomplete output, got %#v", decision)
	}
}

func TestDecideSourceIsDeterministic(t *testing.T) {
	request := Request{
		MediaSourceID:       "source-1",
		ClientProfile:       "web",
		MaxNetworkBitrate:   10_000_000,
		AudioTrackIndex:     1,
		AudioCodec:          "truehd",
		AudioChannels:       8,
		SubtitleTrackIndex:  3,
		SubtitleCodec:       "srt",
		SubtitleTrackActive: true,
	}
	source := SourceFacts{
		MediaSourceID:   "source-1",
		Container:       "matroska,webm",
		VideoCodec:      "h264",
		AudioCodec:      "truehd",
		AudioChannels:   8,
		SubtitleCodec:   "srt",
		SubtitleActive:  true,
		Bitrate:         8_000_000,
		Probed:          true,
		Width:           1920,
		Height:          1080,
		SubtitleStreams: 1,
	}

	first := NewService().DecideSource(context.Background(), request, source)
	second := NewService().DecideSource(context.Background(), request, source)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected deterministic decisions:\nfirst=%#v\nsecond=%#v", first, second)
	}
	if first.DecisionTraceID == "" {
		t.Fatalf("expected decision trace id")
	}
}

// --- issue #56: per-file probe data drives the decision ---

// appleTVRequest is a small builder for the Apple TV 4K capability set used
// across the issue-#56 tests below: HEVC + 10-bit HDR + 60fps.
func appleTVRequest(maxBitDepth int, supportsHDR bool) Request {
	return Request{
		MediaSourceID:    "source-1",
		ClientProfile:    "apple-tv",
		Containers:       []string{"mp4", "mov", "m4v"},
		VideoCodecs:      []string{"h264", "hevc"},
		AudioCodecs:      []string{"aac", "ac3", "eac3", "alac"},
		SubtitleCodecs:   []string{"webvtt", "srt"},
		MaxVideoBitDepth: maxBitDepth,
		MaxFrameRate:     60,
		SupportsHDR:      supportsHDR,
	}
}

func TestDecideSourceHEVCMain8OnAppleTVDirectPlays(t *testing.T) {
	decision := NewService().DecideSource(context.Background(), appleTVRequest(10, true), SourceFacts{
		MediaSourceID: "source-1",
		Container:     "mp4",
		VideoCodec:    "hevc",
		VideoProfile:  "Main",
		VideoBitDepth: 8,
		Width:         1920,
		Height:        1080,
		FrameRate:     24,
		Bitrate:       12_000_000,
		Probed:        true,
	})

	if decision.Mode != DirectPlay {
		t.Fatalf("expected direct play for HEVC Main 8-bit on apple-tv, got %q (%s)", decision.Mode, decision.ReasonCode)
	}
	if decision.ReasonCode != "direct_play_supported" {
		t.Fatalf("expected direct_play_supported, got %q", decision.ReasonCode)
	}
}

func TestDecideSourceHEVCMain10OnAppleTVDirectPlays(t *testing.T) {
	decision := NewService().DecideSource(context.Background(), appleTVRequest(10, true), SourceFacts{
		MediaSourceID: "source-1",
		Container:     "mp4",
		VideoCodec:    "hevc",
		VideoProfile:  "Main 10",
		VideoBitDepth: 10,
		Width:         3840,
		Height:        2160,
		FrameRate:     24,
		Bitrate:       40_000_000,
		Probed:        true,
	})

	if decision.Mode != DirectPlay {
		t.Fatalf("expected direct play for HEVC Main10 on apple-tv, got %q (%s)", decision.Mode, decision.ReasonCode)
	}
	if decision.ReasonCode != "direct_play_supported" {
		t.Fatalf("expected direct_play_supported, got %q", decision.ReasonCode)
	}
}

func TestDecideSourceHEVCMain10OnWebRoutesToTranscode(t *testing.T) {
	// Web profile declares MaxVideoBitDepth=8. A 10-bit source must transcode
	// even when (hypothetically) the container/codec are otherwise direct-play
	// compatible. To isolate the bit-depth gate from the codec gate, we use
	// codec=h264 in mp4 and only flip the bit depth.
	decision := NewService().DecideSource(context.Background(), Request{
		MediaSourceID:    "source-1",
		ClientProfile:    "web",
		Containers:       []string{"mp4", "mov", "webm"},
		VideoCodecs:      []string{"h264", "av1", "vp9"},
		MaxVideoBitDepth: 8,
	}, SourceFacts{
		MediaSourceID: "source-1",
		Container:     "mp4",
		VideoCodec:    "h264",
		VideoProfile:  "High 10",
		VideoBitDepth: 10,
		Width:         1920,
		Height:        1080,
		Bitrate:       8_000_000,
		Probed:        true,
	})

	if decision.Mode != VideoTranscode {
		t.Fatalf("expected video transcode for 10-bit on 8-bit-max client, got %q (%s)", decision.Mode, decision.ReasonCode)
	}
	if decision.ReasonCode != "bit_depth_unsupported_transcode" {
		t.Fatalf("expected bit_depth_unsupported_transcode, got %q", decision.ReasonCode)
	}
}

func TestDecideSourceHDRPassThroughOnHDRClient(t *testing.T) {
	decision := NewService().DecideSource(context.Background(), appleTVRequest(10, true), SourceFacts{
		MediaSourceID: "source-1",
		Container:     "mp4",
		VideoCodec:    "hevc",
		VideoProfile:  "Main 10",
		VideoBitDepth: 10,
		HDR:           "HDR10",
		Width:         3840,
		Height:        2160,
		FrameRate:     24,
		Bitrate:       50_000_000,
		Probed:        true,
	})

	if decision.Mode != DirectPlay {
		t.Fatalf("expected HDR direct play on apple-tv, got %q (%s)", decision.Mode, decision.ReasonCode)
	}
	if decision.ReasonCode != "hdr_pass_through" {
		t.Fatalf("expected hdr_pass_through reason, got %q", decision.ReasonCode)
	}
}

func TestDecideSourceHDRToneMapsOnSDRClient(t *testing.T) {
	// Same source as the pass-through test, but the request has SupportsHDR=false.
	decision := NewService().DecideSource(context.Background(), appleTVRequest(10, false), SourceFacts{
		MediaSourceID: "source-1",
		Container:     "mp4",
		VideoCodec:    "hevc",
		VideoBitDepth: 10,
		HDR:           "HDR10",
		Width:         3840,
		Height:        2160,
		FrameRate:     24,
		Bitrate:       50_000_000,
		Probed:        true,
	})

	if decision.Mode != VideoTranscode {
		t.Fatalf("expected transcode for HDR on SDR client, got %q (%s)", decision.Mode, decision.ReasonCode)
	}
	if decision.ReasonCode != "hdr_tone_map_required" {
		t.Fatalf("expected hdr_tone_map_required, got %q", decision.ReasonCode)
	}
}

func TestDecideSourceFrameRateAboveClientMaxTranscodes(t *testing.T) {
	// 120fps source on a client that maxes out at 60.
	decision := NewService().DecideSource(context.Background(), Request{
		MediaSourceID:    "source-1",
		ClientProfile:    "apple-tv",
		Containers:       []string{"mp4", "mov", "m4v"},
		VideoCodecs:      []string{"h264", "hevc"},
		MaxVideoBitDepth: 10,
		MaxFrameRate:     60,
	}, SourceFacts{
		MediaSourceID: "source-1",
		Container:     "mp4",
		VideoCodec:    "h264",
		VideoBitDepth: 8,
		Width:         1920,
		Height:        1080,
		FrameRate:     120,
		Bitrate:       20_000_000,
		Probed:        true,
	})

	if decision.Mode != VideoTranscode {
		t.Fatalf("expected transcode for 120fps on 60fps-max client, got %q (%s)", decision.Mode, decision.ReasonCode)
	}
	if decision.ReasonCode != "framerate_above_client_max" {
		t.Fatalf("expected framerate_above_client_max, got %q", decision.ReasonCode)
	}
}

func TestDecideSourceFrameRateNoCapDoesNotTranscode(t *testing.T) {
	// A request with no FrameRate cap (MaxFrameRate == 0) must not force a
	// transcode just because the source has a high frame rate.
	decision := NewService().DecideSource(context.Background(), appleTVRequest(10, true), SourceFacts{
		MediaSourceID: "source-1",
		Container:     "mp4",
		VideoCodec:    "hevc",
		VideoBitDepth: 8,
		Width:         1920,
		Height:        1080,
		FrameRate:     60,
		Bitrate:       20_000_000,
		Probed:        true,
	})

	if decision.Mode != DirectPlay {
		t.Fatalf("expected direct play (no framerate cap exceeded), got %q (%s)", decision.Mode, decision.ReasonCode)
	}
}
