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

        // Cached ability data
        abilityCache: {},

        // Party data (from localStorage, shared with party page)
        partyPokemon: [],

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

            // Load party data from shared localStorage
            this.loadPartyData();

            // Restore saved state
            await this.loadState();
        },

        // Load party data from shared localStorage
        loadPartyData() {
            try {
                const saved = localStorage.getItem('nuzlocke_party');
                if (saved) {
                    const state = JSON.parse(saved);
                    this.partyPokemon = state.party || [];
                }
            } catch (e) {
                console.error('Failed to load party data:', e);
            }
        },

        // Save state to localStorage
        saveState() {
            try {
                const state = {
                    selectedPokemonId: this.selectedPokemon ? this.toID(this.selectedPokemon.name) : null,
                    searchQuery: this.searchQuery,
                    showLevelUp: this.showLevelUp,
                    showTM: this.showTM,
                    showEgg: this.showEgg,
                    showTutor: this.showTutor
                };
                localStorage.setItem('moves_state', JSON.stringify(state));
            } catch (e) {
                console.error('Failed to save moves state:', e);
            }
        },

        // Load state from localStorage
        async loadState() {
            try {
                const saved = localStorage.getItem('moves_state');
                if (!saved) return;

                const state = JSON.parse(saved);

                // Restore filters
                if (state.showLevelUp !== undefined) this.showLevelUp = state.showLevelUp;
                if (state.showTM !== undefined) this.showTM = state.showTM;
                if (state.showEgg !== undefined) this.showEgg = state.showEgg;
                if (state.showTutor !== undefined) this.showTutor = state.showTutor;

                // Restore search query
                if (state.searchQuery) this.searchQuery = state.searchQuery;

                // Restore selected Pokemon
                if (state.selectedPokemonId) {
                    await this.selectPokemonById(state.selectedPokemonId);
                }
            } catch (e) {
                console.error('Failed to load moves state:', e);
            }
        },

        // Select Pokemon by ID (for state restoration)
        async selectPokemonById(pokemonId) {
            try {
                const response = await fetch(`/api/pokemon/${pokemonId}/full`);
                const data = await response.json();
                this.selectedPokemon = data.pokemon;
                this.learnset = data.learnset;
                this.typeMatchups = data.typeMatchups;
                this.searchQuery = data.pokemon.name;

                // Pre-load move data
                await this.loadMoveData();
                await this.loadAbilityData();
            } catch (e) {
                console.error('Failed to load Pokemon:', e);
            }
        },

        // Import Pokemon from party
        async importFromParty(partyPokemon) {
            const speciesId = partyPokemon.species.toLowerCase().replace(/[^a-z0-9]/g, '');
            await this.selectPokemonById(speciesId);
            this.saveState();
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

                // Load ability data
                await this.loadAbilityData();

                // Save state after selection
                this.saveState();
            } catch (e) {
                console.error('Failed to load Pokemon:', e);
            }
        },

        // Load ability data for the Pokemon's abilities
        async loadAbilityData() {
            if (!this.selectedPokemon?.abilities) return;

            const promises = [];
            for (const ability of Object.values(this.selectedPokemon.abilities)) {
                const abilityId = this.toID(ability);
                if (!this.abilityCache[abilityId]) {
                    promises.push(this.fetchAbility(abilityId));
                }
            }
            await Promise.all(promises);
        },

        // Fetch a single ability
        async fetchAbility(abilityId) {
            try {
                const response = await fetch(`/api/abilities/${abilityId}`);
                if (response.ok) {
                    this.abilityCache[abilityId] = await response.json();
                }
            } catch (e) {
                console.error(`Failed to load ability ${abilityId}:`, e);
            }
        },

        // Get ability description for tooltip
        getAbilityDesc(abilityName) {
            const abilityId = this.toID(abilityName);
            const ability = this.abilityCache[abilityId];
            return ability?.shortDesc || ability?.desc || '';
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

        // Get move description for tooltip
        getMoveDesc(moveId) {
            const move = this.moveCache[moveId];
            return move?.shortDesc || move?.desc || '';
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
