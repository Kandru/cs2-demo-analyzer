package update

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"
)

// Version is a release version in the repository's YY.MM.BuildNr scheme.
type Version struct {
	Year  int
	Month int
	Build int
}

// ParseVersion parses a release tag such as 26.11.1 or v26.11.1.
func ParseVersion(s string) (Version, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	if s == "" {
		return Version{}, fmt.Errorf("invalid version: empty")
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version %q: expected YY.MM.BuildNr", s)
	}
	var nums [3]int
	for i, part := range parts {
		if part == "" {
			return Version{}, fmt.Errorf("invalid version %q: empty component", s)
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return Version{}, fmt.Errorf("invalid version %q: %q is not a non-negative integer", s, part)
		}
		nums[i] = n
	}
	return Version{Year: nums[0], Month: nums[1], Build: nums[2]}, nil
}

// String returns the version without a leading v.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Year, v.Month, v.Build)
}

// Compare returns -1, 0, or 1 when a is older than, equal to, or newer than b.
func Compare(a, b Version) int {
	if c := cmp.Compare(a.Year, b.Year); c != 0 {
		return c
	}
	if c := cmp.Compare(a.Month, b.Month); c != 0 {
		return c
	}
	return cmp.Compare(a.Build, b.Build)
}
