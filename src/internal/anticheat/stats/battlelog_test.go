package stats

import "testing"

// TestBattlelogChannelsMirrorRegistry guards the report metadata path: a
// channel row must carry the detection's label, unit, description and weight.
func TestBattlelogChannelsMirrorRegistry(t *testing.T) {
	ps := newTestPlayerStats()
	rows := battlelogChannels(ps)
	if len(rows) != len(detections) {
		t.Fatalf("got %d channel rows, want %d", len(rows), len(detections))
	}
	for i, det := range detections {
		row := rows[i]
		if row.ID != det.ID || row.Label != det.Label || row.Unit != det.Unit ||
			row.Description != det.Description || row.Weight != det.Weight {
			t.Fatalf("row %d does not mirror detection %q: %+v", i, det.ID, row)
		}
		if row.Weight <= 0 {
			t.Fatalf("channel %q has non-positive weight", det.ID)
		}
		if row.HasData || row.Zone != "no_data" {
			t.Fatalf("empty stats should produce no-data rows: %+v", row)
		}
	}
}
