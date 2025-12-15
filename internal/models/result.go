package models

import (
	"fmt"
	"strings"
)

// DamageResult holds the result of a damage calculation
type DamageResult struct {
	// Damage values
	Damages []int `json:"damages"` // All 16 damage rolls (85-100%)

	// Damage range
	MinDamage    int     `json:"minDamage"`
	MaxDamage    int     `json:"maxDamage"`
	MinPercent   float64 `json:"minPercent"`
	MaxPercent   float64 `json:"maxPercent"`

	// KO information
	KOChance     *KOChance `json:"ko,omitempty"`

	// Recoil and recovery
	Recoil   *RecoilResult   `json:"recoil,omitempty"`
	Recovery *RecoveryResult `json:"recovery,omitempty"`

	// Description
	Description string `json:"description"`

	// Debug information
	Factors []string `json:"factors,omitempty"` // What affected the calculation
}

// KOChance represents knockout probability
type KOChance struct {
	Chance     float64 `json:"chance"`     // 0.0 to 1.0
	N          int     `json:"n"`          // Number of hits needed (1 = OHKO, 2 = 2HKO, etc.)
	Guaranteed bool    `json:"guaranteed"` // True if 100% chance
	Text       string  `json:"text"`       // Human-readable description
}

// RecoilResult represents recoil damage
type RecoilResult struct {
	Damage  int     `json:"damage"`
	Percent float64 `json:"percent"`
}

// RecoveryResult represents HP recovery
type RecoveryResult struct {
	MinRecovery int     `json:"minRecovery"`
	MaxRecovery int     `json:"maxRecovery"`
	MinPercent  float64 `json:"minPercent"`
	MaxPercent  float64 `json:"maxPercent"`
}

// NewDamageResult creates a new DamageResult from damage rolls
func NewDamageResult(damages []int, defenderMaxHP int) *DamageResult {
	if len(damages) == 0 {
		return &DamageResult{
			Damages: []int{0},
		}
	}

	result := &DamageResult{
		Damages:   damages,
		MinDamage: damages[0],
		MaxDamage: damages[len(damages)-1],
	}

	if defenderMaxHP > 0 {
		result.MinPercent = float64(result.MinDamage) / float64(defenderMaxHP) * 100
		result.MaxPercent = float64(result.MaxDamage) / float64(defenderMaxHP) * 100
	}

	return result
}

// CalculateKO calculates KO probability
func (r *DamageResult) CalculateKO(defenderHP, defenderMaxHP int) {
	if len(r.Damages) == 0 || defenderHP <= 0 {
		return
	}

	// Check for OHKO
	if r.MinDamage >= defenderHP {
		r.KOChance = &KOChance{
			Chance:     1.0,
			N:          1,
			Guaranteed: true,
			Text:       "guaranteed OHKO",
		}
		return
	}

	if r.MaxDamage >= defenderHP {
		// Partial OHKO - count favorable rolls
		favorable := 0
		for _, dmg := range r.Damages {
			if dmg >= defenderHP {
				favorable++
			}
		}
		chance := float64(favorable) / float64(len(r.Damages))
		r.KOChance = &KOChance{
			Chance:     chance,
			N:          1,
			Guaranteed: false,
			Text:       fmt.Sprintf("%.1f%% chance to OHKO", chance*100),
		}
		return
	}

	// Check for 2HKO, 3HKO, etc.
	for n := 2; n <= 4; n++ {
		if r.MinDamage*n >= defenderHP {
			r.KOChance = &KOChance{
				Chance:     1.0,
				N:          n,
				Guaranteed: true,
				Text:       fmt.Sprintf("guaranteed %dHKO", n),
			}
			return
		}
		if r.MaxDamage*n >= defenderHP {
			// Estimate probability (simplified)
			avgDamage := (r.MinDamage + r.MaxDamage) / 2
			if avgDamage*n >= defenderHP {
				r.KOChance = &KOChance{
					Chance:     0.5, // Rough estimate
					N:          n,
					Guaranteed: false,
					Text:       fmt.Sprintf("possible %dHKO", n),
				}
			} else {
				r.KOChance = &KOChance{
					Chance:     0.25,
					N:          n,
					Guaranteed: false,
					Text:       fmt.Sprintf("possible %dHKO (unlikely)", n),
				}
			}
			return
		}
	}

	// No KO in 4 hits
	r.KOChance = &KOChance{
		Chance:     0,
		N:          0,
		Guaranteed: false,
		Text:       "not a KO",
	}
}

// CalculateRecoil calculates recoil damage
func (r *DamageResult) CalculateRecoil(attackerMaxHP, recoilNum, recoilDenom int) {
	if recoilNum == 0 || recoilDenom == 0 {
		return
	}

	// Recoil is based on damage dealt
	avgDamage := (r.MinDamage + r.MaxDamage) / 2
	recoilDamage := avgDamage * recoilNum / recoilDenom

	r.Recoil = &RecoilResult{
		Damage:  recoilDamage,
		Percent: float64(recoilDamage) / float64(attackerMaxHP) * 100,
	}
}

// CalculateRecovery calculates HP recovery from drain moves
func (r *DamageResult) CalculateRecovery(attackerMaxHP, drainNum, drainDenom int) {
	if drainNum == 0 || drainDenom == 0 {
		return
	}

	minRecovery := r.MinDamage * drainNum / drainDenom
	maxRecovery := r.MaxDamage * drainNum / drainDenom

	// Cap at max HP
	if minRecovery > attackerMaxHP {
		minRecovery = attackerMaxHP
	}
	if maxRecovery > attackerMaxHP {
		maxRecovery = attackerMaxHP
	}

	r.Recovery = &RecoveryResult{
		MinRecovery: minRecovery,
		MaxRecovery: maxRecovery,
		MinPercent:  float64(minRecovery) / float64(attackerMaxHP) * 100,
		MaxPercent:  float64(maxRecovery) / float64(attackerMaxHP) * 100,
	}
}

// BuildDescription builds a human-readable damage description
func (r *DamageResult) BuildDescription(attacker, defender *BattlePokemon, move *BattleMove) {
	var parts []string

	// Attacker boosts
	if move.IsPhysical() && attacker.Boosts.Atk != 0 {
		parts = append(parts, fmt.Sprintf("%+d", attacker.Boosts.Atk))
	} else if move.IsSpecial() && attacker.Boosts.SpA != 0 {
		parts = append(parts, fmt.Sprintf("%+d", attacker.Boosts.SpA))
	}

	// Attacker EVs and stat
	if move.IsPhysical() {
		parts = append(parts, fmt.Sprintf("%d Atk", attacker.EVs.Atk))
	} else if move.IsSpecial() {
		parts = append(parts, fmt.Sprintf("%d SpA", attacker.EVs.SpA))
	}

	// Attacker name
	if attacker.SpeciesData != nil {
		parts = append(parts, attacker.SpeciesData.Name)
	} else {
		parts = append(parts, attacker.Species)
	}

	// Move name
	if move.MoveData != nil {
		parts = append(parts, move.MoveData.Name)
	} else {
		parts = append(parts, move.Name)
	}

	parts = append(parts, "vs.")

	// Defender EVs and stat
	if move.GetDefensiveCategory() == "Physical" {
		parts = append(parts, fmt.Sprintf("%d HP / %d Def", defender.EVs.HP, defender.EVs.Def))
	} else {
		parts = append(parts, fmt.Sprintf("%d HP / %d SpD", defender.EVs.HP, defender.EVs.SpD))
	}

	// Defender name
	if defender.SpeciesData != nil {
		parts = append(parts, defender.SpeciesData.Name)
	} else {
		parts = append(parts, defender.Species)
	}

	// Damage range
	damageStr := fmt.Sprintf(": %d-%d (%.1f%% - %.1f%%)",
		r.MinDamage, r.MaxDamage, r.MinPercent, r.MaxPercent)

	// KO chance
	koStr := ""
	if r.KOChance != nil {
		koStr = " -- " + r.KOChance.Text
	}

	r.Description = strings.Join(parts, " ") + damageStr + koStr
}

// AddFactor adds a factor that affected the calculation
func (r *DamageResult) AddFactor(factor string) {
	r.Factors = append(r.Factors, factor)
}
