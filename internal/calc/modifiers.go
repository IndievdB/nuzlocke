package calc

// Common modifier constants (4096 base where 4096 = 1.0x)
const (
	// Base
	ModBase = 4096

	// STAB
	ModSTAB         = 6144 // 1.5x
	ModAdaptability = 8192 // 2.0x (Adaptability STAB)

	// Weather
	ModWeatherBoost = 6144 // 1.5x (Fire in Sun, Water in Rain)
	ModWeatherNerf  = 2048 // 0.5x (Fire in Rain, Water in Sun)

	// Critical hits
	ModCritGen6Plus = 6144 // 1.5x (Gen 6+)
	ModCritGen5     = 8192 // 2.0x (Gen 1-5)

	// Items
	ModLifeOrb      = 5324 // ~1.3x (5324/4096 = 1.2998)
	ModExpertBelt   = 4915 // 1.2x
	ModTypeBoost    = 4915 // 1.2x (Charcoal, etc.)
	ModMetronomeMax = 8192 // 2.0x (max metronome stack)

	// Screens
	ModReflect       = 2048 // 0.5x (singles)
	ModReflectDouble = 2732 // 2/3x (doubles)
	ModLightScreen   = 2048 // 0.5x (singles)
	ModLightScreenDouble = 2732 // 2/3x (doubles)

	// Burn
	ModBurn = 2048 // 0.5x

	// Spread moves (doubles)
	ModSpread = 3072 // 0.75x

	// Type effectiveness
	ModSuperEffective = 8192 // 2.0x
	ModResisted       = 2048 // 0.5x
	ModImmune         = 0    // 0x

	// Abilities (offensive)
	ModHugePower   = 8192 // 2.0x
	ModPurePower   = 8192 // 2.0x
	ModGuts        = 6144 // 1.5x
	ModHustle      = 6144 // 1.5x
	ModFlowerGift  = 6144 // 1.5x
	ModTechnician  = 6144 // 1.5x (for BP <= 60)
	ModPinch       = 6144 // 1.5x (Torrent, Blaze, Overgrow, Swarm when HP <= 1/3)
	ModIronFist    = 4915 // 1.2x
	ModReckless    = 4915 // 1.2x
	ModSheerForce  = 5325 // 1.3x
	ModAnalytic    = 5325 // 1.3x
	ModSandForce   = 5325 // 1.3x
	ModToughClaws  = 5325 // 1.3x
	ModAte         = 4915 // 1.2x (Aerilate, Pixilate, etc.)
	ModMegaLauncher = 6144 // 1.5x
	ModStrongJaw   = 6144 // 1.5x
	ModPunkRock    = 5325 // 1.3x
	ModSteelySpirit = 6144 // 1.5x

	// Abilities (defensive)
	ModFilter       = 3072 // 0.75x
	ModSolidRock    = 3072 // 0.75x
	ModPrismArmor   = 3072 // 0.75x
	ModMultiscale   = 2048 // 0.5x
	ModShadowShield = 2048 // 0.5x
	ModFurCoat      = 8192 // 2.0x defense (effectively 0.5x damage)
	ModIceScales    = 2048 // 0.5x (special only)
	ModFluffyContact = 2048 // 0.5x contact
	ModFluffyFire   = 8192 // 2.0x fire

	// Terrain
	ModTerrainBoost = 5325 // 1.3x (Gen 8+, was 1.5x in Gen 7)
	ModTerrainGen7  = 6144 // 1.5x
	ModMistyHalve   = 2048 // 0.5x (Misty Terrain vs Dragon)
)

// Modifier represents a single damage modifier with context
type Modifier struct {
	Value  int
	Source string // For debugging/display
}

// ModifierChain holds a list of modifiers to be chained
type ModifierChain struct {
	Modifiers []Modifier
}

// NewModifierChain creates a new empty modifier chain
func NewModifierChain() *ModifierChain {
	return &ModifierChain{
		Modifiers: make([]Modifier, 0),
	}
}

// Add adds a modifier to the chain
func (mc *ModifierChain) Add(value int, source string) {
	if value != ModBase { // Only add non-1x modifiers
		mc.Modifiers = append(mc.Modifiers, Modifier{Value: value, Source: source})
	}
}

// AddIf conditionally adds a modifier
func (mc *ModifierChain) AddIf(condition bool, value int, source string) {
	if condition {
		mc.Add(value, source)
	}
}

// Calculate calculates the final chained modifier value
func (mc *ModifierChain) Calculate() int {
	if len(mc.Modifiers) == 0 {
		return ModBase
	}

	values := make([]int, len(mc.Modifiers))
	for i, mod := range mc.Modifiers {
		values[i] = mod.Value
	}
	return ChainModifiers(values)
}

// Apply applies the chained modifier to a damage value
func (mc *ModifierChain) Apply(damage int) int {
	return ApplyChainedModifier(damage, mc.Calculate())
}

// Sources returns a list of modifier sources for debugging
func (mc *ModifierChain) Sources() []string {
	sources := make([]string, len(mc.Modifiers))
	for i, mod := range mc.Modifiers {
		sources[i] = mod.Source
	}
	return sources
}

// TypeEffectivenessModifier returns the 4096-based modifier for type effectiveness
func TypeEffectivenessModifier(multiplier float64) int {
	// Common multipliers
	switch multiplier {
	case 0:
		return ModImmune
	case 0.25:
		return 1024 // 0.25x = 1024/4096
	case 0.5:
		return ModResisted
	case 1:
		return ModBase
	case 2:
		return ModSuperEffective
	case 4:
		return 16384 // 4x = 16384/4096
	default:
		return int(multiplier * float64(ModBase))
	}
}
