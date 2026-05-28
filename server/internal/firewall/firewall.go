// Package firewall ensures inbound TCP traffic to the Xuva server is allowed
// through the host firewall. On Windows this means adding a Windows Defender
// Firewall rule via netsh; on other OSes this is a no-op (Linux/macOS do not
// block inbound by default in the way Windows does, and we don't ship a
// systemd-firewalld manager).
//
// Why this exists: a default-install of Windows blocks all unsolicited
// inbound TCP. Xuva binds to 0.0.0.0:8097 by default but other LAN clients
// (a phone, a tablet, a second laptop) can't reach it until a rule is
// added. Plex/Emby/Jellyfin all do this programmatically on first run; we
// match their behaviour.
package firewall

import (
	"context"
	"log/slog"
)

// EnsureInboundTCP registers (or verifies) a firewall rule that allows
// inbound TCP traffic on the given port. The rule is namespaced by name —
// re-running with the same name + port is idempotent and very cheap (the
// platform-specific implementation should detect an existing rule and skip).
//
// On non-Windows platforms this is a logged no-op; the package does not yet
// manage iptables/firewalld/pf, because Linux server installs usually run
// behind a reverse proxy or have already opened the port at deploy time.
//
// Errors are non-fatal — Xuva still starts; the user just won't be able to
// reach it from other LAN devices until they add the rule manually. The
// error is logged via slog so the cause is recoverable from xuva.ndjson.
func EnsureInboundTCP(ctx context.Context, port int, ruleName string) error {
	return ensureInboundTCP(ctx, port, ruleName)
}

// LogResult is a convenience wrapper that calls EnsureInboundTCP and logs
// the outcome at the appropriate level — info on success/no-op, warn on
// failure with the exact manual fix the user can run as admin.
func LogResult(ctx context.Context, logger *slog.Logger, port int, ruleName string) {
	if logger == nil {
		logger = slog.Default()
	}
	if err := EnsureInboundTCP(ctx, port, ruleName); err != nil {
		logger.Warn("firewall rule add failed",
			"port", port,
			"rule", ruleName,
			"err", err.Error(),
			"manual_fix", manualFixHint(port, ruleName),
		)
		return
	}
	logger.Info("firewall rule ensured", "port", port, "rule", ruleName)
}
