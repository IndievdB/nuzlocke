package calc

import "math"

// Mod4096 is the base value for the 4096 modifier system
const Mod4096 = 4096

// FloorDiv performs integer floor division (a / b rounded toward negative infinity)
// This matches the behavior of integer division in Pokemon games
func FloorDiv(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}

// PokeRound rounds a number using Game Freak's rounding rules:
// - Rounds down at exactly 0.5 (unlike standard rounding which rounds up)
// - Otherwise rounds normally
func PokeRound(n float64) int {
	frac := n - math.Floor(n)
	if frac <= 0.5 {
		return int(math.Floor(n))
	}
	return int(math.Ceil(n))
}

// ApplyModifier applies a 4096-based modifier to a value with pokeRound
func ApplyModifier(value, modifier int) int {
	if modifier == Mod4096 {
		return value
	}
	return PokeRound(float64(value*modifier) / float64(Mod4096))
}

// ChainModifiers chains multiple 4096-based modifiers together
// Each modifier is applied with normal rounding during chaining
func ChainModifiers(modifiers []int) int {
	result := Mod4096
	for _, mod := range modifiers {
		// (result * mod + 0x800) >> 12
		// This is equivalent to: round((result * mod) / 4096)
		result = (result*mod + 0x800) >> 12
	}
	return result
}

// ApplyChainedModifier applies a chained modifier (from ChainModifiers) to a value
func ApplyChainedModifier(value, modifier int) int {
	return ApplyModifier(value, modifier)
}

// Clamp ensures a value is within [min, max]
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ApplyPercentage calculates value * percent / 100 with integer division
func ApplyPercentage(value, percent int) int {
	return FloorDiv(value*percent, 100)
}

// DamageRoll calculates damage for a specific roll (85-100)
func DamageRoll(baseDamage, roll int) int {
	damage := FloorDiv(baseDamage*roll, 100)
	if damage < 1 {
		damage = 1
	}
	return damage
}

// AllDamageRolls returns all 16 damage values (rolls 85-100)
func AllDamageRolls(baseDamage int) []int {
	damages := make([]int, 16)
	for i := 0; i < 16; i++ {
		damages[i] = DamageRoll(baseDamage, 85+i)
	}
	return damages
}
