package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr          string   `json:"httpAddr"`
	DataDir           string   `json:"dataDir"`
	TranscodeDir      string   `json:"transcodeDir,omitempty"`
	DownloadsDir      string   `json:"downloadsDir,omitempty"`
	MetadataDir       string   `json:"metadataDir,omitempty"`
	CacheDir          string   `json:"cacheDir,omitempty"`
	TempDir           string   `json:"tempDir,omitempty"`
	MovieLibraryPath  string   `json:"movieLibraryPath,omitempty"`
	TVLibraryPath     string   `json:"tvLibraryPath,omitempty"`
	FFprobePath       string   `json:"ffprobePath"`
	FFmpegPath        string   `json:"ffmpegPath"`
	OMDbAPIKey        string   `json:"omdbApiKey,omitempty"`
	TMDBAPIKey        string   `json:"tmdbApiKey,omitempty"`
	EventBuffer       int      `json:"eventBuffer"`
	ScanWorkers       int      `json:"scanWorkers"`
	ProbeWorkers      int      `json:"probeWorkers"`
	TranscodeWorkers  int      `json:"transcodeWorkers"`
	GPUWorkers        int      `json:"gpuWorkers"`
	HardwareUnlocked  bool     `json:"hardwareUnlocked,omitempty"`
	PlaybackPolicy    string   `json:"playbackPolicy,omitempty"`
	LibrarySyncMode   string   `json:"librarySyncMode,omitempty"`
	SyncIntervalMins  int      `json:"syncIntervalMins,omitempty"`
	WatchDebounceSecs int      `json:"watchDebounceSecs,omitempty"`
	ProbeBatchLimit   int      `json:"probeBatchLimit,omitempty"`
	AllowedOrigins    []string `json:"allowedOrigins,omitempty"`
	AuthDisabled      bool     `json:"-"`
	AdminUsername     string   `json:"-"`
	AdminPassword     string   `json:"-"`
}

func FromEnv() Config {
	dataDir := envString("VYRDEN_DATA_DIR", "data")
	cfg := Config{
		HTTPAddr:          envString("VYRDEN_HTTP_ADDR", "127.0.0.1:8097"),
		DataDir:           dataDir,
		TranscodeDir:      envString("VYRDEN_TRANSCODE_DIR", filepath.Join(dataDir, "transcode")),
		DownloadsDir:      envString("VYRDEN_DOWNLOADS_DIR", filepath.Join(dataDir, "downloads")),
		MetadataDir:       envString("VYRDEN_METADATA_DIR", filepath.Join(dataDir, "metadata")),
		CacheDir:          envString("VYRDEN_CACHE_DIR", filepath.Join(dataDir, "cache")),
		TempDir:           envString("VYRDEN_TEMP_DIR", filepath.Join(dataDir, "temp")),
		MovieLibraryPath:  envString("VYRDEN_MOVIES_PATH", ""),
		TVLibraryPath:     envString("VYRDEN_TV_PATH", ""),
		FFprobePath:       envString("VYRDEN_FFPROBE_PATH", "ffprobe"),
		FFmpegPath:        envString("VYRDEN_FFMPEG_PATH", "ffmpeg"),
		OMDbAPIKey:        envString("VYRDEN_OMDB_API_KEY", ""),
		TMDBAPIKey:        envString("VYRDEN_TMDB_API_KEY", ""),
		EventBuffer:       envInt("VYRDEN_EVENT_BUFFER", 128),
		ScanWorkers:       envInt("VYRDEN_SCAN_WORKERS", 1),
		ProbeWorkers:      envInt("VYRDEN_PROBE_WORKERS", 2),
		TranscodeWorkers:  envInt("VYRDEN_TRANSCODE_WORKERS", 1),
		GPUWorkers:        envInt("VYRDEN_GPU_WORKERS", 1),
		HardwareUnlocked:  envBool("VYRDEN_HARDWARE_UNLOCKED", false),
		PlaybackPolicy:    envString("VYRDEN_PLAYBACK_POLICY", "original_only"),
		LibrarySyncMode:   envString("VYRDEN_LIBRARY_SYNC_MODE", "daily"),
		SyncIntervalMins:  envInt("VYRDEN_SYNC_INTERVAL_MINS", 1440),
		WatchDebounceSecs: envInt("VYRDEN_WATCH_DEBOUNCE_SECS", 30),
		ProbeBatchLimit:   envInt("VYRDEN_PROBE_BATCH_LIMIT", 50),
		AllowedOrigins:    envCSV("VYRDEN_ALLOWED_ORIGINS", nil),
		AuthDisabled:      envBool("VYRDEN_AUTH_DISABLED", false),
		AdminUsername:     envString("VYRDEN_ADMIN_USERNAME", "admin"),
		AdminPassword:     envString("VYRDEN_ADMIN_PASSWORD", ""),
	}
	if saved, err := LoadFile(dataDir); err == nil {
		cfg = merge(cfg, saved)
	}
	cfg.HTTPAddr = envString("VYRDEN_HTTP_ADDR", cfg.HTTPAddr)
	cfg.DataDir = envString("VYRDEN_DATA_DIR", cfg.DataDir)
	cfg.TranscodeDir = envString("VYRDEN_TRANSCODE_DIR", defaultDir(cfg.TranscodeDir, cfg.DataDir, "transcode"))
	cfg.DownloadsDir = envString("VYRDEN_DOWNLOADS_DIR", defaultDir(cfg.DownloadsDir, cfg.DataDir, "downloads"))
	cfg.MetadataDir = envString("VYRDEN_METADATA_DIR", defaultDir(cfg.MetadataDir, cfg.DataDir, "metadata"))
	cfg.CacheDir = envString("VYRDEN_CACHE_DIR", defaultDir(cfg.CacheDir, cfg.DataDir, "cache"))
	cfg.TempDir = envString("VYRDEN_TEMP_DIR", defaultDir(cfg.TempDir, cfg.DataDir, "temp"))
	cfg.MovieLibraryPath = envString("VYRDEN_MOVIES_PATH", cfg.MovieLibraryPath)
	cfg.TVLibraryPath = envString("VYRDEN_TV_PATH", cfg.TVLibraryPath)
	cfg.FFprobePath = envString("VYRDEN_FFPROBE_PATH", cfg.FFprobePath)
	cfg.FFmpegPath = envString("VYRDEN_FFMPEG_PATH", cfg.FFmpegPath)
	cfg.OMDbAPIKey = envString("VYRDEN_OMDB_API_KEY", cfg.OMDbAPIKey)
	cfg.TMDBAPIKey = envString("VYRDEN_TMDB_API_KEY", cfg.TMDBAPIKey)
	cfg.EventBuffer = envInt("VYRDEN_EVENT_BUFFER", cfg.EventBuffer)
	cfg.ScanWorkers = envInt("VYRDEN_SCAN_WORKERS", cfg.ScanWorkers)
	cfg.ProbeWorkers = envInt("VYRDEN_PROBE_WORKERS", cfg.ProbeWorkers)
	cfg.TranscodeWorkers = envInt("VYRDEN_TRANSCODE_WORKERS", cfg.TranscodeWorkers)
	cfg.GPUWorkers = envInt("VYRDEN_GPU_WORKERS", cfg.GPUWorkers)
	cfg.HardwareUnlocked = envBool("VYRDEN_HARDWARE_UNLOCKED", cfg.HardwareUnlocked)
	cfg.PlaybackPolicy = envString("VYRDEN_PLAYBACK_POLICY", defaultPlaybackPolicy(cfg.PlaybackPolicy))
	cfg.LibrarySyncMode = envString("VYRDEN_LIBRARY_SYNC_MODE", defaultSyncMode(cfg.LibrarySyncMode))
	cfg.SyncIntervalMins = envInt("VYRDEN_SYNC_INTERVAL_MINS", defaultInt(cfg.SyncIntervalMins, 1440))
	cfg.WatchDebounceSecs = envInt("VYRDEN_WATCH_DEBOUNCE_SECS", defaultInt(cfg.WatchDebounceSecs, 30))
	cfg.ProbeBatchLimit = envInt("VYRDEN_PROBE_BATCH_LIMIT", defaultInt(cfg.ProbeBatchLimit, 50))
	cfg.AllowedOrigins = envCSV("VYRDEN_ALLOWED_ORIGINS", cfg.AllowedOrigins)
	cfg.AuthDisabled = envBool("VYRDEN_AUTH_DISABLED", cfg.AuthDisabled)
	cfg.AdminUsername = envString("VYRDEN_ADMIN_USERNAME", cfg.AdminUsername)
	cfg.AdminPassword = envString("VYRDEN_ADMIN_PASSWORD", cfg.AdminPassword)
	return cfg
}

func LoadFile(dataDir string) (Config, error) {
	raw, err := os.ReadFile(filepath.Join(dataDir, "settings.json"))
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func SaveFile(dataDir string, cfg Config) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, "settings.json"), raw, 0o644)
}

func merge(base Config, saved Config) Config {
	if saved.HTTPAddr != "" {
		base.HTTPAddr = saved.HTTPAddr
	}
	if saved.DataDir != "" {
		base.DataDir = saved.DataDir
	}
	if saved.TranscodeDir != "" {
		base.TranscodeDir = saved.TranscodeDir
	}
	if saved.DownloadsDir != "" {
		base.DownloadsDir = saved.DownloadsDir
	}
	if saved.MetadataDir != "" {
		base.MetadataDir = saved.MetadataDir
	}
	if saved.CacheDir != "" {
		base.CacheDir = saved.CacheDir
	}
	if saved.TempDir != "" {
		base.TempDir = saved.TempDir
	}
	if saved.MovieLibraryPath != "" {
		base.MovieLibraryPath = saved.MovieLibraryPath
	}
	if saved.TVLibraryPath != "" {
		base.TVLibraryPath = saved.TVLibraryPath
	}
	if saved.FFprobePath != "" {
		base.FFprobePath = saved.FFprobePath
	}
	if saved.FFmpegPath != "" {
		base.FFmpegPath = saved.FFmpegPath
	}
	if saved.OMDbAPIKey != "" {
		base.OMDbAPIKey = saved.OMDbAPIKey
	}
	if saved.TMDBAPIKey != "" {
		base.TMDBAPIKey = saved.TMDBAPIKey
	}
	if saved.EventBuffer > 0 {
		base.EventBuffer = saved.EventBuffer
	}
	if saved.ScanWorkers > 0 {
		base.ScanWorkers = saved.ScanWorkers
	}
	if saved.ProbeWorkers > 0 {
		base.ProbeWorkers = saved.ProbeWorkers
	}
	if saved.TranscodeWorkers > 0 {
		base.TranscodeWorkers = saved.TranscodeWorkers
	}
	if saved.GPUWorkers > 0 {
		base.GPUWorkers = saved.GPUWorkers
	}
	if saved.HardwareUnlocked {
		base.HardwareUnlocked = saved.HardwareUnlocked
	}
	if saved.PlaybackPolicy != "" {
		base.PlaybackPolicy = saved.PlaybackPolicy
	}
	if saved.LibrarySyncMode != "" {
		base.LibrarySyncMode = saved.LibrarySyncMode
	}
	if saved.SyncIntervalMins > 0 {
		base.SyncIntervalMins = saved.SyncIntervalMins
	}
	if saved.WatchDebounceSecs > 0 {
		base.WatchDebounceSecs = saved.WatchDebounceSecs
	}
	if saved.ProbeBatchLimit > 0 {
		base.ProbeBatchLimit = saved.ProbeBatchLimit
	}
	if len(saved.AllowedOrigins) > 0 {
		base.AllowedOrigins = saved.AllowedOrigins
	}
	return base
}

func defaultSyncMode(value string) string {
	switch value {
	case "manual", "daily", "watch":
		return value
	default:
		return "daily"
	}
}

func defaultPlaybackPolicy(value string) string {
	switch value {
	case "original_only", "light", "full", "cinema":
		return value
	default:
		return "original_only"
	}
}

func defaultInt(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func defaultDir(value string, dataDir string, name string) string {
	if value != "" {
		return value
	}
	return filepath.Join(dataDir, name)
}

func envString(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envCSV(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	output := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			output = append(output, trimmed)
		}
	}
	return output
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
