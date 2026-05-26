package config

import (
	"path/filepath"
	"testing"
)

func TestHTTPAddrLoopbackOnly(t *testing.T) {
	if !HTTPAddrLoopbackOnly("127.0.0.1:8097") {
		t.Fatal("expected loopback IPv4 address to be loopback-only")
	}
	if !HTTPAddrLoopbackOnly("localhost:8097") {
		t.Fatal("expected localhost address to be loopback-only")
	}
	if HTTPAddrLoopbackOnly("0.0.0.0:8097") {
		t.Fatal("expected wildcard address to not be loopback-only")
	}
}

func TestFromEnvDerivesLogDirFromDataDir(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("XUVA_DATA_DIR", dataDir)
	t.Setenv("XUVA_LOG_DIR", "")

	cfg := FromEnv()
	if cfg.LogDir != filepath.Join(dataDir, "logs") {
		t.Fatalf("expected log dir under data dir, got %q", cfg.LogDir)
	}
}

func TestFromEnvHonorsExplicitLogDir(t *testing.T) {
	dataDir := t.TempDir()
	logDir := filepath.Join(t.TempDir(), "logs")
	t.Setenv("XUVA_DATA_DIR", dataDir)
	t.Setenv("XUVA_LOG_DIR", logDir)

	cfg := FromEnv()
	if cfg.LogDir != logDir {
		t.Fatalf("expected explicit log dir, got %q", cfg.LogDir)
	}
}

func TestFromEnvUsesRuntimeHomeForDefaultRuntimeDirs(t *testing.T) {
	runtimeHome := t.TempDir()
	t.Setenv("XUVA_RUNTIME_HOME", runtimeHome)

	cfg := FromEnv()
	if cfg.RuntimeHome != runtimeHome {
		t.Fatalf("expected runtime home to be recorded, got %q", cfg.RuntimeHome)
	}
	if cfg.DataDir != filepath.Join(runtimeHome, "data") {
		t.Fatalf("expected data dir under runtime home, got %q", cfg.DataDir)
	}
	if cfg.LogDir != filepath.Join(runtimeHome, "logs") {
		t.Fatalf("expected log dir under runtime home, got %q", cfg.LogDir)
	}
	if cfg.TranscodeDir != filepath.Join(runtimeHome, "transcode") {
		t.Fatalf("expected transcode dir under runtime home, got %q", cfg.TranscodeDir)
	}
	if cfg.TrailersDir != filepath.Join(runtimeHome, "trailers") {
		t.Fatalf("expected trailers dir under runtime home, got %q", cfg.TrailersDir)
	}
}

func TestFromEnvHonorsExplicitDataDirWithRuntimeHome(t *testing.T) {
	runtimeHome := t.TempDir()
	dataDir := filepath.Join(t.TempDir(), "custom-data")
	t.Setenv("XUVA_RUNTIME_HOME", runtimeHome)
	t.Setenv("XUVA_DATA_DIR", dataDir)

	cfg := FromEnv()
	if cfg.DataDir != dataDir {
		t.Fatalf("expected explicit data dir, got %q", cfg.DataDir)
	}
	if cfg.LogDir != filepath.Join(runtimeHome, "logs") {
		t.Fatalf("expected log dir to stay under runtime home, got %q", cfg.LogDir)
	}
	if cfg.CacheDir != filepath.Join(runtimeHome, "cache") {
		t.Fatalf("expected cache dir to stay under runtime home, got %q", cfg.CacheDir)
	}
}

func TestNormalizeWebOrigin(t *testing.T) {
	got, err := NormalizeWebOrigin("HTTP://media-server.local:8097/settings?x=1")
	if err == nil {
		t.Fatalf("expected path/query origin to be rejected, got %q", got)
	}
	got, err = NormalizeWebOrigin("HTTP://media-server.local:8097")
	if err != nil {
		t.Fatalf("normalize origin: %v", err)
	}
	if got != "http://media-server.local:8097" {
		t.Fatalf("expected normalized origin, got %q", got)
	}
}
