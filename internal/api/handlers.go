package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"nuzlocke/internal/calc"
	"nuzlocke/internal/data"
	"nuzlocke/internal/models"
	"nuzlocke/internal/savefile"
)

// Handler holds the dependencies for HTTP handlers
type Handler struct {
	Store      *data.Store
	Calculator *calc.Calculator
}

// NewHandler creates a new Handler
func NewHandler(store *data.Store) *Handler {
	return &Handler{
		Store:      store,
		Calculator: calc.NewCalculator(store),
	}
}

// CalculateRequest represents the JSON request for damage calculation
type CalculateRequest struct {
	Generation int                   `json:"generation"` // 3 or 5+ (defaults to 9)
	Attacker   *models.BattlePokemon `json:"attacker"`
	Defender   *models.BattlePokemon `json:"defender"`
	Move       *models.BattleMove    `json:"move"`
	Field      *models.Field         `json:"field"`
}

// HandleCalculate handles POST /api/calculate
func (h *Handler) HandleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Attacker == nil || req.Attacker.Species == "" {
		http.Error(w, "Attacker species is required", http.StatusBadRequest)
		return
	}
	if req.Defender == nil || req.Defender.Species == "" {
		http.Error(w, "Defender species is required", http.StatusBadRequest)
		return
	}
	if req.Move == nil || req.Move.Name == "" {
		http.Error(w, "Move name is required", http.StatusBadRequest)
		return
	}

	// Create calculation request
	calcReq := &calc.CalculateRequest{
		Attacker:   req.Attacker,
		Defender:   req.Defender,
		Move:       req.Move,
		Field:      req.Field,
		Generation: req.Generation,
	}

	// Perform calculation
	result := h.Calculator.Calculate(calcReq)

	// Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleListPokemon handles GET /api/pokemon
func (h *Handler) HandleListPokemon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.Store.AllPokemonList())
}

// HandleGetPokemon handles GET /api/pokemon/{id}
func (h *Handler) HandleGetPokemon(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/pokemon/")
	if id == "" {
		http.Error(w, "Pokemon ID is required", http.StatusBadRequest)
		return
	}

	pokemon := h.Store.GetPokemon(id)
	if pokemon == nil {
		http.Error(w, "Pokemon not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pokemon)
}

// HandleListMoves handles GET /api/moves
func (h *Handler) HandleListMoves(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.Store.AllMovesList())
}

// HandleGetMove handles GET /api/moves/{id}
func (h *Handler) HandleGetMove(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/moves/")
	if id == "" {
		http.Error(w, "Move ID is required", http.StatusBadRequest)
		return
	}

	move := h.Store.GetMove(id)
	if move == nil {
		http.Error(w, "Move not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(move)
}

// HandleListItems handles GET /api/items
func (h *Handler) HandleListItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.Store.AllItemsList())
}

// HandleListAbilities handles GET /api/abilities
func (h *Handler) HandleListAbilities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.Store.AllAbilitiesList())
}

// HandleGetAbility handles GET /api/abilities/{id}
func (h *Handler) HandleGetAbility(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/abilities/")
	if id == "" {
		http.Error(w, "Ability ID is required", http.StatusBadRequest)
		return
	}

	ability := h.Store.GetAbility(id)
	if ability == nil {
		http.Error(w, "Ability not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ability)
}

// HandleListNatures handles GET /api/natures
func (h *Handler) HandleListNatures(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.Store.AllNaturesList())
}

// SearchResult represents a search result item
type SearchResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// HandleSearchPokemon handles GET /api/search/pokemon?q=query
func (h *Handler) HandleSearchPokemon(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		json.NewEncoder(w).Encode(map[string][]SearchResult{"results": {}})
		return
	}

	results := h.Store.SearchPokemon(query, 20)
	searchResults := make([]SearchResult, len(results))
	for i, p := range results {
		searchResults[i] = SearchResult{
			ID:   data.ToID(p.Name),
			Name: p.Name,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]SearchResult{"results": searchResults})
}

// HandleSearchMoves handles GET /api/search/moves?q=query
func (h *Handler) HandleSearchMoves(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		json.NewEncoder(w).Encode(map[string][]SearchResult{"results": {}})
		return
	}

	results := h.Store.SearchMoves(query, 20)
	searchResults := make([]SearchResult, len(results))
	for i, m := range results {
		searchResults[i] = SearchResult{
			ID:   data.ToID(m.Name),
			Name: m.Name,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]SearchResult{"results": searchResults})
}

// LearnsetResponse represents the parsed learnset response
type LearnsetResponse struct {
	Pokemon  string               `json:"pokemon"`
	Learnset *data.ParsedLearnset `json:"learnset"`
}

// HandleGetLearnset handles GET /api/pokemon/{id}/learnset
func (h *Handler) HandleGetLearnset(w http.ResponseWriter, r *http.Request) {
	// Extract Pokemon ID from path: /api/pokemon/{id}/learnset
	path := r.URL.Path
	id := strings.TrimPrefix(path, "/api/pokemon/")
	id = strings.TrimSuffix(id, "/learnset")

	if id == "" {
		http.Error(w, "Pokemon ID is required", http.StatusBadRequest)
		return
	}

	pokemon := h.Store.GetPokemon(id)
	if pokemon == nil {
		http.Error(w, "Pokemon not found", http.StatusNotFound)
		return
	}

	learnset := h.Store.GetLearnset(id)

	// If form has no learnset, fall back to base species
	if learnset == nil && pokemon.BaseSpecies != "" {
		baseId := strings.ToLower(strings.ReplaceAll(pokemon.BaseSpecies, " ", ""))
		learnset = h.Store.GetLearnset(baseId)
	}

	// Parse generation from query param (default to 9)
	generation := 9
	if genStr := r.URL.Query().Get("gen"); genStr != "" {
		if g, err := strconv.Atoi(genStr); err == nil && g >= 1 && g <= 9 {
			generation = g
		}
	}

	parsed := data.ParseLearnset(learnset, generation)

	response := LearnsetResponse{
		Pokemon:  id,
		Learnset: parsed,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TypeMatchups represents type effectiveness info
type TypeMatchups struct {
	Weaknesses  map[string]float64 `json:"weaknesses"`
	Resistances map[string]float64 `json:"resistances"`
	Immunities  []string           `json:"immunities"`
}

// FullPokemonResponse represents the full Pokemon data with learnset and type matchups
type FullPokemonResponse struct {
	Pokemon     *data.Pokemon        `json:"pokemon"`
	Learnset    *data.ParsedLearnset `json:"learnset"`
	TypeMatchups *TypeMatchups       `json:"typeMatchups"`
}

// HandleGetPokemonFull handles GET /api/pokemon/{id}/full
func (h *Handler) HandleGetPokemonFull(w http.ResponseWriter, r *http.Request) {
	// Extract Pokemon ID from path: /api/pokemon/{id}/full
	path := r.URL.Path
	id := strings.TrimPrefix(path, "/api/pokemon/")
	id = strings.TrimSuffix(id, "/full")

	if id == "" {
		http.Error(w, "Pokemon ID is required", http.StatusBadRequest)
		return
	}

	pokemon := h.Store.GetPokemon(id)
	if pokemon == nil {
		http.Error(w, "Pokemon not found", http.StatusNotFound)
		return
	}

	learnset := h.Store.GetLearnset(id)

	// Parse generation from query param (default to 9)
	generation := 9
	if genStr := r.URL.Query().Get("gen"); genStr != "" {
		if g, err := strconv.Atoi(genStr); err == nil && g >= 1 && g <= 9 {
			generation = g
		}
	}

	parsed := data.ParseLearnset(learnset, generation)

	// Calculate type matchups
	matchups := calculateTypeMatchups(h.Store, pokemon.Types)

	response := FullPokemonResponse{
		Pokemon:      pokemon,
		Learnset:     parsed,
		TypeMatchups: matchups,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// calculateTypeMatchups calculates type effectiveness against a Pokemon's types
func calculateTypeMatchups(store *data.Store, defenderTypes []string) *TypeMatchups {
	allTypes := []string{
		"Normal", "Fire", "Water", "Electric", "Grass", "Ice",
		"Fighting", "Poison", "Ground", "Flying", "Psychic", "Bug",
		"Rock", "Ghost", "Dragon", "Dark", "Steel", "Fairy",
	}

	matchups := &TypeMatchups{
		Weaknesses:  make(map[string]float64),
		Resistances: make(map[string]float64),
		Immunities:  []string{},
	}

	for _, attackType := range allTypes {
		multiplier := store.GetTypeEffectivenessMultiple(attackType, defenderTypes)

		if multiplier == 0 {
			matchups.Immunities = append(matchups.Immunities, strings.ToLower(attackType))
		} else if multiplier > 1 {
			matchups.Weaknesses[strings.ToLower(attackType)] = multiplier
		} else if multiplier < 1 {
			matchups.Resistances[strings.ToLower(attackType)] = multiplier
		}
	}

	return matchups
}

// MoveDetail contains move information for tooltips
type MoveDetail struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	Power       int    `json:"power"`
	Accuracy    int    `json:"accuracy"`
	PP          int    `json:"pp"`
	Description string `json:"description"`
}

// ItemDetail contains item information for tooltips
type ItemDetail struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AbilityDetail contains ability information for tooltips
type AbilityDetail struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PartyPokemonResponse is the rich response for a party Pokemon
type PartyPokemonResponse struct {
	Species      string                 `json:"species"`
	Nickname     string                 `json:"nickname"`
	Level        int                    `json:"level"`
	Types        []string               `json:"types"`
	Nature       string                 `json:"nature"`
	NatureEffect savefile.NatureEffect  `json:"natureEffect"`
	Ability      *AbilityDetail         `json:"ability,omitempty"`
	Item         *ItemDetail            `json:"item,omitempty"`
	Moves        []MoveDetail           `json:"moves"`
	Stats        savefile.PokemonStats  `json:"stats"`
	IVs          savefile.PokemonStats  `json:"ivs"`
	EVs          savefile.PokemonStats  `json:"evs"`
	CurrentHP    int                    `json:"currentHp"`
	Friendship   int                    `json:"friendship"`
}

// BoxPokemonResponse represents a Pokemon in a PC box with enriched data
type BoxPokemonResponse struct {
	Species      string                `json:"species"`
	Nickname     string                `json:"nickname"`
	Level        int                   `json:"level"`
	Types        []string              `json:"types"`
	Nature       string                `json:"nature"`
	NatureEffect savefile.NatureEffect `json:"natureEffect"`
	Ability      *AbilityDetail        `json:"ability,omitempty"`
	Item         *ItemDetail           `json:"item,omitempty"`
	Moves        []MoveDetail          `json:"moves"`
	Stats        savefile.PokemonStats `json:"stats"`
	IVs          savefile.PokemonStats `json:"ivs"`
	EVs          savefile.PokemonStats `json:"evs"`
	Friendship   int                   `json:"friendship"`
}

// BagItemResponse represents a bag item with resolved name and description
type BagItemResponse struct {
	Name        string `json:"name"`
	Quantity    int    `json:"quantity"`
	Description string `json:"description"`
}

// BagPocketsResponse contains all bag pockets with resolved item names
type BagPocketsResponse struct {
	PCItems   []BagItemResponse `json:"pcItems"`
	Items     []BagItemResponse `json:"items"`
	KeyItems  []BagItemResponse `json:"keyItems"`
	PokeBalls []BagItemResponse `json:"pokeBalls"`
	TMsHMs    []BagItemResponse `json:"tmsHms"`
	Berries   []BagItemResponse `json:"berries"`
}

// ParseSaveResponse is the response for the parse save endpoint
type ParseSaveResponse struct {
	Party []PartyPokemonResponse   `json:"party"`
	Boxes [][]BoxPokemonResponse   `json:"boxes"`
	Bag   *BagPocketsResponse      `json:"bag"`
}

// HandleParseSave handles POST /api/nuzlocke/parse
func (h *Handler) HandleParseSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the save file from the request body
	saveData, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Parse the save file
	result, err := savefile.ParseGen3Save(saveData)
	if err != nil {
		http.Error(w, "Failed to parse save file: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Build rich response
	response := ParseSaveResponse{
		Party: make([]PartyPokemonResponse, 0, len(result.Party)),
		Boxes: make([][]BoxPokemonResponse, len(result.Boxes)),
	}
	for i := range response.Boxes {
		response.Boxes[i] = []BoxPokemonResponse{}
	}

	for _, p := range result.Party {
		pokemon := PartyPokemonResponse{
			Nickname:     p.Nickname,
			Level:        p.Level,
			Nature:       p.Nature,
			NatureEffect: savefile.GetNatureEffect(p.Nature),
			Stats:        p.Stats,
			IVs:          p.IVs,
			EVs:          p.EVs,
			CurrentHP:    p.CurrentHP,
			Friendship:   p.Friendship,
		}

		// Resolve species
		speciesData := h.Store.GetPokemonByNum(p.SpeciesNum)
		if speciesData != nil {
			pokemon.Species = speciesData.Name
			pokemon.Types = speciesData.Types

			// Resolve ability based on slot (0, 1, or 2 for hidden)
			// If requested slot doesn't exist, fall back to slot 0
			slot := "0"
			if p.AbilitySlot == 1 {
				slot = "1"
			} else if p.AbilitySlot == 2 {
				slot = "H"
			}
			abilityName := speciesData.GetAbility(slot)
			if abilityName == "" && slot != "0" {
				// Fallback to slot 0 if the requested slot doesn't exist
				abilityName = speciesData.GetAbility("0")
			}
			if abilityName != "" {
				ability := h.Store.GetAbility(abilityName)
				if ability != nil {
					pokemon.Ability = &AbilityDetail{
						Name:        ability.Name,
						Description: ability.ShortDesc,
					}
				}
			}
		} else {
			pokemon.Species = "Unknown"
		}

		// Resolve item with description
		if p.ItemNum > 0 {
			item := h.Store.GetItemByNum(p.ItemNum)
			if item != nil {
				pokemon.Item = &ItemDetail{
					Name:        item.Name,
					Description: item.Desc,
				}
			}
		}

		// Resolve moves with details
		pokemon.Moves = make([]MoveDetail, 0, len(p.MoveNums))
		for _, moveNum := range p.MoveNums {
			move := h.Store.GetMoveByNum(moveNum)
			if move != nil {
				accuracy := 0
				switch v := move.Accuracy.(type) {
				case int:
					accuracy = v
				case float64:
					accuracy = int(v)
				case bool:
					if v {
						accuracy = 100 // "true" means never-miss
					}
				}
				pokemon.Moves = append(pokemon.Moves, MoveDetail{
					Name:        move.Name,
					Type:        move.Type,
					Category:    move.Category,
					Power:       move.BasePower,
					Accuracy:    accuracy,
					PP:          move.PP,
					Description: move.ShortDesc,
				})
			}
		}

		response.Party = append(response.Party, pokemon)
	}

	// Process box Pokemon
	for boxIdx, box := range result.Boxes {
		for _, p := range box {
			pokemon := BoxPokemonResponse{
				Nickname:     p.Nickname,
				Level:        p.Level,
				Nature:       p.Nature,
				NatureEffect: savefile.GetNatureEffect(p.Nature),
				IVs:          p.IVs,
				EVs:          p.EVs,
				Friendship:   p.Friendship,
			}

			// Resolve species
			speciesData := h.Store.GetPokemonByNum(p.SpeciesNum)
			if speciesData != nil {
				pokemon.Species = speciesData.Name
				pokemon.Types = speciesData.Types

				// Calculate stats from base stats, IVs, EVs, level, nature
				natureData := h.Store.GetNature(strings.ToLower(p.Nature))
				if natureData != nil {
					ivs := models.StatSpread{
						HP: p.IVs.HP, Atk: p.IVs.Attack, Def: p.IVs.Defense,
						SpA: p.IVs.SpAtk, SpD: p.IVs.SpDef, Spe: p.IVs.Speed,
					}
					evs := models.StatSpread{
						HP: p.EVs.HP, Atk: p.EVs.Attack, Def: p.EVs.Defense,
						SpA: p.EVs.SpAtk, SpD: p.EVs.SpDef, Spe: p.EVs.Speed,
					}
					calcStats := models.CalculateAllStats(speciesData.BaseStats, ivs, evs, p.Level, natureData)
					pokemon.Stats = savefile.PokemonStats{
						HP: calcStats.HP, Attack: calcStats.Atk, Defense: calcStats.Def,
						SpAtk: calcStats.SpA, SpDef: calcStats.SpD, Speed: calcStats.Spe,
					}
				}

				// Resolve ability based on slot
				slot := "0"
				if p.AbilitySlot == 1 {
					slot = "1"
				} else if p.AbilitySlot == 2 {
					slot = "H"
				}
				abilityName := speciesData.GetAbility(slot)
				if abilityName == "" && slot != "0" {
					abilityName = speciesData.GetAbility("0")
				}
				if abilityName != "" {
					ability := h.Store.GetAbility(abilityName)
					if ability != nil {
						pokemon.Ability = &AbilityDetail{
							Name:        ability.Name,
							Description: ability.ShortDesc,
						}
					}
				}
			} else {
				pokemon.Species = "Unknown"
			}

			// Resolve item
			if p.ItemNum > 0 {
				item := h.Store.GetItemByNum(p.ItemNum)
				if item != nil {
					pokemon.Item = &ItemDetail{
						Name:        item.Name,
						Description: item.Desc,
					}
				}
			}

			// Resolve moves
			pokemon.Moves = make([]MoveDetail, 0, len(p.MoveNums))
			for _, moveNum := range p.MoveNums {
				move := h.Store.GetMoveByNum(moveNum)
				if move != nil {
					accuracy := 0
					switch v := move.Accuracy.(type) {
					case int:
						accuracy = v
					case float64:
						accuracy = int(v)
					case bool:
						if v {
							accuracy = 100
						}
					}
					pokemon.Moves = append(pokemon.Moves, MoveDetail{
						Name:        move.Name,
						Type:        move.Type,
						Category:    move.Category,
						Power:       move.BasePower,
						Accuracy:    accuracy,
						PP:          move.PP,
						Description: move.ShortDesc,
					})
				}
			}

			response.Boxes[boxIdx] = append(response.Boxes[boxIdx], pokemon)
		}
	}

	// Process bag items
	if result.Bag != nil {
		resolveBagPocket := func(items []savefile.BagItem) []BagItemResponse {
			resolved := make([]BagItemResponse, 0, len(items))
			for _, item := range items {
				itemData := h.Store.GetItemByNum(item.ItemNum)
				if itemData != nil {
					resolved = append(resolved, BagItemResponse{
						Name:        itemData.Name,
						Quantity:    item.Quantity,
						Description: itemData.Desc,
					})
				} else {
					resolved = append(resolved, BagItemResponse{
						Name:        "Unknown Item",
						Quantity:    item.Quantity,
						Description: "",
					})
				}
			}
			return resolved
		}

		response.Bag = &BagPocketsResponse{
			PCItems:   resolveBagPocket(result.Bag.PCItems),
			Items:     resolveBagPocket(result.Bag.Items),
			KeyItems:  resolveBagPocket(result.Bag.KeyItems),
			PokeBalls: resolveBagPocket(result.Bag.PokeBalls),
			TMsHMs:    resolveBagPocket(result.Bag.TMsHMs),
			Berries:   resolveBagPocket(result.Bag.Berries),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
