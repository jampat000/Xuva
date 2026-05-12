package config

import "testing"

func TestFromEnvDevAuthBypassDefaultsOff(t *testing.T) {
	t.Setenv("LORIVO_DEV_AUTH_BYPASS", "")
	cfg := FromEnv()
	if cfg.DevAuthBypass {
		t.Fatal("expected dev auth bypass to be disabled by default")
	}
}

func TestDevAuthBypassActiveRequiresLoopbackAddress(t *testing.T) {
	loopback := Config{HTTPAddr: "127.0.0.1:8097", DevAuthBypass: true}
	if !DevAuthBypassActive(loopback) {
		t.Fatal("expected loopback dev auth bypass to be active")
	}

	localhost := Config{HTTPAddr: "localhost:8097", DevAuthBypass: true}
	if !DevAuthBypassActive(localhost) {
		t.Fatal("expected localhost dev auth bypass to be active")
	}

	lan := Config{HTTPAddr: "0.0.0.0:8097", DevAuthBypass: true}
	if DevAuthBypassActive(lan) {
		t.Fatal("expected non-loopback dev auth bypass to remain disabled")
	}

	disabledAuth := Config{HTTPAddr: "127.0.0.1:8097", DevAuthBypass: true, AuthDisabled: true}
	if DevAuthBypassActive(disabledAuth) {
		t.Fatal("expected dev auth bypass to stay off when auth is fully disabled")
	}
}
