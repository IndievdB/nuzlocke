function movesApp() {
    return {
        // Search state
        searchQuery: '',
        searchResults: [],
        showResults: false,

        // Selected Pokemon data
        selectedPokemon: null,
        learnset: null,
        typeMatchups: null,

        // Move source filters (level-up enabled by default)
        showLevelUp: true,
        showTM: false,
        showEgg: false,
        showTutor: false,

        // Cached move data
        moveCache: {},

        // Initialize
        async init() {
            // Pre-load all moves for quick lookup
            try {
                const response = await fetch('/api/moves');
                const moves = await response.json();
                for (const move of moves) {
                    this.moveCache[move.id] = null; // Will load on demand
                }
            } catch (e) {
                console.error('Failed to load moves list:', e);
            }
        },

        // Search Pokemon
        async searchPokemon() {
            if (this.searchQuery.length < 2) {
                this.searchResults = [];
                return;
            }

            try {
                const response = await fetch(`/api/search/pokemon?q=${encodeURIComponent(this.searchQuery)}`);
                const data = await response.json();
                this.searchResults = data.results || [];
                this.showResults = true;
            } catch (e) {
                console.error('Search failed:', e);
            }
        },

        // Select Pokemon
        async selectPokemon(result) {
            this.searchQuery = result.name;
            this.showResults = false;

            try {
                const response = await fetch(`/api/pokemon/${result.id}/full`);
                const data = await response.json();
                this.selectedPokemon = data.pokemon;
                this.learnset = data.learnset;
                this.typeMatchups = data.typeMatchups;

                // Pre-load move data for all moves in learnset
                await this.loadMoveData();
            } catch (e) {
                console.error('Failed to load Pokemon:', e);
            }
        },

        // Load move data for all moves in learnset
        async loadMoveData() {
            const moveIds = new Set();

            if (this.learnset?.levelup) {
                this.learnset.levelup.forEach(m => moveIds.add(m.move));
            }
            if (this.learnset?.tm) {
                this.learnset.tm.forEach(id => moveIds.add(id));
            }
            if (this.learnset?.egg) {
                this.learnset.egg.forEach(id => moveIds.add(id));
            }
            if (this.learnset?.tutor) {
                this.learnset.tutor.forEach(id => moveIds.add(id));
            }

            // Load each move
            const promises = [];
            for (const moveId of moveIds) {
                if (!this.moveCache[moveId]) {
                    promises.push(this.fetchMove(moveId));
                }
            }

            await Promise.all(promises);
        },

        // Fetch a single move
        async fetchMove(moveId) {
            try {
                const response = await fetch(`/api/moves/${moveId}`);
                if (response.ok) {
                    this.moveCache[moveId] = await response.json();
                }
            } catch (e) {
                console.error(`Failed to load move ${moveId}:`, e);
            }
        },

        // Get move data from cache
        getMoveData(moveId) {
            return this.moveCache[moveId];
        },

        // Get sprite URL
        getSpriteUrl() {
            if (!this.selectedPokemon) return '';
            const id = this.toID(this.selectedPokemon.name);
            return `https://play.pokemonshowdown.com/sprites/gen5/${id}.png`;
        },

        // Handle sprite load error
        handleSpriteError(event) {
            event.target.style.display = 'none';
        },

        // Convert name to ID
        toID(name) {
            return name.toLowerCase().replace(/[^a-z0-9]/g, '');
        },

        // Get stat bar style
        getStatBarStyle(value) {
            const maxStat = 255;
            const percent = Math.min((value / maxStat) * 100, 100);
            let color = '#27ae60'; // Green

            if (value < 50) {
                color = '#e74c3c'; // Red
            } else if (value < 80) {
                color = '#f39c12'; // Orange
            } else if (value < 100) {
                color = '#f1c40f'; // Yellow
            } else if (value >= 150) {
                color = '#3498db'; // Blue
            }

            return {
                width: percent + '%',
                backgroundColor: color
            };
        },

        // Get stat total
        getStatTotal() {
            if (!this.selectedPokemon?.baseStats) return 0;
            const stats = this.selectedPokemon.baseStats;
            return stats.hp + stats.atk + stats.def + stats.spa + stats.spd + stats.spe;
        },

        // Get category class for move
        getCategoryClass(category) {
            if (!category) return '';
            return 'category-' + category.toLowerCase();
        },

        // Sort level-up moves by level
        sortedLevelUpMoves() {
            if (!this.learnset?.levelup) return [];
            return [...this.learnset.levelup].sort((a, b) => a.level - b.level);
        }
    };
}
