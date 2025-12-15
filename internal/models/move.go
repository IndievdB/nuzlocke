package models

import "nuzlocke/internal/data"

// BattleMove represents a move being used in battle
type BattleMove struct {
	// Move identification
	Name     string      `json:"name"`
	MoveData *data.Move  `json:"-"`

	// Override properties (for custom calculations)
	BasePower int    `json:"basePower,omitempty"` // Override base power
	Type      string `json:"type,omitempty"`      // Override type

	// Battle state
	IsCrit       bool `json:"isCrit,omitempty"`
	IsSpread     bool `json:"isSpread,omitempty"`     // Hitting multiple targets
	HitsMultiple bool `json:"hitsMultiple,omitempty"` // Multi-target move in doubles
	UseZMove     bool `json:"useZMove,omitempty"`
	UseMaxMove   bool `json:"useMaxMove,omitempty"`
}

// NewBattleMove creates a new BattleMove
func NewBattleMove(name string) *BattleMove {
	return &BattleMove{
		Name: name,
	}
}

// Initialize populates the move data from the store
func (bm *BattleMove) Initialize(store *data.Store) {
	bm.MoveData = store.GetMove(bm.Name)
}

// GetBasePower returns the base power (override or from data)
func (bm *BattleMove) GetBasePower() int {
	if bm.BasePower > 0 {
		return bm.BasePower
	}
	if bm.MoveData != nil {
		return bm.MoveData.BasePower
	}
	return 0
}

// GetType returns the move type (override or from data)
func (bm *BattleMove) GetType() string {
	if bm.Type != "" {
		return bm.Type
	}
	if bm.MoveData != nil {
		return bm.MoveData.Type
	}
	return "Normal"
}

// GetCategory returns the move category (Physical, Special, Status)
func (bm *BattleMove) GetCategory() string {
	if bm.MoveData != nil {
		return bm.MoveData.Category
	}
	return "Physical"
}

// IsPhysical returns true if the move is physical
func (bm *BattleMove) IsPhysical() bool {
	return bm.GetCategory() == "Physical"
}

// IsSpecial returns true if the move is special
func (bm *BattleMove) IsSpecial() bool {
	return bm.GetCategory() == "Special"
}

// IsStatus returns true if the move is a status move
func (bm *BattleMove) IsStatus() bool {
	return bm.GetCategory() == "Status"
}

// HasFlag checks if the move has a flag
func (bm *BattleMove) HasFlag(flag string) bool {
	if bm.MoveData != nil {
		return bm.MoveData.HasFlag(flag)
	}
	return false
}

// IsPunchMove returns true if this is a punch move
func (bm *BattleMove) IsPunchMove() bool {
	return bm.HasFlag("punch")
}

// IsSoundMove returns true if this is a sound-based move
func (bm *BattleMove) IsSoundMove() bool {
	return bm.HasFlag("sound")
}

// IsBiteMove returns true if this is a biting move
func (bm *BattleMove) IsBiteMove() bool {
	return bm.HasFlag("bite")
}

// IsBulletMove returns true if this is a bullet/ball move
func (bm *BattleMove) IsBulletMove() bool {
	return bm.HasFlag("bullet")
}

// IsContactMove returns true if the move makes contact
func (bm *BattleMove) IsContactMove() bool {
	return bm.HasFlag("contact")
}

// IsRecoilMove returns true if the move has recoil
func (bm *BattleMove) IsRecoilMove() bool {
	if bm.MoveData != nil {
		_, _, hasRecoil := bm.MoveData.GetRecoil()
		return hasRecoil
	}
	return false
}

// IsDrainMove returns true if the move drains HP
func (bm *BattleMove) IsDrainMove() bool {
	if bm.MoveData != nil {
		_, _, hasDrain := bm.MoveData.GetDrain()
		return hasDrain
	}
	return false
}

// GetDrain returns the drain ratio [numerator, denominator]
func (bm *BattleMove) GetDrain() (int, int, bool) {
	if bm.MoveData != nil {
		return bm.MoveData.GetDrain()
	}
	return 0, 0, false
}

// GetRecoil returns the recoil ratio [numerator, denominator]
func (bm *BattleMove) GetRecoil() (int, int, bool) {
	if bm.MoveData != nil {
		return bm.MoveData.GetRecoil()
	}
	return 0, 0, false
}

// GetMultihit returns the hit range [min, max]
func (bm *BattleMove) GetMultihit() (int, int) {
	if bm.MoveData != nil {
		return bm.MoveData.GetMultihit()
	}
	return 1, 1
}

// IsMultiHit returns true if this is a multi-hit move
func (bm *BattleMove) IsMultiHit() bool {
	min, max := bm.GetMultihit()
	return min > 1 || max > 1
}

// WillCrit returns true if this move always crits
func (bm *BattleMove) WillCrit() bool {
	if bm.MoveData != nil {
		return bm.MoveData.WillCrit
	}
	return false
}

// GetDefensiveCategory returns which defense stat to use
// Some moves like Psyshock use physical defense with special attack
func (bm *BattleMove) GetDefensiveCategory() string {
	if bm.MoveData != nil && bm.MoveData.DefensiveCategory != "" {
		return bm.MoveData.DefensiveCategory
	}
	return bm.GetCategory()
}

// HasSecondaryEffect returns true if the move has a secondary effect
func (bm *BattleMove) HasSecondaryEffect() bool {
	if bm.MoveData == nil {
		return false
	}
	return bm.MoveData.Secondary != nil || len(bm.MoveData.Secondaries) > 0
}

// Priority returns the move's priority
func (bm *BattleMove) Priority() int {
	if bm.MoveData != nil {
		return bm.MoveData.Priority
	}
	return 0
}
