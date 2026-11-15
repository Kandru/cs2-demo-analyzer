package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscover(t *testing.T) {
	dir := t.TempDir()
	must := func(path string) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	must(filepath.Join(dir, "a.dem"))
	must(filepath.Join(dir, "nested", "b.dem.gz"))
	must(filepath.Join(dir, "nested", "c.txt"))
	must(filepath.Join(dir, "nested", ".hidden", "d.dem"))

	got, err := discover([]string{dir}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
	got, err = discover([]string{dir}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("non-recursive got %v", got)
	}
}

func TestSplitEnv(t *testing.T) {
	got := splitEnv("/a" + string(os.PathListSeparator) + "/b")
	if len(got) != 2 || got[0] != "/a" || got[1] != "/b" {
		t.Fatalf("got %v", got)
	}
}
