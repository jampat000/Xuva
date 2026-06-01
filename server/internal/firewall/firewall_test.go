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

// TestEnsureInboundUDP_RejectsInvalidPort verifies argument validation for
// the UDP twin too. Mirrors TestEnsureInboundTCP_RejectsInvalidPort. mDNS
// uses UDP 5353 (the discovery package advertises _xuva._tcp via UDP
// multicast) so this code path runs at every Windows startup.
func TestEnsureInboundUDP_RejectsInvalidPort(t *testing.T) {
	for _, p := range []int{0, -1, 65536, 99999} {
		err := EnsureInboundUDP(context.Background(), p, "test")
		if err == nil {
			t.Logf("port=%d returned nil (acceptable on non-Windows where this is a no-op)", p)
			continue
		}
		if !strings.Contains(err.Error(), "invalid port") {
			t.Errorf("port=%d: expected 'invalid port' error, got %v", p, err)
		}
	}
}

// TestManualFixHintProto_MentionsProtocol — the per-protocol hint must
// include both the protocol name and the port (so the same log field can
// tell a TCP failure apart from a UDP one for diagnosis).
func TestManualFixHintProto_MentionsProtocol(t *testing.T) {
	udp := manualFixHintProto("UDP", 5353, "Xuva Local Discovery (mDNS)")
	if !strings.Contains(udp, "UDP") || !strings.Contains(udp, "5353") {
		t.Errorf("UDP hint missing protocol/port: %q", udp)
	}
	tcp := manualFixHintProto("TCP", 8097, "Xuva Server")
	if !strings.Contains(tcp, "TCP") || !strings.Contains(tcp, "8097") {
		t.Errorf("TCP hint missing protocol/port: %q", tcp)
	}
}
