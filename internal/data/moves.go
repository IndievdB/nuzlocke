package data

// Move represents a Pokemon move
type Move struct {
	Num         int              `json:"num"`
	Name        string           `json:"name"`
	Type        string           `json:"type"`
	Category    string           `json:"category"` // Physical, Special, Status
	BasePower   int              `json:"basePower"`
	Accuracy    interface{}      `json:"accuracy"` // Can be int or true (always hits)
	PP          int              `json:"pp"`
	Priority    int              `json:"priority"`
	Flags       map[string]int   `json:"flags,omitempty"`
	Target      string           `json:"target"`
	Secondary   *MoveSecondary   `json:"secondary,omitempty"`
	Secondaries []MoveSecondary  `json:"secondaries,omitempty"`
	CritRatio   int              `json:"critRatio,omitempty"`
	Drain       []int            `json:"drain,omitempty"`       // [numerator, denominator]
	Recoil      []int            `json:"recoil,omitempty"`      // [numerator, denominator]
	Multihit    interface{}      `json:"multihit,omitempty"`    // Can be int or [min, max]
	IgnoreAbility bool           `json:"ignoreAbility,omitempty"`
	IgnoreDefensive bool         `json:"ignoreDefensive,omitempty"`
	IgnoreEvasion bool           `json:"ignoreEvasion,omitempty"`
	IgnoreImmunity interface{}   `json:"ignoreImmunity,omitempty"`
	WillCrit    bool             `json:"willCrit,omitempty"`
	BreaksProtect bool           `json:"breaksProtect,omitempty"`
	DefensiveCategory string     `json:"defensiveCategory,omitempty"` // For moves like Psyshock
	OverrideOffensiveStat string `json:"overrideOffensiveStat,omitempty"`
	OverrideDefensiveStat string `json:"overrideDefensiveStat,omitempty"`
	Desc                  string `json:"desc,omitempty"`
	ShortDesc             string `json:"shortDesc,omitempty"`
}

// MoveSecondary represents secondary effects of a move
type MoveSecondary struct {
	Chance       int               `json:"chance,omitempty"`
	Status       string            `json:"status,omitempty"`
	VolatileStatus string          `json:"volatileStatus,omitempty"`
	Boosts       map[string]int    `json:"boosts,omitempty"`
	Self         *MoveSecondarySelf `json:"self,omitempty"`
}

// MoveSecondarySelf represents self-targeting secondary effects
type MoveSecondarySelf struct {
	Boosts map[string]int `json:"boosts,omitempty"`
}

// IsPhysical returns true if the move is physical
func (m *Move) IsPhysical() bool {
	return m.Category == "Physical"
}

// IsSpecial returns true if the move is special
func (m *Move) IsSpecial() bool {
	return m.Category == "Special"
}

// IsStatus returns true if the move is a status move
func (m *Move) IsStatus() bool {
	return m.Category == "Status"
}

// HasFlag checks if the move has the given flag
func (m *Move) HasFlag(flag string) bool {
	if m.Flags == nil {
		return false
	}
	_, ok := m.Flags[flag]
	return ok
}

// GetAccuracy returns the accuracy as an integer (100 for always-hits moves)
func (m *Move) GetAccuracy() int {
	switch v := m.Accuracy.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case bool:
		if v {
			return 100 // Always hits
		}
		return 0
	default:
		return 100
	}
}

// GetMultihit returns the number of hits as [min, max]
func (m *Move) GetMultihit() (int, int) {
	if m.Multihit == nil {
		return 1, 1
	}
	switch v := m.Multihit.(type) {
	case float64:
		n := int(v)
		return n, n
	case int:
		return v, v
	case []interface{}:
		if len(v) == 2 {
			min, _ := v[0].(float64)
			max, _ := v[1].(float64)
			return int(min), int(max)
		}
	}
	return 1, 1
}

// GetDrain returns the drain ratio as [numerator, denominator], or nil if none
func (m *Move) GetDrain() (int, int, bool) {
	if len(m.Drain) == 2 {
		return m.Drain[0], m.Drain[1], true
	}
	return 0, 0, false
}

// GetRecoil returns the recoil ratio as [numerator, denominator], or nil if none
func (m *Move) GetRecoil() (int, int, bool) {
	if len(m.Recoil) == 2 {
		return m.Recoil[0], m.Recoil[1], true
	}
	return 0, 0, false
}
