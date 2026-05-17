package libraries

import (
	"crypto/sha1"
	"encoding/hex"
	"path/filepath"
	"sort"
	"strings"
	"sync"
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
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	Path            string      `json:"path"`
	Kind            Kind        `json:"kind"`
	StorageType     StorageType `json:"storageType"`
	MetadataSources []string    `json:"metadataSources,omitempty"`
	ArtworkSources  []string    `json:"artworkSources,omitempty"`
}

type WorkerDefaults struct {
	ScanWorkers      int    `json:"scanWorkers"`
	ProbeWorkers     int    `json:"probeWorkers"`
	TranscodeWorkers int    `json:"transcodeWorkers"`
	SyncMode         string `json:"syncMode"`
	Reason           string `json:"reason"`
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
		library.ID = IDFor(library.Kind, library.Path)
	}
	if library.StorageType == "" {
		library.StorageType = DetectStorageType(library.Path)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.libraries[library.ID] = library
}

func (s *Service) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.libraries, id)
}

func (s *Service) Get(id string) (Library, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	library, ok := s.libraries[id]
	return library, ok
}

func IDFor(kind Kind, path string) string {
	hash := sha1.New()
	hash.Write([]byte(kind))
	hash.Write([]byte{0})
	hash.Write([]byte(filepath.Clean(path)))
	return string(kind) + "_" + hex.EncodeToString(hash.Sum(nil))[:12]
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

func DefaultsForStorage(storage StorageType) WorkerDefaults {
	switch storage {
	case StorageNetwork:
		return WorkerDefaults{ScanWorkers: 1, ProbeWorkers: 1, TranscodeWorkers: 1, SyncMode: "watch", Reason: "Network storage benefits from low concurrent reads and debounced change batching."}
	case StorageRemovable:
		return WorkerDefaults{ScanWorkers: 1, ProbeWorkers: 1, TranscodeWorkers: 1, SyncMode: "manual", Reason: "Removable storage can disappear, so background pressure should stay conservative."}
	case StorageMounted:
		return WorkerDefaults{ScanWorkers: 1, ProbeWorkers: 2, TranscodeWorkers: 1, SyncMode: "watch", Reason: "Mounted storage is usually shared or external, so scans should be debounced."}
	case StorageLocal:
		return WorkerDefaults{ScanWorkers: 2, ProbeWorkers: 2, TranscodeWorkers: 1, SyncMode: "daily", Reason: "Local disks can tolerate more scan and probe parallelism."}
	default:
		return WorkerDefaults{ScanWorkers: 1, ProbeWorkers: 1, TranscodeWorkers: 1, SyncMode: "daily", Reason: "Unknown storage uses safe low-overhead defaults."}
	}
}
