package calc

import (
	"nuzlocke/internal/data"
	"nuzlocke/internal/models"
)

// calculateGen3 calculates damage using Gen 3 mechanics
// Move TYPE determines physical/special split (not category)
func (c *Calculator) calculateGen3(req *CalculateRequest) ([]int, []string) {
	attacker := req.Attacker
	defender := req.Defender
	move := req.Move
	field := req.Field

	var factors []string
	factors = append(factors, "Gen 3 mechanics")

	// Get base power
	basePower := c.getBasePower(attacker, defender, move, field)
	if basePower == 0 {
		return []int{0}, factors
	}

	// In Gen 3, physical/special is determined by type, not move category
	isPhysical := data.IsPhysicalInGen3(move.GetType())

	// Get attack and defense stats
	var attack, defense int
	var atkName, defName string

	if isPhysical {
		attack = attacker.GetStat("atk", move.IsCrit, true)
		defense = defender.GetStat("def", move.IsCrit, false)
		atkName = "atk"
		defName = "def"
	} else {
		attack = attacker.GetStat("spa", move.IsCrit, true)
		defense = defender.GetStat("spd", move.IsCrit, false)
		atkName = "spa"
		defName = "spd"
	}

	// Apply stat modifiers (Gen 3 style)
	attack = c.applyAttackModifiersGen3(attack, attacker, isPhysical, field, &factors)
	defense = c.applyDefenseModifiersGen3(defense, defender, isPhysical, field, &factors)

	factors = append(factors, atkName+"/"+defName)

	// Base damage calculation (Gen 3 formula)
	// ((2 * Level / 5 + 2) * BasePower * Attack / Defense) / 50
	level := attacker.Level
	levelFactor := FloorDiv(2*level, 5) + 2
	baseDamage := FloorDiv(levelFactor*basePower*attack, defense)
	baseDamage = FloorDiv(baseDamage, 50)

	// In Gen 3, physical moves add 2, special moves don't (different from later gens)
	// Actually, all moves add 2 in Gen 3 as well
	baseDamage += 2

	// Apply modifiers in Gen 3 order
	baseDamage = c.applyGen3Modifiers(baseDamage, attacker, defender, move, field, &factors)

	// Calculate damage rolls (85-100%)
	// Gen 3 uses the same 85-100% random factor
	damages := AllDamageRolls(baseDamage)

	return damages, factors
}

// applyAttackModifiersGen3 applies Gen 3 attack stat modifiers
func (c *Calculator) applyAttackModifiersGen3(attack int, attacker *models.BattlePokemon, isPhysical bool, field *models.Field, factors *[]string) int {
	// Choice Band (Gen 3 item)
	if isPhysical && attacker.HasItem("choiceband") {
		attack = ApplyModifier(attack, 6144) // 1.5x
		*factors = append(*factors, "Choice Band")
	}

	// Huge Power / Pure Power
	if isPhysical && (attacker.HasAbility("hugepower") || attacker.HasAbility("purepower")) {
		attack = ApplyModifier(attack, 8192) // 2.0x
		*factors = append(*factors, "Huge Power")
	}

	// Guts
	if isPhysical && attacker.HasAbility("guts") && attacker.Status != "" {
		attack = ApplyModifier(attack, 6144) // 1.5x
		*factors = append(*factors, "Guts")
	}

	// Hustle
	if isPhysical && attacker.HasAbility("hustle") {
		attack = ApplyModifier(attack, 6144) // 1.5x
		*factors = append(*factors, "Hustle")
	}

	return attack
}

// applyDefenseModifiersGen3 applies Gen 3 defense stat modifiers
func (c *Calculator) applyDefenseModifiersGen3(defense int, defender *models.BattlePokemon, isPhysical bool, field *models.Field, factors *[]string) int {
	// Marvel Scale
	if isPhysical && defender.HasAbility("marvelscale") && defender.Status != "" {
		defense = ApplyModifier(defense, 6144) // 1.5x
		*factors = append(*factors, "Marvel Scale")
	}

	// Sandstorm SpD boost for Rock types (Gen 4+ only, not Gen 3)
	// Don't apply in Gen 3

	return defense
}

// applyGen3Modifiers applies modifiers in Gen 3 order
func (c *Calculator) applyGen3Modifiers(damage int, attacker, defender *models.BattlePokemon, move *models.BattleMove, field *models.Field, factors *[]string) int {
	isPhysical := data.IsPhysicalInGen3(move.GetType())

	// Gen 3 modifier order:
	// 1. Burn (physical only, 0.5x)
	// 2. Screens (Reflect/Light Screen)
	// 3. Weather
	// 4. Flash Fire (if applicable)
	// 5. Critical hit
	// 6. STAB
	// 7. Type effectiveness
	// Note: Gen 3 doesn't chain these as 4096-based modifiers

	// 1. Burn
	if attacker.IsBurned() && isPhysical && !attacker.HasAbility("guts") {
		damage = FloorDiv(damage, 2)
		*factors = append(*factors, "Burn")
	}

	// 2. Screens
	if !move.IsCrit {
		if isPhysical && field.DefenderSide.Reflect {
			if field.IsDoubles {
				damage = FloorDiv(damage*2, 3)
			} else {
				damage = FloorDiv(damage, 2)
			}
			*factors = append(*factors, "Reflect")
		}
		if !isPhysical && field.DefenderSide.LightScreen {
			if field.IsDoubles {
				damage = FloorDiv(damage*2, 3)
			} else {
				damage = FloorDiv(damage, 2)
			}
			*factors = append(*factors, "Light Screen")
		}
	}

	// 3. Weather
	moveType := move.GetType()
	if field.IsSun() {
		if moveType == "Fire" {
			damage = FloorDiv(damage*3, 2) // 1.5x
			*factors = append(*factors, "Sun (Fire boost)")
		}
		if moveType == "Water" {
			damage = FloorDiv(damage, 2) // 0.5x
			*factors = append(*factors, "Sun (Water nerf)")
		}
	}
	if field.IsRain() {
		if moveType == "Water" {
			damage = FloorDiv(damage*3, 2) // 1.5x
			*factors = append(*factors, "Rain (Water boost)")
		}
		if moveType == "Fire" {
			damage = FloorDiv(damage, 2) // 0.5x
			*factors = append(*factors, "Rain (Fire nerf)")
		}
	}

	// 4. Flash Fire
	if attacker.HasAbility("flashfire") && moveType == "Fire" && attacker.HasVolatile("flashfire") {
		damage = FloorDiv(damage*3, 2) // 1.5x
		*factors = append(*factors, "Flash Fire")
	}

	// 4b. Pinch abilities (Torrent, Blaze, Overgrow, Swarm) - activate at 1/3 HP or less
	if attacker.GetCurrentHPPercent() <= 33.33 {
		if attacker.HasAbility("torrent") && moveType == "Water" {
			damage = FloorDiv(damage*3, 2) // 1.5x
			*factors = append(*factors, "Torrent")
		}
		if attacker.HasAbility("blaze") && moveType == "Fire" {
			damage = FloorDiv(damage*3, 2) // 1.5x
			*factors = append(*factors, "Blaze")
		}
		if attacker.HasAbility("overgrow") && moveType == "Grass" {
			damage = FloorDiv(damage*3, 2) // 1.5x
			*factors = append(*factors, "Overgrow")
		}
		if attacker.HasAbility("swarm") && moveType == "Bug" {
			damage = FloorDiv(damage*3, 2) // 1.5x
			*factors = append(*factors, "Swarm")
		}
	}

	// 5. Critical hit (2x in Gen 3, not 1.5x)
	if move.IsCrit || move.WillCrit() {
		damage *= 2
		*factors = append(*factors, "Critical hit (2x)")
	}

	// 6. STAB
	if c.hasSTAB(attacker, move) {
		damage = FloorDiv(damage*3, 2) // 1.5x
		*factors = append(*factors, "STAB")
	}

	// 7. Type effectiveness
	typeEff := c.getTypeEffectiveness(move, defender)
	if typeEff == 0 {
		*factors = append(*factors, "Immune")
		return 0
	}
	if typeEff == 4 {
		damage *= 4
		*factors = append(*factors, "Super effective (4x)")
	} else if typeEff == 2 {
		damage *= 2
		*factors = append(*factors, "Super effective")
	} else if typeEff == 0.5 {
		damage = FloorDiv(damage, 2)
		*factors = append(*factors, "Not very effective")
	} else if typeEff == 0.25 {
		damage = FloorDiv(damage, 4)
		*factors = append(*factors, "Not very effective (0.25x)")
	}

	// Gen 3 item modifiers (applied after type effectiveness)
	// Type-boosting items
	if attacker.ItemData != nil {
		if boostedType := attacker.ItemData.GetTypeBoost(); boostedType != "" && boostedType == moveType {
			damage = FloorDiv(damage*11, 10) // 1.1x in Gen 3 (not 1.2x like later gens)
			*factors = append(*factors, attacker.ItemData.Name)
		}
	}

	return damage
}

// Modifier constants specific to Gen 3
const (
	ModChoiceBand    = 6144 // 1.5x
	ModChoiceSpecs   = 6144 // 1.5x (doesn't exist in Gen 3 but placeholder)
	ModCritGen3      = 8192 // 2.0x (critical hits are 2x in Gen 1-5)
	ModThickFat      = 2048 // 0.5x
	ModPunkRockDef   = 2048 // Not in Gen 3
)
