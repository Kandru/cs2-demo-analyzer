// Package app wires the CLI, parser, detectors, and reporters together.
package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Kandru/cs2-demo-analyzer/internal/config"
	"github.com/Kandru/cs2-demo-analyzer/internal/store"
	"github.com/Kandru/cs2-demo-analyzer/internal/version"
	"github.com/Kandru/cs2-demo-analyzer/internal/watch"
)

// Run executes the CLI and returns a process exit code.
func Run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return 0
	}
	switch args[0] {
	case "help", "-h", "--help":
		printUsage(os.Stdout)
		return 0
	case "version", "-v", "--version":
		fmt.Fprintln(os.Stdout, version.Info())
		return 0
	case "analyze":
		return runAnalyze(ctx, args[1:])
	case "watch":
		return runWatch(ctx, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		printUsage(os.Stderr)
		return 2
	}
}

func runAnalyze(ctx context.Context, args []string) int {
	cfg := config.Default()
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	commonFlags(fs, &cfg)
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "run 'cs2-battlelog help' for usage")
		return 2
	}
	cfg.Inputs = append(cfg.Inputs, fs.Args()...)
	if err := cfg.Validate(true); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	logger := newLogger(cfg)
	demos, err := discover(cfg.Inputs, cfg.Recursive)
	if err != nil {
		logger.Error("discover demos", "error", err)
		return 1
	}
	if len(demos) == 0 {
		logger.Error("no demos found", "inputs", cfg.Inputs)
		return 1
	}
	st, err := store.Open(cfg.ManifestDir())
	if err != nil {
		logger.Error("open manifest", "error", err)
		return 1
	}
	proc := NewProcessor(cfg, st, logger)
	logger.Info("starting analysis", "demos", len(demos), "workers", cfg.Workers)

	jobs := make(chan string)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var failed, done int
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				if err := proc.Process(ctx, path); err != nil {
					mu.Lock()
					failed++
					mu.Unlock()
				} else {
					mu.Lock()
					done++
					mu.Unlock()
				}
			}
		}()
	}
loop:
	for _, d := range demos {
		select {
		case jobs <- d:
		case <-ctx.Done():
			break loop
		}
	}
	close(jobs)
	wg.Wait()
	_ = proc.Save()

	fmt.Fprintf(os.Stdout, "analyzed %d demo(s), %d failed\n", done, failed)
	if failed > 0 || ctx.Err() != nil {
		return 1
	}
	return 0
}

func runWatch(ctx context.Context, args []string) int {
	cfg := config.Default()
	cfg.Workers = 1
	fs := flag.NewFlagSet("watch", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	commonFlags(fs, &cfg)
	fs.DurationVar(&cfg.Interval, "interval", cfg.Interval, "fallback rescan interval (e.g. 10s)")
	fs.DurationVar(&cfg.Debounce, "debounce", cfg.Debounce, "settle/debounce window before scanning (e.g. 2s)")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "run 'cs2-battlelog help' for usage")
		return 2
	}
	cfg.Inputs = append(cfg.Inputs, fs.Args()...)
	if err := cfg.Validate(true); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	logger := newLogger(cfg)
	st, err := store.Open(cfg.ManifestDir())
	if err != nil {
		logger.Error("open manifest", "error", err)
		return 1
	}
	proc := NewProcessor(cfg, st, logger)
	if err := watch.Run(ctx, cfg, proc, logger); err != nil {
		logger.Error("watcher stopped", "error", err)
		return 1
	}
	return 0
}

func commonFlags(fs *flag.FlagSet, cfg *config.Config) {
	cfg.Inputs = append(cfg.Inputs, splitEnv(os.Getenv("CS2BATTLELOG_INPUT"))...)
	fs.Var((*multiString)(&cfg.Inputs), "input", "input directory or demo file (repeatable)")
	fs.Var((*multiString)(&cfg.Inputs), "i", "alias for --input")
	cfg.Output = envString("CS2BATTLELOG_OUTPUT", cfg.Output)
	fs.StringVar(&cfg.Output, "output", cfg.Output, "output directory (default: next to each demo)")
	fs.StringVar(&cfg.Output, "o", cfg.Output, "alias for --output")
	fs.BoolVar(&cfg.Recursive, "recursive", cfg.Recursive, "scan directories recursively")
	fs.IntVar(&cfg.Workers, "workers", cfg.Workers, "parallel demo workers (memory hungry)")
	fs.BoolVar(&cfg.JSON, "json", envBool("CS2BATTLELOG_JSON", cfg.JSON), "write a JSON sidecar next to the HTML")
	fs.BoolVar(&cfg.Cheat, "cheat", envBool("CS2BATTLELOG_CHEAT", cfg.Cheat), "run the anti-cheat detector (second parse pass)")
	fs.BoolVar(&cfg.Aim, "aim", envBool("CS2BATTLELOG_AIM", cfg.Aim), "collect aim statistics")
	fs.BoolVar(&cfg.Force, "force", cfg.Force, "reprocess demos even if unchanged")
	fs.BoolVar(&cfg.Quiet, "quiet", cfg.Quiet, "only log errors")
	fs.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "debug logging")
	fs.DurationVar(&cfg.Timeout, "timeout", cfg.Timeout, "per-demo processing timeout")
	fs.Usage = func() { printUsage(os.Stderr) }
}

func newLogger(cfg config.Config) *slog.Logger {
	level := slog.LevelInfo
	if cfg.Verbose {
		level = slog.LevelDebug
	}
	if cfg.Quiet {
		level = slog.LevelError
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
}

func discover(inputs []string, recursive bool) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	add := func(path string) {
		if !supportedDemo(path) || seen[path] {
			return
		}
		seen[path] = true
		out = append(out, path)
	}
	for _, in := range inputs {
		info, err := os.Stat(in)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			err := filepath.WalkDir(in, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if d.IsDir() {
					if path != in && (strings.HasPrefix(d.Name(), ".") || d.Name() == ".cs2battlelog") {
						return filepath.SkipDir
					}
					if path != in && !recursive {
						return filepath.SkipDir
					}
					return nil
				}
				add(path)
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			add(in)
		}
	}
	sort.Strings(out)
	return out, nil
}

type multiString []string

func (m *multiString) String() string { return strings.Join(*m, ",") }
func (m *multiString) Set(v string) error {
	if v == "" {
		return nil
	}
	*m = append(*m, v)
	return nil
}

func splitEnv(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, string(os.PathListSeparator))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, strings.TrimSpace(p))
		}
	}
	return out
}

func envString(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, `CS2 Battlelog — analyze CS2 demos into a self-contained battlelog report.

Usage:
  cs2-battlelog analyze [flags] <demo|directory>...
  cs2-battlelog watch   [flags] <directory>...
  cs2-battlelog version
  cs2-battlelog help

Analyze:
  cs2-battlelog analyze --output ./reports stuff/
  cs2-battlelog analyze --json --workers 2 stuff/match.dem

Watch (default: reports written next to each demo):
  cs2-battlelog watch --input /demos --output /reports
  cs2-battlelog watch --input /demos --cheat=false --interval 10s

Flags:
  -i, --input PATH       input directory/file (repeatable; watch accepts positional too)
  -o, --output DIR       output directory (default: next to each demo)
      --recursive        scan directories recursively (default true)
      --workers N        parallel workers (default min(4, CPU))
      --json             also write <demo>.battlelog.json
      --cheat            run anti-cheat detector (default true)
      --aim              collect aim stats (default true)
      --force            reprocess unchanged demos
      --timeout 15m      per-demo timeout
      --quiet / --verbose
  Watch only:
      --interval 5s      fallback rescan interval
      --debounce 2s      settle window for new files

Environment:
  CS2BATTLELOG_INPUT     path-list separated inputs
  CS2BATTLELOG_OUTPUT    output directory
  CS2BATTLELOG_JSON      true/false
  CS2BATTLELOG_CHEAT     true/false
  CS2BATTLELOG_AIM       true/false

Output:
  <demo>.battlelog.html  self-contained dark Battlelog-style report
  <demo>.battlelog.json  optional machine-readable sidecar (--json)
  <output>/.cs2battlelog/manifest.json  watch cache

Exit codes: 0 success, 1 processing failure, 2 usage error.
`)
}
