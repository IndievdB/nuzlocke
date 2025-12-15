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
