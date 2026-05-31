//go:build !windows

package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// keyFileName is the AES-256 key persisted (mode 0600) in the key directory.
const keyFileName = "secret.key"

// fileStore seals values with AES-256-GCM under a key file. Sealed blobs are
// bound to that key: copying the data directory without secret.key makes the
// secrets unrecoverable.
type fileStore struct{ gcm cipher.AEAD }

// New returns a Store backed by an AES-256-GCM key file in keyDir. The key is
// created on first use with restrictive permissions. keyDir must be non-empty
// and writable (typically the server data directory).
func New(keyDir string) (Store, error) {
	if keyDir == "" {
		return nil, errors.New("secret: key directory is required")
	}
	key, err := loadOrCreateKey(keyDir)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &fileStore{gcm: gcm}, nil
}

func loadOrCreateKey(dir string) ([]byte, error) {
	path := filepath.Join(dir, keyFileName)
	switch b, err := os.ReadFile(path); {
	case err == nil:
		if len(b) != 32 {
			return nil, fmt.Errorf("secret: key file %s is corrupt (%d bytes, want 32)", path, len(b))
		}
		return b, nil
	case !errors.Is(err, os.ErrNotExist):
		return nil, fmt.Errorf("secret: read key file: %w", err)
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("secret: create key dir: %w", err)
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("secret: generate key: %w", err)
	}
	// Write atomically (temp + rename) so a crash can't leave a partial key.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, key, 0o600); err != nil {
		return nil, fmt.Errorf("secret: write key: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return nil, fmt.Errorf("secret: finalize key: %w", err)
	}
	return key, nil
}

// Seal returns version || nonce || AES-GCM(ciphertext+tag).
func (s *fileStore) Seal(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("secret: nonce: %w", err)
	}
	out := make([]byte, 0, 1+len(nonce)+len(plaintext)+s.gcm.Overhead())
	out = append(out, blobVersion)
	out = append(out, nonce...)
	return s.gcm.Seal(out, nonce, plaintext, nil), nil
}

func (s *fileStore) Open(sealed []byte) ([]byte, error) {
	ns := s.gcm.NonceSize()
	if len(sealed) < 1+ns {
		return nil, errors.New("secret: sealed blob too short")
	}
	if sealed[0] != blobVersion {
		return nil, fmt.Errorf("secret: unsupported blob version %d", sealed[0])
	}
	nonce := sealed[1 : 1+ns]
	ciphertext := sealed[1+ns:]
	plaintext, err := s.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("secret: open: %w", err)
	}
	return plaintext, nil
}
