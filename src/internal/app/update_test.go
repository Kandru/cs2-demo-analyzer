package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Kandru/cs2-demo-analyzer/internal/update"
)

func TestPrintUpdateStatus(t *testing.T) {
	tests := []struct {
		name string
		st   update.Status
		want string
	}{
		{
			name: "newer available",
			st:   update.Status{Current: "26.11.1", Latest: "26.11.2", Newer: true},
			want: "update available: 26.11.2",
		},
		{
			name: "up to date",
			st:   update.Status{Current: "26.11.2", Latest: "26.11.2"},
			want: "already up to date",
		},
		{
			name: "local build newer",
			st:   update.Status{Current: "26.11.5", Latest: "26.11.2"},
			want: "current build is newer than the latest release (26.11.2)",
		},
		{
			name: "development build",
			st:   update.Status{Current: "dev", Latest: "26.11.2"},
			want: "use --force to install 26.11.2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printUpdateStatus(&buf, tt.st)
			if got := buf.String(); !strings.Contains(got, tt.want) {
				t.Fatalf("output = %q, want substring %q", got, tt.want)
			}
		})
	}
}
