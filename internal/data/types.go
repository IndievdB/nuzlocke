package data

// TypeData represents type effectiveness data
type TypeData struct {
	DamageTaken map[string]int `json:"damageTaken"`
	HPivs       map[string]int `json:"HPivs,omitempty"`
	HPdvs       map[string]int `json:"HPdvs,omitempty"`
}

// Type effectiveness values in damageTaken
// 0 = neutral (1x)
// 1 = super effective (2x) - this type takes 2x damage from that type
// 2 = not very effective (0.5x) - this type takes 0.5x damage from that type
// 3 = immune (0x) - this type is immune to that type

// TypeEffectiveness represents the effectiveness multiplier
type TypeEffectiveness int

const (
	TypeNeutral        TypeEffectiveness = 0
	TypeSuperEffective TypeEffectiveness = 1
	TypeResisted       TypeEffectiveness = 2
	TypeImmune         TypeEffectiveness = 3
)

// GetMultiplier returns the damage multiplier for this effectiveness
func (te TypeEffectiveness) GetMultiplier() float64 {
	switch te {
	case TypeSuperEffective:
		return 2.0
	case TypeResisted:
		return 0.5
	case TypeImmune:
		return 0.0
	default:
		return 1.0
	}
}

// Get4096Multiplier returns the 4096-based multiplier for this effectiveness
func (te TypeEffectiveness) Get4096Multiplier() int {
	switch te {
	case TypeSuperEffective:
		return 8192 // 2.0x
	case TypeResisted:
		return 2048 // 0.5x
	case TypeImmune:
		return 0
	default:
		return 4096 // 1.0x
	}
}

// Gen 3 physical types (for type-based physical/special split)
var Gen3PhysicalTypes = map[string]bool{
	"Normal":   true,
	"Fighting": true,
	"Flying":   true,
	"Ground":   true,
	"Rock":     true,
	"Bug":      true,
	"Ghost":    true,
	"Poison":   true,
	"Steel":    true,
}

// IsPhysicalInGen3 returns true if the type uses physical stats in Gen 3
func IsPhysicalInGen3(typeName string) bool {
	return Gen3PhysicalTypes[typeName]
}

// All Pokemon types
var AllTypes = []string{
	"Normal", "Fire", "Water", "Electric", "Grass", "Ice",
	"Fighting", "Poison", "Ground", "Flying", "Psychic", "Bug",
	"Rock", "Ghost", "Dragon", "Dark", "Steel", "Fairy",
}
