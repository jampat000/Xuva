package streaming

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	ErrMissingToken      = errors.New("stream token is required")
	ErrMalformedToken    = errors.New("stream token is malformed")
	ErrInvalidSignature  = errors.New("stream token signature is invalid")
	ErrExpiredToken      = errors.New("stream token is expired")
	ErrTokenMismatch     = errors.New("stream token does not match playback session")
	ErrStreamLimit       = errors.New("stream limit exceeded")
	ErrSigningKeyMissing = errors.New("stream signing key is missing")
)

type Claims struct {
	KeyID         string `json:"kid"`
	MediaSourceID string `json:"mediaSourceId"`
	SessionID     string `json:"sessionId"`
	UserID        string `json:"userId"`
	DeviceID      string `json:"deviceId"`
	Nonce         string `json:"nonce"`
	ExpiresAt     int64  `json:"exp"`
}

type Expected struct {
	MediaSourceID string
	SessionID     string
	UserID        string
	DeviceID      string
}

type Service struct {
	mu               sync.Mutex
	keys             map[string][]byte
	currentKeyID     string
	ttl              time.Duration
	maxUserStreams   int
	maxDeviceStreams int
	activeByUser     map[string]int
	activeByDevice   map[string]int
}

func NewService() *Service {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		panic(err)
	}
	return NewServiceWithKey("local-1", secret)
}

func NewServiceWithKey(keyID string, secret []byte) *Service {
	if strings.TrimSpace(keyID) == "" {
		keyID = "local-1"
	}
	copied := append([]byte(nil), secret...)
	return &Service{
		keys:             map[string][]byte{keyID: copied},
		currentKeyID:     keyID,
		ttl:              5 * time.Minute,
		maxUserStreams:   4,
		maxDeviceStreams: 2,
		activeByUser:     map[string]int{},
		activeByDevice:   map[string]int{},
	}
}

func (s *Service) SetLimits(maxUserStreams int, maxDeviceStreams int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxUserStreams = maxUserStreams
	s.maxDeviceStreams = maxDeviceStreams
}

func (s *Service) RotateKey(keyID string, secret []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys[keyID] = append([]byte(nil), secret...)
	s.currentKeyID = keyID
}

func (s *Service) Issue(expected Expected, ttl time.Duration) (string, Claims, error) {
	if ttl == 0 {
		ttl = s.ttl
	}
	s.mu.Lock()
	keyID := s.currentKeyID
	secret := append([]byte(nil), s.keys[keyID]...)
	s.mu.Unlock()
	if len(secret) == 0 {
		return "", Claims{}, ErrSigningKeyMissing
	}
	claims := Claims{
		KeyID:         keyID,
		MediaSourceID: expected.MediaSourceID,
		SessionID:     expected.SessionID,
		UserID:        expected.UserID,
		DeviceID:      expected.DeviceID,
		Nonce:         randomNonce(),
		ExpiresAt:     time.Now().UTC().Add(ttl).Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", Claims{}, err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := sign(encodedPayload, secret)
	return encodedPayload + "." + signature, claims, nil
}

func (s *Service) Validate(token string, expected Expected) (Claims, func(), error) {
	if strings.TrimSpace(token) == "" {
		return Claims{}, nil, ErrMissingToken
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return Claims{}, nil, ErrMalformedToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, nil, ErrMalformedToken
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Claims{}, nil, ErrMalformedToken
	}
	s.mu.Lock()
	secret := append([]byte(nil), s.keys[claims.KeyID]...)
	s.mu.Unlock()
	if len(secret) == 0 {
		return Claims{}, nil, ErrSigningKeyMissing
	}
	expectedSignature := sign(parts[0], secret)
	if subtle.ConstantTimeCompare([]byte(expectedSignature), []byte(parts[1])) != 1 {
		return Claims{}, nil, ErrInvalidSignature
	}
	if time.Now().UTC().Unix() > claims.ExpiresAt {
		return Claims{}, nil, ErrExpiredToken
	}
	if !claimsMatch(claims, expected) {
		return Claims{}, nil, ErrTokenMismatch
	}
	release, err := s.acquire(claims)
	if err != nil {
		return Claims{}, nil, err
	}
	return claims, release, nil
}

func (s *Service) acquire(claims Claims) (func(), error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.maxUserStreams > 0 && s.activeByUser[claims.UserID] >= s.maxUserStreams {
		return nil, ErrStreamLimit
	}
	if s.maxDeviceStreams > 0 && s.activeByDevice[claims.UserID+"|"+claims.DeviceID] >= s.maxDeviceStreams {
		return nil, ErrStreamLimit
	}
	s.activeByUser[claims.UserID]++
	s.activeByDevice[claims.UserID+"|"+claims.DeviceID]++
	var once sync.Once
	return func() {
		once.Do(func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			decrement(s.activeByUser, claims.UserID)
			decrement(s.activeByDevice, claims.UserID+"|"+claims.DeviceID)
		})
	}, nil
}

func claimsMatch(claims Claims, expected Expected) bool {
	return claims.MediaSourceID == expected.MediaSourceID &&
		claims.SessionID == expected.SessionID &&
		claims.UserID == expected.UserID &&
		claims.DeviceID == expected.DeviceID
}

func sign(payload string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func decrement(values map[string]int, key string) {
	if values[key] <= 1 {
		delete(values, key)
		return
	}
	values[key]--
}

func randomNonce() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}
