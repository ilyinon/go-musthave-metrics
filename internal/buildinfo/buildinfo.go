package buildinfo

import "log"

// These variables are populated at build time via -ldflags.
// Defaults are empty; Print will render them as "N/A".
var (
	Version string
	Date    string
	Commit  string
)

// Print logs build information to stdout (via log package).
// Missing values are shown as "N/A".
func Print() {
	version := Version
	if version == "" {
		version = "N/A"
	}

	date := Date
	if date == "" {
		date = "N/A"
	}

	commit := Commit
	if commit == "" {
		commit = "N/A"
	}

	log.Printf("Build version: %s", version)
	log.Printf("Build date: %s", date)
	log.Printf("Build commit: %s", commit)
}
