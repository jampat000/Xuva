package subtitles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverSidecars(t *testing.T) {
	dir := t.TempDir()
	media := filepath.Join(dir, "Movie.2024.mkv")
	if err := os.WriteFile(media, []byte("media"), 0o644); err != nil {
		t.Fatalf("write media: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Movie.2024.en.forced.srt"), []byte("sub"), 0o644); err != nil {
		t.Fatalf("write subtitle: %v", err)
	}

	sidecars := DiscoverSidecars(media)
	if len(sidecars) != 1 {
		t.Fatalf("expected one sidecar, got %#v", sidecars)
	}
	if sidecars[0].Format != "srt" || sidecars[0].Language != "en" || !sidecars[0].Forced {
		t.Fatalf("unexpected sidecar: %#v", sidecars[0])
	}
}

func TestPlanConversionForTextSubtitle(t *testing.T) {
	plan := NewService().PlanConversion(Sidecar{Format: "ass"}, "web")

	if plan.Status != "available" {
		t.Fatalf("expected conversion available, got %#v", plan)
	}
	if plan.OutputFormat != "webvtt" || plan.ServerImpact != "low" {
		t.Fatalf("expected WebVTT low-impact output, got %#v", plan)
	}
}

func TestPlanConversionForExistingWebVTT(t *testing.T) {
	plan := NewService().PlanConversion(Sidecar{Format: "vtt"}, "web")

	if plan.Status != "not_required" || plan.ServerImpact != "none" {
		t.Fatalf("expected direct WebVTT plan, got %#v", plan)
	}
}
