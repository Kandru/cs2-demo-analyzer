// Package jsondata renders the machine-readable match sidecar.
package jsondata

import (
	"encoding/json"
	"io"

	"github.com/Kandru/cs2-demo-analyzer/internal/model"
)

// Write encodes the match as indented JSON.
func Write(w io.Writer, m *model.Match) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(m)
}
