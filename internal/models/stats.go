package models

import "nuzlocke/internal/data"

// StatSpread represents a Pokemon's EVs or IVs
type StatSpread struct {
	HP  int `json:"hp"`
	Atk int `json:"atk"`
	Def int `json:"def"`
	SpA int `json:"spa"`
	SpD int `json:"spd"`
	Spe int `json:"spe"`
}

// DefaultIVs returns a stat spread with all IVs at 31
func DefaultIVs() StatSpread {
	return StatSpread{HP: 31, Atk: 31, Def: 31, SpA: 31, SpD: 31, Spe: 31}
}

// DefaultEVs returns a stat spread with all EVs at 0
func DefaultEVs() StatSpread {
	return StatSpread{}
}

// CalculatedStats holds all calculated stats for a Pokemon
type CalculatedStats struct {
	HP  int
	Atk int
	Def int
	SpA int
	SpD int
	Spe int
}

// StatBoosts represents stat boost stages for a Pokemon
type StatBoosts struct {
	Atk int `json:"atk"`
	Def int `json:"def"`
	SpA int `json:"spa"`
	SpD int `json:"spd"`
	Spe int `json:"spe"`
	// Accuracy and evasion are separate mechanics
	Accuracy int `json:"accuracy,omitempty"`
	Evasion  int `json:"evasion,omitempty"`
}

// GetBoost returns the boost for a given stat name
func (b StatBoosts) GetBoost(stat string) int {
	switch stat {
	case "atk":
		return b.Atk
	case "def":
		return b.Def
	case "spa":
		return b.SpA
	case "spd":
		return b.SpD
	case "spe":
		return b.Spe
	default:
		return 0
	}
}

// floorDiv performs integer floor division
func floorDiv(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}

// calculateStat calculates a stat value given base stat, IV, EV, level, and nature
func calculateStat(base, iv, ev, level int, natureModifier int) int {
	stat := floorDiv((2*base+iv+floorDiv(ev, 4))*level, 100) + 5
	stat = floorDiv(stat*natureModifier, 10)
	return stat
}

// calculateHP calculates the HP stat
func calculateHP(base, iv, ev, level int) int {
	if base == 1 { // Shedinja
		return 1
	}
	return floorDiv((2*base+iv+floorDiv(ev, 4))*level, 100) + level + 10
}

// CalculateAllStats calculates all stats for a Pokemon
func CalculateAllStats(baseStats data.BaseStats, ivs, evs StatSpread, level int, nature *data.Nature) CalculatedStats {
	return CalculatedStats{
		HP:  calculateHP(baseStats.HP, ivs.HP, evs.HP, level),
		Atk: calculateStat(baseStats.Atk, ivs.Atk, evs.Atk, level, nature.GetStatModifier("atk")),
		Def: calculateStat(baseStats.Def, ivs.Def, evs.Def, level, nature.GetStatModifier("def")),
		SpA: calculateStat(baseStats.SpA, ivs.SpA, evs.SpA, level, nature.GetStatModifier("spa")),
		SpD: calculateStat(baseStats.SpD, ivs.SpD, evs.SpD, level, nature.GetStatModifier("spd")),
		Spe: calculateStat(baseStats.Spe, ivs.Spe, evs.Spe, level, nature.GetStatModifier("spe")),
	}
}

// Stat boost multipliers for stages -6 to +6
var boostMultipliers = []struct {
	Numerator   int
	Denominator int
}{
	{2, 8}, // -6
	{2, 7}, // -5
	{2, 6}, // -4
	{2, 5}, // -3
	{2, 4}, // -2
	{2, 3}, // -1
	{2, 2}, // 0
	{3, 2}, // +1
	{4, 2}, // +2
	{5, 2}, // +3
	{6, 2}, // +4
	{7, 2}, // +5
	{8, 2}, // +6
}

// ApplyStatBoost applies a stat boost stage to a stat value
func ApplyStatBoost(stat, stage int) int {
	index := stage + 6
	if index < 0 {
		index = 0
	}
	if index > 12 {
		index = 12
	}
	mult := boostMultipliers[index]
	return floorDiv(stat*mult.Numerator, mult.Denominator)
}

// GetModifiedStat returns a stat after applying boosts
// If isCrit is true and the boost is unfavorable to the attacker, ignore it
func GetModifiedStat(baseStat, boost int, isCrit, isAttacker bool) int {
	if isCrit {
		if isAttacker && boost < 0 {
			boost = 0
		}
		if !isAttacker && boost > 0 {
			boost = 0
		}
	}
	return ApplyStatBoost(baseStat, boost)
}
