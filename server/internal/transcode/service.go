package transcode

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/events"
	"github.com/vyrdenhq/vyrden/server/internal/jobs"
	runtimestore "github.com/vyrdenhq/vyrden/server/internal/runtime"
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
	StatusTimeout   Status = "failed-timeout"
	StatusCanceled  Status = "cancelled"
)

type Request struct {
	MediaSourceID  string `json:"mediaSourceId"`
	Mode           Mode   `json:"mode"`
	SourcePath     string `json:"sourcePath,omitempty"`
	Acceleration   string `json:"acceleration,omitempty"`
	VideoEncoder   string `json:"videoEncoder,omitempty"`
	TimeoutSeconds int    `json:"timeoutSeconds,omitempty"`
	MaxAttempts    int    `json:"maxAttempts,omitempty"`
}

type Job struct {
	ID            string    `json:"id"`
	MediaSourceID string    `json:"mediaSourceId"`
	Mode          Mode      `json:"mode"`
	Status        Status    `json:"status"`
	SourcePath    string    `json:"sourcePath,omitempty"`
	OutputPath    string    `json:"outputPath,omitempty"`
	Command       []string  `json:"command,omitempty"`
	Acceleration  string    `json:"acceleration,omitempty"`
	VideoEncoder  string    `json:"videoEncoder,omitempty"`
	Attempts      int       `json:"attempts"`
	MaxAttempts   int       `json:"maxAttempts"`
	Timeout       string    `json:"timeout,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	StartedAt     time.Time `json:"startedAt,omitempty"`
	CompletedAt   time.Time `json:"completedAt,omitempty"`
	Error         string    `json:"error,omitempty"`
	FailureClass  string    `json:"failureClass,omitempty"`
	ReasonCode    string    `json:"reasonCode,omitempty"`
	Remediation   string    `json:"remediation,omitempty"`
}

type Service struct {
	events  *events.Bus
	runtime *runtimestore.Store
	queue   *jobs.Queue
	ffmpeg  string
	outDir  string
	nextID  atomic.Uint64
	mu      sync.RWMutex
	jobs    map[string]Job
	cancels map[string]context.CancelFunc
}

func NewService(eventBus *events.Bus, queue *jobs.Queue, ffmpegPath string, outputDir string) *Service {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	if outputDir == "" {
		outputDir = filepath.Join("data", "transcode")
	}
	return &Service{events: eventBus, queue: queue, ffmpeg: ffmpegPath, outDir: outputDir, jobs: map[string]Job{}, cancels: map[string]context.CancelFunc{}}
}

func NewPersistentService(ctx context.Context, eventBus *events.Bus, queue *jobs.Queue, ffmpegPath string, outputDir string, store *runtimestore.Store) (*Service, error) {
	service := NewService(eventBus, queue, ffmpegPath, outputDir)
	service.runtime = store
	if err := service.recover(ctx); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *Service) Start(ctx context.Context, request Request) (Job, error) {
	if request.MediaSourceID == "" {
		return Job{}, errors.New("media source id is required")
	}
	if request.Mode == "" {
		request.Mode = ModeRemux
	}
	maxAttempts := request.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 2
	}
	job := Job{ID: s.nextJobID(), MediaSourceID: request.MediaSourceID, Mode: request.Mode, SourcePath: request.SourcePath, Acceleration: request.Acceleration, VideoEncoder: request.VideoEncoder, Status: StatusQueued, CreatedAt: time.Now().UTC(), MaxAttempts: maxAttempts}
	if request.TimeoutSeconds > 0 {
		job.Timeout = (time.Duration(request.TimeoutSeconds) * time.Second).String()
	}
	job.OutputPath = filepath.Join(s.outDir, job.ID+".mp4")
	job.Command = s.command(job)
	s.store(job)
	_ = s.persist(context.Background(), job)
	s.publish("transcode.queued", job)
	if err := s.queue.Submit(ctx, func(workerCtx context.Context) {
		job, _ := s.Get(job.ID)
		job.Status = StatusRunning
		job.StartedAt = time.Now().UTC()
		s.store(job)
		_ = s.persist(context.Background(), job)
		s.publish("transcode.started", job)
		ctx := workerCtx
		cancel := func() {}
		if request.TimeoutSeconds > 0 {
			ctx, cancel = context.WithTimeout(workerCtx, time.Duration(request.TimeoutSeconds)*time.Second)
		} else {
			ctx, cancel = context.WithCancel(workerCtx)
		}
		s.storeCancel(job.ID, cancel)
		defer cancel()
		defer s.clearCancel(job.ID)
		result := s.executeWithRetry(ctx, job)
		if result.Err != nil {
			job.Status = statusForFailure(result.Failure)
			job.Error = result.Err.Error()
			job.FailureClass = result.Failure.Class
			job.ReasonCode = result.Failure.ReasonCode
			job.Remediation = result.Failure.Remediation
			job.Attempts = result.Attempts
			_ = os.Remove(job.OutputPath)
		} else {
			job.Status = StatusCompleted
			job.Attempts = result.Attempts
		}
		job.CompletedAt = time.Now().UTC()
		s.store(job)
		_ = s.persist(context.Background(), job)
		if job.Status == StatusCompleted {
			s.publish("transcode.completed", job)
		} else {
			s.publish("transcode.failed", job)
		}
	}); err != nil {
		job.Status = StatusFailed
		job.CompletedAt = time.Now().UTC()
		job.FailureClass = "queue_rejected"
		job.ReasonCode = "transcode_queue_rejected"
		job.Error = err.Error()
		job.Remediation = "Retry after current playback work has capacity."
		s.store(job)
		_ = s.persist(context.Background(), job)
		return Job{}, err
	}
	return job, nil
}

type executeResult struct {
	Attempts int
	Failure  Failure
	Err      error
}

func (s *Service) executeWithRetry(ctx context.Context, job Job) executeResult {
	maxAttempts := job.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	var last Failure
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		job.Attempts = attempt
		s.store(job)
		_ = s.persist(context.Background(), job)
		err := s.execute(ctx, job)
		if err == nil {
			return executeResult{Attempts: attempt}
		}
		last = ClassifyFailure(err.Error(), ctx.Err())
		lastErr = err
		if !last.Retryable || attempt == maxAttempts {
			break
		}
		select {
		case <-ctx.Done():
			last = ClassifyFailure("", ctx.Err())
			lastErr = ctx.Err()
			return executeResult{Attempts: attempt, Failure: last, Err: lastErr}
		case <-time.After(time.Duration(attempt) * 100 * time.Millisecond):
		}
	}
	return executeResult{Attempts: maxAttempts, Failure: last, Err: lastErr}
}

func (s *Service) execute(ctx context.Context, job Job) error {
	if job.SourcePath == "" {
		return errors.New("source path is required")
	}
	if err := os.MkdirAll(filepath.Dir(job.OutputPath), 0o755); err != nil {
		return err
	}
	args := job.Command[1:]
	cmd := exec.CommandContext(ctx, s.ffmpeg, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if len(output) > 0 {
			return errors.New(string(output))
		}
		return err
	}
	return nil
}

func (s *Service) Cancel(id string) (Job, bool) {
	job, ok := s.Get(id)
	if !ok {
		return Job{}, false
	}
	if job.Status != StatusQueued && job.Status != StatusRunning {
		return job, true
	}
	if cancel := s.cancelFor(id); cancel != nil {
		cancel()
	}
	job.Status = StatusCanceled
	job.CompletedAt = time.Now().UTC()
	failure := ClassifyFailure("", context.Canceled)
	job.FailureClass = failure.Class
	job.ReasonCode = failure.ReasonCode
	job.Remediation = failure.Remediation
	job.Error = failure.ReasonText
	_ = os.Remove(job.OutputPath)
	s.store(job)
	_ = s.persist(context.Background(), job)
	s.publish("transcode.cancelled", job)
	return job, true
}

func (s *Service) Cleanup(ctx context.Context, terminalRetention time.Duration) (int64, error) {
	if terminalRetention <= 0 {
		return 0, nil
	}
	return s.runtime.CleanupTerminal(ctx, "transcode", time.Now().UTC().Add(-terminalRetention), string(StatusCompleted), string(StatusFailed), string(StatusTimeout), string(StatusCanceled))
}

func (s *Service) command(job Job) []string {
	args := []string{s.ffmpeg, "-y", "-i", job.SourcePath}
	switch job.Mode {
	case ModeTranscode:
		args = append(args, "-map", "0:v:0", "-map", "0:a?")
		if job.Acceleration == "hardware" && job.VideoEncoder != "" {
			args = append(args, "-c:v", job.VideoEncoder, "-b:v", "8000k", "-maxrate", "9000k", "-bufsize", "18000k")
		} else {
			args = append(args, "-c:v", "libx264", "-preset", "veryfast", "-crf", "20")
		}
		args = append(args, "-c:a", "aac", "-movflags", "+faststart", job.OutputPath)
	default:
		args = append(args, "-map", "0", "-c", "copy", "-movflags", "+faststart", job.OutputPath)
	}
	return args
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

func (s *Service) FindCompleted(mediaSourceID string, mode Mode) (Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, job := range s.jobs {
		if job.MediaSourceID == mediaSourceID && job.Mode == mode && job.Status == StatusCompleted && job.OutputPath != "" {
			if _, err := os.Stat(job.OutputPath); err == nil {
				return job, true
			}
		}
	}
	return Job{}, false
}

func (s *Service) FindActive(mediaSourceID string, mode Mode) (Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, job := range s.jobs {
		if job.MediaSourceID == mediaSourceID && job.Mode == mode && (job.Status == StatusQueued || job.Status == StatusRunning) {
			return job, true
		}
	}
	return Job{}, false
}

func (s *Service) store(job Job) { s.mu.Lock(); defer s.mu.Unlock(); s.jobs[job.ID] = job }
func (s *Service) recover(ctx context.Context) error {
	entities, err := s.runtime.List(ctx, "transcode")
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, entity := range entities {
		var job Job
		if err := json.Unmarshal([]byte(entity.PayloadJSON), &job); err != nil {
			continue
		}
		switch job.Status {
		case StatusQueued, StatusRunning:
			job.Status = StatusFailed
			job.CompletedAt = now
			job.FailureClass = "recovered_after_restart"
			job.ReasonCode = "transcode_recovered_after_restart"
			job.Remediation = "Start the conversion again if this playback still needs it."
			job.Error = "Transcode was interrupted by server shutdown or restart."
			s.store(job)
			_ = s.persist(ctx, job)
		default:
			s.store(job)
		}
	}
	return nil
}

func (s *Service) persist(ctx context.Context, job Job) error {
	if s.runtime == nil {
		return nil
	}
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return s.runtime.Save(ctx, runtimestore.Entity{
		Type:        "transcode",
		ID:          job.ID,
		Status:      string(job.Status),
		PayloadJSON: string(payload),
		CreatedAt:   job.CreatedAt,
		UpdatedAt:   updatedAt(job),
		HeartbeatAt: updatedAt(job),
		CompletedAt: job.CompletedAt,
	})
}
func (s *Service) storeCancel(id string, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancels[id] = cancel
}
func (s *Service) clearCancel(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cancels, id)
}
func (s *Service) cancelFor(id string) context.CancelFunc {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cancels[id]
}
func (s *Service) publish(eventType string, job Job) {
	if s.events != nil {
		s.events.Publish(eventType, job)
	}
}
func (s *Service) nextJobID() string {
	return "work_" + time.Now().UTC().Format("20060102T150405") + "_" + stringID(s.nextID.Add(1))
}

func updatedAt(job Job) time.Time {
	if !job.CompletedAt.IsZero() {
		return job.CompletedAt
	}
	if !job.StartedAt.IsZero() {
		return job.StartedAt
	}
	return job.CreatedAt
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
