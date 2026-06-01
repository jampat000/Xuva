//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
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
// scheduled tasks that run `xuva-server update --scheduled` — one every 6 hours
// and one shortly after boot. Called on each service start, so it self-heals
// missing tasks and applies an opt-out toggle on restart. Idempotent. Runs as
// the LocalSystem service, which has the rights to manage a SYSTEM task.
//
// The tasks are built from schtasks command-line flags rather than an XML
// definition: schtasks constructs a schema-valid task internally, which avoids
// the /create /xml pitfalls (the file must be UTF-16 and its <Settings>
// elements must be in a strict order — both easy to get wrong and the cause of
// the task silently failing to register).
func ensureUpdaterTask(enabled bool) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve own path: %w", err)
	}
	for _, t := range updaterTasks {
		var args []string
		if enabled {
			args = updaterCreateArgs(t.name, exe, t.schedule)
		} else {
			args = updaterDeleteArgs(t.name)
		}
		cmd := exec.Command("schtasks", args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNoWindow}
		out, runErr := cmd.CombinedOutput()
		if runErr != nil && enabled {
			// Removing an already-absent task is fine — disabling stays quiet.
			return fmt.Errorf("schtasks %s: %v: %s", t.name, runErr, strings.TrimSpace(string(out)))
		}
	}
	return nil
}
