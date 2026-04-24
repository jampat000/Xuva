package libraries

import (
	"sort"
	"sync"
)

type Kind string

const (
	KindMovies Kind = "movies"
	KindTV     Kind = "tv"
)

type Library struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Kind Kind   `json:"kind"`
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
