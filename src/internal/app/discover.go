package app

import "strings"

// supportedDemo reports whether path looks like a CS2 demo we can parse.
func supportedDemo(path string) bool {
	p := strings.ToLower(path)
	return strings.HasSuffix(p, ".dem") ||
		strings.HasSuffix(p, ".dem.gz") ||
		strings.HasSuffix(p, ".dem.bz2")
}
