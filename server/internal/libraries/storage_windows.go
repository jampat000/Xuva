//go:build windows

package libraries

import (
	"path/filepath"
	"syscall"
	"unsafe"
)

func windowsDriveStorageType(root string) StorageType {
	absolute, err := filepath.Abs(root)
	if err == nil && len(absolute) >= 3 {
		root = absolute[:3]
	}
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDriveType := kernel32.NewProc("GetDriveTypeW")
	ptr, err := syscall.UTF16PtrFromString(root)
	if err != nil {
		return StorageUnknown
	}
	ret, _, _ := getDriveType.Call(uintptr(unsafe.Pointer(ptr)))
	switch ret {
	case 2:
		return StorageRemovable
	case 3:
		return StorageLocal
	case 4:
		return StorageNetwork
	case 5:
		return StorageRemovable
	default:
		return StorageUnknown
	}
}
