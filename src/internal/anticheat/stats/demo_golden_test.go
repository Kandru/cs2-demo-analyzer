package stats_test

// demo_golden_test.go is an opt-in regression harness for the detector's
// published metrics. It is skipped unless CS2BATTLELOG_GOLDEN_DEMO points at a
// real demo, so the default test run needs no demo files and no network.
//
// Usage:
//
//	CS2BATTLELOG_GOLDEN_DEMO=stuff/match.dem \
//	CS2BATTLELOG_GOLDEN_OUT=/tmp/before.txt go test ./internal/anticheat/stats -run TestGoldenDemoMetrics
//
// Re-run with a different output file after a refactor and diff the two files.
// Only metric category, key, type and value are written; player names are
// deliberately excluded so the dump is safe to share.

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/Kandru/cs2-demo-analyzer/internal/anticheat/analyzer"
	"github.com/Kandru/cs2-demo-analyzer/internal/anticheat/stats"
)

func TestGoldenDemoMetrics(t *testing.T) {
	demoPath := os.Getenv("CS2BATTLELOG_GOLDEN_DEMO")
	if demoPath == "" {
		t.Skip("set CS2BATTLELOG_GOLDEN_DEMO to a demo path")
	}
	outPath := os.Getenv("CS2BATTLELOG_GOLDEN_OUT")
	if outPath == "" {
		t.Fatal("set CS2BATTLELOG_GOLDEN_OUT to the dump path")
	}

	f, err := os.Open(demoPath)
	if err != nil {
		t.Fatalf("open demo: %v", err)
	}
	defer f.Close()

	demoStats, err := analyzer.NewAnalyzer(demoPath).AnalyzeReader(f)
	if err != nil {
		t.Fatalf("analyze demo: %v", err)
	}

	var out []string
	sids := make([]uint64, 0, len(demoStats.Players))
	for sid := range demoStats.Players {
		sids = append(sids, sid)
	}
	sort.Slice(sids, func(i, j int) bool { return sids[i] < sids[j] })

	for _, sid := range sids {
		ps := demoStats.Players[sid]
		cats := make([]string, 0, len(ps.Categories))
		for cat := range ps.Categories {
			cats = append(cats, string(cat))
		}
		sort.Strings(cats)
		for _, cat := range cats {
			keys := make([]string, 0, len(ps.Categories[stats.Category(cat)]))
			for key := range ps.Categories[stats.Category(cat)] {
				keys = append(keys, string(key))
			}
			sort.Strings(keys)
			for _, key := range keys {
				m := ps.Categories[stats.Category(cat)][stats.Key(key)]
				out = append(out, fmt.Sprintf("%d|%s|%s|%s|%v|%v|%v|%v|%s",
					sid, cat, key, m.Type, m.FloatValue, m.IntValue,
					m.DurationValue, m.StringValue, m.Description))
			}
		}
		for _, moment := range ps.Moments {
			out = append(out, fmt.Sprintf("%d|moments|%s|%d|%v|%s|%s|%s",
				sid, moment.Kind, moment.Tick, moment.Value, moment.Unit,
				moment.Opponent, moment.Weapon))
		}
	}

	if err := os.WriteFile(outPath, []byte(strings.Join(out, "\n")+"\n"), 0o644); err != nil {
		t.Fatalf("write dump: %v", err)
	}
	t.Logf("wrote %d metric lines to %s", len(out), outPath)
}
