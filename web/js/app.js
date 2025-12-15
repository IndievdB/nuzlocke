function calculator() {
    return {
        // Generation setting
        generation: '9',

        // Pokemon data
        attacker: createDefaultPokemon(),
        defender: createDefaultPokemon(),

        // Move data
        move: {
            name: '',
            isCrit: false
        },

        // Field conditions
        field: {
            weather: '',
            terrain: '',
            attackerSide: {
                helpingHand: false
            },
            defenderSide: {
                reflect: false,
                lightScreen: false
            }
        },

        // Search state
        attackerSearch: '',
        attackerSearchResults: [],
        showAttackerResults: false,

        defenderSearch: '',
        defenderSearchResults: [],
        showDefenderResults: false,

        moveSearch: '',
        moveSearchResults: [],
        showMoveResults: false,

        // Defender HP percent (for display)
        defenderHPPercent: 100,

        // Calculation result
        result: null,

        // Data caches
        natures: [],

        // Initialize
        async init() {
            // Load natures for dropdown
            try {
                const response = await fetch('/api/natures');
                this.natures = await response.json();
            } catch (e) {
                console.error('Failed to load natures:', e);
            }
        },

        // Search Pokemon
        async searchPokemon(role) {
            const query = role === 'attacker' ? this.attackerSearch : this.defenderSearch;
            if (query.length < 2) {
                if (role === 'attacker') {
                    this.attackerSearchResults = [];
                } else {
                    this.defenderSearchResults = [];
                }
                return;
            }

            try {
                const response = await fetch(`/api/search/pokemon?q=${encodeURIComponent(query)}`);
                const data = await response.json();
                if (role === 'attacker') {
                    this.attackerSearchResults = data.results || [];
                    this.showAttackerResults = true;
                } else {
                    this.defenderSearchResults = data.results || [];
                    this.showDefenderResults = true;
                }
            } catch (e) {
                console.error('Search failed:', e);
            }
        },

        // Select Pokemon
        async selectPokemon(role, result) {
            try {
                const response = await fetch(`/api/pokemon/${result.id}`);
                const pokemon = await response.json();

                const target = role === 'attacker' ? this.attacker : this.defender;
                target.species = result.id;
                target.speciesData = pokemon;

                // Set default ability
                if (pokemon.abilities && pokemon.abilities['0']) {
                    target.ability = pokemon.abilities['0'];
                }

                if (role === 'attacker') {
                    this.attackerSearch = result.name;
                    this.showAttackerResults = false;
                } else {
                    this.defenderSearch = result.name;
                    this.showDefenderResults = false;
                }
            } catch (e) {
                console.error('Failed to load Pokemon:', e);
            }
        },

        // Search Moves
        async searchMoves() {
            if (this.moveSearch.length < 2) {
                this.moveSearchResults = [];
                return;
            }

            try {
                const response = await fetch(`/api/search/moves?q=${encodeURIComponent(this.moveSearch)}`);
                const data = await response.json();
                this.moveSearchResults = data.results || [];
                this.showMoveResults = true;
            } catch (e) {
                console.error('Search failed:', e);
            }
        },

        // Select Move
        selectMove(result) {
            this.move.name = result.id;
            this.moveSearch = result.name;
            this.showMoveResults = false;
        },

        // Check if calculation is possible
        canCalculate() {
            return this.attacker.species && this.defender.species && this.move.name;
        },

        // Perform calculation
        async calculate() {
            if (!this.canCalculate()) return;

            // Build request
            const request = {
                generation: parseInt(this.generation),
                attacker: {
                    species: this.attacker.species,
                    level: this.attacker.level,
                    nature: this.attacker.nature,
                    ability: this.attacker.ability,
                    item: this.attacker.item,
                    status: this.attacker.status,
                    evs: this.attacker.evs,
                    ivs: this.attacker.ivs,
                    boosts: this.attacker.boosts
                },
                defender: {
                    species: this.defender.species,
                    level: this.defender.level,
                    nature: this.defender.nature,
                    ability: this.defender.ability,
                    item: this.defender.item,
                    status: this.defender.status,
                    evs: this.defender.evs,
                    ivs: this.defender.ivs,
                    boosts: this.defender.boosts,
                    currentHP: this.defenderHPPercent < 100 ? Math.floor(this.defender.maxHP * this.defenderHPPercent / 100) : 0
                },
                move: {
                    name: this.move.name,
                    isCrit: this.move.isCrit
                },
                field: this.field
            };

            try {
                const response = await fetch('/api/calculate', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(request)
                });

                if (!response.ok) {
                    const error = await response.text();
                    console.error('Calculation failed:', error);
                    return;
                }

                this.result = await response.json();
            } catch (e) {
                console.error('Calculation failed:', e);
            }
        },

        // Get damage bar style
        getDamageBarStyle() {
            if (!this.result) return {};

            const avg = (this.result.minPercent + this.result.maxPercent) / 2;
            const width = Math.min(avg, 100);

            let color = '#27ae60'; // Green
            if (avg >= 100) {
                color = '#e74c3c'; // Red
            } else if (avg >= 50) {
                color = '#f39c12'; // Orange
            }

            return {
                width: width + '%',
                backgroundColor: color
            };
        }
    };
}

function createDefaultPokemon() {
    return {
        species: '',
        speciesData: null,
        level: 100,
        nature: 'hardy',
        ability: '',
        item: '',
        status: '',
        evs: { hp: 0, atk: 0, def: 0, spa: 0, spd: 0, spe: 0 },
        ivs: { hp: 31, atk: 31, def: 31, spa: 31, spd: 31, spe: 31 },
        boosts: { atk: 0, def: 0, spa: 0, spd: 0, spe: 0 }
    };
}
