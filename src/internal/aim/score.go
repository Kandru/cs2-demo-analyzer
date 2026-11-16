package aim

// score.go converts an anchored value into a 0–100 score and the composite
// rating into a letter grade.

func anchorScore(v, lo, hi float64, higherBetter bool) float64 {
	var s float64
	if higherBetter {
		s = (v - lo) / (hi - lo) * 100
	} else {
		s = (lo - v) / (lo - hi) * 100
	}
	if s < 0 {
		return 0
	}
	if s > 100 {
		return 100
	}
	return s
}

func grade(r float64) string {
	switch {
	case r >= 85:
		return "A+"
	case r >= 75:
		return "A"
	case r >= 65:
		return "B+"
	case r >= 55:
		return "B"
	case r >= 45:
		return "C+"
	case r >= 35:
		return "C"
	case r >= 25:
		return "D"
	default:
		return "F"
	}
}
