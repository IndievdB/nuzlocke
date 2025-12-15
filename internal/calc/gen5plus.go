package calc

import (
	"nuzlocke/internal/models"
)

// calculateGen5Plus calculates damage using Gen 5+ mechanics
// Move category determines physical/special split
func (c *Calculator) calculateGen5Plus(req *CalculateRequest) ([]int, []string) {
	attacker := req.Attacker
	defender := req.Defender
	move := req.Move
	field := req.Field

	var factors []string

	// Get base power
	basePower := c.getBasePower(attacker, defender, move, field)
	if basePower == 0 {
		return []int{0}, factors
	}

	// Get attack and defense stats
	attack, atkName := c.getAttackStat(attacker, move, field)
	defense, defName := c.getDefenseStat(defender, move, field)

	// Apply attack stat modifiers
	attack = c.applyAttackModifiers(attack, attacker, defender, move, field, &factors)

	// Apply defense stat modifiers
	defense = c.applyDefenseModifiers(defense, attacker, defender, move, field, &factors)

	factors = append(factors, atkName+"/"+defName)

	// Base damage calculation
	// ((2 * Level / 5 + 2) * BasePower * Attack / Defense) / 50 + 2
	level := attacker.Level
	levelFactor := FloorDiv(2*level, 5) + 2
	baseDamage := FloorDiv(levelFactor*basePower*attack, defense)
	baseDamage = FloorDiv(baseDamage, 50) + 2

	// Build modifier chain
	modChain := c.buildModifierChainGen5Plus(attacker, defender, move, field, &factors)

	// Apply modifiers
	baseDamage = modChain.Apply(baseDamage)

	// Calculate damage rolls (85-100%)
	damages := AllDamageRolls(baseDamage)

	return damages, factors
}

// applyAttackModifiers applies modifiers to the attack stat
func (c *Calculator) applyAttackModifiers(attack int, attacker, defender *models.BattlePokemon, move *models.BattleMove, field *models.Field, factors *[]string) int {
	// Choice Band / Choice Specs
	if move.IsPhysical() && attacker.HasItem("choiceband") {
		attack = ApplyModifier(attack, ModChoiceBand)
		*factors = append(*factors, "Choice Band")
	}
	if move.IsSpecial() && attacker.HasItem("choicespecs") {
		attack = ApplyModifier(attack, ModChoiceSpecs)
		*factors = append(*factors, "Choice Specs")
	}

	// Huge Power / Pure Power
	if move.IsPhysical() && (attacker.HasAbility("hugepower") || attacker.HasAbility("purepower")) {
		attack = ApplyModifier(attack, ModHugePower)
		*factors = append(*factors, "Huge Power")
	}

	// Guts
	if move.IsPhysical() && attacker.HasAbility("guts") && attacker.Status != "" {
		attack = ApplyModifier(attack, ModGuts)
		*factors = append(*factors, "Guts")
	}

	// Hustle
	if move.IsPhysical() && attacker.HasAbility("hustle") {
		attack = ApplyModifier(attack, ModHustle)
		*factors = append(*factors, "Hustle")
	}

	// Flower Gift (in sun)
	if move.IsPhysical() && attacker.HasAbility("flowergift") && field.IsSun() {
		attack = ApplyModifier(attack, ModFlowerGift)
		*factors = append(*factors, "Flower Gift")
	}

	// Solar Power (in sun, SpA)
	if move.IsSpecial() && attacker.HasAbility("solarpower") && field.IsSun() {
		attack = ApplyModifier(attack, ModFlowerGift) // Same 1.5x
		*factors = append(*factors, "Solar Power")
	}

	// Gorilla Tactics
	if move.IsPhysical() && attacker.HasAbility("gorillatactics") {
		attack = ApplyModifier(attack, ModFlowerGift) // 1.5x
		*factors = append(*factors, "Gorilla Tactics")
	}

	return attack
}

// applyDefenseModifiers applies modifiers to the defense stat
func (c *Calculator) applyDefenseModifiers(defense int, attacker, defender *models.BattlePokemon, move *models.BattleMove, field *models.Field, factors *[]string) int {
	// Assault Vest (1.5x SpD)
	if move.IsSpecial() && defender.HasItem("assaultvest") {
		defense = ApplyModifier(defense, 6144) // 1.5x
		*factors = append(*factors, "Assault Vest")
	}

	// Eviolite (1.5x Def and SpD for NFE Pokemon)
	if defender.HasItem("eviolite") {
		defense = ApplyModifier(defense, 6144) // 1.5x
		*factors = append(*factors, "Eviolite")
	}

	// Fur Coat (2x Def)
	if move.IsPhysical() && defender.HasAbility("furcoat") {
		defense = ApplyModifier(defense, 8192) // 2x
		*factors = append(*factors, "Fur Coat")
	}

	// Marvel Scale (1.5x Def when statused)
	if move.IsPhysical() && defender.HasAbility("marvelscale") && defender.Status != "" {
		defense = ApplyModifier(defense, 6144) // 1.5x
		*factors = append(*factors, "Marvel Scale")
	}

	// Grass Pelt (1.5x Def in Grassy Terrain)
	if move.IsPhysical() && defender.HasAbility("grasspelt") && field.IsGrassyTerrain() {
		defense = ApplyModifier(defense, 6144) // 1.5x
		*factors = append(*factors, "Grass Pelt")
	}

	// Sandstorm SpD boost for Rock types
	if move.IsSpecial() && field.IsSand() && defender.HasType("Rock") {
		defense = ApplyModifier(defense, 6144) // 1.5x
		*factors = append(*factors, "Sandstorm SpD boost")
	}

	return defense
}

// buildModifierChainGen5Plus builds the modifier chain for Gen 5+ damage calculation
func (c *Calculator) buildModifierChainGen5Plus(attacker, defender *models.BattlePokemon, move *models.BattleMove, field *models.Field, factors *[]string) *ModifierChain {
	chain := NewModifierChain()

	// Spread move modifier (doubles)
	if field.IsDoubles && move.HitsMultiple {
		chain.Add(ModSpread, "spread move")
		*factors = append(*factors, "Spread move")
	}

	// Weather
	c.applyWeatherModifiers(chain, move, field, factors)

	// Critical hit
	if move.IsCrit || move.WillCrit() {
		chain.Add(ModCritGen6Plus, "critical hit")
		*factors = append(*factors, "Critical hit")
	}

	// Random factor is applied later via damage rolls

	// STAB
	if c.hasSTAB(attacker, move) {
		if attacker.HasAbility("adaptability") {
			chain.Add(ModAdaptability, "Adaptability STAB")
			*factors = append(*factors, "Adaptability")
		} else {
			chain.Add(ModSTAB, "STAB")
			*factors = append(*factors, "STAB")
		}
	}

	// Type effectiveness
	typeEff := c.getTypeEffectiveness(move, defender)
	if typeEff != 1.0 {
		chain.Add(TypeEffectivenessModifier(typeEff), "type effectiveness")
		if typeEff > 1 {
			*factors = append(*factors, "Super effective")
		} else if typeEff < 1 && typeEff > 0 {
			*factors = append(*factors, "Not very effective")
		} else if typeEff == 0 {
			*factors = append(*factors, "Immune")
		}
	}

	// Burn
	if attacker.IsBurned() && move.IsPhysical() && !attacker.HasAbility("guts") {
		chain.Add(ModBurn, "burn")
		*factors = append(*factors, "Burn")
	}

	// Screens
	if !move.IsCrit {
		if move.IsPhysical() && field.DefenderSide.Reflect {
			if field.IsDoubles {
				chain.Add(ModReflectDouble, "Reflect (doubles)")
			} else {
				chain.Add(ModReflect, "Reflect")
			}
			*factors = append(*factors, "Reflect")
		}
		if move.IsSpecial() && field.DefenderSide.LightScreen {
			if field.IsDoubles {
				chain.Add(ModLightScreenDouble, "Light Screen (doubles)")
			} else {
				chain.Add(ModLightScreen, "Light Screen")
			}
			*factors = append(*factors, "Light Screen")
		}
		if field.DefenderSide.AuroraVeil {
			if field.IsDoubles {
				chain.Add(ModReflectDouble, "Aurora Veil (doubles)")
			} else {
				chain.Add(ModReflect, "Aurora Veil")
			}
			*factors = append(*factors, "Aurora Veil")
		}
	}

	// Item modifiers
	c.applyItemModifiers(chain, attacker, defender, move, typeEff, factors)

	// Ability modifiers
	c.applyAbilityModifiers(chain, attacker, defender, move, typeEff, field, factors)

	// Misty Terrain halves Dragon damage
	if field.IsMistyTerrain() && move.GetType() == "Dragon" {
		chain.Add(ModMistyHalve, "Misty Terrain")
		*factors = append(*factors, "Misty Terrain")
	}

	// Friend Guard (ally ability in doubles)
	if field.IsDoubles && field.DefenderSide.FriendGuard {
		chain.Add(3072, "Friend Guard") // 0.75x
		*factors = append(*factors, "Friend Guard")
	}

	return chain
}

// applyWeatherModifiers adds weather-based damage modifiers
func (c *Calculator) applyWeatherModifiers(chain *ModifierChain, move *models.BattleMove, field *models.Field, factors *[]string) {
	moveType := move.GetType()

	// Sun
	if field.IsSun() {
		if moveType == "Fire" {
			chain.Add(ModWeatherBoost, "Sun boost")
			*factors = append(*factors, "Sun (Fire boost)")
		}
		if moveType == "Water" {
			chain.Add(ModWeatherNerf, "Sun nerf")
			*factors = append(*factors, "Sun (Water nerf)")
		}
	}

	// Rain
	if field.IsRain() {
		if moveType == "Water" {
			chain.Add(ModWeatherBoost, "Rain boost")
			*factors = append(*factors, "Rain (Water boost)")
		}
		if moveType == "Fire" {
			chain.Add(ModWeatherNerf, "Rain nerf")
			*factors = append(*factors, "Rain (Fire nerf)")
		}
	}

	// Strong Winds (reduces SE against Flying)
	if field.Weather == "strongwinds" {
		// Would need defender type check - simplified
	}
}

// applyItemModifiers adds item-based damage modifiers
func (c *Calculator) applyItemModifiers(chain *ModifierChain, attacker, defender *models.BattlePokemon, move *models.BattleMove, typeEff float64, factors *[]string) {
	// Life Orb
	if attacker.HasItem("lifeorb") && !attacker.HasAbility("sheerforce") {
		chain.Add(ModLifeOrb, "Life Orb")
		*factors = append(*factors, "Life Orb")
	}

	// Expert Belt (SE moves)
	if attacker.HasItem("expertbelt") && typeEff > 1 {
		chain.Add(ModExpertBelt, "Expert Belt")
		*factors = append(*factors, "Expert Belt")
	}

	// Type-boosting items
	if attacker.ItemData != nil {
		if boostedType := attacker.ItemData.GetTypeBoost(); boostedType != "" && boostedType == move.GetType() {
			chain.Add(ModTypeBoost, attacker.ItemData.Name)
			*factors = append(*factors, attacker.ItemData.Name)
		}
	}

	// Muscle Band (physical)
	if attacker.HasItem("muscleband") && move.IsPhysical() {
		chain.Add(4505, "Muscle Band") // 1.1x
		*factors = append(*factors, "Muscle Band")
	}

	// Wise Glasses (special)
	if attacker.HasItem("wiseglasses") && move.IsSpecial() {
		chain.Add(4505, "Wise Glasses") // 1.1x
		*factors = append(*factors, "Wise Glasses")
	}
}

// applyAbilityModifiers adds ability-based damage modifiers
func (c *Calculator) applyAbilityModifiers(chain *ModifierChain, attacker, defender *models.BattlePokemon, move *models.BattleMove, typeEff float64, field *models.Field, factors *[]string) {
	// Offensive abilities

	// Sheer Force (moves with secondary effects)
	if attacker.HasAbility("sheerforce") && move.HasSecondaryEffect() {
		chain.Add(ModSheerForce, "Sheer Force")
		*factors = append(*factors, "Sheer Force")
	}

	// Iron Fist (punch moves)
	if attacker.HasAbility("ironfist") && move.IsPunchMove() {
		chain.Add(ModIronFist, "Iron Fist")
		*factors = append(*factors, "Iron Fist")
	}

	// Reckless (recoil moves)
	if attacker.HasAbility("reckless") && move.IsRecoilMove() {
		chain.Add(ModReckless, "Reckless")
		*factors = append(*factors, "Reckless")
	}

	// Tough Claws (contact moves)
	if attacker.HasAbility("toughclaws") && move.IsContactMove() {
		chain.Add(ModToughClaws, "Tough Claws")
		*factors = append(*factors, "Tough Claws")
	}

	// Strong Jaw (biting moves)
	if attacker.HasAbility("strongjaw") && move.IsBiteMove() {
		chain.Add(ModStrongJaw, "Strong Jaw")
		*factors = append(*factors, "Strong Jaw")
	}

	// Mega Launcher (pulse/aura moves)
	if attacker.HasAbility("megalauncher") && move.HasFlag("pulse") {
		chain.Add(ModMegaLauncher, "Mega Launcher")
		*factors = append(*factors, "Mega Launcher")
	}

	// Sand Force (in sandstorm)
	if attacker.HasAbility("sandforce") && field.IsSand() {
		moveType := move.GetType()
		if moveType == "Ground" || moveType == "Rock" || moveType == "Steel" {
			chain.Add(ModSandForce, "Sand Force")
			*factors = append(*factors, "Sand Force")
		}
	}

	// Defensive abilities

	// Filter / Solid Rock / Prism Armor (reduce SE damage)
	if typeEff > 1 {
		if defender.HasAbility("filter") || defender.HasAbility("solidrock") || defender.HasAbility("prismarmor") {
			chain.Add(ModFilter, "Filter/Solid Rock")
			*factors = append(*factors, "Filter/Solid Rock")
		}
	}

	// Multiscale / Shadow Shield (at full HP)
	if defender.IsAtFullHP() {
		if defender.HasAbility("multiscale") || defender.HasAbility("shadowshield") {
			chain.Add(ModMultiscale, "Multiscale")
			*factors = append(*factors, "Multiscale")
		}
	}

	// Ice Scales (reduces special damage)
	if move.IsSpecial() && defender.HasAbility("icescales") {
		chain.Add(ModIceScales, "Ice Scales")
		*factors = append(*factors, "Ice Scales")
	}

	// Fluffy
	if defender.HasAbility("fluffy") {
		if move.IsContactMove() {
			chain.Add(ModFluffyContact, "Fluffy (contact)")
			*factors = append(*factors, "Fluffy (contact)")
		}
		if move.GetType() == "Fire" {
			chain.Add(ModFluffyFire, "Fluffy (Fire)")
			*factors = append(*factors, "Fluffy (Fire)")
		}
	}

	// Punk Rock (reduces sound damage)
	if move.IsSoundMove() && defender.HasAbility("punkrock") {
		chain.Add(ModPunkRockDef, "Punk Rock (defense)")
		*factors = append(*factors, "Punk Rock (defense)")
	}

	// Thick Fat (reduces Fire and Ice)
	if defender.HasAbility("thickfat") {
		moveType := move.GetType()
		if moveType == "Fire" || moveType == "Ice" {
			chain.Add(ModThickFat, "Thick Fat")
			*factors = append(*factors, "Thick Fat")
		}
	}
}
