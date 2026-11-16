package aim

import (
	"math"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
)

// geometry.go holds the view/position math shared by the tracker and the
// preaim metric.

// preaimAngle measures the attacker's crosshair-to-victim angle 200 ms before
// the tick at which the engagement produced damage.
func (t *Tracker) preaimAngle(attackerID, victimID uint64, tick int) (float64, bool) {
	lookback := int(200*t.tickRate/1000.0) + 1
	target := tick - lookback
	ah, vh := t.history[attackerID], t.history[victimID]
	if len(ah) == 0 || len(vh) == 0 {
		return 0, false
	}
	as := nearestAtOrBefore(ah, target)
	vs := nearestAtOrBefore(vh, target)
	if as == nil || vs == nil {
		return 0, false
	}
	view := viewDirectionToVector(as.yaw, as.pitch)
	dx, dy, dz := vs.x-as.x, vs.y-as.y, vs.z-as.z
	mag := math.Sqrt(dx*dx + dy*dy + dz*dz)
	if mag == 0 {
		return 0, false
	}
	dot := (view[0]*dx + view[1]*dy + view[2]*dz) / mag
	if dot > 1 {
		dot = 1
	} else if dot < -1 {
		dot = -1
	}
	ang := math.Acos(dot) * 180 / math.Pi
	if ang <= 0 || ang > 90 {
		return 0, false
	}
	return ang, true
}

func nearestAtOrBefore(hist []posSample, tick int) *posSample {
	for i := len(hist) - 1; i >= 0; i-- {
		if hist[i].tick <= tick {
			return &hist[i]
		}
	}
	return nil
}

func viewToTargetAngle(attacker, victim *common.Player) float64 {
	ap, ok := attacker.PositionEyes()
	if !ok {
		return 0
	}
	vp, ok := victim.PositionEyes()
	if !ok {
		return 0
	}
	view := viewDirectionToVector(float64(attacker.ViewDirectionX()), float64(attacker.ViewDirectionY()))
	dx, dy, dz := vp.X-ap.X, vp.Y-ap.Y, vp.Z-ap.Z
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
	return math.Acos(dot) * 180 / math.Pi
}

func viewDirectionToVector(yawDeg, pitchDeg float64) [3]float64 {
	if pitchDeg > 180 {
		pitchDeg -= 360
	}
	yaw := yawDeg * math.Pi / 180
	pitch := pitchDeg * math.Pi / 180
	cp := math.Cos(pitch)
	return [3]float64{math.Cos(yaw) * cp, math.Sin(yaw) * cp, math.Sin(pitch)}
}
