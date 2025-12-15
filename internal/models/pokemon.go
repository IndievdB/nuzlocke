package models

import (
	"nuzlocke/internal/data"
)

// BattlePokemon represents a Pokemon in battle with all relevant state
type BattlePokemon struct {
	// Species data
	Species     string        `json:"species"`
	SpeciesData *data.Pokemon `json:"-"`

	// Level and stats
	Level int        `json:"level"`
	EVs   StatSpread `json:"evs"`
	IVs   StatSpread `json:"ivs"`

	// Calculated stats
	Stats CalculatedStats `json:"-"`

	// Nature and ability
	Nature      string        `json:"nature"`
	NatureData  *data.Nature  `json:"-"`
	Ability     string        `json:"ability"`
	AbilityData *data.Ability `json:"-"`

	// Item
	Item     string     `json:"item"`
	ItemData *data.Item `json:"-"`

	// Current battle state
	CurrentHP int    `json:"currentHP,omitempty"` // If 0, use max HP
	Status    string `json:"status,omitempty"`    // brn, par, psn, tox, slp, frz
	Boosts    StatBoosts `json:"boosts"`

	// Types (can be overridden by Tera, etc.)
	Types []string `json:"types,omitempty"`

	// Gender (for some calculations)
	Gender string `json:"gender,omitempty"`

	// Volatile conditions
	Volatiles map[string]bool `json:"volatiles,omitempty"`
}

// NewBattlePokemon creates a new BattlePokemon with default values
func NewBattlePokemon(species string) *BattlePokemon {
	return &BattlePokemon{
		Species: species,
		Level:   100,
		EVs:     DefaultEVs(),
		IVs:     DefaultIVs(),
		Boosts:  StatBoosts{},
	}
}

// Initialize populates computed fields from the data store
func (bp *BattlePokemon) Initialize(store *data.Store) {
	// Get species data
	bp.SpeciesData = store.GetPokemon(bp.Species)
	if bp.SpeciesData == nil {
		return
	}

	// Set types from species if not overridden
	if len(bp.Types) == 0 {
		bp.Types = bp.SpeciesData.Types
	}

	// Get nature data
	if bp.Nature == "" {
		bp.Nature = "hardy" // Default neutral nature
	}
	bp.NatureData = store.GetNature(bp.Nature)

	// Get ability data
	if bp.AbilityData == nil && bp.Ability != "" {
		bp.AbilityData = store.GetAbility(bp.Ability)
	}

	// Get item data
	if bp.ItemData == nil && bp.Item != "" {
		bp.ItemData = store.GetItem(bp.Item)
	}

	// Calculate stats
	if bp.NatureData != nil {
		bp.Stats = CalculateAllStats(
			bp.SpeciesData.BaseStats,
			bp.IVs,
			bp.EVs,
			bp.Level,
			bp.NatureData,
		)
	}
}

// GetMaxHP returns the maximum HP
func (bp *BattlePokemon) GetMaxHP() int {
	return bp.Stats.HP
}

// GetCurrentHP returns the current HP (defaults to max if not set)
func (bp *BattlePokemon) GetCurrentHP() int {
	if bp.CurrentHP > 0 {
		return bp.CurrentHP
	}
	return bp.GetMaxHP()
}

// GetCurrentHPPercent returns current HP as a percentage
func (bp *BattlePokemon) GetCurrentHPPercent() float64 {
	maxHP := bp.GetMaxHP()
	if maxHP == 0 {
		return 0
	}
	return float64(bp.GetCurrentHP()) / float64(maxHP) * 100
}

// IsAtFullHP returns true if at full HP
func (bp *BattlePokemon) IsAtFullHP() bool {
	return bp.GetCurrentHP() >= bp.GetMaxHP()
}

// HasType checks if the Pokemon has the given type
func (bp *BattlePokemon) HasType(t string) bool {
	for _, pt := range bp.Types {
		if pt == t {
			return true
		}
	}
	return false
}

// HasAbility checks if the Pokemon has the given ability
func (bp *BattlePokemon) HasAbility(ability string) bool {
	if bp.Ability == "" {
		return false
	}
	return data.ToID(bp.Ability) == data.ToID(ability)
}

// HasItem checks if the Pokemon has the given item
func (bp *BattlePokemon) HasItem(item string) bool {
	if bp.Item == "" {
		return false
	}
	return data.ToID(bp.Item) == data.ToID(item)
}

// IsBurned returns true if the Pokemon has burn status
func (bp *BattlePokemon) IsBurned() bool {
	return bp.Status == "brn"
}

// IsParalyzed returns true if the Pokemon has paralysis status
func (bp *BattlePokemon) IsParalyzed() bool {
	return bp.Status == "par"
}

// HasVolatile checks if the Pokemon has a volatile condition
func (bp *BattlePokemon) HasVolatile(volatile string) bool {
	if bp.Volatiles == nil {
		return false
	}
	return bp.Volatiles[volatile]
}

// GetStat returns a stat value with boosts applied
func (bp *BattlePokemon) GetStat(stat string, isCrit, isAttacker bool) int {
	var baseStat int
	boost := bp.Boosts.GetBoost(stat)

	switch stat {
	case "atk":
		baseStat = bp.Stats.Atk
	case "def":
		baseStat = bp.Stats.Def
	case "spa":
		baseStat = bp.Stats.SpA
	case "spd":
		baseStat = bp.Stats.SpD
	case "spe":
		baseStat = bp.Stats.Spe
	default:
		return 0
	}

	return GetModifiedStat(baseStat, boost, isCrit, isAttacker)
}

// GetWeight returns the Pokemon's weight in kg
func (bp *BattlePokemon) GetWeight() float64 {
	if bp.SpeciesData == nil {
		return 0
	}
	weight := bp.SpeciesData.Weightkg

	// Heavy Metal doubles weight
	if bp.HasAbility("heavymetal") {
		weight *= 2
	}
	// Light Metal halves weight
	if bp.HasAbility("lightmetal") {
		weight /= 2
	}
	// Float Stone halves weight
	if bp.HasItem("floatstone") {
		weight /= 2
	}

	// Minimum weight is 0.1 kg
	if weight < 0.1 {
		weight = 0.1
	}

	return weight
}
