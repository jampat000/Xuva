//go:build windows

package firewall

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ensureInboundTCP / ensureInboundUDP — thin protocol-specific wrappers
// around ensureInbound. Kept as separate symbols so callers (and tests) can
// pin the protocol without threading a string through.
func ensureInboundTCP(ctx context.Context, port int, ruleName string) error {
	return ensureInbound(ctx, "TCP", port, ruleName)
}

func ensureInboundUDP(ctx context.Context, port int, ruleName string) error {
	return ensureInbound(ctx, "UDP", port, ruleName)
}

// ensureInbound uses `netsh advfirewall firewall` to add (or verify) an
// inbound-allow rule for the given protocol+port. netsh is shipped with
// every supported Windows version, doesn't require any PowerShell modules,
// and can be driven non-interactively.
//
// The shape of a Plex-style rule:
//
//	netsh advfirewall firewall add rule name="Xuva Server (8097)" \
//	  dir=in action=allow protocol=TCP localport=8097 \
//	  profile=private,domain enable=yes
//
// We intentionally scope to Private + Domain profiles only — Public is for
// untrusted networks (coffee shop wifi etc.) and exposing the server there
// would be a footgun. Users who really want public access can adjust the
// rule manually.
//
// The function is idempotent: if a rule with the same name already exists,
// we delete it first and re-add. This means a port change in settings still
// produces a correct rule, and stale rules from earlier versions get
// rewritten with current parameters.
func ensureInbound(ctx context.Context, protocol string, port int, ruleName string) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port: %d", port)
	}
	switch protocol {
	case "TCP", "UDP":
	default:
		return fmt.Errorf("invalid protocol: %q (expected TCP or UDP)", protocol)
	}
	displayName := fmt.Sprintf("%s (%d)", ruleName, port)

	// Check whether a rule with our display name already exists, with the
	// expected port. If yes, leave it alone — netsh delete+add churn is
	// expensive (~250ms) and a frequent restart would hammer it.
	exists, sameConfig, err := ruleExists(ctx, displayName, port)
	if err != nil {
		// Errors from the show command are non-fatal — assume missing.
		exists = false
	}
	if exists && sameConfig {
		return nil
	}
	if exists && !sameConfig {
		// Drop the stale rule before re-adding with current parameters.
		if delErr := runNetsh(ctx,
			"advfirewall", "firewall", "delete", "rule",
			"name=", displayName,
		); delErr != nil {
			// Soldier on — the add below will fail with a duplicate-name
			// error if the delete didn't work, and we'll surface that.
			_ = delErr
		}
	}
	return runNetsh(ctx,
		"advfirewall", "firewall", "add", "rule",
		"name="+displayName,
		"dir=in",
		"action=allow",
		"protocol="+protocol,
		"localport="+strconv.Itoa(port),
		"profile=private,domain",
		"enable=yes",
		"description=Xuva media server inbound "+protocol+" (auto-managed)",
	)
}

// ruleExists returns (exists, samePort, err). When `exists` is true and
// `samePort` is false, the caller should delete the stale rule before
// re-adding.
func ruleExists(ctx context.Context, displayName string, port int) (bool, bool, error) {
	out, err := exec.CommandContext(ctx, "netsh", //nolint:gosec // fixed args
		"advfirewall", "firewall", "show", "rule",
		"name="+displayName,
	).CombinedOutput()
	if err != nil {
		// netsh exits non-zero with "No rules match the specified criteria"
		// when the rule is absent. Treat that as "doesn't exist" rather than
		// a hard error.
		if strings.Contains(string(out), "No rules match") {
			return false, false, nil
		}
		return false, false, fmt.Errorf("netsh show: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	// Rule exists — verify the LocalPort line matches.
	needle := fmt.Sprintf("LocalPort:                            %d", port)
	if strings.Contains(string(out), needle) {
		return true, true, nil
	}
	// Fallback: just match the port number after "LocalPort:" to handle
	// locale-translated label widths.
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "LocalPort:") {
			if strings.Contains(line, strconv.Itoa(port)) {
				return true, true, nil
			}
		}
	}
	return true, false, nil
}

func runNetsh(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "netsh", args...) //nolint:gosec // fixed args from caller
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	msg := strings.TrimSpace(string(out))
	if isAccessDenied(msg, err) {
		return errors.New("access denied — Xuva must run as administrator to manage firewall rules. " +
			"Either elevate the Xuva process, or open the port manually (see manual_fix hint).")
	}
	return fmt.Errorf("netsh %s: %w (%s)", args[0], err, msg)
}

func isAccessDenied(out string, err error) bool {
	if err == nil {
		return false
	}
	low := strings.ToLower(out)
	switch {
	case strings.Contains(low, "access is denied"),
		strings.Contains(low, "requires elevation"),
		strings.Contains(low, "the requested operation requires"):
		return true
	}
	return false
}

func manualFixHint(port int, ruleName string) string {
	return manualFixHintProto("TCP", port, ruleName)
}

func manualFixHintProto(protocol string, port int, ruleName string) string {
	return fmt.Sprintf(
		`Run as Administrator: netsh advfirewall firewall add rule name="%s (%d)" dir=in action=allow protocol=%s localport=%d profile=private,domain`,
		ruleName, port, protocol, port,
	)
}
