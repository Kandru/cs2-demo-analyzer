package aim

import (
	"sort"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
)

// helpers.go holds pure numeric and weapon-class helpers shared by the metric
// definitions and the tracker.

func percent(a, b int) float64 {
	if b <= 0 {
		return 0
	}
	return float64(a) / float64(b) * 100
}

func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	cp := append([]float64(nil), xs...)
	sort.Float64s(cp)
	n := len(cp)
	if n%2 == 1 {
		return cp[n/2]
	}
	return (cp[n/2-1] + cp[n/2]) / 2
}

func isGun(t common.EquipmentType) bool {
	switch t.Class() {
	case common.EqClassPistols, common.EqClassSMG, common.EqClassRifle, common.EqClassHeavy:
		return true
	}
	return false
}

func countsForAccuracy(t common.EquipmentType) bool {
	if !isGun(t) {
		return false
	}
	switch t {
	case common.EqSawedOff, common.EqNova, common.EqMag7, common.EqXM1014:
		return false // multiple pellets make hits/shots meaningless here
	}
	return true
}

func isAutomatic(t common.EquipmentType) bool {
	switch t.Class() {
	case common.EqClassSMG:
		return true
	case common.EqClassRifle:
		switch t {
		case common.EqScout, common.EqAWP, common.EqScar20, common.EqG3SG1:
			return false
		}
		return true
	case common.EqClassHeavy:
		return t == common.EqM249 || t == common.EqNegev
	}
	return false
}
