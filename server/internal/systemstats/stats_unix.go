//go:build !windows

package systemstats

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func cpuPercent() float64 {
	first, ok := procCPU()
	if !ok {
		return 0
	}
	time.Sleep(120 * time.Millisecond)
	second, ok := procCPU()
	if !ok {
		return 0
	}
	idle := second.idle - first.idle
	total := second.total - first.total
	if total == 0 {
		return 0
	}
	return (1 - float64(idle)/float64(total)) * 100
}

func memoryStats() MemoryStats {
	raw, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return MemoryStats{}
	}
	values := map[string]uint64{}
	for _, line := range strings.Split(string(raw), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		value, _ := strconv.ParseUint(fields[1], 10, 64)
		values[strings.TrimSuffix(fields[0], ":")] = value * 1024
	}
	total := values["MemTotal"]
	available := values["MemAvailable"]
	if available == 0 {
		available = values["MemFree"]
	}
	return MemoryStats{
		TotalBytes:     total,
		AvailableBytes: available,
		UsedBytes:      total - available,
		UsedPercent:    usedPercent(total, available),
	}
}

func diskStats(name string, path string) DiskStats {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return DiskStats{Name: name, Path: path, Writable: false, Error: err.Error()}
	}
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	return DiskStats{
		Name:        name,
		Path:        path,
		TotalBytes:  total,
		FreeBytes:   free,
		UsedBytes:   total - free,
		UsedPercent: usedPercent(total, free),
		Writable:    writable(path),
	}
}

type cpuTimes struct {
	idle  uint64
	total uint64
}

func procCPU() (cpuTimes, bool) {
	raw, err := os.ReadFile("/proc/stat")
	if err != nil {
		return cpuTimes{}, false
	}
	line := strings.SplitN(string(raw), "\n", 2)[0]
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return cpuTimes{}, false
	}
	var total uint64
	var idle uint64
	for index, field := range fields[1:] {
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return cpuTimes{}, false
		}
		total += value
		if index == 3 || index == 4 {
			idle += value
		}
	}
	return cpuTimes{idle: idle, total: total}, true
}

func networkStats() NetworkStats {
	first, ok := procNet()
	if !ok {
		return NetworkStats{}
	}
	time.Sleep(120 * time.Millisecond)
	second, ok := procNet()
	if !ok {
		return NetworkStats{}
	}
	interval := 0.12
	total := NetworkStats{Interfaces: []NetworkInterfaceStat{}}
	for name, after := range second {
		before, exists := first[name]
		if !exists {
			continue
		}
		rx := rateBps(after.rxBytes, before.rxBytes, interval)
		tx := rateBps(after.txBytes, before.txBytes, interval)
		total.ReceiveBps += rx
		total.TransmitBps += tx
		speedBps := ifaceSpeedBps(name)
		iface := NetworkInterfaceStat{
			Name:         name,
			ReceiveBps:   rx,
			TransmitBps:  tx,
			LinkSpeedBps: speedBps,
		}
		total.Interfaces = append(total.Interfaces, iface)
		if speedBps > total.LinkSpeedBps {
			total.LinkSpeedBps = speedBps
		}
	}
	return total
}

func ifaceSpeedBps(name string) uint64 {
	raw, err := os.ReadFile("/sys/class/net/" + name + "/speed")
	if err != nil {
		return 0
	}
	mbps, err := strconv.ParseInt(strings.TrimSpace(string(raw)), 10, 64)
	if err != nil || mbps <= 0 {
		return 0
	}
	return uint64(mbps) * 1_000_000
}

type netCounters struct {
	rxBytes uint64
	txBytes uint64
}

func procNet() (map[string]netCounters, bool) {
	raw, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return nil, false
	}
	lines := strings.Split(string(raw), "\n")
	output := map[string]netCounters{}
	for _, line := range lines[2:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		if name == "" || name == "lo" {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			continue
		}
		rx, err1 := strconv.ParseUint(fields[0], 10, 64)
		tx, err2 := strconv.ParseUint(fields[8], 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		output[name] = netCounters{rxBytes: rx, txBytes: tx}
	}
	return output, true
}

func rateBps(after uint64, before uint64, seconds float64) uint64 {
	if after < before || seconds <= 0 {
		return 0
	}
	return uint64(float64(after-before) * 8 / seconds)
}

func gpuStats() *GPUStats {
	// NVIDIA: full metrics via nvidia-smi (searches PATH + common install paths).
	if s := nvidiaGPUStats(); s != nil {
		return s
	}
	// AMD: real-time utilisation + VRAM via DRM sysfs (works without ROCm).
	return amdSysfsGPUStats()
}

// amdSysfsGPUStats reads AMD GPU metrics from the DRM sysfs interface.
// This works for any AMDGPU-driven card (GCN 1.0+ / RDNA) without requiring
// the ROCm toolkit. Skips cards that don't expose gpu_busy_percent.
func amdSysfsGPUStats() *GPUStats {
	// Vendor ID 0x1002 = AMD / ATI
	const amdVendorID = "0x1002"
	cards, _ := filepath.Glob("/sys/class/drm/card*/device")
	for _, devPath := range cards {
		vendorRaw, err := os.ReadFile(filepath.Join(devPath, "vendor"))
		if err != nil || strings.TrimSpace(string(vendorRaw)) != amdVendorID {
			continue
		}
		utilRaw, err := os.ReadFile(filepath.Join(devPath, "gpu_busy_percent"))
		if err != nil {
			continue // card present but sysfs metrics not exposed
		}
		util, _ := strconv.ParseFloat(strings.TrimSpace(string(utilRaw)), 64)
		vramUsed := readSysfsUint64(filepath.Join(devPath, "mem_info_vram_used"))
		vramTotal := readSysfsUint64(filepath.Join(devPath, "mem_info_vram_total"))
		return &GPUStats{
			AdapterName:    amdSysfsName(devPath),
			UtilizationPct: util,
			VRAMUsedBytes:  vramUsed,
			VRAMTotalBytes: vramTotal,
		}
	}
	return nil
}

func readSysfsUint64(path string) uint64 {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	v, _ := strconv.ParseUint(strings.TrimSpace(string(raw)), 10, 64)
	return v
}

// amdSysfsName tries to resolve a human-readable product name for the card.
// Falls back to a generic label when the uevent doesn't expose one.
func amdSysfsName(devPath string) string {
	// Some kernels expose the marketing name here.
	if raw, err := os.ReadFile(filepath.Join(devPath, "product_name")); err == nil {
		if name := strings.TrimSpace(string(raw)); name != "" {
			return name
		}
	}
	// drm_name is available on newer kernels.
	if raw, err := os.ReadFile(filepath.Join(devPath, "../drm_name")); err == nil {
		if name := strings.TrimSpace(string(raw)); name != "" {
			return name
		}
	}
	return "AMD Radeon GPU"
}
