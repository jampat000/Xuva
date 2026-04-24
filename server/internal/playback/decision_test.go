package playback

import (
	"context"
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
}

func TestDecideSourceDefersUntilProbe(t *testing.T) {
	decision := NewService().DecideSource(context.Background(), Request{
		MediaSourceID: "source-1",
		ClientProfile: "web",
	}, SourceFacts{MediaSourceID: "source-1"})

	if decision.Mode != DecisionDeferred {
		t.Fatalf("expected deferred decision, got %q", decision.Mode)
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
}
