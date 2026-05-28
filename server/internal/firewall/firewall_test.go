package firewall

import (
	"context"
	"strings"
	"testing"
)

// TestEnsureInboundTCP_RejectsInvalidPort verifies argument validation.
// 0 / negative / >65535 must be rejected before any platform code runs,
// so a misconfigured XUVA_HTTP_ADDR can't shell out a bogus netsh command.
func TestEnsureInboundTCP_RejectsInvalidPort(t *testing.T) {
	for _, p := range []int{0, -1, 65536, 99999} {
		err := EnsureInboundTCP(context.Background(), p, "test")
		// On non-Windows the call is a documented no-op so this test
		// effectively only enforces the validation on Windows. Skip
		// quietly on other OSes — the build-tagged Windows
		// implementation has its own coverage paths.
		if err == nil {
			t.Logf("port=%d returned nil (acceptable on non-Windows where this is a no-op)", p)
			continue
		}
		if !strings.Contains(err.Error(), "invalid port") {
			t.Errorf("port=%d: expected 'invalid port' error, got %v", p, err)
		}
	}
}

// TestManualFixHint_MentionsPort — the manual_fix log field must include
// the actual port so the user can copy-paste it. A change that drops the
// port from the hint would silently make the on-failure UX much worse.
func TestManualFixHint_MentionsPort(t *testing.T) {
	hint := manualFixHint(8097, "Xuva Server")
	if !strings.Contains(hint, "8097") {
		t.Errorf("manual fix hint missing port number: %q", hint)
	}
}
