// Package secret provides encryption-at-rest for sensitive values such as SMB
// share credentials, bound to the local machine.
//
// The sealed output is an opaque, versioned byte blob that callers persist
// (e.g. base64-encoded inside settings.json). Sealing is machine-bound: a data
// directory copied to a different machine cannot decrypt secrets sealed here.
//
//   - Windows: DPAPI with machine scope (CryptProtectData/CryptUnprotectData,
//     CRYPTPROTECT_LOCAL_MACHINE). Any process on the machine can open; the
//     SYSTEM-run service and an admin console process therefore share secrets.
//   - Other platforms: AES-256-GCM under a randomly generated key file
//     (secret.key, mode 0600) kept in the runtime/key directory. This is the
//     documented cross-platform fallback (libsecret remains a future option).
//
// Both implementations are integrity-protected: Open fails on a tampered blob.
package secret

// Store seals and opens secret values. Implementations bind the ciphertext to
// the local machine so a copied data directory can't be trivially decrypted
// elsewhere. Implementations must be safe for concurrent use.
type Store interface {
	// Seal encrypts plaintext, returning an opaque blob safe to persist.
	Seal(plaintext []byte) ([]byte, error)
	// Open decrypts a blob previously produced by Seal on this machine.
	// It returns an error if the blob is corrupt, tampered, or was sealed on
	// a different machine.
	Open(sealed []byte) ([]byte, error)
}

// blobVersion is the first byte of every sealed blob, letting the format evolve
// without ambiguity. Bump only with a corresponding Open compatibility path.
const blobVersion = 0x01
