package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const legacyDefaultServerName = "My Server"

type Config struct {
	ServerName           string `json:"serverName,omitempty"`
	HTTPAddr             string `json:"httpAddr"`
	CanonicalWebOrigin   string `json:"canonicalWebOrigin,omitempty"`
	DiscoveryEnabled     bool   `json:"-"`
	DiscoveryServiceType string `json:"-"`
	// DataDir is resolved at startup and never persisted in settings.json
	// (since the file itself lives inside DataDir). Override via XUVA_DATA_DIR.
	RuntimeHome           string   `json:"-"`
	RuntimeScope          string   `json:"-"`
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
	FpcalcPath            string   `json:"fpcalcPath,omitempty"` // path to fpcalc binary; empty = intro detection disabled
	// API keys: when blank, the four-tier resolver in keys.go falls back to
	// env vars then to build-time embedded defaults. End users do not need
	// to populate these — they exist only as power-user overrides for cases
	// like hitting the shared rate limit.
	OMDbAPIKey     string `json:"omdbApiKey,omitempty"`
	TMDBAPIKey     string `json:"tmdbApiKey,omitempty"`
	FanartTVAPIKey string `json:"fanartTvApiKey,omitempty"`
	// TVDB support has been dropped: their v4 licence requires a per-install
	// subscription, which is incompatible with embed-and-ship UX. TMDB
	// supplies full TV data (episodes, seasons, stills, credits, ratings).
	// The field name is preserved on settings.json to allow forward-clean
	// migration of any existing user file, but is no longer read.
	EventBuffer       int    `json:"eventBuffer"`
	ScanWorkers       int    `json:"scanWorkers"`
	ProbeWorkers      int    `json:"probeWorkers"`
	TranscodeWorkers  int    `json:"transcodeWorkers"`
	GPUWorkers        int    `json:"gpuWorkers"`
	HardwareUnlocked  bool   `json:"hardwareUnlocked,omitempty"`
	PlaybackPolicy    string `json:"playbackPolicy,omitempty"`
	LibrarySyncMode   string `json:"librarySyncMode,omitempty"`
	SyncIntervalMins  int    `json:"syncIntervalMins,omitempty"`
	WatchDebounceSecs int    `json:"watchDebounceSecs,omitempty"`
	ProbeBatchLimit   int    `json:"probeBatchLimit,omitempty"`
	// Per-job automation controls. "Disable..." naming keeps the zero value
	// meaning "enabled", which is backwards-compatible with older settings.json
	// files that predate these fields.
	DisableScanAuto      bool     `json:"disableScanAuto,omitempty"`      // set true to stop automated library scans
	ScanIntervalMins     int      `json:"scanIntervalMins,omitempty"`     // scan cadence; default 15 min, min 5
	DisableMetadataAuto  bool     `json:"disableMetadataAuto,omitempty"`  // set true to stop automated metadata backfill
	MetadataIntervalMins int      `json:"metadataIntervalMins,omitempty"` // metadata cadence; default 360 min (6 h)
	MetadataBatchLimit   int      `json:"metadataBatchLimit,omitempty"`   // 0 = unlimited
	DisableProbeAuto     bool     `json:"disableProbeAuto,omitempty"`     // set true to stop auto-probe after scan
	DisableJobPause      bool     `json:"disableJobPause,omitempty"`      // set true to run jobs even during playback
	AllowedOrigins       []string `json:"allowedOrigins,omitempty"`
	// Region / language settings — captured during the setup wizard.
	Country          string `json:"country,omitempty"`          // ISO 3166-1 alpha-2, e.g. "AU"
	Timezone         string `json:"timezone,omitempty"`         // IANA tz, e.g. "Australia/Sydney"
	MetadataLanguage string `json:"metadataLanguage,omitempty"` // BCP-47 e.g. "en-US", "fr-FR", "de-DE"
	// Playback preferences
	PreferTextSubtitles    bool `json:"preferTextSubtitles,omitempty"`    // prefer SRT/ASS over bitmap subs
	OriginalQualityOnly    bool `json:"originalQualityOnly,omitempty"`    // refuse to transcode video
	DefaultSubtitlesMovies bool `json:"defaultSubtitlesMovies,omitempty"` // start movie playback with subs on
	DefaultSubtitlesTV     bool `json:"defaultSubtitlesTV,omitempty"`     // start TV-episode playback with subs on
	// DisableTrailers suppresses trailer autoplay in the hero carousel. When
	// true the hero always shows the static backdrop image regardless of whether
	// trailer data is available. False (default/zero) means trailers are allowed.
	DisableTrailers bool `json:"disableTrailers,omitempty"`
	SetupComplete   bool `json:"setupComplete,omitempty"`
	// Trailer downloader settings — self-hosted preview videos.
	TrailersEnabled bool   `json:"trailersEnabled,omitempty"`
	TrailersDir     string `json:"trailersDir,omitempty"` // local MP4 cache
	YTDLPPath       string `json:"ytdlpPath,omitempty"`   // yt-dlp binary, defaults to PATH lookup
	TrailerWorkers  int    `json:"trailerWorkers,omitempty"`
	AuthDisabled    bool   `json:"-"`
	AdminUsername   string `json:"-"`
	AdminPassword   string `json:"-"`
	// Logging config — env-only, not persisted in settings.json.
	// LogFormat: "text" (default) or "json" (machine-readable, e.g. for Loki/Datadog)
	// LogLevel: "debug", "info" (default), "warn", "error"
	LogFormat   string `json:"-"`
	LogLevel    string `json:"-"`
	LogDir      string `json:"-"`
	LogMaxMB    int    `json:"-"`
	LogMaxFiles int    `json:"-"`
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
	runtimeHome := strings.TrimSpace(os.Getenv("XUVA_RUNTIME_HOME"))
	runtimeRoot := runtimeHome
	dataDirDefault := defaultDataDir()
	if runtimeRoot != "" {
		dataDirDefault = filepath.Join(runtimeRoot, "data")
	}
	dataDir := envString("XUVA_DATA_DIR", dataDirDefault)
	runtimeDir := func(name string) string {
		if runtimeRoot != "" {
			return filepath.Join(runtimeRoot, name)
		}
		return filepath.Join(dataDir, name)
	}
	cfg := Config{
		ServerName:           envString("XUVA_SERVER_NAME", "Xuva"),
		// Default bind to all interfaces so the personal-media-server use case
		// works out of the box on a LAN. Users who want loopback-only can set
		// XUVA_HTTP_ADDR=127.0.0.1:8097 explicitly. Previously we defaulted to
		// loopback for safety, but that broke autodiscovery (mDNS advertised a
		// LAN IP that the server wasn't actually listening on) and remote LAN
		// browsers, surprising users who never intended a single-machine setup.
		HTTPAddr:             envString("XUVA_HTTP_ADDR", "0.0.0.0:8097"),
		CanonicalWebOrigin:   envString("XUVA_CANONICAL_WEB_ORIGIN", ""),
		DiscoveryEnabled:     envBool("XUVA_DISCOVERY_ENABLED", true),
		DiscoveryServiceType: envString("XUVA_DISCOVERY_SERVICE_TYPE", "_xuva._tcp"),
		RuntimeHome:          runtimeHome,
		RuntimeScope:         envString("XUVA_RUNTIME_SCOPE", ""),
		DataDir:              dataDir,
		TranscodeDir:         envString("XUVA_TRANSCODE_DIR", runtimeDir("transcode")),
		DownloadsDir:         envString("XUVA_DOWNLOADS_DIR", runtimeDir("downloads")),
		MetadataDir:          envString("XUVA_METADATA_DIR", runtimeDir("metadata")),
		CacheDir:             envString("XUVA_CACHE_DIR", runtimeDir("cache")),
		TempDir:              envString("XUVA_TEMP_DIR", runtimeDir("temp")),
		MovieLibraryPath:     envString("XUVA_MOVIES_PATH", ""),
		TVLibraryPath:        envString("XUVA_TV_PATH", ""),
		FFprobePath:          envString("XUVA_FFPROBE_PATH", "ffprobe"),
		FFmpegPath:           envString("XUVA_FFMPEG_PATH", "ffmpeg"),
		FpcalcPath:           envString("XUVA_FPCALC_PATH", ""),
		// Keys default empty here; ResolveProviderKey() in keys.go merges
		// saved + env + embedded after settings.json is loaded below.
		OMDbAPIKey:             "",
		TMDBAPIKey:             "",
		FanartTVAPIKey:         "",
		EventBuffer:            envInt("XUVA_EVENT_BUFFER", 128),
		ScanWorkers:            envInt("XUVA_SCAN_WORKERS", 1),
		ProbeWorkers:           envInt("XUVA_PROBE_WORKERS", 2),
		TranscodeWorkers:       envInt("XUVA_TRANSCODE_WORKERS", 1),
		GPUWorkers:             envInt("XUVA_GPU_WORKERS", 1),
		HardwareUnlocked:       envBool("XUVA_HARDWARE_UNLOCKED", false),
		PlaybackPolicy:         envString("XUVA_PLAYBACK_POLICY", "original_only"),
		LibrarySyncMode:        envString("XUVA_LIBRARY_SYNC_MODE", "daily"),
		SyncIntervalMins:       envInt("XUVA_SYNC_INTERVAL_MINS", 1440),
		WatchDebounceSecs:      envInt("XUVA_WATCH_DEBOUNCE_SECS", 30),
		ProbeBatchLimit:        envInt("XUVA_PROBE_BATCH_LIMIT", 0),
		ScanIntervalMins:       envInt("XUVA_SCAN_INTERVAL_MINS", 15),
		MetadataIntervalMins:   envInt("XUVA_METADATA_INTERVAL_MINS", 360),
		Country:                envString("XUVA_COUNTRY", ""),
		Timezone:               envString("XUVA_TIMEZONE", ""),
		MetadataLanguage:       envString("XUVA_METADATA_LANGUAGE", "en-US"),
		PreferTextSubtitles:    envBool("XUVA_PREFER_TEXT_SUBTITLES", false),
		OriginalQualityOnly:    envBool("XUVA_ORIGINAL_QUALITY_ONLY", false),
		DefaultSubtitlesMovies: envBool("XUVA_DEFAULT_SUBTITLES_MOVIES", false),
		DefaultSubtitlesTV:     envBool("XUVA_DEFAULT_SUBTITLES_TV", false),
		TrailersEnabled:        envBool("XUVA_TRAILERS_ENABLED", true),
		TrailersDir:            envString("XUVA_TRAILERS_DIR", runtimeDir("trailers")),
		YTDLPPath:              envString("XUVA_YTDLP_PATH", "yt-dlp"),
		TrailerWorkers:         envInt("XUVA_TRAILER_WORKERS", 1),
		AllowedOrigins:         envCSV("XUVA_ALLOWED_ORIGINS", nil),
		AuthDisabled:           envBool("XUVA_AUTH_DISABLED", false),
		AdminUsername:          envString("XUVA_ADMIN_USERNAME", "admin"),
		AdminPassword:          envString("XUVA_ADMIN_PASSWORD", ""),
		LogFormat:              envString("XUVA_LOG_FORMAT", "text"),
		LogLevel:               envString("XUVA_LOG_LEVEL", "info"),
		LogDir:                 envString("XUVA_LOG_DIR", runtimeDir("logs")),
		LogMaxMB:               envInt("XUVA_LOG_MAX_MB", 25),
		LogMaxFiles:            envInt("XUVA_LOG_MAX_FILES", 5),
	}
	if saved, err := LoadFile(dataDir); err == nil {
		cfg = merge(cfg, saved)
	}
	cfg.RuntimeHome = envString("XUVA_RUNTIME_HOME", cfg.RuntimeHome)
	cfg.RuntimeScope = envString("XUVA_RUNTIME_SCOPE", cfg.RuntimeScope)
	runtimeRoot = cfg.RuntimeHome
	cfg.HTTPAddr = envString("XUVA_HTTP_ADDR", cfg.HTTPAddr)
	cfg.CanonicalWebOrigin = envString("XUVA_CANONICAL_WEB_ORIGIN", cfg.CanonicalWebOrigin)
	cfg.DiscoveryEnabled = envBool("XUVA_DISCOVERY_ENABLED", cfg.DiscoveryEnabled)
	cfg.DiscoveryServiceType = envString("XUVA_DISCOVERY_SERVICE_TYPE", defaultDiscoveryServiceType(cfg.DiscoveryServiceType))
	cfg.ServerName = envString("XUVA_SERVER_NAME", defaultServerName(cfg.ServerName))
	cfg.DataDir = envString("XUVA_DATA_DIR", cfg.DataDir)
	cfg.TranscodeDir = envString("XUVA_TRANSCODE_DIR", defaultDir(cfg.TranscodeDir, runtimeRoot, cfg.DataDir, "transcode"))
	cfg.DownloadsDir = envString("XUVA_DOWNLOADS_DIR", defaultDir(cfg.DownloadsDir, runtimeRoot, cfg.DataDir, "downloads"))
	cfg.MetadataDir = envString("XUVA_METADATA_DIR", defaultDir(cfg.MetadataDir, runtimeRoot, cfg.DataDir, "metadata"))
	cfg.CacheDir = envString("XUVA_CACHE_DIR", defaultDir(cfg.CacheDir, runtimeRoot, cfg.DataDir, "cache"))
	cfg.TempDir = envString("XUVA_TEMP_DIR", defaultDir(cfg.TempDir, runtimeRoot, cfg.DataDir, "temp"))
	cfg.MovieLibraryPath = envString("XUVA_MOVIES_PATH", cfg.MovieLibraryPath)
	cfg.TVLibraryPath = envString("XUVA_TV_PATH", cfg.TVLibraryPath)
	cfg.FFprobePath = envString("XUVA_FFPROBE_PATH", cfg.FFprobePath)
	cfg.FFmpegPath = envString("XUVA_FFMPEG_PATH", cfg.FFmpegPath)
	cfg.FpcalcPath = envString("XUVA_FPCALC_PATH", cfg.FpcalcPath)
	// Four-tier resolution: settings.json → env → embedded build-time → empty.
	// We pass cfg.* (already merged from saved settings) as `saved`, the env
	// var explicitly, and the embedded default from keys.go.
	cfg.TMDBAPIKey, _ = ResolveProviderKey(cfg.TMDBAPIKey, envString("XUVA_TMDB_API_KEY", ""), DefaultTMDBAPIKey)
	cfg.FanartTVAPIKey, _ = ResolveProviderKey(cfg.FanartTVAPIKey, envString("XUVA_FANARTTV_API_KEY", ""), DefaultFanartTVAPIKey)
	cfg.OMDbAPIKey, _ = ResolveProviderKey(cfg.OMDbAPIKey, envString("XUVA_OMDB_API_KEY", ""), DefaultOMDbAPIKey)

	// Backfill source-order defaults whenever they're empty (covers fresh
	// installs AND legacy installs from before the auto-default landed).
	// The actual default lists live in package metasources to avoid an
	// import cycle — config doesn't depend on metasources, so we duplicate
	// the canonical order here. If you change one, change the other.
	if len(cfg.MovieMetadataSources) == 0 {
		cfg.MovieMetadataSources = defaultMovieMetadataSources()
	}
	if len(cfg.SeriesMetadataSources) == 0 {
		cfg.SeriesMetadataSources = defaultSeriesMetadataSources()
	}
	if len(cfg.MovieArtworkSources) == 0 {
		cfg.MovieArtworkSources = defaultArtworkSources()
	}
	if len(cfg.SeriesArtworkSources) == 0 {
		cfg.SeriesArtworkSources = defaultArtworkSources()
	}
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
	// ProbeBatchLimit: 0 means unlimited; do not apply defaultInt (which would turn 0 into 50).
	cfg.ProbeBatchLimit = envInt("XUVA_PROBE_BATCH_LIMIT", cfg.ProbeBatchLimit)
	cfg.ScanIntervalMins = envInt("XUVA_SCAN_INTERVAL_MINS", defaultInt(cfg.ScanIntervalMins, 15))
	cfg.MetadataIntervalMins = envInt("XUVA_METADATA_INTERVAL_MINS", defaultInt(cfg.MetadataIntervalMins, 360))
	cfg.Country = envString("XUVA_COUNTRY", cfg.Country)
	cfg.Timezone = envString("XUVA_TIMEZONE", cfg.Timezone)
	cfg.MetadataLanguage = envString("XUVA_METADATA_LANGUAGE", defaultString(cfg.MetadataLanguage, "en-US"))
	cfg.PreferTextSubtitles = envBool("XUVA_PREFER_TEXT_SUBTITLES", cfg.PreferTextSubtitles)
	cfg.OriginalQualityOnly = envBool("XUVA_ORIGINAL_QUALITY_ONLY", cfg.OriginalQualityOnly)
	cfg.DefaultSubtitlesMovies = envBool("XUVA_DEFAULT_SUBTITLES_MOVIES", cfg.DefaultSubtitlesMovies)
	cfg.DefaultSubtitlesTV = envBool("XUVA_DEFAULT_SUBTITLES_TV", cfg.DefaultSubtitlesTV)
	cfg.TrailersEnabled = envBool("XUVA_TRAILERS_ENABLED", cfg.TrailersEnabled)
	cfg.TrailersDir = envString("XUVA_TRAILERS_DIR", defaultDir(cfg.TrailersDir, runtimeRoot, cfg.DataDir, "trailers"))
	cfg.YTDLPPath = envString("XUVA_YTDLP_PATH", defaultString(cfg.YTDLPPath, "yt-dlp"))
	cfg.TrailerWorkers = envInt("XUVA_TRAILER_WORKERS", defaultInt(cfg.TrailerWorkers, 1))
	cfg.AllowedOrigins = envCSV("XUVA_ALLOWED_ORIGINS", cfg.AllowedOrigins)
	cfg.AuthDisabled = envBool("XUVA_AUTH_DISABLED", cfg.AuthDisabled)
	cfg.AdminUsername = envString("XUVA_ADMIN_USERNAME", cfg.AdminUsername)
	cfg.AdminPassword = envString("XUVA_ADMIN_PASSWORD", cfg.AdminPassword)
	cfg.LogFormat = envString("XUVA_LOG_FORMAT", cfg.LogFormat)
	cfg.LogLevel = envString("XUVA_LOG_LEVEL", cfg.LogLevel)
	cfg.LogDir = envString("XUVA_LOG_DIR", defaultDir(cfg.LogDir, runtimeRoot, cfg.DataDir, "logs"))
	cfg.LogMaxMB = envInt("XUVA_LOG_MAX_MB", defaultInt(cfg.LogMaxMB, 25))
	cfg.LogMaxFiles = envInt("XUVA_LOG_MAX_FILES", defaultInt(cfg.LogMaxFiles, 5))
	return cfg
}

func LoadFile(dataDir string) (Config, error) {
	raw, err := os.ReadFile(filepath.Join(dataDir, "settings.json"))
	if err != nil {
		return Config{}, err
	}
	// Be tolerant of UTF-8 BOM files produced by some Windows editors/shell
	// commands so we don't silently fall back to defaults like serverName.
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
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
	if saved.CanonicalWebOrigin != "" {
		base.CanonicalWebOrigin = saved.CanonicalWebOrigin
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
	if saved.FpcalcPath != "" {
		base.FpcalcPath = saved.FpcalcPath
	}
	if saved.OMDbAPIKey != "" {
		base.OMDbAPIKey = saved.OMDbAPIKey
	}
	if saved.TMDBAPIKey != "" {
		base.TMDBAPIKey = saved.TMDBAPIKey
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
	if saved.DisableScanAuto {
		base.DisableScanAuto = true
	}
	if saved.ScanIntervalMins > 0 {
		base.ScanIntervalMins = saved.ScanIntervalMins
	}
	if saved.DisableMetadataAuto {
		base.DisableMetadataAuto = true
	}
	if saved.MetadataIntervalMins > 0 {
		base.MetadataIntervalMins = saved.MetadataIntervalMins
	}
	if saved.MetadataBatchLimit > 0 {
		base.MetadataBatchLimit = saved.MetadataBatchLimit
	}
	if saved.DisableProbeAuto {
		base.DisableProbeAuto = true
	}
	if saved.DisableJobPause {
		base.DisableJobPause = true
	}
	if len(saved.AllowedOrigins) > 0 {
		base.AllowedOrigins = saved.AllowedOrigins
	}
	if saved.Country != "" {
		base.Country = saved.Country
	}
	if saved.Timezone != "" {
		base.Timezone = saved.Timezone
	}
	if saved.MetadataLanguage != "" {
		base.MetadataLanguage = saved.MetadataLanguage
	}
	if saved.PreferTextSubtitles {
		base.PreferTextSubtitles = true
	}
	if saved.OriginalQualityOnly {
		base.OriginalQualityOnly = true
	}
	if saved.DefaultSubtitlesMovies {
		base.DefaultSubtitlesMovies = true
	}
	if saved.DefaultSubtitlesTV {
		base.DefaultSubtitlesTV = true
	}
	if saved.SetupComplete {
		base.SetupComplete = true
	}
	if saved.TrailersEnabled {
		base.TrailersEnabled = true
	}
	if saved.TrailersDir != "" {
		base.TrailersDir = saved.TrailersDir
	}
	if saved.YTDLPPath != "" {
		base.YTDLPPath = saved.YTDLPPath
	}
	if saved.TrailerWorkers > 0 {
		base.TrailerWorkers = saved.TrailerWorkers
	}
	return base
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

// defaultMovieMetadataSources / defaultSeriesMetadataSources / defaultArtworkSources
// duplicate the canonical lists from package metasources to avoid an import
// cycle (metasources imports config). Keep these in sync with
// metasources.DefaultSourceOrder() and metasources.DefaultArtworkOrder().
//
// TVDB is intentionally absent — it's a disabled provider in this build.
func defaultMovieMetadataSources() []string {
	return []string{"nfo", "tmdb", "wikipedia", "wikidata", "omdb", "filename"}
}

func defaultSeriesMetadataSources() []string {
	return []string{"nfo", "tmdb", "tvmaze", "wikipedia", "wikidata", "omdb", "filename"}
}

func defaultArtworkSources() []string {
	return []string{"artwork", "nfo", "fanart", "tmdb", "wikipedia", "wikidata"}
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

func NormalizeWebOrigin(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("canonical web origin must be an absolute http:// or https:// URL")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", errors.New("canonical web origin must start with http:// or https://")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("canonical web origin cannot include username, query, or fragment")
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", errors.New("canonical web origin cannot include a path")
	}
	parsed.Scheme = scheme
	parsed.Path = ""
	parsed.RawPath = ""
	parsed.ForceQuery = false
	return strings.TrimRight(parsed.String(), "/"), nil
}

func defaultInt(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

// defaultDir resolves a configured storage directory:
//   - empty value    -> <runtimeHome>/<name> when set, else <dataDir>/<name>
//   - absolute value -> used as-is
//   - relative value -> re-anchored to the runtime base using just the leaf
//     component. This protects against historical settings.json files that
//     stored cwd-relative paths like "data\\transcode" which would otherwise
//     create stray empty directories under whichever cwd the server started in.
func defaultDir(value string, runtimeHome string, dataDir string, name string) string {
	base := dataDir
	if runtimeHome != "" {
		base = runtimeHome
	}
	if value == "" {
		return filepath.Join(base, name)
	}
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(base, filepath.Base(value))
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
