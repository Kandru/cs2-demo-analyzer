// Package store persists the watch manifest and writes reports atomically.
package store

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Kandru/cs2-demo-analyzer/internal/demofile"
)

// ManifestVersion is bumped when the cache file format changes.
const ManifestVersion = 1

// Entry is the cached analysis state for one demo.
type Entry struct {
	Path          string     `json:"path"`
	Size          int64      `json:"size"`
	ModTime       int64      `json:"mtimeUnixNano"`
	Hash          string     `json:"sha256,omitempty"`
	HTMLPath      string     `json:"htmlPath,omitempty"`
	JSONPath      string     `json:"jsonPath,omitempty"`
	AnalyzedAt    time.Time  `json:"analyzedAt,omitempty"`
	ToolVersion   string     `json:"toolVersion,omitempty"`
	ReportOptions string     `json:"reportOptions,omitempty"`
	SchemaVersion int        `json:"schemaVersion,omitempty"`
	Status        string     `json:"status,omitempty"` // ok | error
	Error         string     `json:"error,omitempty"`
	Meta          ReportMeta `json:"meta,omitzero"`
}

// ReportMeta is match-level metadata cached for report index pages.
type ReportMeta struct {
	Map        string    `json:"map,omitempty"`
	ServerName string    `json:"serverName,omitempty"`
	Score      [2]int    `json:"score,omitempty"`
	HasScore   bool      `json:"hasScore,omitempty"`
	PlayedAt   time.Time `json:"playedAt,omitempty"`
}

type manifestFile struct {
	Version int               `json:"version"`
	Entries map[string]*Entry `json:"entries"`
}

// Store is a concurrency-safe manifest bound to one output directory.
type Store struct {
	mu      sync.Mutex
	path    string
	entries map[string]*Entry
	// outputs maps report path -> demo path to detect basename collisions.
	outputs map[string]string
}

// NewMemory returns a store that tracks output names in memory only. Analyze
// uses it so one-shot runs do not create a manifest next to the demos.
func NewMemory() *Store {
	return &Store{
		entries: map[string]*Entry{},
		outputs: map[string]string{},
	}
}

// Open loads (or creates) the manifest under outputDir/.cs2-demo-analyzer/.
func Open(outputDir string) (*Store, error) {
	if outputDir == "" {
		return nil, fmt.Errorf("store: empty output directory")
	}
	dir := filepath.Join(outputDir, ".cs2-demo-analyzer")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create manifest dir: %w", err)
	}
	s := &Store{
		path:    filepath.Join(dir, "manifest.json"),
		entries: map[string]*Entry{},
		outputs: map[string]string{},
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var mf manifestFile
	if err := json.Unmarshal(data, &mf); err != nil {
		// A corrupt manifest must not block analysis; start fresh.
		return s, nil
	}
	if mf.Entries != nil {
		s.entries = mf.Entries
	}
	for p, e := range s.entries {
		if e != nil && e.HTMLPath != "" {
			s.outputs[e.HTMLPath] = p
		}
	}
	return s, nil
}

// CachedFailure reports whether the same file is recorded as failed and should
// not be retried until it changes. Callers handle --force themselves.
func (s *Store) CachedFailure(path string, size int64, modTime time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.entries[path]
	if e == nil {
		return false
	}
	if e.Size != size || e.ModTime != modTime.UnixNano() {
		return false
	}
	return e.Status == "error"
}

// schemaVersion mirrors model.SchemaVersion without importing model (avoids a
// store -> model dependency in callers). Keep in sync; it is recorded in the
// manifest for audit purposes.
const schemaVersion = 1

// Entry returns a copy of the cached entry, if any.
func (s *Store) Entry(path string) (Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.entries[path]
	if e == nil {
		return Entry{}, false
	}
	return *e, true
}

// MarkOK records a successful analysis.
func (s *Store) MarkOK(path string, size int64, modTime time.Time, hash, reportOptions, html, jsonPath, toolVersion string, meta ReportMeta) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[path] = &Entry{
		Path:          path,
		Size:          size,
		ModTime:       modTime.UnixNano(),
		Hash:          hash,
		HTMLPath:      html,
		JSONPath:      jsonPath,
		AnalyzedAt:    time.Now().UTC(),
		ToolVersion:   toolVersion,
		ReportOptions: reportOptions,
		SchemaVersion: schemaVersion,
		Status:        "ok",
		Meta:          meta,
	}
	s.outputs[html] = path
}

// MarkError records a failed analysis. The error is cached so a broken demo is
// not retried on every scan until it changes or --force is used.
func (s *Store) MarkError(path string, size int64, modTime time.Time, err error, toolVersion string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	s.entries[path] = &Entry{
		Path:          path,
		Size:          size,
		ModTime:       modTime.UnixNano(),
		AnalyzedAt:    time.Now().UTC(),
		ToolVersion:   toolVersion,
		SchemaVersion: schemaVersion,
		Status:        "error",
		Error:         msg,
	}
}

// PlanOutput atomically chooses and reserves the report path for a demo. If
// outputDir is empty the report is written next to the demo. Basename
// collisions from different demos are disambiguated with a short source hash.
//
// occupied is optional and is called while the store lock is held. It lets the
// caller mark a candidate as belonging to another demo when the in-memory
// mapping is not available (for example a one-shot analyze run). Keep the
// callback cheap: it typically reads a few hundred bytes of a header.
func (s *Store) PlanOutput(demoPath, outputDir, suffix string, occupied func(string) bool) string {
	dir := outputDir
	if dir == "" {
		dir = filepath.Dir(demoPath)
	}
	base := safeName(demofile.BaseName(demoPath))
	cand := filepath.Join(dir, base+suffix)

	s.mu.Lock()
	defer s.mu.Unlock()
	if src, ok := s.outputs[cand]; ok && src == demoPath {
		return cand
	}
	taken := func(path string) bool {
		if src, ok := s.outputs[path]; ok && src != demoPath {
			return true
		}
		return occupied != nil && occupied(path)
	}
	// Never let a demo report collide with the generated directory index.
	if strings.EqualFold(filepath.Base(cand), "index.html") || taken(cand) {
		h := sha256.Sum256([]byte(demoPath))
		cand = filepath.Join(dir, fmt.Sprintf("%s-%x%s", base, h[:3], suffix))
	}
	s.outputs[cand] = demoPath
	return cand
}

// DemoForReport returns the demo path that produced reportPath, if known.
func (s *Store) DemoForReport(reportPath string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	demo, ok := s.outputs[reportPath]
	return demo, ok && demo != ""
}

// Entries returns a copy of every cached manifest entry.
func (s *Store) Entries() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		if e != nil {
			out = append(out, *e)
		}
	}
	return out
}

// Save writes the manifest atomically.
func (s *Store) Save() error {
	if s.path == "" {
		return nil
	}
	s.mu.Lock()
	data, err := json.MarshalIndent(manifestFile{Version: ManifestVersion, Entries: s.entries}, "", "  ")
	s.mu.Unlock()
	if err != nil {
		return err
	}
	return WriteAtomic(s.path, func(w io.Writer) error {
		_, err := w.Write(data)
		return err
	}, 0o644)
}

// WriteAtomic writes a file via a temp file in the same directory and renames
// it into place, so readers never observe partial reports.
func WriteAtomic(path string, write func(io.Writer) error, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := write(tmp); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if perm != 0 {
		_ = os.Chmod(tmpName, perm)
	}
	return os.Rename(tmpName, path)
}

func safeName(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('_')
		}
	}
	out := b.String()
	out = strings.Trim(out, "._-")
	if out == "" {
		out = "demo"
	}
	if len(out) > 120 {
		out = out[:120]
	}
	return out
}
