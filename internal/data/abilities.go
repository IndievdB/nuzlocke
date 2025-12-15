package data

// Ability represents a Pokemon ability
type Ability struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	Num                  int        `json:"num"`
	Rating               float64    `json:"rating,omitempty"`
	Desc                 string     `json:"desc,omitempty"`
	ShortDesc            string     `json:"shortDesc,omitempty"`
	OnModifyDamage       bool       `json:"onModifyDamage,omitempty"`
	OnModifyAtk          bool       `json:"onModifyAtk,omitempty"`
	OnModifyDef          bool       `json:"onModifyDef,omitempty"`
	OnModifySpA          bool       `json:"onModifySpA,omitempty"`
	OnModifySpD          bool       `json:"onModifySpD,omitempty"`
	OnModifySpe          bool       `json:"onModifySpe,omitempty"`
	OnBasePower          bool       `json:"onBasePower,omitempty"`
	OnSourceModifyDamage bool       `json:"onSourceModifyDamage,omitempty"`
	OnSourceBasePower    bool       `json:"onSourceBasePower,omitempty"`
	OnModifySTAB         bool       `json:"onModifySTAB,omitempty"`
	SuppressWeather      bool       `json:"suppressWeather,omitempty"`
	OnImmunity           bool       `json:"onImmunity,omitempty"`
	OnModifyType         bool       `json:"onModifyType,omitempty"`
	Modifiers            []Modifier `json:"modifiers,omitempty"`
}

// Modifier represents a damage modifier from an ability
type Modifier struct {
	Numerator   int `json:"numerator"`
	Denominator int `json:"denominator"`
}

// Ability modifier constants (4096 base)
const (
	AbilityModBase         = 4096
	AbilityModHugePower    = 8192 // 2.0x
	AbilityModPurePower    = 8192 // 2.0x
	AbilityModGuts         = 6144 // 1.5x
	AbilityModHustle       = 6144 // 1.5x
	AbilityModFlowerGift   = 6144 // 1.5x
	AbilityModTechnician   = 6144 // 1.5x for BP <= 60
	AbilityModIronFist     = 4915 // 1.2x
	AbilityModReckless     = 4915 // 1.2x
	AbilityModSheerForce   = 5325 // 1.3x
	AbilityModAnalytic     = 5325 // 1.3x
	AbilityModSandForce    = 5325 // 1.3x
	AbilityModToughClaws   = 5325 // 1.3x
	AbilityModAerilate     = 4915 // 1.2x
	AbilityModPixilate     = 4915 // 1.2x
	AbilityModRefrigerate  = 4915 // 1.2x
	AbilityModGalvanize    = 4915 // 1.2x
	AbilityModMegaLauncher = 6144 // 1.5x
	AbilityModStrongJaw    = 6144 // 1.5x
	AbilityModPunkRock     = 5325 // 1.3x
	AbilityModSteelySpirit = 6144 // 1.5x

	// Defensive abilities
	AbilityModThickFat     = 2048 // 0.5x
	AbilityModFilter       = 3072 // 0.75x
	AbilityModSolidRock    = 3072 // 0.75x
	AbilityModPrismArmor   = 3072 // 0.75x
	AbilityModMultiscale   = 2048 // 0.5x at full HP
	AbilityModShadowShield = 2048 // 0.5x at full HP
	AbilityModFurCoat      = 2048 // 0.5x (doubles def)
	AbilityModFluffyContact = 2048 // 0.5x contact
	AbilityModFluffyFire   = 8192 // 2.0x fire
	AbilityModIceScales    = 2048 // 0.5x special
	AbilityModPunkRockDef  = 2048 // 0.5x sound
)

// Stat-boosting abilities mapping
var StatBoostAbilities = map[string]struct {
	Stat       string
	Multiplier int
}{
	"hugepower":    {"atk", AbilityModHugePower},
	"purepower":    {"atk", AbilityModPurePower},
	"hustle":       {"atk", AbilityModHustle},
	"guts":         {"atk", AbilityModGuts},
	"gorillaTactics": {"atk", 6144},
	"swordsoftheruin": {"atk", 5325},
	"tabletsofruin": {"def", 3072}, // opponent's def
	"vesselofruin": {"spa", 5325},
	"beadsofruin": {"spd", 3072}, // opponent's spd
}

// Type immunity abilities
var TypeImmunityAbilities = map[string]string{
	"levitate":     "Ground",
	"voltabsorb":   "Electric",
	"lightningrod": "Electric",
	"motordrive":   "Electric",
	"waterabsorb":  "Water",
	"stormdrain":   "Water",
	"dryskin":      "Water",
	"flashfire":    "Fire",
	"sapsipper":    "Grass",
	"eartheater":   "Ground",
}

// Weather abilities
var WeatherAbilities = map[string]string{
	"drought":      "sun",
	"drizzle":      "rain",
	"sandstream":   "sand",
	"snowwarning":  "snow",
	"primordialsea": "heavyrain",
	"desolateland": "harshsun",
	"deltastream":  "strongwinds",
}
