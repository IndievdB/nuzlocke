# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Build
go build -o nuzlocke ./cmd/server

# Run (default port 8080)
./nuzlocke

# Run with options
./nuzlocke -port 8080 -data data -web web

# Development with live reload (requires Air)
air
```

## Data Generation Scripts

```bash
node scripts/convert_showdown_data.js   # Convert Pokemon Showdown data to JSON
python3 scripts/gen_catchrates.py       # Generate catch rate data
```

## Architecture Overview

This is a Pokemon damage calculator and save file parser for nuzlocke runs, built with Go backend and vanilla JS frontend.

### Backend Structure (`internal/`)

- **`data/`** - In-memory data store loaded from JSON files at startup. `Store` struct holds all Pokemon, moves, items, abilities, type chart, natures, learnsets. Provides case-insensitive lookups via `GetPokemon()`, `GetMove()`, etc.

- **`calc/`** - Damage calculation engine supporting Gen 3 and Gen 5+ formulas. Gen 3 uses type-based physical/special split; Gen 5+ uses move category. `Calculator.Calculate()` is the main entry point, returning damage rolls, KO chances, and applied factors.

- **`models/`** - Battle state models: `BattlePokemon` (stats, EVs, IVs, boosts), `BattleMove`, `Field` (weather, terrain, screens). Stat calculations use standard Pokemon formulas with nature modifiers.

- **`savefile/`** - Gen 3 save file parser for Pokemon Emerald (including pokeemerald-expansion). Parses party (100-byte Pokemon) and PC boxes (80-byte Pokemon). Uses personality-based substructure ordering and handles expansion's modified data layout (nickname extensions share bits with experience field).

- **`api/`** - HTTP handlers and routes. Key endpoints: `/api/calculate` (POST), `/api/pokemon/{id}`, `/api/party/parse` (POST for save files).

### Frontend Structure (`web/`)

Vanilla JavaScript with Alpine.js-style reactive patterns. State persisted to localStorage.

- `calculator.html` + `app.js` - Damage calculator with attacker/defender selection
- `party.html` + `party.js` - Save file upload, party and box Pokemon viewer
- `moves.html` + `moves.js` - Move database with learnset filtering by generation

### Data Flow

1. Server loads `data/*.json` into `data.Store` at startup
2. Frontend makes API calls to fetch Pokemon/move data
3. Damage calculations: Frontend POSTs to `/api/calculate`, backend runs formula in `calc/`, returns damage range
4. Save parsing: Frontend POSTs binary `.sav` to `/api/party/parse`, backend decrypts and extracts Pokemon data

## Key Implementation Details

- **Damage modifiers** use 4096-base system (4096 = 1.0x, 6144 = 1.5x) for precision
- **pokeemerald-expansion** support: Species IDs >1000 mapped to national dex, item IDs mapped to Showdown format, experience field masked to 21 bits (upper bits used for extended nicknames)
- **Box Pokemon** don't store calculated stats - they're computed from base stats + IVs + EVs + level + nature
