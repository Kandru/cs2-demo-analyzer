// Package version holds build-time version information injected via ldflags.
package version

import "fmt"

var (
	// Version is the release version (YY.MM.BuildNr), injected at build time.
	Version = "dev"
	// Commit is the git revision, injected at build time.
	Commit = ""
	// Date is the build date, injected at build time.
	Date = ""
)

// String returns a human-readable version line.
func String() string {
	s := Version
	if Commit != "" {
		s += " (" + Commit + ")"
	}
	if Date != "" {
		s += " built " + Date
	}
	return s
}

// Short returns only the version number.
func Short() string { return Version }

// Info returns a multi-line version block.
func Info() string {
	return fmt.Sprintf("cs2-battlelog %s\ncommit: %s\ndate:   %s", Version, Commit, Date)
}
