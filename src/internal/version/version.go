// Package version holds build-time version information injected via ldflags.
package version

import (
	"fmt"
	"strings"
)

// unknownCommit is the placeholder injected when git metadata is unavailable.
const unknownCommit = "unknown"

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

// CacheKey is the version stamp embedded in report artifacts and compared when
// deciding whether a report can be reused. Release builds use the release
// version. Unreleased "dev" builds include the commit so a rebuild does not
// silently reuse stale output. An empty result means reports are never reused.
func CacheKey() string {
	v := strings.TrimSpace(Version)
	if v != "" && !strings.EqualFold(v, "dev") {
		return v
	}
	c := strings.TrimSpace(Commit)
	if c != "" && !strings.EqualFold(c, unknownCommit) {
		return "dev+" + c
	}
	return ""
}

// Info returns a multi-line version block.
func Info() string {
	return fmt.Sprintf("cs2-demo-analyzer %s\ncommit: %s\ndate:   %s", Version, Commit, Date)
}
