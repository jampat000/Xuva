// Package smb provides credentialed access to remote SMB/UNC shares,
// independent of the process's logon identity.
//
// This is the core of the Windows "Server SKU" remote-filesystem story: a
// service running as LocalSystem has no network credentials and cannot reach a
// password-protected \\NAS\share. Validate/Browse authenticate to a UNC path
// with explicit per-share credentials (Windows: WNetAddConnection2 without
// mapping a drive letter), perform the operation, then disconnect.
//
// Credentialed SMB is a Windows-only capability; on other platforms Validate
// and Browse return ErrUnsupported (Linux/Docker reach shares via OS mounts,
// not per-connection credentials). Plain local/already-mounted browsing is
// unchanged and continues to flow through the existing folder-browse handler.
package smb

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Entry is a single immediate child directory of a browsed share path.
type Entry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
}

// ErrUnsupported is returned by Validate/Browse on non-Windows platforms.
var ErrUnsupported = errors.New("smb: credentialed SMB access is only supported on Windows")

// ErrNotUNC is returned when a remote path is not a UNC path (\\host\share...).
var ErrNotUNC = errors.New("smb: path is not a UNC share (expected \\\\host\\share)")

// IsUNC reports whether p looks like a Windows UNC path. Both backslash and
// forward-slash forms are accepted (browsers and JSON often carry "//host/share").
func IsUNC(p string) bool {
	p = strings.TrimSpace(p)
	return strings.HasPrefix(p, `\\`) || strings.HasPrefix(p, "//")
}

// normalizeUNC trims whitespace and converts forward slashes to backslashes so
// the value is a canonical Windows UNC path for the Win32 APIs.
func normalizeUNC(p string) string {
	p = strings.TrimSpace(p)
	p = strings.ReplaceAll(p, "/", `\`)
	return strings.TrimRight(p, `\`)
}

// listDirs returns the immediate subdirectories of root, sorted case-insensitively.
// Shared by the Windows implementation after a connection is established.
func listDirs(root string) ([]Entry, error) {
	dirents, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	out := make([]Entry, 0, len(dirents))
	for _, e := range dirents {
		if !e.IsDir() {
			continue
		}
		out = append(out, Entry{
			Name:  e.Name(),
			Path:  filepath.Join(root, e.Name()),
			IsDir: true,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}
