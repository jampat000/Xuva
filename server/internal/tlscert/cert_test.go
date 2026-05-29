package tlscert

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureGeneratesAndLoads(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	cert, fp, err := Ensure(certPath, keyPath, []string{"example.local", "192.168.1.1"})
	if err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	if len(cert.Certificate) == 0 {
		t.Fatal("expected non-empty certificate chain")
	}
	if !strings.Contains(fp, ":") || len(fp) < 10 {
		t.Errorf("unexpected fingerprint format: %q", fp)
	}

	// Second call should reuse existing files without error.
	cert2, fp2, err := Ensure(certPath, keyPath, nil)
	if err != nil {
		t.Fatalf("Ensure (reload): %v", err)
	}
	if fp2 != fp {
		t.Errorf("fingerprint changed on reload: %q vs %q", fp, fp2)
	}
	if len(cert2.Certificate) == 0 {
		t.Fatal("reloaded cert has empty chain")
	}
}

func TestEnsureCreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "tls", "sub", "cert.pem")
	keyPath := filepath.Join(dir, "tls", "sub", "key.pem")

	_, _, err := Ensure(certPath, keyPath, nil)
	if err != nil {
		t.Fatalf("Ensure with nested dir: %v", err)
	}
	if _, err := os.Stat(certPath); err != nil {
		t.Errorf("cert file not created: %v", err)
	}
}
