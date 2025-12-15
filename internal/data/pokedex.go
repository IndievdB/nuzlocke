package data

// Pokemon represents a Pokemon species from the Pokedex
type Pokemon struct {
	Num         int               `json:"num"`
	Name        string            `json:"name"`
	Types       []string          `json:"types"`
	GenderRatio *GenderRatio      `json:"genderRatio,omitempty"`
	BaseStats   BaseStats         `json:"baseStats"`
	Abilities   map[string]string `json:"abilities"`
	Heightm     float64           `json:"heightm"`
	Weightkg    float64           `json:"weightkg"`
	Color       string            `json:"color"`
	Evos        []string          `json:"evos,omitempty"`
	Prevo       string            `json:"prevo,omitempty"`
	EggGroups   []string          `json:"eggGroups"`
	Tier        string            `json:"tier,omitempty"`
	BaseSpecies string            `json:"baseSpecies,omitempty"`
	Forme       string            `json:"forme,omitempty"`
	Gender      string            `json:"gender,omitempty"`
}

// GenderRatio represents the gender distribution of a Pokemon
type GenderRatio struct {
	M float64 `json:"M"`
	F float64 `json:"F"`
}

// BaseStats represents a Pokemon's base stats
type BaseStats struct {
	HP  int `json:"hp"`
	Atk int `json:"atk"`
	Def int `json:"def"`
	SpA int `json:"spa"`
	SpD int `json:"spd"`
	Spe int `json:"spe"`
}

// GetAbility returns the ability at the given slot (0, 1, or H for hidden)
func (p *Pokemon) GetAbility(slot string) string {
	if ability, ok := p.Abilities[slot]; ok {
		return ability
	}
	return ""
}

// HasType checks if the Pokemon has the given type
func (p *Pokemon) HasType(t string) bool {
	for _, pt := range p.Types {
		if pt == t {
			return true
		}
	}
	return false
}
