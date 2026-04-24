package transcode

import (
	"context"
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/events"
	"github.com/vyrdenhq/vyrden/server/internal/jobs"
)

type Mode string
type Status string

const (
	ModeRemux     Mode = "remux"
	ModeTranscode Mode = "transcode"

	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

type Request struct {
	MediaSourceID string `json:"mediaSourceId"`
	Mode          Mode   `json:"mode"`
}

type Job struct {
	ID            string    `json:"id"`
	MediaSourceID string    `json:"mediaSourceId"`
	Mode          Mode      `json:"mode"`
	Status        Status    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	StartedAt     time.Time `json:"startedAt,omitempty"`
	CompletedAt   time.Time `json:"completedAt,omitempty"`
	Error         string    `json:"error,omitempty"`
}

type Service struct {
	events *events.Bus
	queue  *jobs.Queue
	nextID atomic.Uint64
	mu     sync.RWMutex
	jobs   map[string]Job
}

func NewService(eventBus *events.Bus, queue *jobs.Queue) *Service {
	return &Service{events: eventBus, queue: queue, jobs: map[string]Job{}}
}

func (s *Service) Start(ctx context.Context, request Request) (Job, error) {
	if request.MediaSourceID == "" {
		return Job{}, errors.New("media source id is required")
	}
	if request.Mode == "" {
		request.Mode = ModeRemux
	}
	job := Job{ID: s.nextJobID(), MediaSourceID: request.MediaSourceID, Mode: request.Mode, Status: StatusQueued, CreatedAt: time.Now().UTC()}
	s.store(job)
	s.publish("transcode.queued", job)
	if err := s.queue.Submit(ctx, func(workerCtx context.Context) {
		job, _ := s.Get(job.ID)
		job.Status = StatusRunning
		job.StartedAt = time.Now().UTC()
		s.store(job)
		s.publish("transcode.started", job)
		select {
		case <-workerCtx.Done():
			job.Status = StatusFailed
			job.Error = workerCtx.Err().Error()
		default:
			job.Status = StatusCompleted
		}
		job.CompletedAt = time.Now().UTC()
		s.store(job)
		s.publish("transcode.completed", job)
	}); err != nil {
		return Job{}, err
	}
	return job, nil
}

func (s *Service) Get(id string) (Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[id]
	return job, ok
}

func (s *Service) List() []Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	output := make([]Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		output = append(output, job)
	}
	sort.Slice(output, func(i, j int) bool { return output[i].CreatedAt.After(output[j].CreatedAt) })
	return output
}

func (s *Service) store(job Job)                     { s.mu.Lock(); defer s.mu.Unlock(); s.jobs[job.ID] = job }
func (s *Service) publish(eventType string, job Job) { s.events.Publish(eventType, job) }
func (s *Service) nextJobID() string {
	return "work_" + time.Now().UTC().Format("20060102T150405") + "_" + stringID(s.nextID.Add(1))
}

func stringID(value uint64) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	i := len(digits)
	for value > 0 {
		i--
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[i:])
}
