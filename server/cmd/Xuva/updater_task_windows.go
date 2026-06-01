//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

// autoUpdateEnabled reports whether the unattended self-updater should run. It's
// on by default; an operator opts out via the XUVA_AUTO_UPDATE env var (wins if
// set) or the HKLM\SOFTWARE\Xuva\AutoUpdate registry value (which the MSI writes
// from its XUVA_AUTO_UPDATE property, and an admin can flip post-install).
func autoUpdateEnabled() bool {
	if v, ok := os.LookupEnv("XUVA_AUTO_UPDATE"); ok {
		return !autoUpdateDisabledByValue(v)
	}
	if v, ok := readAutoUpdateRegistry(); ok {
		return !autoUpdateDisabledByValue(v)
	}
	return true
}

func readAutoUpdateRegistry() (string, bool) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Xuva`, registry.QUERY_VALUE)
	if err != nil {
		return "", false
	}
	defer k.Close()
	if s, _, err := k.GetStringValue("AutoUpdate"); err == nil {
		return s, true
	}
	if n, _, err := k.GetIntegerValue("AutoUpdate"); err == nil {
		if n == 0 {
			return "0", true
		}
		return "1", true
	}
	return "", false
}

// ensureUpdaterTask registers (enabled) or removes (disabled) the SYSTEM
// scheduled task that runs `xuva-server update --scheduled`. Called on each
// service start, so it self-heals a missing task and applies an opt-out toggle
// on restart. Idempotent. Runs as the LocalSystem service, which has the rights
// to manage a SYSTEM task.
func ensureUpdaterTask(enabled bool) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve own path: %w", err)
	}
	xmlPath := filepath.Join(filepath.Dir(exe), "XuvaUpdater.xml")
	if enabled {
		if _, statErr := os.Stat(xmlPath); statErr != nil {
			return fmt.Errorf("updater task definition missing at %s: %w", xmlPath, statErr)
		}
	}

	cmd := exec.Command("schtasks", updaterTaskArgs(enabled, xmlPath)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNoWindow}
	out, runErr := cmd.CombinedOutput()
	if runErr != nil {
		if enabled {
			return fmt.Errorf("schtasks create: %v: %s", runErr, strings.TrimSpace(string(out)))
		}
		// Removing an already-absent task is fine — disabling must be idempotent.
		return nil
	}
	return nil
}
