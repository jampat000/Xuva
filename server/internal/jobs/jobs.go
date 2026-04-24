package jobs

import (
	"context"
	"errors"

	"github.com/vyrdenhq/vyrden/server/internal/resources"
)

type Job func(context.Context)

type Queue struct {
	Name    string                  `json:"name"`
	Class   resources.WorkloadClass `json:"class"`
	Workers int                     `json:"workers"`
	jobs    chan Job
}

type Registry struct {
	Scan      *Queue
	Probe     *Queue
	Transcode *Queue
}

func NewRegistry(manager *resources.Manager) *Registry {
	limits := manager.Limits()
	return &Registry{
		Scan:      NewQueue("scan", resources.Background, limits.ScanWorkers),
		Probe:     NewQueue("probe", resources.Background, limits.ProbeWorkers),
		Transcode: NewQueue("transcode", resources.PlaybackCritical, limits.TranscodeWorkers),
	}
}

func NewQueue(name string, class resources.WorkloadClass, workers int) *Queue {
	if workers < 1 {
		workers = 1
	}
	return &Queue{
		Name:    name,
		Class:   class,
		Workers: workers,
		jobs:    make(chan Job, workers*4),
	}
}

func (q *Queue) Submit(ctx context.Context, job Job) error {
	if job == nil {
		return errors.New("nil job")
	}
	select {
	case q.jobs <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Registry) Snapshot() []map[string]any {
	queues := []*Queue{r.Scan, r.Probe, r.Transcode}
	output := make([]map[string]any, 0, len(queues))
	for _, queue := range queues {
		output = append(output, map[string]any{
			"name":    queue.Name,
			"class":   queue.Class,
			"workers": queue.Workers,
			"queued":  len(queue.jobs),
		})
	}
	return output
}
