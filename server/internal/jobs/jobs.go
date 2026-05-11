package jobs

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/jampat000/Lorivo/server/internal/resources"
)

type Job func(context.Context)

type Queue struct {
	Name    string                  `json:"name"`
	Class   resources.WorkloadClass `json:"class"`
	Workers int                     `json:"workers"`
	jobs    chan Job
	started sync.Once
	active  atomic.Int64
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

func (q *Queue) Start(ctx context.Context) {
	q.started.Do(func() {
		for i := 0; i < q.Workers; i++ {
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case job := <-q.jobs:
						q.active.Add(1)
						job(ctx)
						q.active.Add(-1)
					}
				}
			}()
		}
	})
}

func (r *Registry) Start(ctx context.Context) {
	r.Scan.Start(ctx)
	r.Probe.Start(ctx)
	r.Transcode.Start(ctx)
}

func (r *Registry) Snapshot() []map[string]any {
	queues := []*Queue{r.Scan, r.Probe, r.Transcode}
	output := make([]map[string]any, 0, len(queues))
	for _, queue := range queues {
		active := queue.active.Load()
		output = append(output, map[string]any{
			"name":              queue.Name,
			"class":             queue.Class,
			"workers":           queue.Workers,
			"queued":            len(queue.jobs),
			"active":            active,
			"workerUtilization": float64(active) / float64(queue.Workers),
		})
	}
	return output
}
