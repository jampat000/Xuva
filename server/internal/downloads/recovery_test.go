package downloads

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jampat000/Lorivo/server/internal/database"
	"github.com/jampat000/Lorivo/server/internal/events"
	"github.com/jampat000/Lorivo/server/internal/jobs"
	"github.com/jampat000/Lorivo/server/internal/resources"
	runtimestore "github.com/jampat000/Lorivo/server/internal/runtime"
)

func TestPersistentServiceFailsInterruptedDownloadJob(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	store := runtimestore.NewStore(db)
	created := time.Now().UTC().Add(-time.Minute)
	job := Job{ID: "download-interrupted", MediaSourceID: "source-1", TargetProfile: ProfileBalanced, Status: StatusRunning, CreatedAt: created, StartedAt: created}
	payload, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("marshal job: %v", err)
	}
	if err := store.Save(ctx, runtimestore.Entity{Type: "download", ID: job.ID, Status: string(job.Status), PayloadJSON: string(payload), CreatedAt: created, UpdatedAt: created, HeartbeatAt: created}); err != nil {
		t.Fatalf("seed job: %v", err)
	}

	queue := jobs.NewQueue("download-test", resources.Background, 1)
	service, err := NewPersistentService(ctx, events.NewBus(8), queue, "ffmpeg", t.TempDir(), store)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	recovered, ok := service.Get(job.ID)
	if !ok {
		t.Fatal("expected interrupted download to remain visible after recovery")
	}
	if recovered.Status != StatusFailed || recovered.Error == "" {
		t.Fatalf("expected failed recovered download, got %#v", recovered)
	}
}
