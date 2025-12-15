package models

// Field represents the battle field conditions
type Field struct {
	// Weather
	Weather string `json:"weather,omitempty"` // sun, rain, sand, snow, hail, harshsun, heavyrain, strongwinds

	// Terrain
	Terrain string `json:"terrain,omitempty"` // electric, grassy, misty, psychic

	// Game type
	IsDoubles bool `json:"isDoubles,omitempty"`

	// Side conditions
	AttackerSide SideConditions `json:"attackerSide"`
	DefenderSide SideConditions `json:"defenderSide"`

	// Gravity, Magic Room, Wonder Room
	Gravity    bool `json:"gravity,omitempty"`
	MagicRoom  bool `json:"magicRoom,omitempty"`
	WonderRoom bool `json:"wonderRoom,omitempty"`

	// Generation-specific settings
	Generation int `json:"generation,omitempty"` // 3 = Gen 3 mechanics, 5+ = Gen 5+ mechanics
}

// SideConditions represents conditions on one side of the field
type SideConditions struct {
	// Screens
	Reflect     bool `json:"reflect,omitempty"`
	LightScreen bool `json:"lightScreen,omitempty"`
	AuroraVeil  bool `json:"auroraVeil,omitempty"`

	// Entry hazards
	Spikes      int  `json:"spikes,omitempty"`      // 0-3 layers
	StealthRock bool `json:"stealthRock,omitempty"`
	ToxicSpikes int  `json:"toxicSpikes,omitempty"` // 0-2 layers
	StickyWeb   bool `json:"stickyWeb,omitempty"`

	// Other
	Tailwind       bool `json:"tailwind,omitempty"`
	HelpingHand    bool `json:"helpingHand,omitempty"`
	FriendGuard    bool `json:"friendGuard,omitempty"`
	Battery        bool `json:"battery,omitempty"`
	PowerSpot      bool `json:"powerSpot,omitempty"`
}

// NewField creates a new field with default values
func NewField() *Field {
	return &Field{
		Generation:   9, // Default to current gen
		AttackerSide: SideConditions{},
		DefenderSide: SideConditions{},
	}
}

// IsGen3 returns true if using Gen 3 mechanics
func (f *Field) IsGen3() bool {
	return f.Generation == 3
}

// IsGen5Plus returns true if using Gen 5+ mechanics
func (f *Field) IsGen5Plus() bool {
	return f.Generation >= 5
}

// HasWeather returns true if any weather is active
func (f *Field) HasWeather() bool {
	return f.Weather != ""
}

// IsSun returns true if sun/harsh sun is active
func (f *Field) IsSun() bool {
	return f.Weather == "sun" || f.Weather == "harshsun"
}

// IsRain returns true if rain/heavy rain is active
func (f *Field) IsRain() bool {
	return f.Weather == "rain" || f.Weather == "heavyrain"
}

// IsSand returns true if sandstorm is active
func (f *Field) IsSand() bool {
	return f.Weather == "sand"
}

// IsSnow returns true if snow/hail is active
func (f *Field) IsSnow() bool {
	return f.Weather == "snow" || f.Weather == "hail"
}

// HasTerrain returns true if any terrain is active
func (f *Field) HasTerrain() bool {
	return f.Terrain != ""
}

// IsElectricTerrain returns true if electric terrain is active
func (f *Field) IsElectricTerrain() bool {
	return f.Terrain == "electric"
}

// IsGrassyTerrain returns true if grassy terrain is active
func (f *Field) IsGrassyTerrain() bool {
	return f.Terrain == "grassy"
}

// IsMistyTerrain returns true if misty terrain is active
func (f *Field) IsMistyTerrain() bool {
	return f.Terrain == "misty"
}

// IsPsychicTerrain returns true if psychic terrain is active
func (f *Field) IsPsychicTerrain() bool {
	return f.Terrain == "psychic"
}

// HasScreen returns true if the defender has an active screen for the given category
func (f *Field) HasScreen(category string) bool {
	if f.DefenderSide.AuroraVeil {
		return true
	}
	if category == "Physical" {
		return f.DefenderSide.Reflect
	}
	if category == "Special" {
		return f.DefenderSide.LightScreen
	}
	return false
}
