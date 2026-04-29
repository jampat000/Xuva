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
