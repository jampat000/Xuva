package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	HTTPAddr         string `json:"httpAddr"`
	DataDir          string `json:"dataDir"`
	MovieLibraryPath string `json:"movieLibraryPath,omitempty"`
	TVLibraryPath    string `json:"tvLibraryPath,omitempty"`
	FFprobePath      string `json:"ffprobePath"`
	FFmpegPath       string `json:"ffmpegPath"`
	EventBuffer      int    `json:"eventBuffer"`
	ScanWorkers      int    `json:"scanWorkers"`
	ProbeWorkers     int    `json:"probeWorkers"`
	TranscodeWorkers int    `json:"transcodeWorkers"`
	GPUWorkers       int    `json:"gpuWorkers"`
}

func FromEnv() Config {
	dataDir := envString("VYRDEN_DATA_DIR", "data")
	cfg := Config{
		HTTPAddr:         envString("VYRDEN_HTTP_ADDR", "127.0.0.1:8097"),
		DataDir:          dataDir,
		MovieLibraryPath: envString("VYRDEN_MOVIES_PATH", ""),
		TVLibraryPath:    envString("VYRDEN_TV_PATH", ""),
		FFprobePath:      envString("VYRDEN_FFPROBE_PATH", "ffprobe"),
		FFmpegPath:       envString("VYRDEN_FFMPEG_PATH", "ffmpeg"),
		EventBuffer:      envInt("VYRDEN_EVENT_BUFFER", 128),
		ScanWorkers:      envInt("VYRDEN_SCAN_WORKERS", 1),
		ProbeWorkers:     envInt("VYRDEN_PROBE_WORKERS", 2),
		TranscodeWorkers: envInt("VYRDEN_TRANSCODE_WORKERS", 1),
		GPUWorkers:       envInt("VYRDEN_GPU_WORKERS", 1),
	}
	if saved, err := LoadFile(dataDir); err == nil {
		cfg = merge(cfg, saved)
	}
	cfg.HTTPAddr = envString("VYRDEN_HTTP_ADDR", cfg.HTTPAddr)
	cfg.DataDir = envString("VYRDEN_DATA_DIR", cfg.DataDir)
	cfg.MovieLibraryPath = envString("VYRDEN_MOVIES_PATH", cfg.MovieLibraryPath)
	cfg.TVLibraryPath = envString("VYRDEN_TV_PATH", cfg.TVLibraryPath)
	cfg.FFprobePath = envString("VYRDEN_FFPROBE_PATH", cfg.FFprobePath)
	cfg.FFmpegPath = envString("VYRDEN_FFMPEG_PATH", cfg.FFmpegPath)
	cfg.EventBuffer = envInt("VYRDEN_EVENT_BUFFER", cfg.EventBuffer)
	cfg.ScanWorkers = envInt("VYRDEN_SCAN_WORKERS", cfg.ScanWorkers)
	cfg.ProbeWorkers = envInt("VYRDEN_PROBE_WORKERS", cfg.ProbeWorkers)
	cfg.TranscodeWorkers = envInt("VYRDEN_TRANSCODE_WORKERS", cfg.TranscodeWorkers)
	cfg.GPUWorkers = envInt("VYRDEN_GPU_WORKERS", cfg.GPUWorkers)
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
	return base
}

func envString(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
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
