package probes

import (
	"context"
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/catalog"
	"github.com/vyrdenhq/vyrden/server/internal/events"
	"github.com/vyrdenhq/vyrden/server/internal/jobs"
	"github.com/vyrdenhq/vyrden/server/internal/probe"
)

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

type Request struct {
	MediaSourceID string `json:"mediaSourceId,omitempty"`
	Limit         int    `json:"limit,omitempty"`
}

type Job struct {
	ID            string    `json:"id"`
	Status        Status    `json:"status"`
	MediaSourceID string    `json:"mediaSourceId,omitempty"`
	Limit         int       `json:"limit"`
	CreatedAt     time.Time `json:"createdAt"`
	StartedAt     time.Time `json:"startedAt,omitempty"`
	CompletedAt   time.Time `json:"completedAt,omitempty"`
	Total         int       `json:"total"`
	Completed     int       `json:"completed"`
	Failed        int       `json:"failed"`
	LastPath      string    `json:"lastPath,omitempty"`
	Error         string    `json:"error,omitempty"`
}

type Service struct {
	events  *events.Bus
	queue   *jobs.Queue
	catalog *catalog.Service
	probe   *probe.Service

	nextID atomic.Uint64
	mu     sync.RWMutex
	jobs   map[string]Job
}

func NewService(eventBus *events.Bus, queue *jobs.Queue, catalogService *catalog.Service, probeService *probe.Service) *Service {
	return &Service{events: eventBus, queue: queue, catalog: catalogService, probe: probeService, jobs: map[string]Job{}}
}

func (s *Service) Start(ctx context.Context, request Request) (Job, error) {
	if request.Limit <= 0 || request.Limit > 500 {
		request.Limit = 50
	}
	if request.MediaSourceID == "" && request.Limit < 1 {
		return Job{}, errors.New("probe limit must be positive")
	}
	job := Job{ID: s.nextJobID(), Status: StatusQueued, MediaSourceID: request.MediaSourceID, Limit: request.Limit, CreatedAt: time.Now().UTC()}
	s.store(job)
	s.publish("probe.queued", job)
	if err := s.queue.Submit(ctx, func(workerCtx context.Context) { s.run(workerCtx, job.ID, request) }); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()
		job.CompletedAt = time.Now().UTC()
		s.store(job)
		s.publish("probe.failed", job)
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

func (s *Service) run(ctx context.Context, id string, request Request) {
	job, _ := s.Get(id)
	job.Status = StatusRunning
	job.StartedAt = time.Now().UTC()
	s.store(job)
	s.publish("probe.started", job)

	var sources []catalog.MediaSourceItem
	var err error
	if request.MediaSourceID != "" {
		var source catalog.MediaSourceItem
		var ok bool
		source, ok, err = s.catalog.GetMediaSource(ctx, request.MediaSourceID)
		if ok {
			sources = []catalog.MediaSourceItem{source}
		}
	} else {
		sources, err = s.catalog.ListMediaSources(ctx, request.Limit, true)
	}
	if err != nil {
		s.fail(id, err)
		return
	}
	job.Total = len(sources)
	s.store(job)
	for _, source := range sources {
		if err := ctx.Err(); err != nil {
			s.fail(id, err)
			return
		}
		job, _ = s.Get(id)
		job.LastPath = source.RelPath
		s.store(job)
		result, err := s.probe.Probe(ctx, source.Path)
		if err != nil {
			job.Failed++
		} else {
			_ = s.catalog.SaveProbe(ctx, source.ID, catalog.ProbeResult{Container: result.Container, DurationSeconds: result.DurationSeconds, Bitrate: result.Bitrate, VideoCodec: result.VideoCodec, Width: result.Width, Height: result.Height, AudioStreams: result.AudioStreams, SubtitleStreams: result.SubtitleStreams, RawJSON: result.RawJSON})
			job.Completed++
		}
		s.store(job)
		s.publish("probe.progress", job)
	}
	job.Status = StatusCompleted
	job.CompletedAt = time.Now().UTC()
	s.store(job)
	s.publish("probe.completed", job)
}

func (s *Service) fail(id string, err error) {
	job, ok := s.Get(id)
	if !ok {
		return
	}
	job.Status = StatusFailed
	job.Error = err.Error()
	job.CompletedAt = time.Now().UTC()
	s.store(job)
	s.publish("probe.failed", job)
}

func (s *Service) store(job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
}

func (s *Service) publish(eventType string, job Job) { s.events.Publish(eventType, job) }

func (s *Service) nextJobID() string {
	return "probe_" + time.Now().UTC().Format("20060102T150405") + "_" + stringID(s.nextID.Add(1))
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
