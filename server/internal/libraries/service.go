package libraries

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

type Kind string
type StorageType string

const (
	KindMovies Kind = "movies"
	KindTV     Kind = "tv"

	StorageLocal     StorageType = "local"
	StorageRemovable StorageType = "removable"
	StorageNetwork   StorageType = "network"
	StorageMounted   StorageType = "mounted"
	StorageUnknown   StorageType = "unknown"
)

type Library struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Path        string      `json:"path"`
	Kind        Kind        `json:"kind"`
	StorageType StorageType `json:"storageType"`
}

type Service struct {
	mu        sync.RWMutex
	libraries map[string]Library
}

func NewService() *Service {
	return &Service{libraries: make(map[string]Library)}
}

func (s *Service) Set(library Library) {
	if library.ID == "" {
		return
	}
	if library.StorageType == "" {
		library.StorageType = DetectStorageType(library.Path)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.libraries[library.ID] = library
}

func (s *Service) List() []Library {
	s.mu.RLock()
	defer s.mu.RUnlock()

	output := make([]Library, 0, len(s.libraries))
	for _, library := range s.libraries {
		output = append(output, library)
	}
	sort.Slice(output, func(i, j int) bool {
		return output[i].ID < output[j].ID
	})
	return output
}

func DetectStorageType(path string) StorageType {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return StorageUnknown
	}
	normalized := strings.ReplaceAll(trimmed, "/", "\\")
	lower := strings.ToLower(normalized)

	if strings.HasPrefix(normalized, `\\`) {
		return StorageNetwork
	}
	if strings.HasPrefix(lower, `smb:\`) || strings.HasPrefix(lower, `nfs:\`) || strings.HasPrefix(lower, `\\?\unc\`) {
		return StorageNetwork
	}
	if len(normalized) >= 3 && normalized[1] == ':' && normalized[2] == '\\' {
		if driveType := windowsDriveStorageType(normalized[:3]); driveType != StorageUnknown {
			return driveType
		}
		return StorageLocal
	}
	if strings.HasPrefix(trimmed, "/mnt/") || strings.HasPrefix(trimmed, "/media/") || strings.HasPrefix(trimmed, "/Volumes/") {
		return StorageMounted
	}
	if strings.HasPrefix(trimmed, "/") {
		return StorageLocal
	}
	return StorageUnknown
}

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
