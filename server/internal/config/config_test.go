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
