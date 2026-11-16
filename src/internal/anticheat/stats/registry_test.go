package stats

import "testing"

// TestDetectionRegistryContract pins the public channel surface. Changing a
// weight, mode, or ID is a calibration decision and must update this test plus
// the matching docs/detections page.
func TestDetectionRegistryContract(t *testing.T) {
	want := []struct {
		id     string
		weight float64
		mode   Mode
	}{
		{"hs", 0.18, ModePositiveOnly},
		{"snap", 0.10, ModePositiveOnly},
		{"reaction", 0.10, ModeBidirectional},
		{"ttd_sub100", 0.10, ModePositiveOnly},
		{"recoil", 0.10, ModePositiveOnly},
		{"pre_fov", 0.22, ModeBidirectional},
		{"pre_fov_presence", 0.10, ModePositiveOnly},
		{"attention", 0.06, ModePositiveOnly},
		{"back_killed", 0.06, ModePositiveOnly},
		{"decoupling", 0.10, ModePositiveOnly},
	}
	if len(detections) != len(want) {
		t.Fatalf("got %d detections, want %d", len(detections), len(want))
	}
	seen := map[string]bool{}
	for i, d := range detections {
		w := want[i]
		if d.ID != w.id || d.Weight != w.weight || d.Mode != w.mode {
			t.Fatalf("detection %d = %s/%v/%v, want %s/%v/%v",
				i, d.ID, d.Weight, d.Mode, w.id, w.weight, w.mode)
		}
		if seen[d.ID] {
			t.Fatalf("duplicate detection ID %q", d.ID)
		}
		seen[d.ID] = true
		if d.Label == "" || d.Unit == "" || d.Description == "" || d.Doc == "" {
			t.Fatalf("detection %q is missing report metadata", d.ID)
		}
		if d.Evaluate == nil && d.LobbyEvaluate == nil {
			t.Fatalf("detection %q has no evaluator", d.ID)
		}
		if d.Raw == nil {
			t.Fatalf("detection %q has no raw extractor", d.ID)
		}
	}
}

func TestModifierRegistryContract(t *testing.T) {
	want := []string{
		"wingman_boost",
		"competitive_boost",
		"position_discount",
		"evidence_stacking_boost",
		"wallhack_co_occurrence_boost",
		"ttd_sub100_high_floor",
		"sniper_overrides",
	}
	if len(modifiers) != len(want) {
		t.Fatalf("got %d modifiers, want %d", len(modifiers), len(want))
	}
	for i, m := range modifiers {
		if m.ID != want[i] {
			t.Fatalf("modifier %d = %q, want %q", i, m.ID, want[i])
		}
		if m.Doc == "" || m.Apply == nil || m.Publish == nil || m.Adjustments == nil {
			t.Fatalf("modifier %q is incomplete", m.ID)
		}
	}
}

func TestDetectionsHandleEmptyStats(t *testing.T) {
	ps := newTestPlayerStats()
	for _, d := range detections {
		if d.Evaluate != nil {
			ch := d.Evaluate(ps)
			if ch.HasData {
				t.Fatalf("%s reported data for empty stats", d.ID)
			}
		}
	}
}

func TestEvidenceKindsAreRegistered(t *testing.T) {
	for _, kind := range []string{"snap", "ttd", "ttd_sub100", "preaim", "recoil", "wallbang", "smoke", "noscope"} {
		spec := evidenceSpec(kind)
		if spec == nil {
			t.Fatalf("missing evidence spec for %q", kind)
		}
		if spec.Label == "" || spec.Rank < 1 || spec.Severity == nil || spec.ValueText == nil {
			t.Fatalf("incomplete evidence spec for %q", kind)
		}
	}
	if got := evidenceSpec("unknown-kind"); got != nil {
		t.Fatalf("unknown kind returned %+v", got)
	}
}
