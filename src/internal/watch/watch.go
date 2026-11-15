// Package watch discovers new demo files and feeds them to the analysis
// processor. It combines fsnotify (fast reaction) with periodic rescans
// (reliability on network mounts) and a settle window so partially written
// demos are never parsed.
package watch

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/Kandru/cs2-demo-analyzer/internal/config"
)

// Processor is the analysis pipeline used by the watcher.
type Processor interface {
	Process(ctx context.Context, path string) error
	Skip(path string, size int64, mod time.Time, force bool) bool
	Save() error
}

type stability struct {
	size  int64
	mod   time.Time
	count int
}

// Runner watches the configured inputs.
type Runner struct {
	cfg  config.Config
	proc Processor
	log  *slog.Logger

	mu      sync.Mutex
	stable  map[string]*stability
	queued  map[string]bool
	jobs    chan string
	wg      sync.WaitGroup
}

// Run blocks until ctx is cancelled.
func Run(ctx context.Context, cfg config.Config, proc Processor, log *slog.Logger) error {
	r := &Runner{
		cfg:    cfg,
		proc:   proc,
		log:    log,
		stable: map[string]*stability{},
		queued: map[string]bool{},
		jobs:   make(chan string, 128),
	}
	for i := 0; i < cfg.Workers; i++ {
		r.wg.Add(1)
		go r.worker(ctx)
	}

	watcher, werr := fsnotify.NewWatcher()
	if werr != nil {
		log.Warn("fsnotify unavailable; falling back to polling", "error", werr)
	}
	if watcher != nil {
		for _, in := range cfg.Inputs {
			if err := addRecursive(watcher, in); err != nil {
				log.Warn("could not watch input; polling still active", "input", in, "error", err)
			}
		}
	}

	wake := make(chan struct{}, 1)
	if watcher != nil {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case ev, ok := <-watcher.Events:
					if !ok {
						return
					}
					if ev.Op&fsnotify.Create != 0 {
						if info, err := os.Stat(ev.Name); err == nil && info.IsDir() && cfg.Recursive {
							_ = addRecursive(watcher, ev.Name)
						}
					}
					select {
					case wake <- struct{}{}:
					default:
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Warn("watcher error", "error", err)
				}
			}
		}()
	}

	log.Info("watching for demos", "inputs", cfg.Inputs, "output", outputDescription(cfg), "workers", cfg.Workers, "interval", cfg.Interval)
	r.scan(ctx)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			if watcher != nil {
				_ = watcher.Close()
			}
			close(r.jobs)
			r.wg.Wait()
			if err := r.proc.Save(); err != nil {
				log.Warn("could not save manifest on shutdown", "error", err)
			}
			log.Info("watcher stopped")
			return nil
		case <-ticker.C:
			r.scan(ctx)
		case <-wake:
			// Debounce bursts of fs events.
			timer := time.NewTimer(cfg.Debounce)
			select {
			case <-timer.C:
				r.scan(ctx)
			case <-ctx.Done():
				timer.Stop()
			}
		}
	}
}

func outputDescription(cfg config.Config) string {
	if cfg.Output == "" {
		return "alongside each demo"
	}
	return cfg.Output
}

func (r *Runner) scan(ctx context.Context) {
	for _, in := range r.cfg.Inputs {
		info, err := os.Stat(in)
		if err != nil {
			r.log.Warn("input unavailable", "input", in, "error", err)
			continue
		}
		if !info.IsDir() {
			if supportedDemo(in) {
				r.consider(ctx, in, info)
			}
			continue
		}
		err = filepath.WalkDir(in, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				name := d.Name()
				if path != in && (!r.cfg.Recursive || strings.HasPrefix(name, ".") || name == ".cs2battlelog") {
					if strings.HasPrefix(name, ".") || name == ".cs2battlelog" {
						return filepath.SkipDir
					}
					if !r.cfg.Recursive {
						return filepath.SkipDir
					}
				}
				return nil
			}
			if supportedDemo(path) {
				if fi, err := d.Info(); err == nil {
					r.consider(ctx, path, fi)
				}
			}
			return nil
		})
		if err != nil {
			r.log.Warn("scan failed", "input", in, "error", err)
		}
	}
}

func (r *Runner) consider(ctx context.Context, path string, info os.FileInfo) {
	if r.proc.Skip(path, info.Size(), info.ModTime(), r.cfg.Force) {
		r.mu.Lock()
		delete(r.stable, path)
		r.mu.Unlock()
		return
	}
	r.mu.Lock()
	st := r.stable[path]
	if st == nil || st.size != info.Size() || !st.mod.Equal(info.ModTime()) {
		r.stable[path] = &stability{size: info.Size(), mod: info.ModTime(), count: 1}
		r.mu.Unlock()
		return
	}
	st.count++
	ready := st.count >= 2 || r.cfg.Debounce == 0
	if !ready {
		r.mu.Unlock()
		return
	}
	if r.queued[path] {
		r.mu.Unlock()
		return
	}
	r.queued[path] = true
	r.mu.Unlock()

	select {
	case r.jobs <- path:
		r.log.Info("queued demo", "demo", path, "size", humanSize(info.Size()))
	case <-ctx.Done():
	default:
		r.mu.Lock()
		delete(r.queued, path)
		r.mu.Unlock()
		r.log.Warn("analysis queue full; will retry on next scan", "demo", path)
	}
}

func (r *Runner) worker(ctx context.Context) {
	defer r.wg.Done()
	for path := range r.jobs {
		if ctx.Err() != nil {
			r.done(path)
			continue
		}
		if err := r.proc.Process(ctx, path); err != nil {
			// Process logs the details; keep the watcher alive.
			r.log.Warn("demo processing failed", "demo", path, "error", err)
		}
		r.done(path)
	}
}

func (r *Runner) done(path string) {
	r.mu.Lock()
	delete(r.queued, path)
	delete(r.stable, path)
	r.mu.Unlock()
}

func addRecursive(w *fsnotify.Watcher, root string) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return w.Add(filepath.Dir(root))
	}
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if path != root && (strings.HasPrefix(d.Name(), ".") || d.Name() == ".cs2battlelog") {
			return filepath.SkipDir
		}
		if err := w.Add(path); err != nil {
			return nil
		}
		return nil
	})
}

func supportedDemo(path string) bool {
	p := strings.ToLower(path)
	return strings.HasSuffix(p, ".dem") ||
		strings.HasSuffix(p, ".dem.gz") ||
		strings.HasSuffix(p, ".dem.bz2")
}

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return formatInt(n) + " B"
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return formatFloat(float64(n)/float64(div), 1) + string("KMGTPE"[exp]) + "iB"
}

func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func formatFloat(v float64, digits int) string {
	pow := int64(1)
	for i := 0; i < digits; i++ {
		pow *= 10
	}
	n := int64(v*float64(pow) + 0.5)
	whole := n / pow
	frac := n % pow
	s := formatInt(whole) + "."
	for p := pow / 10; p > 0; p /= 10 {
		s += string(byte('0' + (frac/p)%10))
	}
	return s
}
