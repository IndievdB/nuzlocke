package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"nuzlocke/internal/calc"
	"nuzlocke/internal/data"
	"nuzlocke/internal/models"
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
	parsed := data.ParseLearnset(learnset, 9) // Default to Gen 9

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
	parsed := data.ParseLearnset(learnset, 9)

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
