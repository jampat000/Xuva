//go:build !windows

package main

import (
	"context"
	"fmt"
	"os"
)

// On non-Windows builds the MSI self-updater is inert: there is no MSI to apply.
// The check/verify path (runUpdateCheck) still compiles and is exercised by unit
// tests with a fake applier; only the real apply is Windows-only. Docker/Linux
// installs upgrade by replacing the image/package.

func newUpdateApplier() updateApplier { return stubApplier{} }

type stubApplier struct{}

func (stubApplier) apply(_ context.Context, _ string, _ string) error {
	return fmt.Errorf("MSI self-update is only supported on Windows")
}

func runApplyStage(_ string) int {
	fmt.Fprintln(os.Stderr, "xuva update --apply is only supported on Windows")
	return 1
}
