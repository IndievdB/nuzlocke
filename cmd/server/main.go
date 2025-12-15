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
	log.Printf("Loaded %d Pokemon, %d moves, %d items, %d abilities, %d natures",
		len(store.Pokedex),
		len(store.Moves),
		len(store.Items),
		len(store.Abilities),
		len(store.Natures),
	)

	// Create handler
	handler := api.NewHandler(store)

	// Setup routes
	mux := http.NewServeMux()
	handler.SetupRoutes(mux)

	// Serve static files
	fs := http.FileServer(http.Dir(absWebDir))
	mux.Handle("/", fs)

	// Apply middleware
	finalHandler := api.Chain(mux, api.CORS, api.Logging)

	// Start server
	addr := ":" + *port
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, finalHandler); err != nil {
		log.Fatal("Server failed:", err)
	}
}
