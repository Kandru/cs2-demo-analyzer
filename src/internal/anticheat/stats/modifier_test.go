package stats

import "testing"

func TestWingmanBoost(t *testing.T) {
	ps := newTestPlayerStats()
	ps.AddMetric(gameInfoCategory, Key("game_mode"), Metric{Type: MetricString, StringValue: "Wingman"})
	metricInt(t, ps, killsCategory, Key("total_kills"), 12)
	st := &scoreState{score: 10, ps: ps}
	if a := applyWingmanBoost(st); !a.Fired || st.score != 18 || a.Detail != "kills=12" {
		t.Fatalf("unexpected wingman result: score=%v applied=%+v", st.score, a)
	}
	ps.AddMetric(gameInfoCategory, Key("game_mode"), Metric{Type: MetricString, StringValue: "Competitive"})
	st = &scoreState{score: 10, ps: ps}
	if a := applyWingmanBoost(st); a.Fired || st.score != 10 {
		t.Fatalf("non-wingman should not boost: %+v", a)
	}
}

func TestCompetitiveBoost(t *testing.T) {
	ps := newTestPlayerStats()
	ps.AddMetric(gameInfoCategory, Key("game_mode"), Metric{Type: MetricString, StringValue: "Competitive"})
	metricInt(t, ps, killsCategory, Key("total_kills"), 45)
	metricInt(t, ps, gameInfoCategory, Key("round_count"), 25)
	st := &scoreState{score: 10, ps: ps}
	if a := applyCompetitiveBoost(st); !a.Fired || st.score != 12 {
		t.Fatalf("unexpected competitive result: score=%v applied=%+v", st.score, a)
	}
	metricInt(t, ps, gameInfoCategory, Key("round_count"), 31)
	st = &scoreState{score: 10, ps: ps}
	if a := applyCompetitiveBoost(st); a.Fired {
		t.Fatal("overtime round count should not boost")
	}
}

func TestPositionDiscount(t *testing.T) {
	ps := newTestPlayerStats()
	metricFloat(t, ps, scoreboardCategory, Key("position_factor"), 0.5)
	st := &scoreState{score: 10, ps: ps}
	a := applyPositionDiscount(st)
	if !a.Fired || a.Value != 0.1 || st.score != 9 {
		t.Fatalf("unexpected position result: score=%v applied=%+v", st.score, a)
	}
}

func TestEvidenceStacking(t *testing.T) {
	ps := newTestPlayerStats()
	st := &scoreState{
		score: 10,
		ps:    ps,
		channels: []Channel{
			{HasData: true, Score: 0.6, Confidence: 0.6},
			{HasData: true, Score: 0.7, Confidence: 0.7},
			{HasData: true, Score: 0.8, Confidence: 0.8},
			{HasData: true, Score: 0.1, Confidence: 0.1},
		},
	}
	a := applyEvidenceStacking(st)
	if !a.Fired || a.Count != 3 || st.score != 14 {
		t.Fatalf("unexpected stacking result: score=%v applied=%+v", st.score, a)
	}
}

func TestWallhackCooccurrence(t *testing.T) {
	ps := newTestPlayerStats()
	metricFloat(t, ps, behavioralCategory, Key("back_kill_given_pct"), 10)
	metricInt(t, ps, behavioralCategory, Key("back_kill_given_total_kills"), 5)
	st := &scoreState{
		score:    10,
		ps:       ps,
		channels: []Channel{{ID: detPreFOV.ID, HasData: true, Score: 0.8, Confidence: 0.8}},
	}
	if a := applyWallhackCoOccurrenceBoost(st); !a.Fired || st.score != 12 {
		t.Fatalf("unexpected cooccurrence result: score=%v applied=%+v", st.score, a)
	}
	// A strong pre-FOV reading without back-kills must not fire.
	delete(ps.Categories[behavioralCategory], Key("back_kill_given_pct"))
	st = &scoreState{score: 10, ps: ps, channels: []Channel{{ID: detPreFOV.ID, HasData: true, Score: 0.8, Confidence: 0.8}}}
	if a := applyWallhackCoOccurrenceBoost(st); a.Fired {
		t.Fatal("pre-FOV alone should not fire")
	}
}

func TestSniperOverrides(t *testing.T) {
	ps := newTestPlayerStats()
	metricInt(t, ps, sniperCategory, Key("sniper_wallbang_kills"), 11)
	st := &scoreState{score: 42, ps: ps}
	a := applySniperOverrides(st)
	if !a.Fired || st.score != 100 || len(a.Names) != 1 || a.Names[0] != "sniper_wallbang_override" {
		t.Fatalf("unexpected wallbang override: score=%v applied=%+v", st.score, a)
	}

	ps2 := newTestPlayerStats()
	metricInt(t, ps2, sniperCategory, Key("scout_kills"), 11)
	metricFloat(t, ps2, sniperCategory, Key("scout_hs_rate"), 85)
	st2 := &scoreState{score: 42, ps: ps2}
	a2 := applySniperOverrides(st2)
	if !a2.Fired || st2.score != 100 || len(a2.Names) != 1 || a2.Names[0] != "scout_precision_override" {
		t.Fatalf("unexpected scout override: score=%v applied=%+v", st2.score, a2)
	}
}
