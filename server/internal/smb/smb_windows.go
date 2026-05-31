//go:build windows

package smb

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const resourcetypeDisk = 0x1 // RESOURCETYPE_DISK

var (
	modmpr                     = windows.NewLazySystemDLL("mpr.dll")
	procWNetAddConnection2W    = modmpr.NewProc("WNetAddConnection2W")
	procWNetCancelConnection2W = modmpr.NewProc("WNetCancelConnection2W")
)

// netResource mirrors the Win32 NETRESOURCEW structure (only the fields we set).
type netResource struct {
	Scope       uint32
	Type        uint32
	DisplayType uint32
	Usage       uint32
	LocalName   *uint16
	RemoteName  *uint16
	Comment     *uint16
	Provider    *uint16
}

// connect authenticates to a UNC remote (\\host\share) with explicit
// credentials WITHOUT mapping a drive letter (lpLocalName = NULL). It returns a
// disconnect func that tears the session down. An empty username/password binds
// with NULL, i.e. the caller's current identity.
//
// Note: WNetAddConnection2 is a synchronous, non-cancelable call that can block
// for the OS network timeout (~tens of seconds) when the host is unreachable.
// Callers that need a deadline should run this on a goroutine and stop waiting
// on it; the syscall itself will unwind when the OS gives up.
func connect(remote, username, password string) (func(), error) {
	rn, err := windows.UTF16PtrFromString(remote)
	if err != nil {
		return nil, fmt.Errorf("smb: encode remote: %w", err)
	}
	nr := netResource{Type: resourcetypeDisk, RemoteName: rn}

	// Credential pointer semantics for WNetAddConnection2W:
	//   lpPassword == NULL  -> use the caller's *default* credentials
	//   lpPassword == L""    -> an *explicit empty* password
	// These are different. When a username is supplied we must pass an explicit
	// password pointer even when it's blank, otherwise e.g. `guest` with an
	// empty password is sent with the machine's default credentials and the
	// server rejects it with ERROR_LOGON_FAILURE (1326). This is exactly how
	// `net use \\host\share /user:guest ""` succeeds where NULL would fail.
	var userPtr, passPtr *uint16
	if username != "" {
		if userPtr, err = windows.UTF16PtrFromString(username); err != nil {
			return nil, fmt.Errorf("smb: encode username: %w", err)
		}
		// Explicit username => explicit password (empty string allowed).
		if passPtr, err = windows.UTF16PtrFromString(password); err != nil {
			return nil, fmt.Errorf("smb: encode password: %w", err)
		}
	} else if password != "" {
		// No username but an explicit password: pass it through as-is.
		if passPtr, err = windows.UTF16PtrFromString(password); err != nil {
			return nil, fmt.Errorf("smb: encode password: %w", err)
		}
	}
	// Both empty => userPtr/passPtr stay nil (NULL) => default credentials.

	r, _, _ := procWNetAddConnection2W.Call(
		uintptr(unsafe.Pointer(&nr)),
		uintptr(unsafe.Pointer(passPtr)),
		uintptr(unsafe.Pointer(userPtr)),
		0, // dwFlags: no CONNECT_UPDATE_PROFILE — ephemeral, not persisted
	)
	if r != 0 { // anything non-zero is a Win32 error code (NO_ERROR == 0)
		return nil, fmt.Errorf("smb: connect to %s failed: %w", remote, syscall.Errno(r))
	}

	disconnect := func() {
		// fForce = TRUE: drop even if the share is in use by this process.
		procWNetCancelConnection2W.Call(uintptr(unsafe.Pointer(rn)), 0, 1)
	}
	return disconnect, nil
}

// Validate authenticates to remote with the given credentials and confirms the
// share is reachable and listable, then disconnects.
func Validate(remote, username, password string) error {
	if !IsUNC(remote) {
		return ErrNotUNC
	}
	remote = normalizeUNC(remote)
	disconnect, err := connect(remote, username, password)
	if err != nil {
		return err
	}
	defer disconnect()
	if _, err := listDirs(remote); err != nil {
		return fmt.Errorf("smb: connected but cannot list %s: %w", remote, err)
	}
	return nil
}

// Browse authenticates to remote and returns its immediate subdirectories, then
// disconnects.
func Browse(remote, username, password string) ([]Entry, error) {
	if !IsUNC(remote) {
		return nil, ErrNotUNC
	}
	remote = normalizeUNC(remote)
	disconnect, err := connect(remote, username, password)
	if err != nil {
		return nil, err
	}
	defer disconnect()
	return listDirs(remote)
}
