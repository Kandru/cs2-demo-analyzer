// Package report owns the canonical report document that both the HTML page
// and the optional JSON sidecar are built from.
//
// Every artifact starts with a small, self-describing header so the analyzer
// can decide whether an existing file is still current without parsing the
// whole document or relying on a sidecar manifest.
package report

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kandru/cs2-demo-analyzer/internal/model"
)

// HeaderFormat is bumped when the meaning or encoding of Header changes in a
// way that must invalidate cached artifacts.
const HeaderFormat = 1

// HeaderPrefix starts the metadata comment written into HTML reports.
const HeaderPrefix = "<!--cs2-demo-analyzer:"

// HeaderSuffix closes the metadata comment.
const HeaderSuffix = "-->"

// headerReadLimit bounds how much of an HTML report is inspected for the
// metadata comment. The header is always written before the document body.
const headerReadLimit = 16 * 1024

// Header identifies the artifact and the inputs that produced it. Fields that
// affect rendering are compared by Matches; presentation-only fields (map,
// server, score, timestamps) are informational.
type Header struct {
	Format          int       `json:"format"`
	ToolVersion     string    `json:"toolVersion"`
	SchemaVersion   int       `json:"schemaVersion"`
	Options         string    `json:"options,omitempty"`
	DemoKey         string    `json:"demoKey"`
	DemoSize        int64     `json:"demoSize"`
	DemoModUnixNano int64     `json:"demoModUnixNano"`
	GeneratedAt     time.Time `json:"generatedAt"`

	// Match summary used by report index pages. These do not affect freshness.
	Map        string    `json:"map,omitempty"`
	ServerName string    `json:"serverName,omitempty"`
	Score      [2]int    `json:"score,omitempty"`
	HasScore   bool      `json:"hasScore,omitempty"`
	PlayedAt   time.Time `json:"playedAt,omitempty"`
}

// Matches reports whether an existing artifact header describes the same
// report as want. Presentation-only fields and GeneratedAt are ignored.
func (h Header) Matches(want Header) bool {
	return h.Format == want.Format &&
		h.ToolVersion == want.ToolVersion &&
		h.SchemaVersion == want.SchemaVersion &&
		h.Options == want.Options &&
		h.DemoKey == want.DemoKey &&
		h.DemoSize == want.DemoSize &&
		h.DemoModUnixNano == want.DemoModUnixNano
}

// Options identifies the report-affecting flags. It must be deterministic: it
// is embedded in every artifact header and compared on later runs.
type Options struct {
	Cheat        bool
	Aim          bool
	DemoDownload bool
	Index        bool
}

// String returns the canonical header value for these options.
func (o Options) String() string {
	return fmt.Sprintf("cheat=%t;aim=%t;demo-download=%t;index=%t", o.Cheat, o.Aim, o.DemoDownload, o.Index)
}

// Document is the canonical report. The embedded match keeps the JSON sidecar
// flat and backwards compatible; Header is always the first member.
type Document struct {
	Header Header `json:"header"`
	*model.Match
}

// New builds a document from a match and header.
func New(m *model.Match, h Header) *Document {
	return &Document{Header: h, Match: m}
}

// WriteJSON writes the canonical JSON document with the header first.
func WriteJSON(w io.Writer, d *Document) error {
	if d == nil || d.Match == nil {
		return fmt.Errorf("report: nil document")
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(d)
}

// HeaderComment renders the metadata line written at the top of HTML reports.
// The JSON encoding HTML-escapes '<', '>' and '&', so the result cannot close
// the comment early; the explicit check protects future encoder changes.
func HeaderComment(h Header) (string, error) {
	raw, err := json.Marshal(h)
	if err != nil {
		return "", fmt.Errorf("marshal report header: %w", err)
	}
	s := string(raw)
	if strings.Contains(s, HeaderSuffix) || strings.ContainsAny(s, "\r\n") {
		return "", fmt.Errorf("report header contains unsafe comment characters")
	}
	return HeaderPrefix + s + HeaderSuffix, nil
}

// DemoKey returns a short, stable fingerprint for a demo path. It lets an
// artifact prove which source file produced it without embedding a local path.
func DemoKey(demoPath string) string {
	abs, err := filepath.Abs(demoPath)
	if err != nil {
		abs = demoPath
	}
	sum := sha256.Sum256([]byte(abs))
	return hex.EncodeToString(sum[:8])
}

// ReadHeader reads only the metadata header of an HTML or JSON report.
func ReadHeader(path string) (Header, error) {
	f, err := os.Open(path)
	if err != nil {
		return Header{}, err
	}
	defer f.Close()
	return ReadHeaderFrom(f)
}

// Fresh reports whether path exists and carries a header matching want.
func Fresh(path string, want Header) bool {
	h, err := ReadHeader(path)
	return err == nil && h.Matches(want)
}

// ReadHeaderFrom probes the start of a report stream. HTML reports start with
// a comment; JSON reports start with an object whose first key is "header".
func ReadHeaderFrom(r io.Reader) (Header, error) {
	br := bufio.NewReaderSize(r, headerReadLimit)
	peek, err := br.Peek(1024)
	if err != nil && len(peek) == 0 {
		return Header{}, fmt.Errorf("read report header: %w", err)
	}
	trimmed := bytes.TrimLeft(bytes.TrimPrefix(peek, []byte{0xEF, 0xBB, 0xBF}), " \t\r\n")
	if len(trimmed) == 0 {
		return Header{}, fmt.Errorf("read report header: empty file")
	}
	switch trimmed[0] {
	case '<':
		return readHTMLHeader(br)
	case '{':
		return readJSONHeader(br)
	default:
		return Header{}, fmt.Errorf("read report header: unrecognized artifact")
	}
}

func readHTMLHeader(r io.Reader) (Header, error) {
	data, err := io.ReadAll(io.LimitReader(r, headerReadLimit))
	if err != nil {
		return Header{}, fmt.Errorf("read html header: %w", err)
	}
	start := bytes.Index(data, []byte(HeaderPrefix))
	if start < 0 {
		return Header{}, fmt.Errorf("read html header: metadata comment missing")
	}
	rest := data[start+len(HeaderPrefix):]
	end := bytes.Index(rest, []byte(HeaderSuffix))
	if end < 0 {
		return Header{}, fmt.Errorf("read html header: metadata comment unterminated")
	}
	var h Header
	if err := json.Unmarshal(rest[:end], &h); err != nil {
		return Header{}, fmt.Errorf("read html header: %w", err)
	}
	return h, nil
}

func readJSONHeader(r io.Reader) (Header, error) {
	dec := json.NewDecoder(r)
	tok, err := dec.Token()
	if err != nil {
		return Header{}, fmt.Errorf("read json header: %w", err)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return Header{}, fmt.Errorf("read json header: expected object")
	}
	tok, err = dec.Token()
	if err != nil {
		return Header{}, fmt.Errorf("read json header: %w", err)
	}
	key, ok := tok.(string)
	if !ok || key != "header" {
		return Header{}, fmt.Errorf("read json header: header must be the first key")
	}
	var h Header
	if err := dec.Decode(&h); err != nil {
		return Header{}, fmt.Errorf("read json header: %w", err)
	}
	return h, nil
}
