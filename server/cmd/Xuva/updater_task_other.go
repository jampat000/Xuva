//go:build !windows

package main

import "os"

// On non-Windows builds there's no Task Scheduler or registry. Auto-update is on
// by default and only the XUVA_AUTO_UPDATE env var can disable it (exercised by
// the cross-platform tests); ensureUpdaterTask is a no-op since the scheduled
// task is a Windows-only concept.

func autoUpdateEnabled() bool {
	if v, ok := os.LookupEnv("XUVA_AUTO_UPDATE"); ok {
		return !autoUpdateDisabledByValue(v)
	}
	return true
}

func ensureUpdaterTask(_ bool) error { return nil }
