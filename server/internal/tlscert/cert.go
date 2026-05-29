// Package tlscert handles TLS certificate provisioning for Xuva's optional
// HTTPS listener. It generates a self-signed ECDSA certificate the first time
// TLS is enabled and reuses it on subsequent starts.
package tlscert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const certValidYears = 10

// Ensure loads a TLS certificate from certPath/keyPath. If either file does not
// exist, a new self-signed ECDSA-P256 certificate is generated covering the
// supplied hostnames and IP addresses and written to both paths.
//
// Returns the certificate, a human-readable SHA-256 fingerprint (colon-separated
// hex pairs), and any error. The fingerprint is suitable for logging so users can
// pin it in their browser.
func Ensure(certPath, keyPath string, hosts []string) (tls.Certificate, string, error) {
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		if genErr := generate(certPath, keyPath, hosts); genErr != nil {
			return tls.Certificate{}, "", fmt.Errorf("generate self-signed cert: %w", genErr)
		}
	}
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return tls.Certificate{}, "", fmt.Errorf("load tls cert: %w", err)
	}
	fp, err := fingerprint(cert)
	if err != nil {
		return cert, "", nil //nolint:nilerr
	}
	return cert, fp, nil
}

// fingerprint returns a colon-separated uppercase SHA-256 hex fingerprint of the leaf cert.
func fingerprint(cert tls.Certificate) (string, error) {
	if len(cert.Certificate) == 0 {
		return "", fmt.Errorf("empty certificate chain")
	}
	sum := sha256.Sum256(cert.Certificate[0])
	pairs := make([]string, len(sum))
	for i, b := range sum {
		pairs[i] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(pairs, ":"), nil
}

func generate(certPath, keyPath string, hosts []string) error {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"Xuva Self-Signed"},
			CommonName:   "xuva",
		},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().AddDate(certValidYears, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Add all provided hosts as DNS names or IP SANs.
	tmpl.DNSNames = append(tmpl.DNSNames, "localhost")
	tmpl.IPAddresses = append(tmpl.IPAddresses, net.ParseIP("127.0.0.1"), net.ParseIP("::1"))
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			tmpl.IPAddresses = append(tmpl.IPAddresses, ip)
		} else if h != "" && h != "localhost" {
			tmpl.DNSNames = append(tmpl.DNSNames, h)
		}
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(certPath), 0o700); err != nil {
		return err
	}

	certFile, err := os.OpenFile(certPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		certFile.Close()
		return err
	}
	if err := certFile.Close(); err != nil {
		return err
	}

	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	privDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		keyFile.Close()
		return err
	}
	if err := pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privDER}); err != nil {
		keyFile.Close()
		return err
	}
	return keyFile.Close()
}
