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

        // Attacker's learnset and move details cache
        attackerLearnset: null,
        moveCache: {},

        // Form variants
        attackerForms: [],
        defenderForms: [],

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

        // Search Pokemon (filters out form variants)
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

                // Filter out form variants (those with baseSpecies are variants)
                // We need to fetch each Pokemon to check, but for performance we'll
                // filter client-side by checking if name contains '-' patterns for forms
                const results = (data.results || []).filter(p => {
                    // Keep base forms: no hyphen, or specific exceptions
                    const name = p.name;
                    // Common form suffixes to filter out
                    const formPatterns = ['-Mega', '-Alola', '-Galar', '-Hisui', '-Paldea',
                                          '-Gmax', '-Totem', '-Starter', '-Origin', '-Therian',
                                          '-Black', '-White', '-Crowned', '-Eternamax'];
                    return !formPatterns.some(pattern => name.includes(pattern));
                });

                if (role === 'attacker') {
                    this.attackerSearchResults = results;
                    this.showAttackerResults = true;
                } else {
                    this.defenderSearchResults = results;
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
                target.selectedForm = '';

                // Set default ability (first one)
                if (pokemon.abilities && pokemon.abilities['0']) {
                    target.ability = pokemon.abilities['0'];
                }

                if (role === 'attacker') {
                    this.attackerSearch = result.name;
                    this.showAttackerResults = false;
                    // Load learnset for attacker
                    await this.loadAttackerLearnset(result.id);
                    // Load forms
                    await this.loadForms('attacker', pokemon.name);
                } else {
                    this.defenderSearch = result.name;
                    this.showDefenderResults = false;
                    // Load forms
                    await this.loadForms('defender', pokemon.name);
                }

                // Clear move selection when attacker changes
                if (role === 'attacker') {
                    this.move.name = '';
                }
            } catch (e) {
                console.error('Failed to load Pokemon:', e);
            }
        },

        // Load forms for a Pokemon
        async loadForms(role, baseName) {
            try {
                // Search for forms with this base name
                const response = await fetch(`/api/search/pokemon?q=${encodeURIComponent(baseName)}`);
                const data = await response.json();

                // Filter to get only forms of this Pokemon
                const forms = (data.results || []).filter(p => {
                    // Include the base form or forms that start with baseName-
                    return p.name === baseName || p.name.startsWith(baseName + '-');
                });

                if (role === 'attacker') {
                    this.attackerForms = forms;
                } else {
                    this.defenderForms = forms;
                }
            } catch (e) {
                console.error('Failed to load forms:', e);
                if (role === 'attacker') {
                    this.attackerForms = [];
                } else {
                    this.defenderForms = [];
                }
            }
        },

        // Select a form variant
        async selectForm(role, formId) {
            if (!formId) return;

            try {
                const response = await fetch(`/api/pokemon/${formId}`);
                const pokemon = await response.json();

                const target = role === 'attacker' ? this.attacker : this.defender;
                target.species = formId;
                target.speciesData = pokemon;
                target.selectedForm = formId;

                // Update ability to first of new form
                if (pokemon.abilities && pokemon.abilities['0']) {
                    target.ability = pokemon.abilities['0'];
                }

                // Update search display
                if (role === 'attacker') {
                    this.attackerSearch = pokemon.name;
                    // Reload learnset for new form
                    await this.loadAttackerLearnset(formId);
                } else {
                    this.defenderSearch = pokemon.name;
                }
            } catch (e) {
                console.error('Failed to load form:', e);
            }
        },

        // Load attacker's learnset
        async loadAttackerLearnset(pokemonId) {
            try {
                const response = await fetch(`/api/pokemon/${pokemonId}/learnset`);
                const data = await response.json();
                this.attackerLearnset = data.learnset;

                // Pre-fetch move details for all moves in learnset
                await this.prefetchMoveDetails();
            } catch (e) {
                console.error('Failed to load learnset:', e);
                this.attackerLearnset = null;
            }
        },

        // Prefetch move details for learnset
        async prefetchMoveDetails() {
            if (!this.attackerLearnset) return;

            const moveIds = new Set();

            // Collect all move IDs
            if (this.attackerLearnset.levelup) {
                this.attackerLearnset.levelup.forEach(m => moveIds.add(m.move));
            }
            if (this.attackerLearnset.tm) {
                this.attackerLearnset.tm.forEach(m => moveIds.add(m));
            }
            if (this.attackerLearnset.tutor) {
                this.attackerLearnset.tutor.forEach(m => moveIds.add(m));
            }
            if (this.attackerLearnset.egg) {
                this.attackerLearnset.egg.forEach(m => moveIds.add(m));
            }

            // Fetch details for moves not in cache
            const fetchPromises = [];
            for (const moveId of moveIds) {
                if (!this.moveCache[moveId]) {
                    fetchPromises.push(
                        fetch(`/api/moves/${moveId}`)
                            .then(r => r.json())
                            .then(move => { this.moveCache[moveId] = move; })
                            .catch(() => {})
                    );
                }
            }

            // Wait for all fetches (limit concurrency)
            await Promise.all(fetchPromises);
        },

        // Get move details from cache
        getMoveDetails(moveId) {
            return this.moveCache[moveId] || null;
        },

        // Format move for display
        formatMove(moveId, level = null) {
            const move = this.getMoveDetails(moveId);
            if (!move) return moveId;

            const power = move.basePower || '-';
            const type = move.type || '???';
            const category = move.category ? move.category.charAt(0) : '?';

            if (level !== null) {
                return `Lv.${level} ${move.name} (${type}/${category}/${power})`;
            }
            return `${move.name} (${type}/${category}/${power})`;
        },

        // Select move from learnset
        selectLearnsetMove(moveId) {
            const move = this.getMoveDetails(moveId);
            this.move.name = moveId;
            if (move) {
                // Store display name for reference
                this.selectedMoveName = move.name;
            }
        },

        // Get abilities list for a Pokemon
        getAbilitiesList(role) {
            const target = role === 'attacker' ? this.attacker : this.defender;
            if (!target.speciesData || !target.speciesData.abilities) {
                return [];
            }

            const abilities = [];
            const abilityMap = target.speciesData.abilities;

            if (abilityMap['0']) {
                abilities.push({ slot: '0', name: abilityMap['0'], hidden: false });
            }
            if (abilityMap['1']) {
                abilities.push({ slot: '1', name: abilityMap['1'], hidden: false });
            }
            if (abilityMap['H']) {
                abilities.push({ slot: 'H', name: abilityMap['H'], hidden: true });
            }

            return abilities;
        },

        // Swap attacker and defender
        swapPokemon() {
            // Swap Pokemon data
            const tempAttacker = { ...this.attacker };
            const tempDefender = { ...this.defender };

            this.attacker = tempDefender;
            this.defender = tempAttacker;

            // Swap search text
            const tempSearch = this.attackerSearch;
            this.attackerSearch = this.defenderSearch;
            this.defenderSearch = tempSearch;

            // Swap forms
            const tempForms = this.attackerForms;
            this.attackerForms = this.defenderForms;
            this.defenderForms = tempForms;

            // Clear move and reload learnset for new attacker
            this.move.name = '';
            if (this.attacker.species) {
                this.loadAttackerLearnset(this.attacker.species);
            } else {
                this.attackerLearnset = null;
            }

            // Clear result
            this.result = null;
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
        selectedForm: '',
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
