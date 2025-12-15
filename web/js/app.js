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

        // Get critical hit chance based on generation
        getCritChance() {
            // Gen 6+: 1/24 (4.17%), Gen 2-5: 1/16 (6.25%), Gen 1: varies
            if (this.generation === '3') {
                return '6.25%';
            }
            return '4.17%';
        },

        // Get min damage bar style (green portion)
        getMinBarStyle() {
            if (!this.result) return {};
            const width = Math.min(this.result.minPercent, 100);
            return { width: width + '%' };
        },

        // Get range bar style (yellow portion between min and max)
        getRangeBarStyle() {
            if (!this.result) return {};
            const rangeWidth = Math.min(this.result.maxPercent, 100) - Math.min(this.result.minPercent, 100);
            return { width: Math.max(rangeWidth, 0) + '%' };
        },

        // Get defender's current HP based on percentage and max HP
        getDefenderCurrentHP() {
            if (!this.result || !this.result.maxDamage) return 0;

            // Calculate max HP from the damage percentage info
            // maxDamage / maxPercent * 100 = maxHP
            const maxHP = Math.round(this.result.maxDamage / this.result.maxPercent * 100);
            return Math.floor(maxHP * this.defenderHPPercent / 100);
        },

        // Get best case KO (using max damage rolls)
        getBestCaseKO() {
            if (!this.result || !this.result.maxDamage) return '-';

            const currentHP = this.getDefenderCurrentHP();
            if (currentHP <= 0) return '-';

            const hitsNeeded = Math.ceil(currentHP / this.result.maxDamage);
            return hitsNeeded === 1 ? 'OHKO' : hitsNeeded + 'HKO';
        },

        // Get worst case KO (using min damage rolls)
        getWorstCaseKO() {
            if (!this.result || !this.result.minDamage) return '-';

            const currentHP = this.getDefenderCurrentHP();
            if (currentHP <= 0) return '-';

            const hitsNeeded = Math.ceil(currentHP / this.result.minDamage);
            return hitsNeeded === 1 ? 'OHKO' : hitsNeeded + 'HKO';
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
