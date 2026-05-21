package pairing

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusExpired  = "expired"
	StatusDenied   = "denied"
)

var (
	ErrNotFound = errors.New("pairing request not found")
	ErrExpired  = errors.New("pairing request expired")
	ErrClosed   = errors.New("pairing request is no longer pending")
)

type Request struct {
	ID            string     `json:"id"`
	Code          string     `json:"code,omitempty"`
	DeviceName    string     `json:"deviceName"`
	ClientProfile string     `json:"clientProfile"`
	DeviceID      string     `json:"deviceId,omitempty"`
	Auth          *AuthGrant `json:"auth,omitempty"`
	Status        string     `json:"status"`
	ApprovedBy    string     `json:"approvedBy,omitempty"`
	ExpiresAt     time.Time  `json:"expiresAt"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type AuthGrant struct {
	Method       string    `json:"method"`
	SessionToken string    `json:"sessionToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type CreateRequest struct {
	DeviceName    string `json:"deviceName"`
	ClientProfile string `json:"clientProfile"`
	DeviceID      string `json:"deviceId"`
}

type Service struct {
	ttl  time.Duration
	mu   sync.RWMutex
	byID map[string]Request
}

func NewService() *Service {
	return &Service{
		ttl:  10 * time.Minute,
		byID: map[string]Request{},
	}
}

func (s *Service) Create(request CreateRequest) (Request, error) {
	now := time.Now().UTC()
	deviceName := strings.TrimSpace(request.DeviceName)
	if deviceName == "" {
		deviceName = "Apple TV"
	}
	clientProfile := strings.TrimSpace(request.ClientProfile)
	if clientProfile == "" {
		clientProfile = "apple-tv"
	}
	code, err := randomCode()
	if err != nil {
		return Request{}, err
	}
	item := Request{
		ID:            "pair_" + uuid.NewString(),
		Code:          code,
		DeviceName:    deviceName,
		ClientProfile: clientProfile,
		DeviceID:      strings.TrimSpace(request.DeviceID),
		Status:        StatusPending,
		ExpiresAt:     now.Add(s.ttl),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.mu.Lock()
	s.byID[item.ID] = item
	s.mu.Unlock()
	return item, nil
}

func (s *Service) List() []Request {
	s.expireOld()
	s.mu.RLock()
	defer s.mu.RUnlock()
	output := make([]Request, 0, len(s.byID))
	for _, item := range s.byID {
		output = append(output, publicRequest(item, false))
	}
	sort.Slice(output, func(i, j int) bool { return output[i].CreatedAt.After(output[j].CreatedAt) })
	return output
}

func (s *Service) Get(id string) (Request, bool) {
	s.expireOld()
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.byID[id]
	if !ok {
		return Request{}, false
	}
	return publicRequest(item, true), true
}

func (s *Service) Approve(id string, approvedBy string) (Request, error) {
	return s.close(id, StatusApproved, approvedBy, "")
}

func (s *Service) ApproveWithDeviceID(id string, approvedBy string, deviceID string) (Request, error) {
	return s.close(id, StatusApproved, approvedBy, deviceID)
}

func (s *Service) Deny(id string, approvedBy string) (Request, error) {
	return s.close(id, StatusDenied, approvedBy, "")
}

func (s *Service) AttachAuthGrant(id string, grant AuthGrant) (Request, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.byID[id]
	if !ok {
		return Request{}, ErrNotFound
	}
	if item.Status != StatusApproved {
		return Request{}, ErrClosed
	}
	item.Auth = &grant
	item.UpdatedAt = time.Now().UTC()
	s.byID[id] = item
	return publicRequest(item, true), nil
}

func (s *Service) close(id string, status string, approvedBy string, deviceID string) (Request, error) {
	s.expireOld()
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.byID[id]
	if !ok {
		return Request{}, ErrNotFound
	}
	if item.Status == StatusExpired {
		return Request{}, ErrExpired
	}
	if item.Status != StatusPending {
		return Request{}, ErrClosed
	}
	item.Status = status
	item.ApprovedBy = strings.TrimSpace(approvedBy)
	item.UpdatedAt = time.Now().UTC()
	if status == StatusApproved {
		item.DeviceID = strings.TrimSpace(firstNonEmpty(item.DeviceID, deviceID))
		if item.DeviceID == "" {
			item.DeviceID = "device_" + uuid.NewString()
		}
	}
	s.byID[id] = item
	return publicRequest(item, true), nil
}

func (s *Service) expireOld() {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, item := range s.byID {
		if item.Status == StatusPending && now.After(item.ExpiresAt) {
			item.Status = StatusExpired
			item.UpdatedAt = now
			s.byID[id] = item
		}
	}
}

func publicRequest(item Request, includeAuth bool) Request {
	if item.Status != StatusPending {
		item.Code = ""
	}
	if !includeAuth {
		item.Auth = nil
	}
	return item
}

func randomCode() (string, error) {
	var bytes [4]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	value := int(bytes[0])<<24 | int(bytes[1])<<16 | int(bytes[2])<<8 | int(bytes[3])
	if value < 0 {
		value = -value
	}
	return fmt.Sprintf("%06d", value%1000000), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
