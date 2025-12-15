package calc

import (
	"nuzlocke/internal/data"
	"nuzlocke/internal/models"
)

// Calculator performs damage calculations
type Calculator struct {
	Store *data.Store
}

// NewCalculator creates a new Calculator with the given data store
func NewCalculator(store *data.Store) *Calculator {
	return &Calculator{Store: store}
}

// CalculateRequest represents a damage calculation request
type CalculateRequest struct {
	Attacker   *models.BattlePokemon `json:"attacker"`
	Defender   *models.BattlePokemon `json:"defender"`
	Move       *models.BattleMove    `json:"move"`
	Field      *models.Field         `json:"field"`
	Generation int                   `json:"generation"` // 3 = Gen 3, 5+ = Gen 5+ mechanics
}

// Calculate performs a damage calculation
func (c *Calculator) Calculate(req *CalculateRequest) *models.DamageResult {
	// Initialize all entities with data from store
	req.Attacker.Initialize(c.Store)
	req.Defender.Initialize(c.Store)
	req.Move.Initialize(c.Store)

	// Set default field if not provided
	if req.Field == nil {
		req.Field = models.NewField()
	}

	// Override generation from request if specified
	if req.Generation > 0 {
		req.Field.Generation = req.Generation
	}

	// Check for status moves
	if req.Move.IsStatus() {
		return &models.DamageResult{
			Damages:     []int{0},
			Description: "Status moves deal no damage",
		}
	}

	// Use appropriate formula based on generation
	var damages []int
	var factors []string

	if req.Field.IsGen3() {
		damages, factors = c.calculateGen3(req)
	} else {
		damages, factors = c.calculateGen5Plus(req)
	}

	// Build result
	result := models.NewDamageResult(damages, req.Defender.GetMaxHP())
	result.Factors = factors

	// Calculate KO chance
	result.CalculateKO(req.Defender.GetCurrentHP(), req.Defender.GetMaxHP())

	// Calculate recoil
	if num, denom, hasRecoil := req.Move.GetRecoil(); hasRecoil {
		result.CalculateRecoil(req.Attacker.GetMaxHP(), num, denom)
	}

	// Life Orb recoil
	if req.Attacker.HasItem("lifeorb") && !req.Attacker.HasAbility("sheerforce") {
		recoilDamage := req.Attacker.GetMaxHP() / 10
		result.Recoil = &models.RecoilResult{
			Damage:  recoilDamage,
			Percent: float64(recoilDamage) / float64(req.Attacker.GetMaxHP()) * 100,
		}
	}

	// Calculate recovery for drain moves
	if num, denom, hasDrain := req.Move.GetDrain(); hasDrain {
		result.CalculateRecovery(req.Attacker.GetMaxHP(), num, denom)
	}

	// Build description
	result.BuildDescription(req.Attacker, req.Defender, req.Move)

	return result
}

// getAttackStat returns the attack stat to use (Atk or SpA)
func (c *Calculator) getAttackStat(attacker *models.BattlePokemon, move *models.BattleMove, field *models.Field) (int, string) {
	var stat int
	var statName string

	// Determine which stat to use based on move category or overrides
	if field.IsGen3() {
		// Gen 3: Type determines physical/special
		if data.IsPhysicalInGen3(move.GetType()) {
			stat = attacker.GetStat("atk", move.IsCrit, true)
			statName = "atk"
		} else {
			stat = attacker.GetStat("spa", move.IsCrit, true)
			statName = "spa"
		}
	} else {
		// Gen 4+: Category determines physical/special
		if move.IsPhysical() {
			stat = attacker.GetStat("atk", move.IsCrit, true)
			statName = "atk"
		} else {
			stat = attacker.GetStat("spa", move.IsCrit, true)
			statName = "spa"
		}
	}

	return stat, statName
}

// getDefenseStat returns the defense stat to use (Def or SpD)
func (c *Calculator) getDefenseStat(defender *models.BattlePokemon, move *models.BattleMove, field *models.Field) (int, string) {
	var stat int
	var statName string

	// Handle moves like Psyshock that use physical defense with special attack
	defCategory := move.GetDefensiveCategory()

	if field.IsGen3() {
		// Gen 3: Type determines physical/special
		if data.IsPhysicalInGen3(move.GetType()) {
			stat = defender.GetStat("def", move.IsCrit, false)
			statName = "def"
		} else {
			stat = defender.GetStat("spd", move.IsCrit, false)
			statName = "spd"
		}
	} else {
		// Gen 4+: Category determines physical/special
		if defCategory == "Physical" {
			stat = defender.GetStat("def", move.IsCrit, false)
			statName = "def"
		} else {
			stat = defender.GetStat("spd", move.IsCrit, false)
			statName = "spd"
		}
	}

	return stat, statName
}

// hasSTAB returns true if the attacker gets STAB for the move
func (c *Calculator) hasSTAB(attacker *models.BattlePokemon, move *models.BattleMove) bool {
	return attacker.HasType(move.GetType())
}

// getTypeEffectiveness returns the type effectiveness multiplier
func (c *Calculator) getTypeEffectiveness(move *models.BattleMove, defender *models.BattlePokemon) float64 {
	return c.Store.GetTypeEffectivenessMultiple(move.GetType(), defender.Types)
}

// getBasePower returns the move's base power after modifications
func (c *Calculator) getBasePower(attacker *models.BattlePokemon, defender *models.BattlePokemon, move *models.BattleMove, field *models.Field) int {
	bp := move.GetBasePower()
	if bp == 0 {
		return 0
	}

	// Technician: 1.5x for moves with base power <= 60
	if attacker.HasAbility("technician") && bp <= 60 {
		bp = ApplyModifier(bp, ModTechnician)
	}

	// Analytic: 1.3x if moving last (we'll assume yes for calculator purposes if set)
	// This would need battle state to determine properly

	// Terrain boosts
	if field.IsElectricTerrain() && move.GetType() == "Electric" {
		if field.Generation >= 8 {
			bp = ApplyModifier(bp, ModTerrainBoost)
		} else {
			bp = ApplyModifier(bp, ModTerrainGen7)
		}
	}
	if field.IsGrassyTerrain() && move.GetType() == "Grass" {
		if field.Generation >= 8 {
			bp = ApplyModifier(bp, ModTerrainBoost)
		} else {
			bp = ApplyModifier(bp, ModTerrainGen7)
		}
	}
	if field.IsPsychicTerrain() && move.GetType() == "Psychic" {
		if field.Generation >= 8 {
			bp = ApplyModifier(bp, ModTerrainBoost)
		} else {
			bp = ApplyModifier(bp, ModTerrainGen7)
		}
	}

	// Helping Hand: 1.5x
	if field.AttackerSide.HelpingHand {
		bp = ApplyModifier(bp, 6144) // 1.5x
	}

	return bp
}
