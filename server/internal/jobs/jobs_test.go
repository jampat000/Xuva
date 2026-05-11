package jobs

import (
	"testing"

	"github.com/jampat000/Lorivo/server/internal/resources"
)

func TestRegistryUsesSeparateQueues(t *testing.T) {
	manager := resources.NewManager(resources.Limits{
		ScanWorkers:      1,
		ProbeWorkers:     2,
		TranscodeWorkers: 1,
	})
	registry := NewRegistry(manager)

	if registry.Scan == registry.Probe || registry.Scan == registry.Transcode || registry.Probe == registry.Transcode {
		t.Fatal("expected separate scan, probe, and transcode queues")
	}
	if registry.Scan.Class != resources.Background {
		t.Fatalf("expected scan queue to be background, got %q", registry.Scan.Class)
	}
	if registry.Probe.Class != resources.Background {
		t.Fatalf("expected probe queue to be background, got %q", registry.Probe.Class)
	}
	if registry.Transcode.Class != resources.PlaybackCritical {
		t.Fatalf("expected transcode queue to be playback critical, got %q", registry.Transcode.Class)
	}
}

func TestQueueNormalizesInvalidWorkerCount(t *testing.T) {
	queue := NewQueue("scan", resources.Background, 0)

	if queue.Workers != 1 {
		t.Fatalf("expected workers to normalize to 1, got %d", queue.Workers)
	}
}
