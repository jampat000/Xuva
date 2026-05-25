package transcode

import (
	"testing"
	"time"
)

// TestReaperCancelsStaleJob verifies the orphan-ffmpeg safety net: a job
// whose LastAccessedAt is older than the idle timeout gets cancelled when
// ReapIdleJobs runs. This is what stops ffmpeg from burning CPU when a
// user navigates away from the "Preparing…" panel.
func TestReaperCancelsStaleJob(t *testing.T) {
	service := newTestService(t, "ffmpeg")

	// Seed a fake job directly so we don't actually start ffmpeg. The reaper
	// only inspects status + LastAccessedAt — the job's other fields don't
	// matter for this test.
	stale := Job{
		ID:             "work-stale",
		MediaSourceID:  "source-1",
		Mode:           ModeRemuxAudio,
		Status:         StatusRunning,
		CreatedAt:      time.Now().UTC().Add(-5 * time.Minute),
		LastAccessedAt: time.Now().UTC().Add(-5 * time.Minute),
	}
	fresh := Job{
		ID:             "work-fresh",
		MediaSourceID:  "source-2",
		Mode:           ModeRemuxAudio,
		Status:         StatusRunning,
		CreatedAt:      time.Now().UTC(),
		LastAccessedAt: time.Now().UTC(),
	}
	service.store(stale)
	service.store(fresh)

	cancelled := service.ReapIdleJobs(60 * time.Second)
	if len(cancelled) != 1 || cancelled[0] != "work-stale" {
		t.Fatalf("expected reaper to cancel only the stale job, got %v", cancelled)
	}

	// FindActive on the stale source returns nothing (cancelled job is no
	// longer in StatusRunning/StatusQueued).
	if _, ok := service.FindActive("source-1", ModeRemuxAudio, 0); ok {
		t.Fatalf("stale job should have been cancelled, but FindActive returned it")
	}
	// Fresh job survives.
	if _, ok := service.FindActive("source-2", ModeRemuxAudio, 0); !ok {
		t.Fatalf("fresh job should survive reaping")
	}
}

// TestReaperIgnoresCompletedJobs makes sure we don't try to cancel a job
// that's already finished — Cancel on a completed job would be a no-op
// here, but adding state to track it would be wasted work.
func TestReaperIgnoresCompletedJobs(t *testing.T) {
	service := newTestService(t, "ffmpeg")
	done := Job{
		ID:             "work-done",
		MediaSourceID:  "source-3",
		Mode:           ModeRemux,
		Status:         StatusCompleted,
		CreatedAt:      time.Now().UTC().Add(-10 * time.Minute),
		LastAccessedAt: time.Now().UTC().Add(-10 * time.Minute),
	}
	service.store(done)
	cancelled := service.ReapIdleJobs(60 * time.Second)
	if len(cancelled) != 0 {
		t.Fatalf("expected reaper to skip completed jobs, got %v", cancelled)
	}
}

// TestFindActiveTouchesLastAccessedAt confirms the polling pathway:
// every time a client polls /api/playback/route and matches an active
// job, that job's LastAccessedAt bumps forward, keeping the reaper at bay.
func TestFindActiveTouchesLastAccessedAt(t *testing.T) {
	service := newTestService(t, "ffmpeg")
	stale := Job{
		ID:             "work-pollable",
		MediaSourceID:  "source-4",
		Mode:           ModeRemuxAudio,
		Status:         StatusRunning,
		CreatedAt:      time.Now().UTC().Add(-5 * time.Minute),
		LastAccessedAt: time.Now().UTC().Add(-5 * time.Minute),
	}
	service.store(stale)

	// Simulate a poll — should bump LastAccessedAt to ~now.
	if _, ok := service.FindActive("source-4", ModeRemuxAudio, 0); !ok {
		t.Fatalf("FindActive should find the job")
	}
	got, _ := service.Get("work-pollable")
	if time.Since(got.LastAccessedAt) > time.Second {
		t.Fatalf("FindActive should bump LastAccessedAt to ~now, got %v ago", time.Since(got.LastAccessedAt))
	}

	// Reaper should leave it alone now that it's freshly touched.
	cancelled := service.ReapIdleJobs(60 * time.Second)
	if len(cancelled) != 0 {
		t.Fatalf("freshly-polled job should survive reaper, got %v", cancelled)
	}
}
