package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vyrdenhq/vyrden/server/internal/database"
)

const (
	SessionCookieName = "vyrden_session"
	CSRFCookieName    = "vyrden_csrf"
)

var (
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrCSRF               = errors.New("invalid csrf token")
	ErrLocked             = errors.New("login temporarily locked")
)

type Principal struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	CSRFToken string    `json:"csrfToken,omitempty"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

type ResolvedSession struct {
	Principal Principal
	Session   Session
	Token     string
	Rotated   bool
}

type contextKey string

const resolvedSessionKey contextKey = "auth.resolvedSession"

type Service struct {
	db               *sql.DB
	disabled         bool
	sessionTTL       time.Duration
	rotationInterval time.Duration
	lockoutThreshold int
	lockoutWindow    time.Duration

	mu             sync.Mutex
	unknownAttempt map[string]attempt
}

type attempt struct {
	count       int
	lastFailed  time.Time
	lockedUntil time.Time
}

type userRecord struct {
	ID               string
	Username         string
	DisplayName      string
	Role             string
	PasswordHash     string
	LockedUntil      string
	FailedLoginCount int
	LastFailedAt     string
}

type BootstrapOptions struct {
	Username string
	Password string
}

func NewService(databaseService *database.Service, disabled bool) *Service {
	return &Service{
		db:               databaseService.DB(),
		disabled:         disabled,
		sessionTTL:       24 * time.Hour,
		rotationInterval: 30 * time.Minute,
		lockoutThreshold: 5,
		lockoutWindow:    15 * time.Minute,
		unknownAttempt:   map[string]attempt{},
	}
}

func (s *Service) Disabled() bool {
	return s == nil || s.disabled
}

func (s *Service) Bootstrap(ctx context.Context, options BootstrapOptions) error {
	if s.Disabled() {
		return nil
	}
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE username <> '' AND password_hash <> ''`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	username := normalizeUsername(options.Username)
	if username == "" {
		username = "admin"
	}
	password := options.Password
	if strings.TrimSpace(password) == "" {
		password = randomToken(24)
		slog.Warn("auth bootstrap password generated", "username", username, "password", password)
	}
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	now := timestamp(time.Now())
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO users(id, username, display_name, role, password_hash, password_updated_at, updated_at)
		VALUES(?, ?, ?, 'admin', ?, ?, ?)
	`, "admin", username, "Administrator", hash, now, now)
	if err != nil {
		return err
	}
	slog.Info("auth bootstrap user created", "username", username)
	return nil
}

func (s *Service) CreateUser(ctx context.Context, username string, password string, displayName string, role string) (Principal, error) {
	if s.Disabled() {
		return Principal{}, ErrUnauthorized
	}
	username = normalizeUsername(username)
	if username == "" {
		return Principal{}, errors.New("username is required")
	}
	role = normalizeRole(role)
	if role == "" {
		return Principal{}, errors.New("role is required")
	}
	if strings.TrimSpace(displayName) == "" {
		displayName = username
	}
	hash, err := hashPassword(password)
	if err != nil {
		return Principal{}, err
	}
	now := timestamp(time.Now())
	id := "user_" + uuid.NewString()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO users(id, username, display_name, role, password_hash, password_updated_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?)
	`, id, username, displayName, role, hash, now, now)
	if err != nil {
		return Principal{}, err
	}
	return Principal{ID: id, Username: username, DisplayName: displayName, Role: role}, nil
}

func normalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin":
		return "admin"
	case "standard":
		return "standard"
	default:
		return ""
	}
}

func (s *Service) Authenticate(ctx context.Context, username string, password string, remoteAddr string, userAgent string) (Principal, Session, string, error) {
	if s.Disabled() {
		return Principal{}, Session{}, "", ErrUnauthorized
	}
	key := s.unknownKey(username, remoteAddr)
	if until, locked := s.checkUnknownLockout(key); locked {
		slog.Warn("auth login blocked", "username", normalizeUsername(username), "remote_addr", remoteAddr, "locked_until", until.Format(time.RFC3339))
		return Principal{}, Session{}, "", lockoutError(until)
	}
	user, ok, err := s.lookupUser(ctx, username)
	if err != nil {
		return Principal{}, Session{}, "", err
	}
	if !ok || user.PasswordHash == "" {
		if until, locked := s.recordUnknownFailure(key); locked {
			slog.Warn("auth login locked", "username", normalizeUsername(username), "remote_addr", remoteAddr, "locked_until", until.Format(time.RFC3339))
			return Principal{}, Session{}, "", lockoutError(until)
		}
		slog.Warn("auth login failed", "username", normalizeUsername(username), "remote_addr", remoteAddr, "reason", "unknown user")
		return Principal{}, Session{}, "", ErrInvalidCredentials
	}
	if until, locked := parseTimestamp(user.LockedUntil); locked && until.After(time.Now().UTC()) {
		slog.Warn("auth login blocked", "username", user.Username, "remote_addr", remoteAddr, "locked_until", until.Format(time.RFC3339))
		return Principal{}, Session{}, "", lockoutError(until)
	}
	valid, err := verifyPassword(user.PasswordHash, password)
	if err != nil {
		return Principal{}, Session{}, "", err
	}
	if !valid {
		until, err := s.recordFailedLogin(ctx, user)
		if err != nil {
			return Principal{}, Session{}, "", err
		}
		if !until.IsZero() {
			slog.Warn("auth login locked", "username", user.Username, "remote_addr", remoteAddr, "locked_until", until.Format(time.RFC3339))
			return Principal{}, Session{}, "", lockoutError(until)
		}
		slog.Warn("auth login failed", "username", user.Username, "remote_addr", remoteAddr, "reason", "password mismatch")
		return Principal{}, Session{}, "", ErrInvalidCredentials
	}
	if err := s.clearFailedLogin(ctx, user.ID); err != nil {
		return Principal{}, Session{}, "", err
	}
	s.clearUnknownFailure(key)
	principal := Principal{ID: user.ID, Username: user.Username, DisplayName: user.DisplayName, Role: user.Role}
	session, token, err := s.issueSession(ctx, principal, remoteAddr, userAgent)
	if err != nil {
		return Principal{}, Session{}, "", err
	}
	slog.Info("auth login success", "username", principal.Username, "remote_addr", remoteAddr, "session_id", session.ID)
	return principal, session, token, nil
}

func (s *Service) Resolve(ctx context.Context, token string, remoteAddr string, userAgent string) (ResolvedSession, error) {
	if s.Disabled() {
		return ResolvedSession{}, ErrUnauthorized
	}
	id, secret, err := parseToken(token)
	if err != nil {
		return ResolvedSession{}, ErrUnauthorized
	}
	row := struct {
		SessionID   string
		UserID      string
		Username    string
		DisplayName string
		Role        string
		SecretHash  string
		CSRFToken   string
		CreatedAt   string
		LastSeenAt  string
		ExpiresAt   string
		RevokedAt   string
	}{}
	err = s.db.QueryRowContext(ctx, `
		SELECT s.id, s.user_id, u.username, u.display_name, u.role, s.secret_hash, s.csrf_token, s.created_at, s.last_seen_at, s.expires_at, s.revoked_at
		FROM auth_sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.id = ?
	`, id).Scan(&row.SessionID, &row.UserID, &row.Username, &row.DisplayName, &row.Role, &row.SecretHash, &row.CSRFToken, &row.CreatedAt, &row.LastSeenAt, &row.ExpiresAt, &row.RevokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ResolvedSession{}, ErrUnauthorized
	}
	if err != nil {
		return ResolvedSession{}, err
	}
	if strings.TrimSpace(row.RevokedAt) != "" {
		return ResolvedSession{}, ErrUnauthorized
	}
	if subtle.ConstantTimeCompare([]byte(row.SecretHash), []byte(hashSecret(secret))) != 1 {
		return ResolvedSession{}, ErrUnauthorized
	}
	now := time.Now().UTC()
	expiresAt, ok := parseTimestamp(row.ExpiresAt)
	if !ok || now.After(expiresAt) {
		_ = s.revokeSessionID(ctx, row.SessionID)
		return ResolvedSession{}, ErrUnauthorized
	}
	createdAt, _ := parseTimestamp(row.CreatedAt)
	rotated := false
	nextToken := token
	nextCSRF := row.CSRFToken
	if createdAt.IsZero() {
		createdAt = now
	}
	if now.Sub(createdAt) >= s.rotationInterval {
		newSecret := randomToken(32)
		nextToken = row.SessionID + "." + newSecret
		nextCSRF = randomToken(32)
		rotated = true
		if _, err := s.db.ExecContext(ctx, `
			UPDATE auth_sessions
			SET secret_hash = ?, csrf_token = ?, created_at = ?, last_seen_at = ?, expires_at = ?, remote_addr = ?, user_agent = ?
			WHERE id = ?
		`, hashSecret(newSecret), nextCSRF, timestamp(now), timestamp(now), timestamp(now.Add(s.sessionTTL)), remoteAddr, userAgent, row.SessionID); err != nil {
			return ResolvedSession{}, err
		}
	} else {
		if _, err := s.db.ExecContext(ctx, `
			UPDATE auth_sessions
			SET last_seen_at = ?, expires_at = ?, remote_addr = ?, user_agent = ?
			WHERE id = ?
		`, timestamp(now), timestamp(now.Add(s.sessionTTL)), remoteAddr, userAgent, row.SessionID); err != nil {
			return ResolvedSession{}, err
		}
	}
	return ResolvedSession{
		Principal: Principal{ID: row.UserID, Username: row.Username, DisplayName: row.DisplayName, Role: row.Role},
		Session: Session{
			ID:        row.SessionID,
			UserID:    row.UserID,
			CSRFToken: nextCSRF,
			ExpiresAt: now.Add(s.sessionTTL),
			CreatedAt: createdAt,
		},
		Token:   nextToken,
		Rotated: rotated,
	}, nil
}

func (s *Service) Revoke(ctx context.Context, token string) error {
	if s.Disabled() {
		return nil
	}
	id, _, err := parseToken(token)
	if err != nil {
		return ErrUnauthorized
	}
	return s.revokeSessionID(ctx, id)
}

func (s *Service) ValidateCSRF(session ResolvedSession, cookieToken string, headerToken string) error {
	if s.Disabled() {
		return nil
	}
	if cookieToken == "" || headerToken == "" {
		return ErrCSRF
	}
	if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(session.Session.CSRFToken)) != 1 {
		return ErrCSRF
	}
	if subtle.ConstantTimeCompare([]byte(headerToken), []byte(session.Session.CSRFToken)) != 1 {
		return ErrCSRF
	}
	return nil
}

func (s *Service) lookupUser(ctx context.Context, username string) (userRecord, bool, error) {
	name := normalizeUsername(username)
	var user userRecord
	err := s.db.QueryRowContext(ctx, `
		SELECT id, username, display_name, role, password_hash, locked_until, failed_login_count, last_failed_at
		FROM users
		WHERE username = ? COLLATE NOCASE
	`, name).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Role, &user.PasswordHash, &user.LockedUntil, &user.FailedLoginCount, &user.LastFailedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return userRecord{}, false, nil
	}
	if err != nil {
		return userRecord{}, false, err
	}
	return user, true, nil
}

func (s *Service) recordFailedLogin(ctx context.Context, user userRecord) (time.Time, error) {
	now := time.Now().UTC()
	lastFailed, _ := parseTimestamp(user.LastFailedAt)
	count := user.FailedLoginCount
	if lastFailed.IsZero() || now.Sub(lastFailed) > s.lockoutWindow {
		count = 0
	}
	count++
	lockUntil := time.Time{}
	if count >= s.lockoutThreshold {
		lockUntil = now.Add(s.lockoutWindow)
		count = 0
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE users
		SET failed_login_count = ?, last_failed_at = ?, locked_until = ?, updated_at = ?
		WHERE id = ?
	`, count, timestamp(now), formatMaybeTime(lockUntil), timestamp(now), user.ID)
	return lockUntil, err
}

func (s *Service) clearFailedLogin(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users
		SET failed_login_count = 0, last_failed_at = '', locked_until = '', updated_at = ?
		WHERE id = ?
	`, timestamp(time.Now().UTC()), userID)
	return err
}

func (s *Service) issueSession(ctx context.Context, principal Principal, remoteAddr string, userAgent string) (Session, string, error) {
	now := time.Now().UTC()
	sessionID := "auth_" + uuid.NewString()
	secret := randomToken(32)
	csrfToken := randomToken(32)
	token := sessionID + "." + secret
	expiresAt := now.Add(s.sessionTTL)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO auth_sessions(id, user_id, secret_hash, csrf_token, remote_addr, user_agent, created_at, last_seen_at, expires_at, revoked_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, '')
	`, sessionID, principal.ID, hashSecret(secret), csrfToken, remoteAddr, userAgent, timestamp(now), timestamp(now), timestamp(expiresAt))
	if err != nil {
		return Session{}, "", err
	}
	return Session{ID: sessionID, UserID: principal.ID, CSRFToken: csrfToken, ExpiresAt: expiresAt, CreatedAt: now}, token, nil
}

func (s *Service) revokeSessionID(ctx context.Context, sessionID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE auth_sessions
		SET revoked_at = ?, expires_at = ?
		WHERE id = ? AND revoked_at = ''
	`, timestamp(time.Now().UTC()), timestamp(time.Now().UTC()), sessionID)
	return err
}

func (s *Service) checkUnknownLockout(key string) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := s.unknownAttempt[key]
	if !entry.lockedUntil.IsZero() && time.Now().UTC().Before(entry.lockedUntil) {
		return entry.lockedUntil, true
	}
	if !entry.lockedUntil.IsZero() && time.Now().UTC().After(entry.lockedUntil) {
		delete(s.unknownAttempt, key)
	}
	return time.Time{}, false
}

func (s *Service) recordUnknownFailure(key string) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	entry := s.unknownAttempt[key]
	if !entry.lastFailed.IsZero() && now.Sub(entry.lastFailed) > s.lockoutWindow {
		entry = attempt{}
	}
	entry.count++
	entry.lastFailed = now
	if entry.count >= s.lockoutThreshold {
		entry.lockedUntil = now.Add(s.lockoutWindow)
		s.unknownAttempt[key] = entry
		return entry.lockedUntil, true
	}
	s.unknownAttempt[key] = entry
	return time.Time{}, false
}

func (s *Service) clearUnknownFailure(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.unknownAttempt, key)
}

func (s *Service) unknownKey(username string, remoteAddr string) string {
	return normalizeUsername(username) + "|" + strings.TrimSpace(remoteAddr)
}

func parseToken(token string) (string, string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", ErrUnauthorized
	}
	return parts[0], parts[1], nil
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func normalizeUsername(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func randomToken(size int) string {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func timestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTimestamp(value string) (time.Time, bool) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func formatMaybeTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return timestamp(value)
}

type lockoutError time.Time

func (e lockoutError) Error() string {
	return fmt.Sprintf("%s until %s", ErrLocked.Error(), time.Time(e).Format(time.RFC3339))
}

func LockoutUntil(err error) (time.Time, bool) {
	var typed lockoutError
	if errors.As(err, &typed) {
		return time.Time(typed), true
	}
	return time.Time{}, false
}

func ContextWithResolvedSession(ctx context.Context, resolved ResolvedSession) context.Context {
	return context.WithValue(ctx, resolvedSessionKey, resolved)
}

func ResolvedSessionFromContext(ctx context.Context) (ResolvedSession, bool) {
	value := ctx.Value(resolvedSessionKey)
	resolved, ok := value.(ResolvedSession)
	return resolved, ok
}
