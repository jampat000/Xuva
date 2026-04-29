package probes

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/database"
	"github.com/vyrdenhq/vyrden/server/internal/events"
	"github.com/vyrdenhq/vyrden/server/internal/jobs"
	"github.com/vyrdenhq/vyrden/server/internal/resources"
	runtimestore "github.com/vyrdenhq/vyrden/server/internal/runtime"
)

func TestPersistentServiceFailsInterruptedProbeJob(t *testing.T) {
	ctx := context.Background()
	store, closeDB := testRuntimeStore(t)
	defer closeDB()
	created := time.Now().UTC().Add(-time.Minute)
	job := Job{ID: "probe-interrupted", Status: StatusRunning, Limit: 50, CreatedAt: created, StartedAt: created}
	payload, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("marshal job: %v", err)
	}
	if err := store.Save(ctx, runtimestore.Entity{Type: "probe", ID: job.ID, Status: string(job.Status), PayloadJSON: string(payload), CreatedAt: created, UpdatedAt: created, HeartbeatAt: created}); err != nil {
		t.Fatalf("seed job: %v", err)
	}

	queue := jobs.NewQueue("probe-test", resources.Background, 1)
	service, err := NewPersistentService(ctx, events.NewBus(8), queue, nil, nil, store)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	recovered, ok := service.Get(job.ID)
	if !ok {
		t.Fatal("expected interrupted probe to remain visible after recovery")
	}
	if recovered.Status != StatusFailed || recovered.Error == "" {
		t.Fatalf("expected failed recovered probe, got %#v", recovered)
	}
}

func testRuntimeStore(t *testing.T) (*runtimestore.Store, func()) {
	t.Helper()
	db, err := database.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return runtimestore.NewStore(db), func() { _ = db.Close() }
}
