package aim

import (
	"math"
	"testing"
)

func TestAnchorScore(t *testing.T) {
	cases := []struct {
		name string
		v    float64
		lo   float64
		hi   float64
		high bool
		want float64
	}{
		{"higher midpoint", 20, 10, 30, true, 50},
		{"higher below", 5, 10, 30, true, 0},
		{"higher above", 40, 10, 30, true, 100},
		{"lower midpoint", 400, 600, 200, false, 50},
		{"lower below", 100, 600, 200, false, 100},
		{"lower above", 700, 600, 200, false, 0},
	}
	for _, c := range cases {
		if got := anchorScore(c.v, c.lo, c.hi, c.high); math.Abs(got-c.want) > 0.001 {
			t.Errorf("%s: got %.2f want %.2f", c.name, got, c.want)
		}
	}
}

func TestBuildAimStats(t *testing.T) {
	st := &playerState{
		shots:        100,
		hits:         25,
		headHits:     10,
		ttd:          []float64{300, 350, 400},
		reaction:     []float64{250, 300, 350},
		preaim:       []float64{6, 8, 10, 12},
		counterShots: 100,
		counterGood:  60,
		sprayAtt:     40,
		sprayHits:    20,
	}
	a := buildAimStats(st, true)
	if a == nil {
		t.Fatal("nil aim stats")
	}
	if a.Rating <= 0 || a.Rating > 100 {
		t.Fatalf("rating out of range: %.2f", a.Rating)
	}
	if a.Grade == "" {
		t.Fatal("missing grade")
	}
	want := map[string]string{
		"damage_speed":   "350 ms",
		"reaction":       "300 ms",
		"preaim":         "9.0°",
		"accuracy":       "25.0%",
		"head_accuracy":  "40.0%",
		"counter_strafe": "60.0%",
		"spray":          "50.0%",
	}
	for _, m := range a.Metrics {
		if m.Display != want[m.ID] {
			t.Errorf("%s: got %q want %q", m.ID, m.Display, want[m.ID])
		}
	}
}

func TestNoHurtDisablesAccuracy(t *testing.T) {
	st := &playerState{shots: 100, hits: 0, counterShots: 100, counterGood: 50}
	a := buildAimStats(st, false)
	for _, m := range a.Metrics {
		if m.ID == "accuracy" && m.Display != "—" {
			t.Fatalf("accuracy should be unavailable without damage events, got %q", m.Display)
		}
	}
	if a.Rating != 0 || a.Grade != "" {
		t.Fatalf("rating should be empty when fewer than 3 metrics have data, got %.1f %q", a.Rating, a.Grade)
	}
}

func TestMedian(t *testing.T) {
	if got := median(nil); got != 0 {
		t.Fatalf("nil median: %v", got)
	}
	if got := median([]float64{3, 1, 2}); got != 2 {
		t.Fatalf("odd median: %v", got)
	}
	if got := median([]float64{4, 1, 3, 2}); got != 2.5 {
		t.Fatalf("even median: %v", got)
	}
}
