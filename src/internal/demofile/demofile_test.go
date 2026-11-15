package demofile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsSupported(t *testing.T) {
	for _, p := range []string{"a.dem", "A.DEM.GZ", "x.dem.bz2"} {
		if !IsSupported(p) {
			t.Errorf("%q should be supported", p)
		}
	}
	for _, p := range []string{"a.txt", "a.dem.info", "a", "a.gz"} {
		if IsSupported(p) {
			t.Errorf("%q should not be supported", p)
		}
	}
}

func TestBaseName(t *testing.T) {
	cases := map[string]string{
		"match.dem":         "match",
		"match.dem.gz":      "match",
		"match.dem.bz2":     "match",
		"match.gz":          "match",
		"match.bz2":         "match",
		"match.demo":        "match.demo",
		"archive/match.DEM": "match",
	}
	for path, want := range cases {
		if got := BaseName(path); got != want {
			t.Errorf("BaseName(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestFindSibling(t *testing.T) {
	dir := t.TempDir()
	demo := filepath.Join(dir, "match.dem.gz")
	if err := os.WriteFile(demo, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := FindSibling(filepath.Join(dir, "match.html")); got != demo {
		t.Fatalf("got %q, want %q", got, demo)
	}
	if got := FindSibling(filepath.Join(dir, "missing.html")); got != "" {
		t.Fatalf("unexpected sibling: %q", got)
	}
}
