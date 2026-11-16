package stats

import (
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
)

// intMetric returns an integer metric value, or 0 when absent.
func intMetric(ps *PlayerStats, cat Category, k Key) int64 {
	m, _ := ps.GetMetric(cat, k)
	return m.IntValue
}

// addIntMetric adds n to an integer metric, creating it when absent.
func addIntMetric(ps *PlayerStats, cat Category, k Key, n int64) {
	if m, ok := ps.GetMetric(cat, k); ok {
		m.IntValue += n
		m.Type = MetricInteger
		ps.AddMetric(cat, k, m)
		return
	}
	ps.AddMetric(cat, k, Metric{Type: MetricInteger, IntValue: n})
}

// getWeaponName returns a readable name for a weapon.
func getWeaponName(weapon *common.Equipment) string {
	if weapon == nil {
		return "Unknown"
	}
	if name := weapon.String(); name != "" && name != "Unknown" {
		return name
	}
	return weaponTypeToString(weapon.Type)
}

// weaponTypeToString is the fallback mapping used when the demo does not carry
// a display name for the weapon.
func weaponTypeToString(weaponType common.EquipmentType) string {
	switch weaponType {
	case common.EqAK47:
		return "ak47"
	case common.EqM4A4:
		return "m4a4"
	case common.EqM4A1:
		return "m4a1"
	case common.EqFamas:
		return "famas"
	case common.EqGalil:
		return "galil"
	case common.EqMP7:
		return "mp7"
	case common.EqMP9:
		return "mp9"
	case common.EqP90:
		return "p90"
	case common.EqUMP:
		return "ump"
	case common.EqNegev:
		return "negev"
	case common.EqM249:
		return "m249"
	case common.EqSG556:
		return "sg556"
	case common.EqAUG:
		return "aug"
	default:
		return "unknown"
	}
}
