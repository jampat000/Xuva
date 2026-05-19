// Package thumbnails generates thumbnail sprite sheets and WebVTT chapter
// tracks for media files. Results are stored in cacheDir/thumbnails/{id}/
// and served via the API for scrubber-preview in the web player.
package thumbnails

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// IntervalSeconds is how often thumbnails are sampled from the video.
	IntervalSeconds = 10
	// ThumbWidth / ThumbHeight are the dimensions of each thumbnail cell.
	ThumbWidth  = 160
	ThumbHeight = 90
	// GridCols is the number of columns in the sprite sheet.
	GridCols = 20
)

// Status describes whether thumbnails have been generated for a media source.
type Status struct {
	Generated      bool      `json:"generated"`
	GeneratedAt    time.Time `json:"generatedAt,omitempty"`
	SpriteURL      string    `json:"spriteUrl,omitempty"`
	VTTURL         string    `json:"vttUrl,omitempty"`
	ChaptersURL    string    `json:"chaptersUrl,omitempty"`
	HasChapters    bool      `json:"hasChapters"`
	ThumbnailCount int       `json:"thumbnailCount"`
	Error          string    `json:"error,omitempty"`
}

type statusFile struct {
	Status
	MediaSourceID string `json:"mediaSourceId"`
	SourcePath    string `json:"sourcePath"`
}

// Service manages thumbnail sprite generation.
type Service struct {
	cacheDir    string
	ffmpegPath  string
	ffprobePath string
	mu          sync.Mutex
	running     map[string]bool
}

// New returns a new Service that stores thumbnails under cacheDir/thumbnails/.
// ffmpegPath and ffprobePath should be absolute paths (or empty to use PATH).
func New(cacheDir, ffmpegPath, ffprobePath string) *Service {
	return &Service{
		cacheDir:    cacheDir,
		ffmpegPath:  ffmpegPath,
		ffprobePath: ffprobePath,
		running:     make(map[string]bool),
	}
}

// thumbDir returns the cache subdirectory for the given media source ID.
func (s *Service) thumbDir(mediaSourceID string) string {
	return filepath.Join(s.cacheDir, "thumbnails", mediaSourceID)
}

// spritePath returns the absolute path to the sprite JPEG.
func (s *Service) spritePath(mediaSourceID string) string {
	return filepath.Join(s.thumbDir(mediaSourceID), "sprite.jpg")
}

// vttPath returns the absolute path to the WebVTT thumbnail track.
func (s *Service) vttPath(mediaSourceID string) string {
	return filepath.Join(s.thumbDir(mediaSourceID), "thumbnails.vtt")
}

// chaptersPath returns the absolute path to the WebVTT chapter track.
func (s *Service) chaptersPath(mediaSourceID string) string {
	return filepath.Join(s.thumbDir(mediaSourceID), "chapters.vtt")
}

// statusPath returns the absolute path to the JSON status sidecar.
func (s *Service) statusPath(mediaSourceID string) string {
	return filepath.Join(s.thumbDir(mediaSourceID), "status.json")
}

// GetStatus returns the current generation status for a media source.
func (s *Service) GetStatus(mediaSourceID string) Status {
	s.mu.Lock()
	inProgress := s.running[mediaSourceID]
	s.mu.Unlock()

	if data, err := os.ReadFile(s.statusPath(mediaSourceID)); err == nil {
		var sf statusFile
		if json.Unmarshal(data, &sf) == nil {
			st := sf.Status
			if inProgress {
				st.Generated = false // still generating
			}
			return st
		}
	}
	return Status{}
}

// SpritePath returns the filesystem path to the sprite image, or an error if
// thumbnails have not been generated for this media source.
func (s *Service) SpritePath(mediaSourceID string) (string, error) {
	p := s.spritePath(mediaSourceID)
	if _, err := os.Stat(p); err != nil {
		return "", errors.New("sprite not available")
	}
	return p, nil
}

// VTTPath returns the filesystem path to the WebVTT thumbnail track, or an
// error if thumbnails have not been generated.
func (s *Service) VTTPath(mediaSourceID string) (string, error) {
	p := s.vttPath(mediaSourceID)
	if _, err := os.Stat(p); err != nil {
		return "", errors.New("thumbnail VTT not available")
	}
	return p, nil
}

// ChaptersPath returns the filesystem path to the WebVTT chapter track if
// chapters were found, or an error otherwise.
func (s *Service) ChaptersPath(mediaSourceID string) (string, error) {
	p := s.chaptersPath(mediaSourceID)
	if _, err := os.Stat(p); err != nil {
		return "", errors.New("chapters not available")
	}
	return p, nil
}

// Generate starts background thumbnail generation for the given media source.
// It is a no-op if generation is already in progress or already completed.
// The provided context is used only to check cancellation at submission time;
// the actual generation runs in a goroutine with context.Background().
func (s *Service) Generate(ctx context.Context, mediaSourceID, sourcePath string, durationSeconds float64) error {
	if mediaSourceID == "" || sourcePath == "" {
		return errors.New("mediaSourceID and sourcePath are required")
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	s.mu.Lock()
	if s.running[mediaSourceID] {
		s.mu.Unlock()
		return nil // already in progress
	}
	// Check if already done
	if _, err := os.Stat(s.spritePath(mediaSourceID)); err == nil {
		s.mu.Unlock()
		return nil // already generated
	}
	s.running[mediaSourceID] = true
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.running, mediaSourceID)
			s.mu.Unlock()
		}()
		_ = s.generate(context.Background(), mediaSourceID, sourcePath, durationSeconds)
	}()
	return nil
}

// GenerateSync runs thumbnail generation synchronously and returns when done.
// Useful for testing and explicit one-shot generation.
func (s *Service) GenerateSync(ctx context.Context, mediaSourceID, sourcePath string, durationSeconds float64) error {
	if mediaSourceID == "" || sourcePath == "" {
		return errors.New("mediaSourceID and sourcePath are required")
	}
	s.mu.Lock()
	if s.running[mediaSourceID] {
		s.mu.Unlock()
		return errors.New("generation already in progress")
	}
	s.running[mediaSourceID] = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.running, mediaSourceID)
		s.mu.Unlock()
	}()
	return s.generate(ctx, mediaSourceID, sourcePath, durationSeconds)
}

func (s *Service) generate(ctx context.Context, mediaSourceID, sourcePath string, durationSeconds float64) error {
	dir := s.thumbDir(mediaSourceID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create thumbnail cache dir: %w", err)
	}

	interval := float64(IntervalSeconds)
	if durationSeconds <= 0 {
		durationSeconds = 3600 // assume 1 hour if unknown
	}
	totalThumbs := int(math.Ceil(durationSeconds / interval))
	if totalThumbs < 1 {
		totalThumbs = 1
	}

	gridCols := GridCols
	if totalThumbs < gridCols {
		gridCols = totalThumbs
	}
	gridRows := int(math.Ceil(float64(totalThumbs) / float64(gridCols)))

	// ── Generate sprite sheet with ffmpeg ─────────────────────────────────
	spritePath := s.spritePath(mediaSourceID)
	ffmpeg := s.ffmpegPath
	if ffmpeg == "" {
		ffmpeg = "ffmpeg"
	}
	tileFilter := fmt.Sprintf("fps=1/%d,scale=%d:%d,tile=%dx%d",
		IntervalSeconds, ThumbWidth, ThumbHeight, gridCols, gridRows)
	cmd := exec.CommandContext(ctx, ffmpeg,
		"-i", sourcePath,
		"-vf", tileFilter,
		"-frames:v", "1",
		"-q:v", "5",
		"-y",
		spritePath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		// Save error status
		_ = s.saveStatus(mediaSourceID, sourcePath, statusFile{
			Status: Status{
				Generated: false,
				Error:     strings.TrimSpace(string(out)),
			},
		})
		return fmt.Errorf("ffmpeg sprite generation: %w; output: %s", err, string(out))
	}

	// ── Write WebVTT thumbnail track ──────────────────────────────────────
	vtt := buildThumbnailVTT(totalThumbs, interval, gridCols, ThumbWidth, ThumbHeight, "sprite.jpg")
	if err := os.WriteFile(s.vttPath(mediaSourceID), []byte(vtt), 0o644); err != nil {
		return fmt.Errorf("write thumbnail VTT: %w", err)
	}

	// ── Extract and write chapters (best-effort) ──────────────────────────
	hasChapters := false
	chapters, chapErr := s.extractChapters(ctx, sourcePath)
	if chapErr == nil && len(chapters) > 0 {
		chVTT := buildChapterVTT(chapters)
		if err := os.WriteFile(s.chaptersPath(mediaSourceID), []byte(chVTT), 0o644); err == nil {
			hasChapters = true
		}
	}

	// ── Save status sidecar ───────────────────────────────────────────────
	st := statusFile{
		MediaSourceID: mediaSourceID,
		SourcePath:    sourcePath,
		Status: Status{
			Generated:      true,
			GeneratedAt:    time.Now().UTC(),
			SpriteURL:      "/api/media-sources/" + mediaSourceID + "/thumbnails/sprite.jpg",
			VTTURL:         "/api/media-sources/" + mediaSourceID + "/thumbnails/thumbnails.vtt",
			HasChapters:    hasChapters,
			ThumbnailCount: totalThumbs,
		},
	}
	if hasChapters {
		st.ChaptersURL = "/api/media-sources/" + mediaSourceID + "/thumbnails/chapters.vtt"
	}
	return s.saveStatus(mediaSourceID, sourcePath, st)
}

func (s *Service) saveStatus(mediaSourceID, sourcePath string, st statusFile) error {
	st.MediaSourceID = mediaSourceID
	st.SourcePath = sourcePath
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.statusPath(mediaSourceID), data, 0o644)
}

// ── WebVTT sprite builder ──────────────────────────────────────────────────

func buildThumbnailVTT(totalThumbs int, interval float64, gridCols, thumbW, thumbH int, spriteFile string) string {
	var sb strings.Builder
	sb.WriteString("WEBVTT\n\n")
	for i := 0; i < totalThumbs; i++ {
		start := float64(i) * interval
		end := start + interval
		col := i % gridCols
		row := i / gridCols
		x := col * thumbW
		y := row * thumbH
		sb.WriteString(vttTimestamp(start))
		sb.WriteString(" --> ")
		sb.WriteString(vttTimestamp(end))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("%s#xywh=%d,%d,%d,%d\n\n", spriteFile, x, y, thumbW, thumbH))
	}
	return sb.String()
}

func vttTimestamp(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	ms := int((seconds-math.Floor(seconds))*1000)
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}

// ── Chapter extraction via ffprobe ─────────────────────────────────────────

type ffprobeChapter struct {
	ID        int64             `json:"id"`
	StartTime string            `json:"start_time"`
	EndTime   string            `json:"end_time"`
	Tags      map[string]string `json:"tags"`
}

type ffprobeChaptersOutput struct {
	Chapters []ffprobeChapter `json:"chapters"`
}

func (s *Service) extractChapters(ctx context.Context, sourcePath string) ([]ffprobeChapter, error) {
	ffprobe := s.ffprobePath
	if ffprobe == "" {
		ffprobe = "ffprobe"
	}
	cmd := exec.CommandContext(ctx, ffprobe,
		"-v", "quiet",
		"-print_format", "json",
		"-show_chapters",
		sourcePath,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var payload ffprobeChaptersOutput
	if err := json.Unmarshal(out, &payload); err != nil {
		return nil, err
	}
	return payload.Chapters, nil
}

func buildChapterVTT(chapters []ffprobeChapter) string {
	var sb strings.Builder
	sb.WriteString("WEBVTT\n\n")
	for i, ch := range chapters {
		start, _ := strconv.ParseFloat(ch.StartTime, 64)
		end, _ := strconv.ParseFloat(ch.EndTime, 64)
		title := ch.Tags["title"]
		if title == "" {
			title = fmt.Sprintf("Chapter %d", i+1)
		}
		sb.WriteString(fmt.Sprintf("%d\n", i+1))
		sb.WriteString(vttTimestamp(start))
		sb.WriteString(" --> ")
		sb.WriteString(vttTimestamp(end))
		sb.WriteString("\n")
		sb.WriteString(title)
		sb.WriteString("\n\n")
	}
	return sb.String()
}
