//go:build windows

package systemstats

import (
	"math"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetSystemTimes       = kernel32.NewProc("GetSystemTimes")
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
