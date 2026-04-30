package adaptive

import (
	"strings"
	"testing"
)

func TestBuildPlanChoosesAdaptiveForConstrainedRemoteRoute(t *testing.T) {
	plan := BuildPlan(Request{
		MediaSourceID:     "source_1",
		ClientProfile:     "web",
		RouteType:         "remote",
		SourceBitrate:     61_000_000,
		MaxNetworkBitrate: 10_000_000,
		Width:             3840,
		Height:            2160,
		SupportsHLS:       true,
	})
	if !plan.Enabled || plan.ReasonCode != "adaptive_remote_route" {
		t.Fatalf("expected adaptive plan, got %#v", plan)
	}
	if len(plan.Variants) < 2 {
		t.Fatalf("expected bitrate ladder, got %#v", plan.Variants)
	}
	if plan.Variants[0].Bitrate > 10_000_000 {
		t.Fatalf("expected variants under network limit, got %#v", plan.Variants[0])
	}
}

func TestBuildPlanFallsBackWhenClientDoesNotSupportHLS(t *testing.T) {
	plan := BuildPlan(Request{
		MediaSourceID:     "source_1",
		RouteType:         "remote",
		SourceBitrate:     61_000_000,
		MaxNetworkBitrate: 10_000_000,
		Width:             3840,
		Height:            2160,
	})
	if plan.Enabled || plan.ReasonCode != "client_no_hls" {
		t.Fatalf("expected unsupported client fallback, got %#v", plan)
	}
}

func TestPlaylistsAreGeneratedForSelectedVariants(t *testing.T) {
	plan := BuildPlan(Request{MediaSourceID: "source_1", RouteType: "remote", SourceBitrate: 40_000_000, MaxNetworkBitrate: 9_000_000, Width: 1920, Height: 1080, SupportsHLS: true})
	master := MasterPlaylist(plan)
	if !strings.Contains(master, "#EXT-X-STREAM-INF") || !strings.Contains(master, "variant-1080p.m3u8") {
		t.Fatalf("unexpected master playlist: %s", master)
	}
	media, ok := MediaPlaylist(plan, "720p")
	if !ok || !strings.Contains(media, "#EXT-X-TARGETDURATION") {
		t.Fatalf("unexpected media playlist: %s", media)
	}
}

func TestNormalizeTelemetryAddsCorrelationAndEventPrefix(t *testing.T) {
	event := NormalizeTelemetry(Telemetry{Event: "stall", StallMS: 1200}, "req_abc")
	if event.Event != "adaptive.stall" || event.CorrelationID != "req_abc" || event.CreatedAt == "" {
		t.Fatalf("unexpected telemetry: %#v", event)
	}
}
