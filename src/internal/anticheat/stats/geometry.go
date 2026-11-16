package stats

import (
	"math"
	"sort"
)

// geometry.go holds the shared vector/angle math used by the behavioral
// collector and the aim tracker equivalents. Keeping it in one place avoids
// two slightly different implementations drifting apart.

// viewDirectionToVector converts CS2 yaw/pitch (deg) to a unit direction
// vector.
//
// CS2's ViewDirectionY is documented as 270..90 (with 270 == -90). Treating
// pitch as a signed angle in [-90, 90] requires normalizing the 270 case.
func viewDirectionToVector(yawDeg, pitchDeg float64) [3]float64 {
	if pitchDeg > 180.0 {
		pitchDeg -= 360.0
	}
	yaw := yawDeg * math.Pi / 180.0
	pitch := pitchDeg * math.Pi / 180.0
	cp := math.Cos(pitch)
	return [3]float64{
		math.Cos(yaw) * cp,
		math.Sin(yaw) * cp,
		math.Sin(pitch),
	}
}

// angleBetweenViewAndTarget returns the angle (deg) between a unit view vector
// and the vector from view origin to the target position.
func angleBetweenViewAndTarget(view [3]float64, ox, oy, oz, tx, ty, tz float64) float64 {
	dx := tx - ox
	dy := ty - oy
	dz := tz - oz
	mag := math.Sqrt(dx*dx + dy*dy + dz*dz)
	if mag == 0 {
		return 0
	}
	dot := (view[0]*dx + view[1]*dy + view[2]*dz) / mag
	if dot > 1 {
		dot = 1
	} else if dot < -1 {
		dot = -1
	}
	return math.Acos(dot) * 180.0 / math.Pi
}

// median returns the median of a non-empty slice without mutating the input.
func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	cp := make([]float64, len(xs))
	copy(cp, xs)
	sort.Float64s(cp)
	n := len(cp)
	if n%2 == 1 {
		return cp[n/2]
	}
	return (cp[n/2-1] + cp[n/2]) / 2.0
}
