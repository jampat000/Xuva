//go:build windows

package systemstats

import (
	"math"
	"strings"
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
		// Speed is in bps; 0xFFFFFFFF is the Windows sentinel for >4.2Gbps
		// (need GetIfEntry2 for those — we just report 0/unknown for now).
		var speedBps uint64
		if row.Speed != 0 && row.Speed != 0xFFFFFFFF {
			speedBps = uint64(row.Speed)
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
