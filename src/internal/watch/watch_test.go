package watch

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kandru/cs2-demo-analyzer/internal/config"
)

type fakeProc struct {
	processed chan string
}

func (f *fakeProc) Process(ctx context.Context, path string) error {
	f.processed <- path
	return nil
}
func (f *fakeProc) Skip(path string, size int64, mod time.Time, force bool) bool { return false }
func (f *fakeProc) Save() error                                                  { return nil }

func TestScanSettlesFiles(t *testing.T) {
	dir := t.TempDir()
	demo := filepath.Join(dir, "match.dem")
	if err := os.WriteFile(demo, []byte("demo"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Inputs = []string{dir}
	cfg.Debounce = time.Millisecond
	proc := &fakeProc{processed: make(chan string, 4)}
	r := &Runner{
		cfg:    cfg,
		proc:   proc,
		stable: map[string]*stability{},
		queued: map[string]bool{},
		jobs:   make(chan string, 4),
		log:    slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}
	r.scan(context.Background())
	select {
	case p := <-r.jobs:
		t.Fatalf("queued before settling: %s", p)
	default:
	}
	r.scan(context.Background())
	select {
	case p := <-r.jobs:
		if p != demo {
			t.Fatalf("queued %q want %q", p, demo)
		}
	case <-time.After(time.Second):
		t.Fatal("file was not queued after settling")
	}
}
