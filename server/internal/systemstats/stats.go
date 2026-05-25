package systemstats

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Snapshot struct {
	CollectedAt string       `json:"collectedAt"`
	CPU         CPUStats     `json:"cpu"`
	Memory      MemoryStats  `json:"memory"`
	Process     ProcessStats `json:"process"`
	Network     NetworkStats `json:"network"`
	Disks       []DiskStats  `json:"disks"`
	GPU         *GPUStats    `json:"gpu,omitempty"`
}

// GPUStats carries real hardware GPU metrics (adapter name, utilization,
// VRAM). Fields are zero/empty when the relevant data is unavailable.
type GPUStats struct {
	AdapterName    string  `json:"adapterName,omitempty"`
	UtilizationPct float64 `json:"utilizationPct"`
	VRAMUsedBytes  uint64  `json:"vramUsedBytes"`
	VRAMTotalBytes uint64  `json:"vramTotalBytes"`
}

type CPUStats struct {
	Percent float64 `json:"percent"`
	Cores   int     `json:"cores"`
}

type MemoryStats struct {
	TotalBytes     uint64  `json:"totalBytes"`
	AvailableBytes uint64  `json:"availableBytes"`
	UsedBytes      uint64  `json:"usedBytes"`
	UsedPercent    float64 `json:"usedPercent"`
}

type ProcessStats struct {
	GoAllocBytes uint64 `json:"goAllocBytes"`
	GoSysBytes   uint64 `json:"goSysBytes"`
	Goroutines   int    `json:"goroutines"`
}

type NetworkStats struct {
	ReceiveBps   uint64                 `json:"receiveBps"`
	TransmitBps  uint64                 `json:"transmitBps"`
	LinkSpeedBps uint64                 `json:"linkSpeedBps"` // max interface speed; 0 if unknown
	Interfaces   []NetworkInterfaceStat `json:"interfaces"`
}

type NetworkInterfaceStat struct {
	Name         string `json:"name"`
	ReceiveBps   uint64 `json:"receiveBps"`
	TransmitBps  uint64 `json:"transmitBps"`
	LinkSpeedBps uint64 `json:"linkSpeedBps"` // 0 if unknown
}

type DiskStats struct {
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	TotalBytes     uint64  `json:"totalBytes"`
	FreeBytes      uint64  `json:"freeBytes"`
	UsedBytes      uint64  `json:"usedBytes"`
	UsedPercent    float64 `json:"usedPercent"`
	Writable       bool    `json:"writable"`
	Error          string  `json:"error,omitempty"`
	SharedWithData bool    `json:"sharedWithData"`
}

func Collect(paths map[string]string) Snapshot {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	output := Snapshot{
		CollectedAt: time.Now().UTC().Format(time.RFC3339),
		CPU: CPUStats{
			Percent: cpuPercent(),
			Cores:   runtime.NumCPU(),
		},
		Memory: memoryStats(),
		Process: ProcessStats{
			GoAllocBytes: mem.Alloc,
			GoSysBytes:   mem.Sys,
			Goroutines:   runtime.NumGoroutine(),
		},
		Network: networkStats(),
		Disks:   []DiskStats{},
		GPU:     gpuStats(),
	}
	dataRoot := volumeRoot(paths["data"])
	for name, path := range paths {
		if path == "" {
			continue
		}
		disk := diskStats(name, path)
		disk.SharedWithData = dataRoot != "" && volumeRoot(path) == dataRoot
		output.Disks = append(output.Disks, disk)
	}
	return output
}

func writable(path string) bool {
	if path == "" {
		return false
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return false
	}
	file, err := os.CreateTemp(path, ".xuva-write-*")
	if err != nil {
		return false
	}
	name := file.Name()
	_ = file.Close()
	_ = os.Remove(name)
	return true
}

func volumeRoot(path string) string {
	if path == "" {
		return ""
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	volume := filepath.VolumeName(abs)
	if volume != "" {
		return volume
	}
	return string(filepath.Separator)
}

// nvidiaGPUStats queries nvidia-smi for the first GPU's name, utilization,
// and VRAM. Returns nil when nvidia-smi is not installed or the query fails.
func nvidiaGPUStats() *GPUStats {
	smi := findNvidiaSmi()
	if smi == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, smi,
		"--query-gpu=name,utilization.gpu,memory.used,memory.total",
		"--format=csv,noheader,nounits",
	).Output()
	if err != nil {
		return nil
	}
	// Take the first GPU if multiple are present.
	line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	parts := strings.SplitN(line, ",", 4)
	if len(parts) < 4 {
		return nil
	}
	util, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	memUsedMiB, _ := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 64)
	memTotalMiB, _ := strconv.ParseUint(strings.TrimSpace(parts[3]), 10, 64)
	return &GPUStats{
		AdapterName:    strings.TrimSpace(parts[0]),
		UtilizationPct: util,
		VRAMUsedBytes:  memUsedMiB * 1024 * 1024,
		VRAMTotalBytes: memTotalMiB * 1024 * 1024,
	}
}

// findNvidiaSmi returns the full path to nvidia-smi, searching PATH first then
// common install directories (handles the frequent case where the NVIDIA NVSMI
// directory is not added to PATH on Windows).
func findNvidiaSmi() string {
	if path, err := exec.LookPath("nvidia-smi"); err == nil {
		return path
	}
	for _, candidate := range []string{
		`C:\Program Files\NVIDIA Corporation\NVSMI\nvidia-smi.exe`,
		`C:\Windows\System32\nvidia-smi.exe`,
		`/usr/bin/nvidia-smi`,
		`/usr/local/bin/nvidia-smi`,
		`/usr/lib/nvidia/bin/nvidia-smi`,
		`/opt/cuda/bin/nvidia-smi`,
	} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func usedPercent(total uint64, free uint64) float64 {
	if total == 0 {
		return 0
	}
	used := total - free
	return float64(used) / float64(total) * 100
}
