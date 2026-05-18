package config

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const legacyDefaultServerName = "My Server"

type Config struct {
	ServerName            string   `json:"serverName,omitempty"`
	HTTPAddr              string   `json:"httpAddr"`
	DiscoveryEnabled      bool     `json:"-"`
	DiscoveryServiceType  string   `json:"-"`
	// DataDir is resolved at startup and never persisted in settings.json
	// (since the file itself lives inside DataDir). Override via XUVA_DATA_DIR.
	DataDir               string   `json:"-"`
	TranscodeDir          string   `json:"transcodeDir,omitempty"`
	DownloadsDir          string   `json:"downloadsDir,omitempty"`
	MetadataDir           string   `json:"metadataDir,omitempty"`
	CacheDir              string   `json:"cacheDir,omitempty"`
	TempDir               string   `json:"tempDir,omitempty"`
	MovieLibraryPath      string   `json:"movieLibraryPath,omitempty"`
	TVLibraryPath         string   `json:"tvLibraryPath,omitempty"`
	MovieMetadataSources  []string `json:"movieMetadataSources,omitempty"`
	SeriesMetadataSources []string `json:"seriesMetadataSources,omitempty"`
	MovieArtworkSources   []string `json:"movieArtworkSources,omitempty"`
	SeriesArtworkSources  []string `json:"seriesArtworkSources,omitempty"`
	FFprobePath           string   `json:"ffprobePath"`
	FFmpegPath            string   `json:"ffmpegPath"`
	OMDbAPIKey            string   `json:"omdbApiKey,omitempty"`
	TMDBAPIKey            string   `json:"tmdbApiKey,omitempty"`
	TVDBAPIKey            string   `json:"tvdbApiKey,omitempty"`
	FanartTVAPIKey        string   `json:"fanartTvApiKey,omitempty"`
	EventBuffer           int      `json:"eventBuffer"`
	ScanWorkers           int      `json:"scanWorkers"`
	ProbeWorkers          int      `json:"probeWorkers"`
	TranscodeWorkers      int      `json:"transcodeWorkers"`
	GPUWorkers            int      `json:"gpuWorkers"`
	HardwareUnlocked      bool     `json:"hardwareUnlocked,omitempty"`
	PlaybackPolicy        string   `json:"playbackPolicy,omitempty"`
	LibrarySyncMode       string   `json:"librarySyncMode,omitempty"`
	SyncIntervalMins      int      `json:"syncIntervalMins,omitempty"`
	WatchDebounceSecs     int      `json:"watchDebounceSecs,omitempty"`
	ProbeBatchLimit       int      `json:"probeBatchLimit,omitempty"`
	AllowedOrigins        []string `json:"allowedOrigins,omitempty"`
	AuthDisabled          bool     `json:"-"`
	AdminUsername         string   `json:"-"`
	AdminPassword         string   `json:"-"`
}

// defaultDataDir resolves the data directory to a stable absolute path that
// does not depend on the process working directory.
//
//   - In development (`go run …`), walk up from cwd until a `go.mod` is found
//     and anchor `data/` to that module root. This keeps `server/data/` no
//     matter whether you launched from `server/` or `server/cmd/Xuva/`.
//   - For a built binary, fall back to `<exe-dir>/data`.
//   - As a last resort, return the literal "data" so the legacy behaviour
//     (cwd-relative) still works.
//
// Always overridable via the XUVA_DATA_DIR environment variable.
func defaultDataDir() string {
	if cwd, err := os.Getwd(); err == nil {
		dir := cwd
		for i := 0; i < 12; i++ {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return filepath.Join(dir, "data")
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	if exe, err := os.Executable(); err == nil {
		// Skip `go run` temp executables; their dir is unstable.
		exeDir := filepath.Dir(exe)
		if !strings.Contains(strings.ToLower(exeDir), string(filepath.Separator)+"go-build") {
			return filepath.Join(exeDir, "data")
		}
	}
	return "data"
}

func FromEnv() Config {
	dataDir := envString("XUVA_DATA_DIR", defaultDataDir())
	cfg := Config{
		ServerName:           envString("XUVA_SERVER_NAME", "Xuva"),
		HTTPAddr:             envString("XUVA_HTTP_ADDR", "127.0.0.1:8097"),
		DiscoveryEnabled:     envBool("XUVA_DISCOVERY_ENABLED", true),
		DiscoveryServiceType: envString("XUVA_DISCOVERY_SERVICE_TYPE", "_xuva._tcp"),
		DataDir:              dataDir,
		TranscodeDir:         envString("XUVA_TRANSCODE_DIR", filepath.Join(dataDir, "transcode")),
		DownloadsDir:         envString("XUVA_DOWNLOADS_DIR", filepath.Join(dataDir, "downloads")),
		MetadataDir:          envString("XUVA_METADATA_DIR", filepath.Join(dataDir, "metadata")),
		CacheDir:             envString("XUVA_CACHE_DIR", filepath.Join(dataDir, "cache")),
		TempDir:              envString("XUVA_TEMP_DIR", filepath.Join(dataDir, "temp")),
		MovieLibraryPath:     envString("XUVA_MOVIES_PATH", ""),
		TVLibraryPath:        envString("XUVA_TV_PATH", ""),
		FFprobePath:          envString("XUVA_FFPROBE_PATH", "ffprobe"),
		FFmpegPath:           envString("XUVA_FFMPEG_PATH", "ffmpeg"),
		OMDbAPIKey:           envString("XUVA_OMDB_API_KEY", ""),
		TMDBAPIKey:           envString("XUVA_TMDB_API_KEY", ""),
		TVDBAPIKey:           envString("XUVA_TVDB_API_KEY", ""),
		FanartTVAPIKey:       envString("XUVA_FANARTTV_API_KEY", ""),
		EventBuffer:          envInt("XUVA_EVENT_BUFFER", 128),
		ScanWorkers:          envInt("XUVA_SCAN_WORKERS", 1),
		ProbeWorkers:         envInt("XUVA_PROBE_WORKERS", 2),
		TranscodeWorkers:     envInt("XUVA_TRANSCODE_WORKERS", 1),
		GPUWorkers:           envInt("XUVA_GPU_WORKERS", 1),
		HardwareUnlocked:     envBool("XUVA_HARDWARE_UNLOCKED", false),
		PlaybackPolicy:       envString("XUVA_PLAYBACK_POLICY", "original_only"),
		LibrarySyncMode:      envString("XUVA_LIBRARY_SYNC_MODE", "daily"),
		SyncIntervalMins:     envInt("XUVA_SYNC_INTERVAL_MINS", 1440),
		WatchDebounceSecs:    envInt("XUVA_WATCH_DEBOUNCE_SECS", 30),
		ProbeBatchLimit:      envInt("XUVA_PROBE_BATCH_LIMIT", 50),
		AllowedOrigins:       envCSV("XUVA_ALLOWED_ORIGINS", nil),
		AuthDisabled:         envBool("XUVA_AUTH_DISABLED", false),
		AdminUsername:        envString("XUVA_ADMIN_USERNAME", "admin"),
		AdminPassword:        envString("XUVA_ADMIN_PASSWORD", ""),
	}
	if saved, err := LoadFile(dataDir); err == nil {
		cfg = merge(cfg, saved)
	}
	cfg.HTTPAddr = envString("XUVA_HTTP_ADDR", cfg.HTTPAddr)
	cfg.DiscoveryEnabled = envBool("XUVA_DISCOVERY_ENABLED", cfg.DiscoveryEnabled)
	cfg.DiscoveryServiceType = envString("XUVA_DISCOVERY_SERVICE_TYPE", defaultDiscoveryServiceType(cfg.DiscoveryServiceType))
	cfg.ServerName = envString("XUVA_SERVER_NAME", defaultServerName(cfg.ServerName))
	cfg.DataDir = envString("XUVA_DATA_DIR", cfg.DataDir)
	cfg.TranscodeDir = envString("XUVA_TRANSCODE_DIR", defaultDir(cfg.TranscodeDir, cfg.DataDir, "transcode"))
	cfg.DownloadsDir = envString("XUVA_DOWNLOADS_DIR", defaultDir(cfg.DownloadsDir, cfg.DataDir, "downloads"))
	cfg.MetadataDir = envString("XUVA_METADATA_DIR", defaultDir(cfg.MetadataDir, cfg.DataDir, "metadata"))
	cfg.CacheDir = envString("XUVA_CACHE_DIR", defaultDir(cfg.CacheDir, cfg.DataDir, "cache"))
	cfg.TempDir = envString("XUVA_TEMP_DIR", defaultDir(cfg.TempDir, cfg.DataDir, "temp"))
	cfg.MovieLibraryPath = envString("XUVA_MOVIES_PATH", cfg.MovieLibraryPath)
	cfg.TVLibraryPath = envString("XUVA_TV_PATH", cfg.TVLibraryPath)
	cfg.FFprobePath = envString("XUVA_FFPROBE_PATH", cfg.FFprobePath)
	cfg.FFmpegPath = envString("XUVA_FFMPEG_PATH", cfg.FFmpegPath)
	cfg.OMDbAPIKey = envString("XUVA_OMDB_API_KEY", cfg.OMDbAPIKey)
	cfg.TMDBAPIKey = envString("XUVA_TMDB_API_KEY", cfg.TMDBAPIKey)
	cfg.TVDBAPIKey = envString("XUVA_TVDB_API_KEY", cfg.TVDBAPIKey)
	cfg.FanartTVAPIKey = envString("XUVA_FANARTTV_API_KEY", cfg.FanartTVAPIKey)
	cfg.EventBuffer = envInt("XUVA_EVENT_BUFFER", cfg.EventBuffer)
	cfg.ScanWorkers = envInt("XUVA_SCAN_WORKERS", cfg.ScanWorkers)
	cfg.ProbeWorkers = envInt("XUVA_PROBE_WORKERS", cfg.ProbeWorkers)
	cfg.TranscodeWorkers = envInt("XUVA_TRANSCODE_WORKERS", cfg.TranscodeWorkers)
	cfg.GPUWorkers = envInt("XUVA_GPU_WORKERS", cfg.GPUWorkers)
	cfg.HardwareUnlocked = envBool("XUVA_HARDWARE_UNLOCKED", cfg.HardwareUnlocked)
	cfg.PlaybackPolicy = envString("XUVA_PLAYBACK_POLICY", defaultPlaybackPolicy(cfg.PlaybackPolicy))
	cfg.LibrarySyncMode = envString("XUVA_LIBRARY_SYNC_MODE", defaultSyncMode(cfg.LibrarySyncMode))
	cfg.SyncIntervalMins = envInt("XUVA_SYNC_INTERVAL_MINS", defaultInt(cfg.SyncIntervalMins, 1440))
	cfg.WatchDebounceSecs = envInt("XUVA_WATCH_DEBOUNCE_SECS", defaultInt(cfg.WatchDebounceSecs, 30))
	cfg.ProbeBatchLimit = envInt("XUVA_PROBE_BATCH_LIMIT", defaultInt(cfg.ProbeBatchLimit, 50))
	cfg.AllowedOrigins = envCSV("XUVA_ALLOWED_ORIGINS", cfg.AllowedOrigins)
	cfg.AuthDisabled = envBool("XUVA_AUTH_DISABLED", cfg.AuthDisabled)
	cfg.AdminUsername = envString("XUVA_ADMIN_USERNAME", cfg.AdminUsername)
	cfg.AdminPassword = envString("XUVA_ADMIN_PASSWORD", cfg.AdminPassword)
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
	if saved.ServerName != "" {
		base.ServerName = saved.ServerName
	}
	if saved.HTTPAddr != "" {
		base.HTTPAddr = saved.HTTPAddr
	}
	// DataDir is intentionally never merged from saved settings — it's
	// resolved at startup and json:"-" guarantees saved.DataDir is "".
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
	if len(saved.MovieMetadataSources) > 0 {
		base.MovieMetadataSources = append([]string(nil), saved.MovieMetadataSources...)
	}
	if len(saved.SeriesMetadataSources) > 0 {
		base.SeriesMetadataSources = append([]string(nil), saved.SeriesMetadataSources...)
	}
	if len(saved.MovieArtworkSources) > 0 {
		base.MovieArtworkSources = append([]string(nil), saved.MovieArtworkSources...)
	}
	if len(saved.SeriesArtworkSources) > 0 {
		base.SeriesArtworkSources = append([]string(nil), saved.SeriesArtworkSources...)
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
	if saved.TVDBAPIKey != "" {
		base.TVDBAPIKey = saved.TVDBAPIKey
	}
	if saved.FanartTVAPIKey != "" {
		base.FanartTVAPIKey = saved.FanartTVAPIKey
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

func Merge(base Config, saved Config) Config {
	return merge(base, saved)
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

func defaultServerName(value string) string {
	if strings.TrimSpace(value) == legacyDefaultServerName {
		return "Xuva"
	}
	if normalized, err := NormalizeServerName(value); err == nil {
		return normalized
	}
	return "Xuva"
}

func defaultDiscoveryServiceType(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "_xuva._tcp"
	}
	return trimmed
}

func NormalizeServerName(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errors.New("server name is required")
	}
	if utf8.RuneCountInString(trimmed) > 50 {
		return "", errors.New("server name must be 50 characters or fewer")
	}
	for _, char := range trimmed {
		if unicode.IsControl(char) {
			return "", errors.New("server name cannot include control characters")
		}
	}
	return trimmed, nil
}

func defaultInt(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

// defaultDir resolves a configured storage directory:
//   - empty value          -> <dataDir>/<name>
//   - absolute value       -> used as-is
//   - relative value       -> re-anchored to dataDir using just the leaf
//     component. This protects against historical settings.json files that
//     stored cwd-relative paths like "data\\transcode" which would otherwise
//     create stray empty directories under whichever cwd the server started in.
func defaultDir(value string, dataDir string, name string) string {
	if value == "" {
		return filepath.Join(dataDir, name)
	}
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(dataDir, filepath.Base(value))
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

func HTTPAddrLoopbackOnly(addr string) bool {
	return loopbackHTTPAddr(addr)
}

func loopbackHTTPAddr(addr string) bool {
	host := strings.TrimSpace(addr)
	if host == "" {
		return false
	}
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}
	host = strings.TrimSpace(strings.Trim(host, "[]"))
	if host == "" {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}
