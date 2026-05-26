package logging

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigureWritesStructuredLogFile(t *testing.T) {
	dir := t.TempDir()
	closer, err := Configure(Config{Format: "text", Level: "info", Dir: dir})
	if err != nil {
		t.Fatalf("configure logging: %v", err)
	}
	t.Cleanup(func() { _ = closer.Close() })

	slog.Info("runtime ready", "component", "test")
	if err := closer.Close(); err != nil {
		t.Fatalf("close log: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "xuva.ndjson"))
	if err != nil {
		t.Fatalf("read structured log: %v", err)
	}
	line := strings.TrimSpace(string(raw))
	if line == "" {
		t.Fatal("expected structured log line")
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		t.Fatalf("structured log is not json: %v", err)
	}
	if payload["msg"] != "runtime ready" {
		t.Fatalf("expected message in structured log, got %#v", payload)
	}
	if payload["component"] != "test" {
		t.Fatalf("expected attributes in structured log, got %#v", payload)
	}
}

func TestConfigureRotatesStructuredLogFile(t *testing.T) {
	dir := t.TempDir()
	closer, err := Configure(Config{Format: "text", Level: "info", Dir: dir, MaxMB: 1, MaxFiles: 2})
	if err != nil {
		t.Fatalf("configure logging: %v", err)
	}

	for i := 0; i < 3000; i++ {
		slog.Info("large log payload", "payload", strings.Repeat("x", 1000), "index", i)
	}
	if err := closer.Close(); err != nil {
		t.Fatalf("close log: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "xuva.ndjson")); err != nil {
		t.Fatalf("expected current log file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "xuva.ndjson.1")); err != nil {
		t.Fatalf("expected rotated log file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "xuva.ndjson.3")); !os.IsNotExist(err) {
		t.Fatalf("expected old rotations beyond max files to be absent, got %v", err)
	}
}
