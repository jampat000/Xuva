//go:build windows

package systemstats

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetSystemTimes       = kernel32.NewProc("GetSystemTimes")
	iphlpapi                 = syscall.NewLazyDLL("iphlpapi.dll")
	procGetIfTable           = iphlpapi.NewProc("GetIfTable")
)

type memoryStatusEx struct {
	length               uint32
	memoryLoad           uint32
	totalPhys            uint64
	availPhys            uint64
	totalPageFile        uint64
	availPageFile        uint64
	totalVirtual         uint64
	availVirtual         uint64
	availExtendedVirtual uint64
}

func cpuPercent() float64 {
	first, ok := systemTimes()
	if !ok {
		return 0
	}
	time.Sleep(120 * time.Millisecond)
	second, ok := systemTimes()
	if !ok {
		return 0
	}
	idle := second.idle - first.idle
	total := second.total - first.total
	if total == 0 {
		return 0
	}
	value := (1 - float64(idle)/float64(total)) * 100
	return math.Max(0, math.Min(100, value))
}

func memoryStats() MemoryStats {
	var status memoryStatusEx
	status.length = uint32(unsafe.Sizeof(status))
	ok, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&status)))
	if ok == 0 {
		return MemoryStats{}
	}
	used := status.totalPhys - status.availPhys
	return MemoryStats{
		TotalBytes:     status.totalPhys,
		AvailableBytes: status.availPhys,
		UsedBytes:      used,
		UsedPercent:    usedPercent(status.totalPhys, status.availPhys),
	}
}

func diskStats(name string, path string) DiskStats {
	var freeAvailable, total, free uint64
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return DiskStats{Name: name, Path: path, Writable: false, Error: err.Error()}
	}
	if err := windows.GetDiskFreeSpaceEx(ptr, &freeAvailable, &total, &free); err != nil {
		return DiskStats{Name: name, Path: path, Writable: false, Error: err.Error()}
	}
	return DiskStats{
		Name:        name,
		Path:        path,
		TotalBytes:  total,
		FreeBytes:   freeAvailable,
		UsedBytes:   total - freeAvailable,
		UsedPercent: usedPercent(total, freeAvailable),
		Writable:    writable(path),
	}
}

type cpuTimes struct {
	idle  uint64
	total uint64
}

func systemTimes() (cpuTimes, bool) {
	var idle, kernel, user windows.Filetime
	ok, _, _ := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&idle)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	)
	if ok == 0 {
		return cpuTimes{}, false
	}
	idleValue := filetimeUint64(idle)
	kernelValue := filetimeUint64(kernel)
	userValue := filetimeUint64(user)
	return cpuTimes{idle: idleValue, total: kernelValue + userValue}, true
}

func filetimeUint64(value windows.Filetime) uint64 {
	return uint64(value.HighDateTime)<<32 + uint64(value.LowDateTime)
}

type mibIfRow struct {
	Name            [256]uint16
	Index           uint32
	Type            uint32
	MTU             uint32
	Speed           uint32
	PhysAddrLen     uint32
	PhysAddr        [8]byte
	AdminStatus     uint32
	OperStatus      uint32
	LastChange      uint32
	InOctets        uint32
	InUcastPkts     uint32
	InNUcastPkts    uint32
	InDiscards      uint32
	InErrors        uint32
	InUnknownProtos uint32
	OutOctets       uint32
	OutUcastPkts    uint32
	OutNUcastPkts   uint32
	OutDiscards     uint32
	OutErrors       uint32
	OutQLen         uint32
	DescrLen        uint32
	Descr           [256]byte
}

type netCounters struct {
	name         string
	rxBytes      uint64
	txBytes      uint64
	linkSpeedBps uint64
}

func networkStats() NetworkStats {
	first, ok := ifTable()
	if !ok {
		return NetworkStats{}
	}
	time.Sleep(120 * time.Millisecond)
	second, ok := ifTable()
	if !ok {
		return NetworkStats{}
	}
	interval := 0.12
	total := NetworkStats{Interfaces: []NetworkInterfaceStat{}}
	for index, after := range second {
		before, exists := first[index]
		if !exists {
			continue
		}
		rx := rateBps(after.rxBytes, before.rxBytes, interval)
		tx := rateBps(after.txBytes, before.txBytes, interval)
		total.ReceiveBps += rx
		total.TransmitBps += tx
		iface := NetworkInterfaceStat{
			Name:         after.name,
			ReceiveBps:   rx,
			TransmitBps:  tx,
			LinkSpeedBps: after.linkSpeedBps,
		}
		total.Interfaces = append(total.Interfaces, iface)
		if after.linkSpeedBps > total.LinkSpeedBps {
			total.LinkSpeedBps = after.linkSpeedBps
		}
	}
	// If the summed traffic across all interfaces exceeds the single-interface max
	// speed reported by GetIfTable (uint32 Speed field, capped at ~4.29 Gbps), it
	// means one or more high-speed NICs (10/25/40 Gbps) returned Speed=0xFFFFFFFF
	// and were excluded from the max. Signal "unknown" so the UI doesn't show a
	// misleading red alert.
	if total.LinkSpeedBps > 0 &&
		(total.ReceiveBps > total.LinkSpeedBps || total.TransmitBps > total.LinkSpeedBps) {
		total.LinkSpeedBps = 0
	}
	return total
}

func ifTable() (map[uint32]netCounters, bool) {
	size := uint32(0)
	procGetIfTable.Call(0, uintptr(unsafe.Pointer(&size)), 0)
	if size == 0 {
		return nil, false
	}
	buffer := make([]byte, size)
	result, _, _ := procGetIfTable.Call(uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&size)), 0)
	if result != 0 {
		return nil, false
	}
	count := *(*uint32)(unsafe.Pointer(&buffer[0]))
	rowSize := unsafe.Sizeof(mibIfRow{})
	base := uintptr(unsafe.Pointer(&buffer[0])) + unsafe.Sizeof(count)
	output := map[uint32]netCounters{}
	for i := uint32(0); i < count; i++ {
		row := (*mibIfRow)(unsafe.Pointer(base + uintptr(i)*rowSize))
		if row == nil || row.Type == 24 {
			continue
		}
		name := windows.UTF16ToString(row.Name[:])
		if name == "" {
			name = string(row.Descr[:row.DescrLen])
		}
		// Speed is a 32-bit field; 0xFFFFFFFF is the Windows sentinel for
		// >4.2 Gbps and many modern drivers return 0 for this field. Fall
		// back to the PowerShell Get-NetAdapter cache for accurate data.
		var speedBps uint64
		if row.Speed != 0 && row.Speed != 0xFFFFFFFF {
			speedBps = uint64(row.Speed)
		}
		if speedBps == 0 {
			speedBps = linkSpeedForIndex(row.Index)
		}
		output[row.Index] = netCounters{
			name:         strings.TrimSpace(name),
			rxBytes:      uint64(row.InOctets),
			txBytes:      uint64(row.OutOctets),
			linkSpeedBps: speedBps,
		}
	}
	return output, true
}

func rateBps(after uint64, before uint64, seconds float64) uint64 {
	if after < before || seconds <= 0 {
		return 0
	}
	return uint64(float64(after-before) * 8 / seconds)
}

// ─── Link-speed cache ─────────────────────────────────────────────────────────
//
// GetIfTable (MIB_IFROW.Speed) is a 32-bit field and returns 0 or 0xFFFFFFFF for
// modern NICs whose speed exceeds 4.2 Gbps or whose driver doesn't populate the
// legacy field. We fall back to Get-NetAdapter (PowerShell) which reads the
// NDIS-reported link speed accurately, and cache the result for 5 minutes since
// link speed never changes during normal operation.

var (
	linkSpeedCacheMu   sync.Mutex
	linkSpeedCache     map[uint32]uint64
	linkSpeedCacheTime time.Time
)

// linkSpeedForIndex returns the link speed (bps) for the given interface index,
// using a cached PowerShell lookup. Returns 0 if unavailable.
func linkSpeedForIndex(index uint32) uint64 {
	linkSpeedCacheMu.Lock()
	defer linkSpeedCacheMu.Unlock()
	if time.Since(linkSpeedCacheTime) > 5*time.Minute || linkSpeedCache == nil {
		if m := psNetLinkSpeeds(); m != nil {
			linkSpeedCache = m
			linkSpeedCacheTime = time.Now()
		}
	}
	if linkSpeedCache == nil {
		return 0
	}
	return linkSpeedCache[index]
}

// psNetLinkSpeeds shells out to PowerShell to get per-adapter link speeds keyed
// by interface index. Returns nil on any error (caller uses the cached value).
func psNetLinkSpeeds() map[uint32]uint64 {
	if _, err := os.Stat(`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`); err != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Select ifIndex (matches MIB_IFROW.Index) and Speed (NDIS 64-bit bps).
	cmd := `Get-NetAdapter | Where-Object {$_.Status -eq 'Up'} | ` +
		`Select-Object @{n='i';e={$_.ifIndex}},@{n='s';e={$_.Speed}} | ` +
		`ConvertTo-Json -Compress`
	out, err := exec.CommandContext(ctx,
		"powershell", "-NoProfile", "-NonInteractive", "-Command", cmd,
	).Output()
	if err != nil {
		return nil
	}
	type entry struct {
		I float64 `json:"i"`
		S float64 `json:"s"`
	}
	var arr []entry
	if err := json.Unmarshal(bytes.TrimSpace(out), &arr); err != nil {
		// Single adapter returns an object, not an array.
		var single entry
		if err2 := json.Unmarshal(bytes.TrimSpace(out), &single); err2 != nil {
			return nil
		}
		arr = []entry{single}
	}
	result := make(map[uint32]uint64, len(arr))
	for _, a := range arr {
		if a.I > 0 && a.S > 0 {
			result[uint32(a.I)] = uint64(a.S)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func gpuStats() *GPUStats {
	// nvidia-smi gives full real-time metrics (NVIDIA cards, incl. path search).
	if s := nvidiaGPUStats(); s != nil {
		return s
	}
	// For AMD / Intel discrete GPUs on Windows: try wmic then PowerShell CIM.
	// These don't give live utilisation, but we surface adapter name + VRAM.
	if s := wmicGPUStats(); s != nil {
		return s
	}
	return psGPUStats()
}

// wmicGPUStats queries Win32_VideoController via wmic for the adapter name
// and VRAM (AdapterRAM — uint32 field, capped at ~4 GB for cards that only
// report via legacy BIOS; accurate for most integrated and older discrete GPUs).
// wmic is deprecated in Windows 11 but still present on most machines.
func wmicGPUStats() *GPUStats {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx,
		"wmic", "path", "Win32_VideoController",
		"get", "Name,AdapterRAM", "/format:list",
	).Output()
	if err != nil {
		return nil
	}
	var name string
	var vramBytes uint64
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Name=") {
			n := strings.TrimSpace(strings.TrimPrefix(line, "Name="))
			if n != "" && !strings.Contains(n, "Microsoft Basic") && name == "" {
				name = n
			}
		}
		if strings.HasPrefix(line, "AdapterRAM=") && vramBytes == 0 {
			r, _ := strconv.ParseUint(strings.TrimSpace(strings.TrimPrefix(line, "AdapterRAM=")), 10, 64)
			if r > 0 {
				vramBytes = r
			}
		}
	}
	if name == "" {
		return nil
	}
	return &GPUStats{AdapterName: name, VRAMTotalBytes: vramBytes}
}

// psGPUStats is a fallback for machines where wmic is not available (e.g. a
// stripped Windows 11 install). It shells out to PowerShell CIM cmdlets.
func psGPUStats() *GPUStats {
	// Guard: only attempt if powershell.exe exists.
	if _, err := os.Stat(`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`); err != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := `Get-CimInstance -ClassName Win32_VideoController | ` +
		`Where-Object {$_.Name -notlike '*Microsoft*'} | ` +
		`Select-Object -First 1 Name,AdapterRAM | ` +
		`ConvertTo-Json -Compress`
	out, err := exec.CommandContext(ctx,
		"powershell", "-NoProfile", "-NonInteractive", "-Command", cmd,
	).Output()
	if err != nil {
		return nil
	}
	var result struct {
		Name       string  `json:"Name"`
		AdapterRAM float64 `json:"AdapterRAM"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(out), &result); err != nil {
		return nil
	}
	if strings.TrimSpace(result.Name) == "" {
		return nil
	}
	return &GPUStats{
		AdapterName:    strings.TrimSpace(result.Name),
		VRAMTotalBytes: uint64(result.AdapterRAM),
	}
}
