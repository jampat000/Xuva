package buildinfo

import "strings"

// These variables are populated by release builds through -ldflags -X.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = ""
)

type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date,omitempty"`
}

func Current() Info {
	version := strings.TrimSpace(Version)
	if version == "" {
		version = "dev"
	}
	commit := strings.TrimSpace(Commit)
	if commit == "" {
		commit = "unknown"
	}
	return Info{
		Version: version,
		Commit:  commit,
		Date:    strings.TrimSpace(Date),
	}
}
