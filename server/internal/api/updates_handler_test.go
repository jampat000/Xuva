package api

import "testing"

func TestClassifyUpdateAsset(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		// MSI (Windows Service install) — the SYSTEM updater targets these.
		{"xuva-server-v1.2.3.msi", "windows-msi"},
		{"xuva-server-v1.2.3.msi.sha256", "windows-msi-checksum"},
		{"XUVA-SERVER-V1.2.3.MSI", "windows-msi"},
		// NSIS installer (desktop/tray) — the existing launcher applies these.
		{"xuva-v1.2.3-win-x64.exe", "windows-installer"},
		{"xuva-v1.2.3-win-x64.exe.sha256", "windows-installer-checksum"},
		// Portable zip.
		{"xuva-v1.2.3-win-x64.zip", "windows-portable"},
		{"xuva-v1.2.3-win-x64.zip.sha256", "windows-portable-checksum"},
		// Unknown.
		{"checksums.txt", "asset"},
		{"", "asset"},
		{"  xuva-server-v1.2.3.msi  ", "windows-msi"},
	}
	for _, tc := range cases {
		if got := classifyUpdateAsset(tc.name); got != tc.want {
			t.Errorf("classifyUpdateAsset(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}

// The .msi/.exe checksum suffixes must win over the bare .msi/.exe cases, or a
// sidecar would be mistaken for the installer itself and handed to the applier.
func TestClassifyUpdateAssetChecksumPrecedence(t *testing.T) {
	if got := classifyUpdateAsset("foo.msi.sha256"); got == "windows-msi" {
		t.Fatalf("foo.msi.sha256 classified as the installer, not its checksum")
	}
}

func TestCompareReleaseVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.2.4", "v1.2.3", 1},
		{"v1.2.3", "v1.2.4", -1},
		{"v1.2.3", "v1.2.3", 0},
		{"v1.3.0", "v1.2.9", 1},
		{"v2.0.0", "v1.9.9", 1},
		// "dev" (the un-ldflag'd build) is unparseable — never an upgrade target.
		// This is exactly the bug C0 fixes by injecting a real version into the MSI.
		{"v1.2.3", "dev", 0},
		{"dev", "v1.2.3", 0},
		{"1.2.4", "1.2.3", 1}, // tolerant of a missing leading "v"
		{"garbage", "v1.0.0", 0},
	}
	for _, tc := range cases {
		if got := compareReleaseVersions(tc.a, tc.b); got != tc.want {
			t.Errorf("compareReleaseVersions(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}
