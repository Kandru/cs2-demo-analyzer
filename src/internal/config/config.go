// Package config contains the runtime configuration shared by analyze/watch.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Config is the validated runtime configuration.
type Config struct {
	Inputs    []string
	Output    string
	Recursive bool
	Workers   int
	JSON      bool
	Cheat     bool
	Aim       bool
	Index     bool
	DemoLinks bool
	Force     bool
	Update    bool
	Quiet     bool
	Verbose   bool
	Timeout   time.Duration

	// Watch-only settings.
	Interval time.Duration
	Debounce time.Duration
}

// Default returns the default configuration.
func Default() Config {
	workers := runtime.NumCPU()
	if workers > 4 {
		workers = 4
	}
	if workers < 1 {
		workers = 1
	}
	return Config{
		Recursive: true,
		Workers:   workers,
		JSON:      false,
		Cheat:     true,
		Aim:       true,
		Timeout:   15 * time.Minute,
		Interval:  5 * time.Second,
		Debounce:  2 * time.Second,
	}
}

// Validate checks the configuration and normalizes paths.
func (c *Config) Validate(requireInputs bool) error {
	if requireInputs && len(c.Inputs) == 0 {
		return fmt.Errorf("at least one input directory or demo file is required")
	}
	if c.Workers < 1 {
		return fmt.Errorf("workers must be >= 1")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be > 0")
	}
	if c.Interval <= 0 {
		return fmt.Errorf("interval must be > 0")
	}
	for i, in := range c.Inputs {
		abs, err := filepath.Abs(in)
		if err != nil {
			return fmt.Errorf("input %q: %w", in, err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return fmt.Errorf("input %q: %w", in, err)
		}
		if !info.IsDir() && info.Mode().IsRegular() {
			// Single file is allowed for analyze.
		} else if !info.IsDir() {
			return fmt.Errorf("input %q is not a file or directory", in)
		}
		c.Inputs[i] = abs
	}
	if c.Output != "" {
		abs, err := filepath.Abs(c.Output)
		if err != nil {
			return fmt.Errorf("output: %w", err)
		}
		if err := os.MkdirAll(abs, 0o755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
		c.Output = abs
	}
	return nil
}

// ManifestDir returns where the watch manifest should live.
func (c *Config) ManifestDir() string {
	if c.Output != "" {
		return c.Output
	}
	if len(c.Inputs) > 0 {
		if info, err := os.Stat(c.Inputs[0]); err == nil && info.IsDir() {
			return c.Inputs[0]
		}
	}
	wd, _ := os.Getwd()
	return wd
}
