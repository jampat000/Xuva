package config

import "testing"

func TestHTTPAddrLoopbackOnly(t *testing.T) {
	if !HTTPAddrLoopbackOnly("127.0.0.1:8097") {
		t.Fatal("expected loopback IPv4 address to be loopback-only")
	}
	if !HTTPAddrLoopbackOnly("localhost:8097") {
		t.Fatal("expected localhost address to be loopback-only")
	}
	if HTTPAddrLoopbackOnly("0.0.0.0:8097") {
		t.Fatal("expected wildcard address to not be loopback-only")
	}
}

func TestNormalizeWebOrigin(t *testing.T) {
	got, err := NormalizeWebOrigin("HTTP://media-server.local:8097/settings?x=1")
	if err == nil {
		t.Fatalf("expected path/query origin to be rejected, got %q", got)
	}
	got, err = NormalizeWebOrigin("HTTP://media-server.local:8097")
	if err != nil {
		t.Fatalf("normalize origin: %v", err)
	}
	if got != "http://media-server.local:8097" {
		t.Fatalf("expected normalized origin, got %q", got)
	}
}
