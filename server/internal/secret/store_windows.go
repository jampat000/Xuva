//go:build windows

package secret

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// cryptProtectLocalMachine seals to the machine rather than the calling user,
// so the SYSTEM-run service and an elevated admin console process can both open
// the same secret. (Without it, DPAPI binds to the user profile.)
const cryptProtectLocalMachine = 0x4

var (
	modcrypt32             = windows.NewLazySystemDLL("crypt32.dll")
	modkernel32            = windows.NewLazySystemDLL("kernel32.dll")
	procCryptProtectData   = modcrypt32.NewProc("CryptProtectData")
	procCryptUnprotectData = modcrypt32.NewProc("CryptUnprotectData")
	procLocalFree          = modkernel32.NewProc("LocalFree")
)

// dataBlob mirrors the Win32 DATA_BLOB structure.
type dataBlob struct {
	cbData uint32
	pbData *byte
}

func newBlob(b []byte) *dataBlob {
	if len(b) == 0 {
		return &dataBlob{}
	}
	return &dataBlob{cbData: uint32(len(b)), pbData: &b[0]}
}

// copyOut copies the DPAPI-allocated buffer into a Go-owned slice. Must be
// called before the corresponding LocalFree.
func (b *dataBlob) copyOut() []byte {
	if b.cbData == 0 || b.pbData == nil {
		return nil
	}
	out := make([]byte, b.cbData)
	copy(out, unsafe.Slice(b.pbData, b.cbData))
	return out
}

// dpapiStore seals values with Windows DPAPI at machine scope.
type dpapiStore struct{}

// New returns a DPAPI-backed Store. keyDir is unused on Windows (DPAPI manages
// the master key); the parameter is kept for a uniform cross-platform signature.
func New(keyDir string) (Store, error) { return dpapiStore{}, nil }

func (dpapiStore) Seal(plaintext []byte) ([]byte, error) {
	in := newBlob(plaintext)
	var out dataBlob
	r, _, callErr := procCryptProtectData.Call(
		uintptr(unsafe.Pointer(in)),
		0, // szDataDescr
		0, // pOptionalEntropy
		0, // pvReserved
		0, // pPromptStruct
		uintptr(cryptProtectLocalMachine),
		uintptr(unsafe.Pointer(&out)),
	)
	if r == 0 {
		return nil, fmt.Errorf("secret: CryptProtectData failed: %w", callErr)
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(out.pbData)))
	// version || DPAPI blob
	sealed := make([]byte, 0, 1+int(out.cbData))
	sealed = append(sealed, blobVersion)
	sealed = append(sealed, out.copyOut()...)
	return sealed, nil
}

func (dpapiStore) Open(sealed []byte) ([]byte, error) {
	if len(sealed) < 1 {
		return nil, errors.New("secret: sealed blob too short")
	}
	if sealed[0] != blobVersion {
		return nil, fmt.Errorf("secret: unsupported blob version %d", sealed[0])
	}
	in := newBlob(sealed[1:])
	var out dataBlob
	r, _, callErr := procCryptUnprotectData.Call(
		uintptr(unsafe.Pointer(in)),
		0, // ppszDataDescr
		0, // pOptionalEntropy
		0, // pvReserved
		0, // pPromptStruct
		uintptr(cryptProtectLocalMachine),
		uintptr(unsafe.Pointer(&out)),
	)
	if r == 0 {
		return nil, fmt.Errorf("secret: CryptUnprotectData failed (corrupt, tampered, or sealed on another machine): %w", callErr)
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(out.pbData)))
	return out.copyOut(), nil
}
