//go:build !windows

package smb

// Validate reports that credentialed SMB access is unavailable on this platform.
// Non-Windows deployments (Linux/Docker) reach shares through OS-level mounts
// rather than per-connection credentials.
func Validate(remote, username, password string) error {
	return ErrUnsupported
}

// Browse reports that credentialed SMB access is unavailable on this platform.
func Browse(remote, username, password string) ([]Entry, error) {
	return nil, ErrUnsupported
}
