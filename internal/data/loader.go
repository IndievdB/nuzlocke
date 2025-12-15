package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Store holds all loaded Pokemon data
type Store struct {
	Pokedex   map[string]*Pokemon
	Moves     map[string]*Move
	Items     map[string]*Item
	Abilities map[string]*Ability
	TypeChart map[string]*TypeData
	Natures   map[string]*Nature
	Learnsets map[string]*Learnset

	// Index maps for case-insensitive lookups
	pokedexIndex   map[string]string
	movesIndex     map[string]string
	itemsIndex     map[string]string
	abilitiesIndex map[string]string
	naturesIndex   map[string]string
}

// NewStore creates a new empty Store
func NewStore() *Store {
	return &Store{
		Pokedex:        make(map[string]*Pokemon),
		Moves:          make(map[string]*Move),
		Items:          make(map[string]*Item),
		Abilities:      make(map[string]*Ability),
		TypeChart:      make(map[string]*TypeData),
		Natures:        make(map[string]*Nature),
		Learnsets:      make(map[string]*Learnset),
		pokedexIndex:   make(map[string]string),
		movesIndex:     make(map[string]string),
		itemsIndex:     make(map[string]string),
		abilitiesIndex: make(map[string]string),
		naturesIndex:   make(map[string]string),
	}
}

// LoadFromDirectory loads all JSON data files from the given directory
func LoadFromDirectory(dir string) (*Store, error) {
	store := NewStore()

	// Load pokedex
	if err := store.loadPokedex(filepath.Join(dir, "pokedex.json")); err != nil {
		return nil, fmt.Errorf("loading pokedex: %w", err)
	}

	// Load moves
	if err := store.loadMoves(filepath.Join(dir, "moves.json")); err != nil {
		return nil, fmt.Errorf("loading moves: %w", err)
	}

	// Load items
	if err := store.loadItems(filepath.Join(dir, "items.json")); err != nil {
		return nil, fmt.Errorf("loading items: %w", err)
	}

	// Load abilities
	if err := store.loadAbilities(filepath.Join(dir, "abilities.json")); err != nil {
		return nil, fmt.Errorf("loading abilities: %w", err)
	}

	// Load type chart
	if err := store.loadTypeChart(filepath.Join(dir, "typechart.json")); err != nil {
		return nil, fmt.Errorf("loading typechart: %w", err)
	}

	// Load natures
	if err := store.loadNatures(filepath.Join(dir, "natures.json")); err != nil {
		return nil, fmt.Errorf("loading natures: %w", err)
	}

	// Load learnsets
	if err := store.loadLearnsets(filepath.Join(dir, "learnsets.json")); err != nil {
		return nil, fmt.Errorf("loading learnsets: %w", err)
	}

	// Load catch rates (optional, doesn't fail if missing)
	_ = store.loadCatchRates(filepath.Join(dir, "catchrates.json"))

	return store, nil
}

// ToID converts a name to a lowercase ID (removes spaces, special chars)
// Exported for use by other packages
func ToID(name string) string {
	return toID(name)
}

// toID converts a name to a lowercase ID (removes spaces, special chars)
func toID(name string) string {
	name = strings.ToLower(name)
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func (s *Store) loadPokedex(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &s.Pokedex); err != nil {
		return err
	}

	// Build index
	for id, pokemon := range s.Pokedex {
		s.pokedexIndex[toID(pokemon.Name)] = id
		s.pokedexIndex[id] = id
	}

	return nil
}

func (s *Store) loadMoves(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &s.Moves); err != nil {
		return err
	}

	// Build index
	for id, move := range s.Moves {
		s.movesIndex[toID(move.Name)] = id
		s.movesIndex[id] = id
	}

	return nil
}

func (s *Store) loadItems(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &s.Items); err != nil {
		return err
	}

	// Build index
	for id, item := range s.Items {
		s.itemsIndex[toID(item.Name)] = id
		s.itemsIndex[id] = id
	}

	return nil
}

func (s *Store) loadAbilities(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &s.Abilities); err != nil {
		return err
	}

	// Build index
	for id, ability := range s.Abilities {
		s.abilitiesIndex[toID(ability.Name)] = id
		s.abilitiesIndex[id] = id
	}

	return nil
}

func (s *Store) loadTypeChart(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.TypeChart)
}

func (s *Store) loadNatures(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &s.Natures); err != nil {
		return err
	}

	// Build index
	for id, nature := range s.Natures {
		s.naturesIndex[toID(nature.Name)] = id
		s.naturesIndex[id] = id
	}

	return nil
}

func (s *Store) loadLearnsets(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.Learnsets)
}

func (s *Store) loadCatchRates(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var catchRates map[string]int
	if err := json.Unmarshal(data, &catchRates); err != nil {
		return err
	}

	// Apply catch rates to Pokemon by their dex number
	for _, pokemon := range s.Pokedex {
		numStr := fmt.Sprintf("%d", pokemon.Num)
		if rate, ok := catchRates[numStr]; ok {
			pokemon.CatchRate = rate
		}
	}

	return nil
}

// GetPokemon returns a Pokemon by name or ID (case-insensitive)
func (s *Store) GetPokemon(nameOrID string) *Pokemon {
	id := toID(nameOrID)
	if realID, ok := s.pokedexIndex[id]; ok {
		return s.Pokedex[realID]
	}
	return nil
}

// GetMove returns a Move by name or ID (case-insensitive)
func (s *Store) GetMove(nameOrID string) *Move {
	id := toID(nameOrID)
	if realID, ok := s.movesIndex[id]; ok {
		return s.Moves[realID]
	}
	return nil
}

// GetItem returns an Item by name or ID (case-insensitive)
func (s *Store) GetItem(nameOrID string) *Item {
	id := toID(nameOrID)
	if realID, ok := s.itemsIndex[id]; ok {
		return s.Items[realID]
	}
	return nil
}

// GetAbility returns an Ability by name or ID (case-insensitive)
func (s *Store) GetAbility(nameOrID string) *Ability {
	id := toID(nameOrID)
	if realID, ok := s.abilitiesIndex[id]; ok {
		return s.Abilities[realID]
	}
	return nil
}

// GetNature returns a Nature by name or ID (case-insensitive)
func (s *Store) GetNature(nameOrID string) *Nature {
	id := toID(nameOrID)
	if realID, ok := s.naturesIndex[id]; ok {
		return s.Natures[realID]
	}
	return nil
}

// GetLearnset returns a Pokemon's learnset by species ID
func (s *Store) GetLearnset(pokemonID string) *Learnset {
	id := toID(pokemonID)
	if learnset, ok := s.Learnsets[id]; ok {
		return learnset
	}
	return nil
}

// GetTypeEffectiveness returns the effectiveness of attackType vs defenderType
// Returns: 0 = neutral, 1 = super effective, 2 = resisted, 3 = immune
func (s *Store) GetTypeEffectiveness(attackType, defenderType string) TypeEffectiveness {
	defenderID := strings.ToLower(defenderType)
	if typeData, ok := s.TypeChart[defenderID]; ok {
		if eff, ok := typeData.DamageTaken[attackType]; ok {
			return TypeEffectiveness(eff)
		}
	}
	return TypeNeutral
}

// GetTypeEffectivenessMultiple calculates effectiveness against multiple types
func (s *Store) GetTypeEffectivenessMultiple(attackType string, defenderTypes []string) float64 {
	multiplier := 1.0
	for _, defType := range defenderTypes {
		eff := s.GetTypeEffectiveness(attackType, defType)
		multiplier *= eff.GetMultiplier()
	}
	return multiplier
}

// SearchPokemon returns Pokemon matching the search query
func (s *Store) SearchPokemon(query string, limit int) []*Pokemon {
	query = strings.ToLower(query)
	var results []*Pokemon

	for _, pokemon := range s.Pokedex {
		if strings.Contains(strings.ToLower(pokemon.Name), query) {
			results = append(results, pokemon)
			if len(results) >= limit {
				break
			}
		}
	}

	return results
}

// SearchMoves returns Moves matching the search query
func (s *Store) SearchMoves(query string, limit int) []*Move {
	query = strings.ToLower(query)
	var results []*Move

	for _, move := range s.Moves {
		if strings.Contains(strings.ToLower(move.Name), query) {
			results = append(results, move)
			if len(results) >= limit {
				break
			}
		}
	}

	return results
}

// AllPokemonList returns a list of all Pokemon for autocomplete
func (s *Store) AllPokemonList() []map[string]string {
	var list []map[string]string
	for id, pokemon := range s.Pokedex {
		list = append(list, map[string]string{
			"id":   id,
			"name": pokemon.Name,
		})
	}
	return list
}

// AllMovesList returns a list of all Moves for autocomplete
func (s *Store) AllMovesList() []map[string]string {
	var list []map[string]string
	for id, move := range s.Moves {
		list = append(list, map[string]string{
			"id":   id,
			"name": move.Name,
		})
	}
	return list
}

// AllItemsList returns a list of all Items for autocomplete
func (s *Store) AllItemsList() []map[string]string {
	var list []map[string]string
	for id, item := range s.Items {
		list = append(list, map[string]string{
			"id":   id,
			"name": item.Name,
			"desc": item.Desc,
		})
	}
	return list
}

// AllAbilitiesList returns a list of all Abilities for autocomplete
func (s *Store) AllAbilitiesList() []map[string]string {
	var list []map[string]string
	for id, ability := range s.Abilities {
		list = append(list, map[string]string{
			"id":   id,
			"name": ability.Name,
		})
	}
	return list
}

// AllNaturesList returns a list of all Natures for autocomplete
func (s *Store) AllNaturesList() []map[string]string {
	var list []map[string]string
	for id, nature := range s.Natures {
		list = append(list, map[string]string{
			"id":   id,
			"name": nature.Name,
		})
	}
	return list
}

// GetPokemonByNum returns a Pokemon by its national dex number
// Prefers base forms (without hyphens) over form variants
func (s *Store) GetPokemonByNum(num int) *Pokemon {
	var fallback *Pokemon
	for _, pokemon := range s.Pokedex {
		if pokemon.Num == num {
			// Prefer base form (no hyphen in name) over variants
			if !strings.Contains(pokemon.Name, "-") {
				return pokemon
			}
			// Keep first variant as fallback
			if fallback == nil {
				fallback = pokemon
			}
		}
	}
	return fallback
}

// GetMoveByNum returns a Move by its number
func (s *Store) GetMoveByNum(num int) *Move {
	for _, move := range s.Moves {
		if move.Num == num {
			return move
		}
	}
	return nil
}

// GetItemByNum returns an Item by its number
// Prefers battle items (with flingBasePower) over key items when there are duplicates
func (s *Store) GetItemByNum(num int) *Item {
	var fallback *Item
	for _, item := range s.Items {
		if item.Num == num {
			// Battle items have flingBasePower, key items don't
			if item.FlingBasePower > 0 {
				return item
			}
			// Keep first match as fallback if no battle item found
			if fallback == nil {
				fallback = item
			}
		}
	}
	return fallback
}
