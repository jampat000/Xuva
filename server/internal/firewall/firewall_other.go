//go:build !windows

package firewall

import (
	"context"
	"fmt"
)

// ensureInboundTCP is a no-op on non-Windows platforms. Linux/macOS servers
// either run behind a reverse proxy (port already opened at deploy) or use
// iptables/firewalld/pf that this package does not yet manage. Returning
// nil so callers don't treat this as a failure on those platforms.
func ensureInboundTCP(_ context.Context, _ int, _ string) error {
	return nil
}

// manualFixHint produces a best-effort suggestion for the user on non-Windows
// platforms. Linux distros vary so this is necessarily generic.
func manualFixHint(port int, _ string) string {
	return fmt.Sprintf(
		"Allow TCP %d through your host firewall (e.g. `sudo ufw allow %d/tcp` on Ubuntu, "+
			"or `sudo firewall-cmd --add-port=%d/tcp --permanent && sudo firewall-cmd --reload` on Fedora).",
		port, port, port,
	)
}
