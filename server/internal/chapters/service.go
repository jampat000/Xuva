package chapters

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"
)

// Chapters holds detected chapter markers for one media source.
type Chapters struct {
	MediaSourceID string   `json:"mediaSourceId"`
	Intro         *Segment `json:"intro,omitempty"`
	Credits       *Segment `json:"credits,omitempty"`
	AnalyzedAt    string   `json:"analyzedAt,omitempty"`
}

// EpisodeInput describes one episode for season-level analysis.
type EpisodeInput struct {
	MediaSourceID string
	Path          string
	Duration      float64
}

// Service manages chapter detection and persistence.
type Service struct {
	db       *sql.DB
	analyzer *Analyzer
	mu       sync.Mutex
	running  map[string]bool
}

// NewService returns a new Service backed by the given database and tool paths.
func NewService(db *sql.DB, ffmpegPath, fpcalcPath string) *Service {
	return &Service{
		db:      db,
		analyzer: New(ffmpegPath, fpcalcPath),
		running: make(map[string]bool),
	}
}

// IntroEnabled reports whether fpcalc is configured for intro detection.
func (s *Service) IntroEnabled() bool {
	return s.analyzer.IntroEnabled()
}

// Get returns the stored chapters for the given media source ID, if any.
func (s *Service) Get(ctx context.Context, mediaSourceID string) (Chapters, bool, error) {
	var ch Chapters
	ch.MediaSourceID = mediaSourceID
	var introStart, introEnd, creditsStart sql.NullFloat64
	err := s.db.QueryRowContext(ctx, `
		SELECT intro_start, intro_end, credits_start, analyzed_at
		FROM media_source_chapters
		WHERE media_source_id = ?
	`, mediaSourceID).Scan(&introStart, &introEnd, &creditsStart, &ch.AnalyzedAt)
	if err == sql.ErrNoRows {
		return Chapters{}, false, nil
	}
	if err != nil {
		return Chapters{}, false, err
	}
	if introStart.Valid && introEnd.Valid && introEnd.Float64 > introStart.Float64 {
		ch.Intro = &Segment{Start: introStart.Float64, End: introEnd.Float64}
	}
	if creditsStart.Valid && creditsStart.Float64 >= 0 {
		ch.Credits = &Segment{Start: creditsStart.Float64}
	}
	return ch, true, nil
}

func (s *Service) upsert(ctx context.Context, ch Chapters) error {
	introStart := -1.0
	introEnd := -1.0
	creditsStart := -1.0
	if ch.Intro != nil {
		introStart = ch.Intro.Start
		introEnd = ch.Intro.End
	}
	if ch.Credits != nil {
		creditsStart = ch.Credits.Start
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if ch.AnalyzedAt == "" {
		ch.AnalyzedAt = now
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO media_source_chapters(media_source_id, intro_start, intro_end, credits_start, analyzed_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?)
		ON CONFLICT(media_source_id) DO UPDATE SET
			intro_start  = excluded.intro_start,
			intro_end    = excluded.intro_end,
			credits_start = excluded.credits_start,
			analyzed_at  = excluded.analyzed_at,
			updated_at   = excluded.updated_at
	`, ch.MediaSourceID, introStart, introEnd, creditsStart, ch.AnalyzedAt, now)
	return err
}

// AnalyzeCredits runs credits detection for a single media source and saves
// the result. If a credits marker already exists it is not re-analysed.
// Safe to call concurrently — duplicate calls for the same ID are no-ops.
func (s *Service) AnalyzeCredits(ctx context.Context, mediaSourceID, path string, durationSecs float64) {
	s.mu.Lock()
	key := "credits:" + mediaSourceID
	if s.running[key] {
		s.mu.Unlock()
		return
	}
	s.running[key] = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.running, key)
		s.mu.Unlock()
	}()

	existing, ok, _ := s.Get(ctx, mediaSourceID)
	if ok && existing.Credits != nil {
		return // already done
	}

	credits, err := s.analyzer.DetectCredits(ctx, path, durationSecs)
	if err != nil {
		slog.Warn("chapters: credits detection failed", "mediaSourceId", mediaSourceID, "err", err)
		return
	}
	ch := Chapters{MediaSourceID: mediaSourceID, Credits: credits}
	if ok {
		ch.Intro = existing.Intro
	}
	if err := s.upsert(ctx, ch); err != nil {
		slog.Warn("chapters: save failed", "mediaSourceId", mediaSourceID, "err", err)
	}
}

// AnalyzeSeason fingerprints all episodes and detects their shared intro.
// Credits are NOT overwritten by this call — use AnalyzeCredits for credits.
// Safe to call concurrently — duplicate season calls are no-ops.
func (s *Service) AnalyzeSeason(ctx context.Context, episodes []EpisodeInput) {
	if len(episodes) < 2 || !s.analyzer.IntroEnabled() {
		return
	}
	seasonKey := "season"
	for _, ep := range episodes {
		seasonKey += ":" + ep.MediaSourceID
	}
	s.mu.Lock()
	if s.running[seasonKey] {
		s.mu.Unlock()
		return
	}
	s.running[seasonKey] = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.running, seasonKey)
		s.mu.Unlock()
	}()

	var fps []EpisodeFingerprint
	for _, ep := range episodes {
		fp, err := s.analyzer.Fingerprint(ctx, ep.Path, FingerprintLengthSecs)
		if err != nil {
			slog.Debug("chapters: fingerprint failed", "mediaSourceId", ep.MediaSourceID, "err", err)
			continue
		}
		fps = append(fps, EpisodeFingerprint{MediaSourceID: ep.MediaSourceID, Data: fp})
	}
	if len(fps) < 2 {
		return
	}

	intros := FindIntros(fps)
	for _, ep := range episodes {
		existing, _, _ := s.Get(ctx, ep.MediaSourceID)
		seg, hasIntro := intros[ep.MediaSourceID]
		ch := Chapters{MediaSourceID: ep.MediaSourceID, Credits: existing.Credits}
		if hasIntro {
			ch.Intro = &Segment{Start: seg.Start, End: seg.End}
		}
		if err := s.upsert(ctx, ch); err != nil {
			slog.Warn("chapters: save intro failed", "mediaSourceId", ep.MediaSourceID, "err", err)
		}
	}
}
