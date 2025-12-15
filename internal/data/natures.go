package data

// Nature represents a Pokemon nature and its stat modifiers
type Nature struct {
	Name  string `json:"name"`
	Plus  string `json:"plus,omitempty"`  // Stat that gets +10%
	Minus string `json:"minus,omitempty"` // Stat that gets -10%
}

// IsNeutral returns true if the nature has no stat modifications
func (n *Nature) IsNeutral() bool {
	return n.Plus == "" && n.Minus == ""
}

// GetStatMultiplier returns the multiplier for a given stat (1.1, 1.0, or 0.9)
func (n *Nature) GetStatMultiplier(stat string) float64 {
	if n.Plus == stat {
		return 1.1
	}
	if n.Minus == stat {
		return 0.9
	}
	return 1.0
}

// Nature modifier constants (for integer math)
const (
	NaturePlus  = 11 // Multiply by 11, divide by 10
	NatureNeutral = 10
	NatureMinus = 9
)

// GetStatModifier returns the integer modifier for a stat (11, 10, or 9)
// Use: stat = stat * modifier / 10
func (n *Nature) GetStatModifier(stat string) int {
	if n.Plus == stat {
		return NaturePlus
	}
	if n.Minus == stat {
		return NatureMinus
	}
	return NatureNeutral
}

// All nature names for validation
var AllNatures = []string{
	"Adamant", "Bashful", "Bold", "Brave", "Calm",
	"Careful", "Docile", "Gentle", "Hardy", "Hasty",
	"Impish", "Jolly", "Lax", "Lonely", "Mild",
	"Modest", "Naive", "Naughty", "Quiet", "Quirky",
	"Rash", "Relaxed", "Sassy", "Serious", "Timid",
}
