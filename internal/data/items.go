package data

// Item represents a held item
type Item struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Num              int             `json:"num"`
	Gen              int             `json:"gen,omitempty"`
	Desc             string          `json:"desc,omitempty"`
	FlingBasePower   int             `json:"flingBasePower,omitempty"`
	NaturalGift      *NaturalGift    `json:"naturalGift,omitempty"`
	Boosts           map[string]int  `json:"boosts,omitempty"`
	OnModifyDamage   bool            `json:"onModifyDamage,omitempty"`
	OnModifyAtk      bool            `json:"onModifyAtk,omitempty"`
	OnModifyDef      bool            `json:"onModifyDef,omitempty"`
	OnModifySpA      bool            `json:"onModifySpA,omitempty"`
	OnModifySpD      bool            `json:"onModifySpD,omitempty"`
	OnModifySpe      bool            `json:"onModifySpe,omitempty"`
	OnBasePower      bool            `json:"onBasePower,omitempty"`
}

// NaturalGift represents the Natural Gift move data for berries
type NaturalGift struct {
	BasePower int    `json:"basePower"`
	Type      string `json:"type"`
}

// Item modifier constants (4096 base)
const (
	ItemModBase          = 4096
	ItemModLifeOrb       = 5324  // ~1.3x
	ItemModExpertBelt    = 4915  // 1.2x
	ItemModTypeBoost     = 4915  // 1.2x (Charcoal, Mystic Water, etc.)
	ItemModChoiceBand    = 6144  // 1.5x (applied to stat, not damage)
	ItemModChoiceSpecs   = 6144  // 1.5x
	ItemModMuscleBand    = 4505  // 1.1x
	ItemModWiseGlasses   = 4505  // 1.1x
	ItemModMetronome     = 4096  // Base, increases per use
	ItemModAdamantOrb    = 4915  // 1.2x for Dialga
	ItemModLustrousOrb   = 4915  // 1.2x for Palkia
	ItemModGriseousOrb   = 4915  // 1.2x for Giratina
	ItemModSoulDew       = 4915  // 1.2x for Latios/Latias (Gen 7+)
)

// Type-boosting items mapping
var TypeBoostingItems = map[string]string{
	"charcoal":      "Fire",
	"mysticwater":   "Water",
	"miracleseed":   "Grass",
	"magnet":        "Electric",
	"nevermeltice":  "Ice",
	"blackbelt":     "Fighting",
	"poisonbarb":    "Poison",
	"softsand":      "Ground",
	"sharpbeak":     "Flying",
	"twistedspoon":  "Psychic",
	"silverpowder":  "Bug",
	"hardstone":     "Rock",
	"spelltag":      "Ghost",
	"dragonfang":    "Dragon",
	"blackglasses":  "Dark",
	"metalcoat":     "Steel",
	"silkscarf":     "Normal",
	"pixieplate":    "Fairy",
	// Plates
	"flameplate":    "Fire",
	"splashplate":   "Water",
	"meadowplate":   "Grass",
	"zapplate":      "Electric",
	"icicleplate":   "Ice",
	"fistplate":     "Fighting",
	"toxicplate":    "Poison",
	"earthplate":    "Ground",
	"skyplate":      "Flying",
	"mindplate":     "Psychic",
	"insectplate":   "Bug",
	"stoneplate":    "Rock",
	"spookyplate":   "Ghost",
	"dracoplate":    "Dragon",
	"dreadplate":    "Dark",
	"ironplate":     "Steel",
	// Incenses
	"seaincense":    "Water",
	"roseincense":   "Grass",
	"oddincense":    "Psychic",
	"rockincense":   "Rock",
	"waveincense":   "Water",
}

// GetTypeBoost returns the type this item boosts, or empty string if none
func (i *Item) GetTypeBoost() string {
	if t, ok := TypeBoostingItems[i.ID]; ok {
		return t
	}
	return ""
}
