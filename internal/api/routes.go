package api

import (
	"net/http"
	"strings"
)

// SetupRoutes configures HTTP routes
func (h *Handler) SetupRoutes(mux *http.ServeMux) {
	// API routes
	mux.HandleFunc("/api/calculate", h.HandleCalculate)
	mux.HandleFunc("/api/pokemon/", h.routePokemon)
	mux.HandleFunc("/api/pokemon", h.HandleListPokemon)
	mux.HandleFunc("/api/moves/", h.routeMoves)
	mux.HandleFunc("/api/moves", h.HandleListMoves)
	mux.HandleFunc("/api/items", h.HandleListItems)
	mux.HandleFunc("/api/abilities", h.HandleListAbilities)
	mux.HandleFunc("/api/natures", h.HandleListNatures)
	mux.HandleFunc("/api/search/pokemon", h.HandleSearchPokemon)
	mux.HandleFunc("/api/search/moves", h.HandleSearchMoves)
}

// routePokemon routes Pokemon requests based on path
func (h *Handler) routePokemon(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/pokemon/")
	if id == "" {
		h.HandleListPokemon(w, r)
		return
	}
	h.HandleGetPokemon(w, r)
}

// routeMoves routes Move requests based on path
func (h *Handler) routeMoves(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/moves/")
	if id == "" {
		h.HandleListMoves(w, r)
		return
	}
	h.HandleGetMove(w, r)
}
