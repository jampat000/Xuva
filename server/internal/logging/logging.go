package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Config struct {
	Format   string
	Level    string
	Dir      string
	MaxMB    int
	MaxFiles int
}

// Configure installs console logging plus an optional structured JSONL file.
// The JSONL file is the durable diagnostics surface for packaged installs.
func Configure(cfg Config) (io.Closer, error) {
	var slogLevel slog.Level
	switch strings.ToLower(strings.TrimSpace(cfg.Level)) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn", "warning":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: slogLevel}
	handlers := []slog.Handler{consoleHandler(cfg.Format, opts)}
	var closer io.Closer = noopCloser{}

	logDir := strings.TrimSpace(cfg.Dir)
	if logDir != "" {
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			slog.SetDefault(slog.New(multiHandler(handlers)))
			return closer, err
		}
		file, err := newRotatingFile(filepath.Join(logDir, "xuva.ndjson"), cfg.MaxMB, cfg.MaxFiles)
		if err != nil {
			slog.SetDefault(slog.New(multiHandler(handlers)))
			return closer, err
		}
		closer = file
		handlers = append(handlers, slog.NewJSONHandler(file, opts))
	}

	slog.SetDefault(slog.New(multiHandler(handlers)))
	return closer, nil
}

func consoleHandler(format string, opts *slog.HandlerOptions) slog.Handler {
	if strings.EqualFold(strings.TrimSpace(format), "json") {
		return slog.NewJSONHandler(os.Stderr, opts)
	}
	return slog.NewTextHandler(os.Stderr, opts)
}

type noopCloser struct{}

func (noopCloser) Close() error { return nil }

type rotatingFile struct {
	mu       sync.Mutex
	path     string
	maxBytes int64
	maxFiles int
	file     *os.File
	size     int64
}

func newRotatingFile(path string, maxMB int, maxFiles int) (*rotatingFile, error) {
	if maxMB < 1 {
		maxMB = 25
	}
	if maxFiles < 1 {
		maxFiles = 5
	}
	writer := &rotatingFile{
		path:     path,
		maxBytes: int64(maxMB) * 1024 * 1024,
		maxFiles: maxFiles,
	}
	if err := writer.open(); err != nil {
		return nil, err
	}
	return writer, nil
}

func (w *rotatingFile) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.open(); err != nil {
			return 0, err
		}
	}
	if w.size > 0 && w.size+int64(len(p)) > w.maxBytes {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

func (w *rotatingFile) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

func (w *rotatingFile) open() error {
	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return err
	}
	w.file = file
	w.size = stat.Size()
	return nil
}

func (w *rotatingFile) rotate() error {
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
		w.file = nil
	}
	for i := w.maxFiles - 1; i >= 1; i-- {
		src := w.path + "." + strconv.Itoa(i)
		dst := w.path + "." + strconv.Itoa(i+1)
		if i == w.maxFiles-1 {
			_ = os.Remove(dst)
		}
		if _, err := os.Stat(src); err == nil {
			if err := os.Rename(src, dst); err != nil {
				return err
			}
		}
	}
	if _, err := os.Stat(w.path); err == nil {
		if err := os.Rename(w.path, w.path+".1"); err != nil {
			return err
		}
	}
	return w.open()
}

type multiHandler []slog.Handler

func (h multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h multiHandler) Handle(ctx context.Context, record slog.Record) error {
	var firstErr error
	for _, handler := range h {
		if !handler.Enabled(ctx, record.Level) {
			continue
		}
		if err := handler.Handle(ctx, record.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (h multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make(multiHandler, 0, len(h))
	for _, handler := range h {
		next = append(next, handler.WithAttrs(attrs))
	}
	return next
}

func (h multiHandler) WithGroup(name string) slog.Handler {
	next := make(multiHandler, 0, len(h))
	for _, handler := range h {
		next = append(next, handler.WithGroup(name))
	}
	return next
}
