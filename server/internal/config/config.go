package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr         string `json:"httpAddr"`
	DataDir          string `json:"dataDir"`
	EventBuffer      int    `json:"eventBuffer"`
	ScanWorkers      int    `json:"scanWorkers"`
	ProbeWorkers     int    `json:"probeWorkers"`
	TranscodeWorkers int    `json:"transcodeWorkers"`
	GPUWorkers       int    `json:"gpuWorkers"`
}

func FromEnv() Config {
	return Config{
		HTTPAddr:         envString("VYRDEN_HTTP_ADDR", "127.0.0.1:8097"),
		DataDir:          envString("VYRDEN_DATA_DIR", "data"),
		EventBuffer:      envInt("VYRDEN_EVENT_BUFFER", 128),
		ScanWorkers:      envInt("VYRDEN_SCAN_WORKERS", 1),
		ProbeWorkers:     envInt("VYRDEN_PROBE_WORKERS", 2),
		TranscodeWorkers: envInt("VYRDEN_TRANSCODE_WORKERS", 1),
		GPUWorkers:       envInt("VYRDEN_GPU_WORKERS", 1),
	}
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
