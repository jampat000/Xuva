package sessions

import (
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/events"
)

type Session struct {
	ID            string    `json:"id"`
	UserID        string    `json:"userId"`
	DeviceID      string    `json:"deviceId"`
	MediaSourceID string    `json:"mediaSourceId"`
	Title         string    `json:"title,omitempty"`
	ArtworkURL    string    `json:"artworkUrl,omitempty"`
	SourceName    string    `json:"sourceName,omitempty"`
	QualityLabel  string    `json:"qualityLabel,omitempty"`
	Container     string    `json:"container,omitempty"`
	VideoCodec    string    `json:"videoCodec,omitempty"`
	Bitrate       int64     `json:"bitrate,omitempty"`
	ClientProfile string    `json:"clientProfile,omitempty"`
	Route         string    `json:"route,omitempty"`
	ServerImpact  string    `json:"serverImpact,omitempty"`
	Mode          string    `json:"mode"`
	Status        string    `json:"status"`
	Progress      float64   `json:"progressSeconds"`
	Duration      float64   `json:"durationSeconds"`
	StartedAt     time.Time `json:"startedAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type StartRequest struct {
	UserID          string  `json:"userId"`
	DeviceID        string  `json:"deviceId"`
	MediaSourceID   string  `json:"mediaSourceId"`
	Title           string  `json:"title"`
	ArtworkURL      string  `json:"artworkUrl"`
	SourceName      string  `json:"sourceName"`
	QualityLabel    string  `json:"qualityLabel"`
	Container       string  `json:"container"`
	VideoCodec      string  `json:"videoCodec"`
	Bitrate         int64   `json:"bitrate"`
	ClientProfile   string  `json:"clientProfile"`
	Route           string  `json:"route"`
	ServerImpact    string  `json:"serverImpact"`
	Mode            string  `json:"mode"`
	ProgressSeconds float64 `json:"progressSeconds"`
	DurationSeconds float64 `json:"durationSeconds"`
}

type UpdateRequest struct {
	ProgressSeconds float64 `json:"progressSeconds"`
	DurationSeconds float64 `json:"durationSeconds"`
	Status          string  `json:"status"`
}

type Service struct {
	events *events.Bus
	nextID atomic.Uint64
	mu     sync.RWMutex
	items  map[string]Session
}

func NewService(eventBus *events.Bus) *Service {
	return &Service{events: eventBus, items: map[string]Session{}}
}

func (s *Service) Start(request StartRequest) (Session, error) {
	if request.MediaSourceID == "" {
		return Session{}, errors.New("media source id is required")
	}
	if request.UserID == "" {
		request.UserID = "local"
	}
	if request.DeviceID == "" {
		request.DeviceID = "web"
	}
	if request.Mode == "" {
		request.Mode = "direct"
	}
	now := time.Now().UTC()
	session := Session{
		ID:            s.nextSessionID(),
		UserID:        request.UserID,
		DeviceID:      request.DeviceID,
		MediaSourceID: request.MediaSourceID,
		Title:         request.Title,
		ArtworkURL:    request.ArtworkURL,
		SourceName:    request.SourceName,
		QualityLabel:  request.QualityLabel,
		Container:     request.Container,
		VideoCodec:    request.VideoCodec,
		Bitrate:       request.Bitrate,
		ClientProfile: request.ClientProfile,
		Route:         request.Route,
		ServerImpact:  request.ServerImpact,
		Mode:          request.Mode,
		Status:        "playing",
		Progress:      nonNegative(request.ProgressSeconds),
		Duration:      nonNegative(request.DurationSeconds),
		StartedAt:     now,
		UpdatedAt:     now,
	}
	s.store(session)
	s.publish("session.started", session)
	return session, nil
}

func (s *Service) Update(id string, request UpdateRequest) (Session, bool) {
	s.mu.Lock()
	session, ok := s.items[id]
	if !ok {
		s.mu.Unlock()
		return Session{}, false
	}
	session.Progress = nonNegative(request.ProgressSeconds)
	session.Duration = nonNegative(request.DurationSeconds)
	if request.Status != "" {
		session.Status = request.Status
	}
	session.UpdatedAt = time.Now().UTC()
	s.items[id] = session
	s.mu.Unlock()
	s.publish("session.updated", session)
	return session, true
}

func (s *Service) Stop(id string) (Session, bool) {
	s.mu.Lock()
	session, ok := s.items[id]
	if !ok {
		s.mu.Unlock()
		return Session{}, false
	}
	session.Status = "stopped"
	session.UpdatedAt = time.Now().UTC()
	delete(s.items, id)
	s.mu.Unlock()
	s.publish("session.stopped", session)
	return session, true
}

func (s *Service) List() []Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	output := make([]Session, 0, len(s.items))
	for _, item := range s.items {
		output = append(output, item)
	}
	sort.Slice(output, func(i, j int) bool { return output[i].UpdatedAt.After(output[j].UpdatedAt) })
	return output
}

func (s *Service) store(session Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[session.ID] = session
}

func (s *Service) publish(eventType string, session Session) {
	if s.events != nil {
		s.events.Publish(eventType, session)
	}
}

func (s *Service) nextSessionID() string {
	return "session_" + time.Now().UTC().Format("20060102T150405") + "_" + stringID(s.nextID.Add(1))
}

func nonNegative(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
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
