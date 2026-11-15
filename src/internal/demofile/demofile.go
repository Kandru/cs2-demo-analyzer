// Package demofile recognizes CS2 demo filenames and derives report names.
package demofile

import (
	"os"
	"path/filepath"
	"strings"
)

// suffixes are stripped in order, longest first, when naming a demo.
var suffixes = [...]string{".dem.gz", ".dem.bz2", ".dem", ".gz", ".bz2"}

// IsSupported reports whether path has a supported demo extension.
func IsSupported(path string) bool {
	p := strings.ToLower(path)
	return strings.HasSuffix(p, ".dem") ||
		strings.HasSuffix(p, ".dem.gz") ||
		strings.HasSuffix(p, ".dem.bz2")
}

// BaseName returns the report base name for a demo path, stripping a supported
// extension and archive suffix.
func BaseName(path string) string {
	base := filepath.Base(path)
	lower := strings.ToLower(base)
	for _, ext := range suffixes {
		if strings.HasSuffix(lower, ext) {
			return base[:len(base)-len(ext)]
		}
	}
	return base
}

// FindSibling returns a demo next to a report that uses the same base name.
func FindSibling(reportPath string) string {
	base := strings.TrimSuffix(reportPath, filepath.Ext(reportPath))
	for _, ext := range suffixes {
		cand := base + ext
		if info, err := os.Stat(cand); err == nil && info.Mode().IsRegular() {
			return cand
		}
	}
	return ""
}
