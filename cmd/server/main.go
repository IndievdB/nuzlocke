package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"nuzlocke/internal/api"
	"nuzlocke/internal/data"
)

func main() {
	// Command line flags
	port := flag.String("port", "8080", "Server port")
	dataDir := flag.String("data", "data", "Data directory containing JSON files")
	webDir := flag.String("web", "web", "Web directory containing static files")
	flag.Parse()

	// Get absolute paths
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get working directory:", err)
	}

	absDataDir := filepath.Join(workDir, *dataDir)
	absWebDir := filepath.Join(workDir, *webDir)

	// Load data
	log.Printf("Loading data from %s...", absDataDir)
	store, err := data.LoadFromDirectory(absDataDir)
	if err != nil {
		log.Fatal("Failed to load data:", err)
	}
	log.Printf("Loaded %d Pokemon, %d moves, %d items, %d abilities, %d natures, %d learnsets",
		len(store.Pokedex),
		len(store.Moves),
		len(store.Items),
		len(store.Abilities),
		len(store.Natures),
		len(store.Learnsets),
	)

	// Create handler
	handler := api.NewHandler(store)

	// Setup routes
	mux := http.NewServeMux()
	handler.SetupRoutes(mux)

	// Serve specific pages
	mux.HandleFunc("/calculator", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(absWebDir, "calculator.html"))
	})
	mux.HandleFunc("/moves", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(absWebDir, "moves.html"))
	})
	mux.HandleFunc("/party", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(absWebDir, "party.html"))
	})
	mux.HandleFunc("/catch", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(absWebDir, "catch.html"))
	})
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(absWebDir, "items.html"))
	})
	mux.HandleFunc("/matchup", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(absWebDir, "matchup.html"))
	})
	mux.HandleFunc("/map", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(absWebDir, "map.html"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/calculator", http.StatusFound)
			return
		}
		// Serve static files for other paths
		http.FileServer(http.Dir(absWebDir)).ServeHTTP(w, r)
	})

	// Apply middleware
	finalHandler := api.Chain(mux, api.CORS, api.Logging)

	// Start server
	addr := ":" + *port
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, finalHandler); err != nil {
		log.Fatal("Server failed:", err)
	}
}
