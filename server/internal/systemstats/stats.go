package systemstats

import (
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type Snapshot struct {
	CollectedAt string       `json:"collectedAt"`
	CPU         CPUStats     `json:"cpu"`
	Memory      MemoryStats  `json:"memory"`
	Process     ProcessStats `json:"process"`
	Network     NetworkStats `json:"network"`
	Disks       []DiskStats  `json:"disks"`
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
	ReceiveBps  uint64                 `json:"receiveBps"`
	TransmitBps uint64                 `json:"transmitBps"`
	Interfaces  []NetworkInterfaceStat `json:"interfaces"`
}

type NetworkInterfaceStat struct {
	Name        string `json:"name"`
	ReceiveBps  uint64 `json:"receiveBps"`
	TransmitBps uint64 `json:"transmitBps"`
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

func usedPercent(total uint64, free uint64) float64 {
	if total == 0 {
		return 0
	}
	used := total - free
	return float64(used) / float64(total) * 100
}
