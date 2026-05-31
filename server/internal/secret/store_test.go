package secret

import (
	"bytes"
	"testing"
)

// These tests are platform-agnostic: on non-Windows they exercise the
// AES-256-GCM file store, on Windows the DPAPI store. Both must satisfy the
// same contract (round-trip, no plaintext leakage, tamper rejection, machine
// stability across Store instances).

func TestSealOpenRoundTrip(t *testing.T) {
	st, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	cases := [][]byte{
		nil,
		[]byte(""),
		[]byte("hunter2"),
		[]byte(`{"username":"media","password":"p@ss w/ spaces & symbols #1"}`),
		bytes.Repeat([]byte{0xAB}, 10_000),
	}
	for _, pt := range cases {
		sealed, err := st.Seal(pt)
		if err != nil {
			t.Fatalf("Seal(%d bytes): %v", len(pt), err)
		}
		if len(pt) > 0 && bytes.Contains(sealed, pt) {
			t.Errorf("sealed blob leaks plaintext for %d-byte input", len(pt))
		}
		got, err := st.Open(sealed)
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		if !bytes.Equal(got, pt) && !(len(got) == 0 && len(pt) == 0) {
			t.Errorf("round-trip mismatch: got %q want %q", got, pt)
		}
	}
}

func TestOpenRejectsTamperedBlob(t *testing.T) {
	st, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sealed, err := st.Seal([]byte("secret-credential"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	tampered := append([]byte(nil), sealed...)
	tampered[len(tampered)-1] ^= 0xFF // flip a bit in the auth tag / ciphertext
	if _, err := st.Open(tampered); err == nil {
		t.Fatal("expected Open to fail on a tampered blob")
	}
}

func TestOpenRejectsBadVersion(t *testing.T) {
	st, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sealed, err := st.Seal([]byte("x"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	sealed[0] = 0xFE
	if _, err := st.Open(sealed); err == nil {
		t.Fatal("expected Open to reject an unknown version byte")
	}
}

// A new Store over the same key directory must open blobs sealed by an earlier
// instance. On the file store this proves the key persists; on DPAPI it holds
// because machine-scope blobs are independent of any per-process state.
func TestStableAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	a, err := New(dir)
	if err != nil {
		t.Fatalf("New a: %v", err)
	}
	sealed, err := a.Seal([]byte("durable"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	b, err := New(dir)
	if err != nil {
		t.Fatalf("New b: %v", err)
	}
	got, err := b.Open(sealed)
	if err != nil {
		t.Fatalf("Open with second instance: %v", err)
	}
	if string(got) != "durable" {
		t.Errorf("got %q, want %q", got, "durable")
	}
}
