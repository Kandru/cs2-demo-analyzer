package stats

const (
	decouplingID     = "decoupling"
	decouplingWeight = 0.10
	decouplingMode   = ModePositiveOnly
)

// evaluateDecoupling scores the fight-vs-idle decoupling:
// attention_median − pre_fov_median. Wallhackers concentrate during
// engagements but their crosshair drifts during chill moments; legit players
// are consistent. Ramp 8°→22°, positive-only. Silent if either half is
// missing.
func evaluateDecoupling(ps *PlayerStats) Channel {
	preFOVN, hasFOVN := psGetInt(ps, channelCategoryBehavioral, Key("pre_fov_aim_samples"))
	attN, hasAttN := psGetInt(ps, channelCategoryBehavioral, Key("nearest_enemy_angle_samples"))
	if !hasFOVN || preFOVN <= 0 || !hasAttN || attN <= 0 {
		return Channel{ID: decouplingID, Weight: decouplingWeight, Mode: decouplingMode}
	}
	preFOVMed, _ := psGetFloat(ps, channelCategoryBehavioral, Key("pre_fov_aim_median_deg"))
	attMed, _ := psGetFloat(ps, channelCategoryBehavioral, Key("nearest_enemy_angle_median_deg"))
	delta := attMed - preFOVMed
	score := linearScore(delta, 8.0, 22.0)

	// Confidence requires both halves trustworthy.
	confFOV := sqrtConfidence(preFOVN, 15)
	confAtt := linearConfidence(attN, 200)
	conf := confFOV
	if confAtt < conf {
		conf = confAtt
	}
	return Channel{
		ID:         decouplingID,
		Score:      score,
		Confidence: conf,
		Raw:        delta,
		SampleN:    preFOVN, // the limiting half gates confidence
		Weight:     decouplingWeight,
		Zone:       zoneFor(score),
		Mode:       decouplingMode,
		HasData:    true,
	}
}

func rawDecoupling(ps *PlayerStats) (float64, int, bool) {
	preFOV, preFOVOK := psGetFloat(ps, channelCategoryBehavioral, Key("pre_fov_aim_median_deg"))
	att, attOK := psGetFloat(ps, channelCategoryBehavioral, Key("nearest_enemy_angle_median_deg"))
	n, _ := psGetInt(ps, channelCategoryBehavioral, Key("pre_fov_aim_samples"))
	return att - preFOV, int(n), preFOVOK && attOK
}

var detDecoupling = &Detection{
	ID:          "decoupling",
	Label:       "Fight/idle decoupling",
	Unit:        "°",
	Description: "Attention angle minus pre-FOV angle.",
	Doc:         "docs/detections/decoupling.md",
	Weight:      0.10,
	Mode:        ModePositiveOnly,
	Evaluate:    evaluateDecoupling,
	Raw:         rawDecoupling,
	Narrative: &NarrativeSpec{
		ScoreBased: true,
		Ascending:  true,
		Mild:       0.30,
		Strong:     0.45,
		Blatant:    0.60,
		Sentence: func(raw float64, n int64, adj string) string {
			return "Fight-vs-idle attention split is " + adj + " decoupled — focused during engagements, drifting toward unseen enemies between them."
		},
	},
}
