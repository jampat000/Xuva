package scanner

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type LibraryKind string

const (
	KindMovies LibraryKind = "movies"
	KindTV     LibraryKind = "tv"
)

var ErrMissingRoot = errors.New("scan root is required")

type Request struct {
	Kind          LibraryKind    `json:"kind"`
	Root          string         `json:"root"`
	IncludeHidden bool           `json:"includeHidden"`
	Progress      func(Progress) `json:"-"`
}

type Progress struct {
	Kind         LibraryKind `json:"kind"`
	Root         string      `json:"root"`
	TotalFiles   int         `json:"totalFiles"`
	MediaFiles   int         `json:"mediaFiles"`
	IgnoredFiles int         `json:"ignoredFiles"`
	LastPath     string      `json:"lastPath,omitempty"`
}

type FileCandidate struct {
	Path       string    `json:"path"`
	RelPath    string    `json:"relPath"`
	Name       string    `json:"name"`
	Extension  string    `json:"extension"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

type ScanError struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

type Summary struct {
	Kind         LibraryKind    `json:"kind"`
	Root         string         `json:"root"`
	StartedAt    time.Time      `json:"startedAt"`
	CompletedAt  time.Time      `json:"completedAt"`
	DurationMS   int64          `json:"durationMs"`
	TotalFiles   int            `json:"totalFiles"`
	MediaFiles   int            `json:"mediaFiles"`
	IgnoredFiles int            `json:"ignoredFiles"`
	ErrorCount   int            `json:"errorCount"`
	Extensions   map[string]int `json:"extensions"`
}

type Result struct {
	Summary
	Files  []FileCandidate `json:"files"`
	Errors []ScanError     `json:"errors,omitempty"`
}

type Service struct {
	mediaExtensions map[string]struct{}
	skipDirs        map[string]struct{}
}

func NewService() *Service {
	return &Service{
		mediaExtensions: defaultMediaExtensions(),
		skipDirs:        defaultSkipDirs(),
	}
}

func (s *Service) Scan(ctx context.Context, request Request) (Result, error) {
	startedAt := time.Now().UTC()
	result := Result{
		Summary: Summary{
			Kind:       request.Kind,
			Root:       request.Root,
			StartedAt:  startedAt,
			Extensions: make(map[string]int),
		},
	}

	root := strings.TrimSpace(request.Root)
	if root == "" {
		result.finish(startedAt)
		return result, ErrMissingRoot
	}

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		result.finish(startedAt)
		return result, err
	}
	result.Root = absoluteRoot

	info, err := os.Stat(absoluteRoot)
	if err != nil {
		result.finish(startedAt)
		return result, err
	}
	if !info.IsDir() {
		result.finish(startedAt)
		return result, errors.New("scan root must be a directory")
	}

	err = filepath.WalkDir(absoluteRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if walkErr != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, ScanError{Path: path, Error: walkErr.Error()})
			return nil
		}
		if entry.IsDir() {
			if path == absoluteRoot {
				return nil
			}
			if !request.IncludeHidden && isHiddenName(entry.Name()) {
				return filepath.SkipDir
			}
			if _, skip := s.skipDirs[strings.ToLower(entry.Name())]; skip {
				return filepath.SkipDir
			}
			return nil
		}

		result.TotalFiles++
		if !request.IncludeHidden && isHiddenName(entry.Name()) {
			result.IgnoredFiles++
			return nil
		}

		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if _, ok := s.mediaExtensions[extension]; !ok {
			result.IgnoredFiles++
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, ScanError{Path: path, Error: err.Error()})
			return nil
		}
		relPath, err := filepath.Rel(absoluteRoot, path)
		if err != nil {
			relPath = entry.Name()
		}

		result.MediaFiles++
		result.Extensions[extension]++
		candidate := FileCandidate{
			Path:       path,
			RelPath:    relPath,
			Name:       entry.Name(),
			Extension:  extension,
			Size:       info.Size(),
			ModifiedAt: info.ModTime().UTC(),
		}
		result.Files = append(result.Files, candidate)
		if request.Progress != nil {
			request.Progress(Progress{
				Kind:         request.Kind,
				Root:         absoluteRoot,
				TotalFiles:   result.TotalFiles,
				MediaFiles:   result.MediaFiles,
				IgnoredFiles: result.IgnoredFiles,
				LastPath:     candidate.RelPath,
			})
		}
		return nil
	})
	if err != nil {
		result.finish(startedAt)
		return result, err
	}

	sort.Slice(result.Files, func(i, j int) bool {
		return result.Files[i].RelPath < result.Files[j].RelPath
	})
	result.finish(startedAt)
	return result, nil
}

func (r *Result) finish(startedAt time.Time) {
	completedAt := time.Now().UTC()
	r.StartedAt = startedAt
	r.CompletedAt = completedAt
	r.DurationMS = completedAt.Sub(startedAt).Milliseconds()
}

func isHiddenName(name string) bool {
	return strings.HasPrefix(name, ".")
}

func defaultMediaExtensions() map[string]struct{} {
	extensions := []string{
		".3g2", ".3gp", ".asf", ".avi", ".divx", ".dv", ".f4v", ".flv",
		".iso", ".m2ts", ".m4v", ".mkv", ".mov", ".mp4", ".mpeg", ".mpg",
		".mts", ".ogm", ".ogv", ".rm", ".rmvb", ".strm", ".ts", ".vob",
		".webm", ".wmv",
	}
	output := make(map[string]struct{}, len(extensions))
	for _, extension := range extensions {
		output[extension] = struct{}{}
	}
	return output
}

func defaultSkipDirs() map[string]struct{} {
	dirs := []string{
		"#recycle", "$recycle.bin", "@eadir", ".git", ".stfolder", ".sync",
		"lost+found", "node_modules", "system volume information",
	}
	output := make(map[string]struct{}, len(dirs))
	for _, dir := range dirs {
		output[dir] = struct{}{}
	}
	return output
}
