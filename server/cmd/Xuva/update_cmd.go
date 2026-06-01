package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jampat000/Xuva/server/internal/buildinfo"
	"github.com/jampat000/Xuva/server/internal/updater"
)

// updateApplier applies a verified MSI. It's an interface so the check/verify
// orchestration can be unit-tested with a fake on any OS, while the real
// implementation (self-copy + msiexec) lives in the Windows-tagged file.
type updateApplier interface {
	// apply hands off a verified MSI for installation. On Windows it re-execs a
	// temp copy of this binary (so the installed exe isn't locked during the
	// MajorUpgrade) which runs msiexec; returning nil means the handoff started.
	apply(ctx context.Context, msiPath, version string) error
}

// handleUpdateCommand implements `xuva-server update [--apply <msi>]`. It runs as
// a normal (for the service self-updater, SYSTEM) console process and exits.
// Returns true when an update verb was handled so main() stops.
//
// Two stages keep the updater from locking the file it's replacing (Chrome/Edge
// pattern — the process that runs the installer is never the one being upgraded):
//   - `update`               — stage 1: registered as the scheduled-task action;
//     checks the feed, downloads+verifies the MSI, then re-execs a temp self-copy.
//   - `update --apply <msi>` — stage 2: the temp copy runs msiexec and waits.
func handleUpdateCommand(args []string) bool {
	if len(args) < 2 || args[1] != "update" {
		return false
	}
	if len(args) >= 4 && args[2] == "--apply" {
		os.Exit(runApplyStage(args[3]))
	}
	if err := runUpdateCheck(context.Background(), newUpdateApplier()); err != nil {
		fmt.Fprintf(os.Stderr, "xuva update: %v\n", err)
		os.Exit(1)
	}
	return true
}

// runUpdateCheck is stage 1: poll the release feed, and if a newer release ships
// a Windows MSI, download it, verify its SHA-256 (fail closed if unverifiable),
// and hand the verified path to the applier. OS-agnostic so it's unit-testable.
func runUpdateCheck(ctx context.Context, applier updateApplier) error {
	current := buildinfo.Current().Version
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	rel, err := updater.FetchLatest(ctx)
	if err != nil {
		return err
	}
	if updater.CompareVersions(rel.Version, current) <= 0 {
		slog.Info("xuva update: already up to date", "current", current, "latest", rel.Version)
		return nil
	}
	asset, ok := updater.FindAsset(rel.Assets, "windows-msi")
	if !ok {
		return fmt.Errorf("release %s does not include a Windows MSI", rel.Version)
	}

	dir := filepath.Join(updateDownloadRoot(), "updates", updater.SafePathName(rel.Version))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create update directory: %w", err)
	}
	msiPath := filepath.Join(dir, asset.Name)
	if err := updater.Download(ctx, asset.URL, msiPath); err != nil {
		return err
	}

	// Prefer the feed's asset digest; fall back to a published .msi.sha256
	// sidecar. An MSI we cannot verify is never applied.
	expected := updater.ExpectedSHA(asset)
	if expected == "" {
		if sum, ok := updater.FindAsset(rel.Assets, "windows-msi-checksum"); ok && sum.URL != "" {
			sumPath := filepath.Join(dir, sum.Name)
			if err := updater.Download(ctx, sum.URL, sumPath); err == nil {
				expected = updater.ParseSHA256File(sumPath)
			}
		}
	}
	if expected == "" {
		_ = os.Remove(msiPath)
		return fmt.Errorf("release %s has no checksum for the MSI; refusing to apply an unverified update", rel.Version)
	}
	actual, err := updater.SHA256File(msiPath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(expected, actual) {
		_ = os.Remove(msiPath)
		return fmt.Errorf("MSI checksum mismatch for %s (expected %s, got %s)", rel.Version, expected, actual)
	}

	slog.Info("xuva update: verified newer release, applying",
		"current", current, "latest", rel.Version, "msi", msiPath)
	return applier.apply(ctx, msiPath, rel.Version)
}

// updateDownloadRoot is where the verb stages downloads. It mirrors the service
// runtime home (XUVA_RUNTIME_HOME, else C:\ProgramData\Xuva) so a default and a
// relocated install both keep update artifacts alongside the rest of the state.
// The scheduled task usually runs without the service's per-service env, so the
// ProgramData fallback is the common path.
func updateDownloadRoot() string {
	if r := strings.TrimSpace(os.Getenv("XUVA_RUNTIME_HOME")); r != "" {
		return r
	}
	if pd := strings.TrimSpace(os.Getenv("PROGRAMDATA")); pd != "" {
		return filepath.Join(pd, "Xuva")
	}
	return filepath.Join(os.TempDir(), "Xuva")
}
