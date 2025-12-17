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
        abilityCache: {},
        itemCache: {},

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

        // Item search state
        attackerItemResults: [],
        defenderItemResults: [],
        showAttackerItemResults: false,
        showDefenderItemResults: false,

        // HP percent (for abilities like Torrent, Blaze, etc.)
        attackerHPPercent: 100,
        defenderHPPercent: 100,

        // Calculation result
        result: null,

        // Data caches
        natures: [],
        allItems: null, // Cached items list

        // Party data (from localStorage, shared with party page)
        partyPokemon: [],

        // Initialize
        async init() {
            // Load natures for dropdown
            try {
                const response = await fetch('/api/natures');
                this.natures = await response.json();
            } catch (e) {
                console.error('Failed to load natures:', e);
            }
            // Preload items for autocomplete and descriptions
            this.loadItems();
            // Load party data from shared localStorage
            this.loadPartyData();

            // Check for matchup-to-calculator data (from clicking matchup card)
            const matchupData = localStorage.getItem('matchup_to_calculator');
            if (matchupData) {
                localStorage.removeItem('matchup_to_calculator');
                // Clear existing state first to ensure clean load
                localStorage.removeItem('calculator_state');
                await this.loadFromMatchup(JSON.parse(matchupData));
            } else {
                // Restore saved state
                await this.loadState();
            }
        },

        // Load data from matchup page
        async loadFromMatchup(data) {
            try {
                // Reset state first
                this.attacker = createDefaultPokemon();
                this.defender = createDefaultPokemon();
                this.attackerForms = [];
                this.defenderForms = [];
                this.attackerLearnset = null;
                this.move = { name: '', isCrit: false };
                this.result = null;

                // Set generation
                this.generation = String(data.generation || 9);

                // Load attacker (party Pokemon)
                if (data.attacker?.species) {
                    const speciesId = data.attacker.species.toLowerCase().replace(/[^a-z0-9]/g, '');
                    const response = await fetch(`/api/pokemon/${speciesId}`);
                    const pokemon = await response.json();

                    this.attacker.species = speciesId;
                    this.attacker.speciesData = pokemon;
                    this.attacker.level = data.attacker.level || 50;
                    this.attacker.nature = data.attacker.nature?.toLowerCase() || 'hardy';
                    this.attacker.ability = data.attacker.ability || pokemon.abilities?.['0'] || '';
                    this.attacker.item = data.attacker.item || '';

                    // IVs and EVs from party data
                    if (data.attacker.ivs) {
                        this.attacker.ivs = {
                            hp: data.attacker.ivs.hp ?? data.attacker.ivs.HP ?? 31,
                            atk: data.attacker.ivs.atk ?? data.attacker.ivs.attack ?? 31,
                            def: data.attacker.ivs.def ?? data.attacker.ivs.defense ?? 31,
                            spa: data.attacker.ivs.spa ?? data.attacker.ivs.spAtk ?? 31,
                            spd: data.attacker.ivs.spd ?? data.attacker.ivs.spDef ?? 31,
                            spe: data.attacker.ivs.spe ?? data.attacker.ivs.speed ?? 31
                        };
                        this.attacker.unknownIvs = false;
                    }
                    if (data.attacker.evs) {
                        this.attacker.evs = {
                            hp: data.attacker.evs.hp ?? data.attacker.evs.HP ?? 0,
                            atk: data.attacker.evs.atk ?? data.attacker.evs.attack ?? 0,
                            def: data.attacker.evs.def ?? data.attacker.evs.defense ?? 0,
                            spa: data.attacker.evs.spa ?? data.attacker.evs.spAtk ?? 0,
                            spd: data.attacker.evs.spd ?? data.attacker.evs.spDef ?? 0,
                            spe: data.attacker.evs.spe ?? data.attacker.evs.speed ?? 0
                        };
                        this.attacker.unknownEvs = false;
                    }

                    this.attackerSearch = pokemon.name;
                    await this.loadAttackerLearnset(speciesId);
                    await this.loadForms('attacker', pokemon.baseSpecies || pokemon.name);

                    // Set first known move if available
                    if (data.attacker.moves?.length > 0) {
                        const moveId = data.attacker.moves[0].name.toLowerCase().replace(/[^a-z0-9]/g, '');
                        this.move.name = moveId;
                    }
                }

                // Load defender (enemy Pokemon)
                if (data.defender?.species) {
                    const speciesId = data.defender.species.toLowerCase().replace(/[^a-z0-9]/g, '');
                    const response = await fetch(`/api/pokemon/${speciesId}`);
                    const pokemon = await response.json();

                    this.defender.species = speciesId;
                    this.defender.speciesData = pokemon;
                    this.defender.level = data.defender.level || 50;
                    this.defender.nature = 'hardy';
                    this.defender.ability = pokemon.abilities?.['0'] || '';
                    this.defender.item = '';

                    // EVs/IVs from matchup config
                    if (data.defender.evs) {
                        this.defender.evs = data.defender.evs;
                    }
                    if (data.defender.ivs) {
                        this.defender.ivs = data.defender.ivs;
                    }
                    this.defender.unknownEvs = data.defender.unknownEvs ?? true;
                    this.defender.unknownIvs = data.defender.unknownIvs ?? true;

                    this.defenderSearch = pokemon.name;
                    await this.loadForms('defender', pokemon.baseSpecies || pokemon.name);
                }

                // Save state after loading
                this.saveState();

            } catch (e) {
                console.error('Failed to load from matchup:', e);
            }
        },

        // Load party data from shared localStorage
        loadPartyData() {
            try {
                const saved = localStorage.getItem('nuzlocke_party');
                if (saved) {
                    const state = JSON.parse(saved);
                    this.partyPokemon = [...(state.party || []), ...(state.boxes || []).flat()];
                }
            } catch (e) {
                console.error('Failed to load party data:', e);
            }
        },

        // Get known moves for the current attacker (if they match a party Pokemon)
        getAttackerKnownMoves() {
            if (!this.attacker.species || !this.partyPokemon.length) {
                return [];
            }
            // Find a party Pokemon with matching species (case-insensitive)
            const attackerSpecies = this.attacker.species.toLowerCase();
            const partyMatch = this.partyPokemon.find(p =>
                p.species && p.species.toLowerCase() === attackerSpecies
            );
            if (partyMatch && partyMatch.moves && partyMatch.moves.length > 0) {
                return partyMatch.moves.map(m => ({
                    id: m.name.toLowerCase().replace(/[^a-z0-9]/g, ''),
                    name: m.name,
                    type: m.type,
                    category: m.category,
                    power: m.power
                }));
            }
            return [];
        },

        // Save calculator state to localStorage
        saveState() {
            try {
                const state = {
                    generation: this.generation,
                    attacker: this.serializePokemon(this.attacker),
                    defender: this.serializePokemon(this.defender),
                    attackerSearch: this.attackerSearch,
                    defenderSearch: this.defenderSearch,
                    move: this.move,
                    field: this.field,
                    attackerHPPercent: this.attackerHPPercent,
                    defenderHPPercent: this.defenderHPPercent
                };
                localStorage.setItem('calculator_state', JSON.stringify(state));
            } catch (e) {
                console.error('Failed to save calculator state:', e);
            }
        },

        // Clear all saved calculator state
        clearState() {
            localStorage.removeItem('calculator_state');
            this.attacker = createDefaultPokemon();
            this.defender = createDefaultPokemon();
            this.attackerSearch = '';
            this.defenderSearch = '';
            this.attackerForms = [];
            this.defenderForms = [];
            this.attackerLearnset = null;
            this.move = { name: '', isCrit: false };
            this.field = {
                weather: '',
                terrain: '',
                attackerSide: { helpingHand: false },
                defenderSide: { reflect: false, lightScreen: false }
            };
            this.attackerHPPercent = 100;
            this.defenderHPPercent = 100;
            this.result = null;
        },

        // Serialize Pokemon state (excluding non-serializable speciesData)
        serializePokemon(pokemon) {
            return {
                species: pokemon.species,
                selectedForm: pokemon.selectedForm,
                level: pokemon.level,
                nature: pokemon.nature,
                ability: pokemon.ability,
                item: pokemon.item,
                status: pokemon.status,
                evs: pokemon.evs,
                ivs: pokemon.ivs,
                boosts: pokemon.boosts,
                unknownEvs: pokemon.unknownEvs,
                unknownIvs: pokemon.unknownIvs
            };
        },

        // Load calculator state from localStorage
        async loadState() {
            try {
                const saved = localStorage.getItem('calculator_state');
                if (!saved) return;

                const state = JSON.parse(saved);

                // Restore generation
                if (state.generation) this.generation = state.generation;

                // Restore field settings
                if (state.field) this.field = state.field;

                // Restore HP percents
                if (state.attackerHPPercent !== undefined) this.attackerHPPercent = state.attackerHPPercent;
                if (state.defenderHPPercent !== undefined) this.defenderHPPercent = state.defenderHPPercent;

                // Restore attacker
                if (state.attacker && state.attacker.species) {
                    await this.restorePokemon('attacker', state.attacker, state.attackerSearch);
                }

                // Restore defender
                if (state.defender && state.defender.species) {
                    await this.restorePokemon('defender', state.defender, state.defenderSearch);
                }

                // Restore move after attacker is loaded
                if (state.move) this.move = state.move;

            } catch (e) {
                console.error('Failed to load calculator state:', e);
            }
        },

        // Restore a Pokemon from saved state
        async restorePokemon(role, savedPokemon, searchText) {
            try {
                const response = await fetch(`/api/pokemon/${savedPokemon.species}`);
                const pokemon = await response.json();

                const target = role === 'attacker' ? this.attacker : this.defender;
                target.species = savedPokemon.species;
                target.speciesData = pokemon;
                target.selectedForm = savedPokemon.selectedForm || '';
                target.level = savedPokemon.level;
                target.nature = savedPokemon.nature;
                target.ability = savedPokemon.ability;
                target.item = savedPokemon.item;
                target.status = savedPokemon.status;
                target.evs = savedPokemon.evs;
                target.ivs = savedPokemon.ivs;
                target.boosts = savedPokemon.boosts;
                target.unknownEvs = savedPokemon.unknownEvs;
                target.unknownIvs = savedPokemon.unknownIvs;

                if (role === 'attacker') {
                    this.attackerSearch = searchText || pokemon.name;
                    await this.loadAttackerLearnset(savedPokemon.species);
                    await this.loadForms('attacker', pokemon.baseSpecies || pokemon.name);
                } else {
                    this.defenderSearch = searchText || pokemon.name;
                    await this.loadForms('defender', pokemon.baseSpecies || pokemon.name);
                }
            } catch (e) {
                console.error('Failed to restore Pokemon:', e);
            }
        },

        // Import Pokemon from party
        async importFromParty(role, partyPokemon) {
            try {
                console.log('Importing from party:', partyPokemon); // Debug
                console.log('Party Pokemon IVs:', partyPokemon.ivs); // Debug
                console.log('Party Pokemon EVs:', partyPokemon.evs); // Debug

                // Find species ID
                const speciesId = partyPokemon.species.toLowerCase().replace(/[^a-z0-9]/g, '');
                const response = await fetch(`/api/pokemon/${speciesId}`);
                const pokemon = await response.json();

                const target = role === 'attacker' ? this.attacker : this.defender;
                target.species = speciesId;
                target.speciesData = pokemon;
                target.selectedForm = '';
                target.level = partyPokemon.level;
                target.nature = partyPokemon.nature.toLowerCase();
                target.ability = partyPokemon.ability?.name || '';
                target.item = partyPokemon.item?.name || '';
                target.status = '';

                // Import IVs and EVs
                target.ivs = {
                    hp: partyPokemon.ivs?.hp ?? 31,
                    atk: partyPokemon.ivs?.attack ?? 31,
                    def: partyPokemon.ivs?.defense ?? 31,
                    spa: partyPokemon.ivs?.spAtk ?? 31,
                    spd: partyPokemon.ivs?.spDef ?? 31,
                    spe: partyPokemon.ivs?.speed ?? 31
                };
                target.evs = {
                    hp: partyPokemon.evs?.hp ?? 0,
                    atk: partyPokemon.evs?.attack ?? 0,
                    def: partyPokemon.evs?.defense ?? 0,
                    spa: partyPokemon.evs?.spAtk ?? 0,
                    spd: partyPokemon.evs?.spDef ?? 0,
                    spe: partyPokemon.evs?.speed ?? 0
                };
                console.log('Imported IVs:', target.ivs); // Debug
                console.log('Imported EVs:', target.evs); // Debug

                target.boosts = { atk: 0, def: 0, spa: 0, spd: 0, spe: 0 };
                target.unknownEvs = false;
                target.unknownIvs = false;

                if (role === 'attacker') {
                    this.attackerSearch = partyPokemon.species;
                    await this.loadAttackerLearnset(speciesId);
                    await this.loadForms('attacker', pokemon.baseSpecies || pokemon.name);

                    // Set first known move as default
                    const knownMoves = this.getAttackerKnownMoves();
                    if (knownMoves.length > 0) {
                        this.move.name = knownMoves[0].id;
                    }
                } else {
                    this.defenderSearch = partyPokemon.species;
                    await this.loadForms('defender', pokemon.baseSpecies || pokemon.name);
                }

                // Save state after import
                this.saveState();

            } catch (e) {
                console.error('Failed to import from party:', e);
            }
        },

        // Get sprite URL for a Pokemon
        getSpriteUrl(role) {
            const target = role === 'attacker' ? this.attacker : this.defender;
            if (!target.speciesData) return '';
            const name = target.speciesData.name.toLowerCase()
                .replace(/[^a-z0-9-]/g, '')
                .replace(/^(.+)-mega(-[xy])?$/, 'mega$1$2')  // Handle mega forms
                .replace(/^(.+)-alola$/, '$1-alola')
                .replace(/^(.+)-galar$/, '$1-galar');
            return `https://play.pokemonshowdown.com/sprites/gen5/${name}.png`;
        },

        // Handle sprite load errors
        handleSpriteError(event) {
            event.target.src = 'https://play.pokemonshowdown.com/sprites/gen5/0.png';
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

                // Save state after selection
                this.saveState();
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

                // Save state after form change
                this.saveState();
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

        // Sort level-up moves by level
        sortedLevelUpMoves() {
            if (!this.attackerLearnset?.levelup) return [];
            return [...this.attackerLearnset.levelup].sort((a, b) => a.level - b.level);
        },

        // Sort TM moves alphabetically by move name
        sortedTMMoves() {
            if (!this.attackerLearnset?.tm) return [];
            return [...this.attackerLearnset.tm].sort((a, b) => {
                const nameA = this.getMoveDetails(a)?.name || a;
                const nameB = this.getMoveDetails(b)?.name || b;
                return nameA.localeCompare(nameB);
            });
        },

        // Sort tutor moves alphabetically by move name
        sortedTutorMoves() {
            if (!this.attackerLearnset?.tutor) return [];
            return [...this.attackerLearnset.tutor].sort((a, b) => {
                const nameA = this.getMoveDetails(a)?.name || a;
                const nameB = this.getMoveDetails(b)?.name || b;
                return nameA.localeCompare(nameB);
            });
        },

        // Sort egg moves alphabetically by move name
        sortedEggMoves() {
            if (!this.attackerLearnset?.egg) return [];
            return [...this.attackerLearnset.egg].sort((a, b) => {
                const nameA = this.getMoveDetails(a)?.name || a;
                const nameB = this.getMoveDetails(b)?.name || b;
                return nameA.localeCompare(nameB);
            });
        },

        // Get move details from cache
        getMoveDetails(moveId) {
            return this.moveCache[moveId] || null;
        },

        // Fetch and cache ability data
        async fetchAbility(abilityName) {
            if (!abilityName) return null;
            const key = abilityName.toLowerCase().replace(/\s+/g, '');
            if (this.abilityCache[key]) {
                return this.abilityCache[key];
            }
            try {
                const response = await fetch(`/api/abilities/${key}`);
                if (response.ok) {
                    const ability = await response.json();
                    this.abilityCache[key] = ability;
                    return ability;
                }
            } catch (e) {
                console.error('Failed to fetch ability:', e);
            }
            return null;
        },

        // Get ability description (from cache or fetch)
        getAbilityDescription(role) {
            const target = role === 'attacker' ? this.attacker : this.defender;
            if (!target.ability) return '';
            const key = target.ability.toLowerCase().replace(/\s+/g, '');
            const cached = this.abilityCache[key];
            if (cached) {
                return cached.shortDesc || cached.desc || '';
            }
            // Trigger fetch if not cached
            this.fetchAbility(target.ability);
            return '';
        },

        // Get item description
        getItemDescription(role) {
            const target = role === 'attacker' ? this.attacker : this.defender;
            if (!target.item || !this.allItems) return '';

            // Find item in cached list
            const itemLower = target.item.toLowerCase();
            const item = this.allItems.find(i => i.name.toLowerCase() === itemLower);
            return item?.desc || '';
        },

        // Load and cache all items
        async loadItems() {
            if (this.allItems) return this.allItems;
            try {
                const response = await fetch('/api/items');
                this.allItems = await response.json();
                return this.allItems;
            } catch (e) {
                console.error('Failed to load items:', e);
                return [];
            }
        },

        // Search items for autocomplete
        async searchItems(role) {
            const target = role === 'attacker' ? this.attacker : this.defender;
            const query = target.item;

            if (!query || query.length < 2) {
                if (role === 'attacker') {
                    this.attackerItemResults = [];
                    this.showAttackerItemResults = false;
                } else {
                    this.defenderItemResults = [];
                    this.showDefenderItemResults = false;
                }
                return;
            }

            const items = await this.loadItems();

            // Filter items that match the query
            const queryLower = query.toLowerCase();
            const results = items
                .filter(item => item.name.toLowerCase().includes(queryLower))
                .slice(0, 8); // Limit results

            if (role === 'attacker') {
                this.attackerItemResults = results;
                this.showAttackerItemResults = results.length > 0;
            } else {
                this.defenderItemResults = results;
                this.showDefenderItemResults = results.length > 0;
            }
        },

        // Select an item from autocomplete
        selectItem(role, item) {
            const target = role === 'attacker' ? this.attacker : this.defender;
            target.item = item.name;
            if (role === 'attacker') {
                this.showAttackerItemResults = false;
            } else {
                this.showDefenderItemResults = false;
            }
            this.saveState();
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
            // Save state after move selection
            this.saveState();
        },

        // Set all IVs to a value
        setAllIvs(role, value) {
            const target = role === 'attacker' ? this.attacker : this.defender;
            target.ivs = { hp: value, atk: value, def: value, spa: value, spd: value, spe: value };
        },

        // Set all EVs to a value
        setAllEvs(role, value) {
            const target = role === 'attacker' ? this.attacker : this.defender;
            target.evs = { hp: value, atk: value, def: value, spa: value, spd: value, spe: value };
        },

        // Set nature by stat boost
        setNature(role, boostStat) {
            const target = role === 'attacker' ? this.attacker : this.defender;
            // Common competitive natures for each boost
            const natureMap = {
                'atk': 'adamant',    // +Atk -SpA
                'spa': 'modest',     // +SpA -Atk
                'def': 'bold',       // +Def -Atk (for special attackers) or impish for physical
                'spd': 'calm',       // +SpD -Atk (for special attackers) or careful for physical
                'spe': 'jolly'       // +Spe -SpA
            };
            target.nature = natureMap[boostStat] || 'hardy';
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

        // Check if any unknowns are set
        hasUnknowns() {
            return this.attacker.unknownEvs || this.attacker.unknownIvs ||
                   this.defender.unknownEvs || this.defender.unknownIvs;
        },

        // Build Pokemon request with specific EV/IV overrides
        buildPokemonRequest(pokemon, evOverride = null, ivOverride = null) {
            return {
                species: pokemon.species,
                level: pokemon.level,
                nature: pokemon.nature,
                ability: pokemon.ability,
                item: pokemon.item,
                status: pokemon.status,
                evs: evOverride || pokemon.evs,
                ivs: ivOverride || pokemon.ivs,
                boosts: pokemon.boosts
            };
        },

        // Perform calculation
        async calculate() {
            if (!this.canCalculate()) return;

            // If no unknowns, do a single calculation
            if (!this.hasUnknowns()) {
                await this.calculateSingle();
                return;
            }

            // With unknowns, calculate best and worst case scenarios
            await this.calculateWithUnknowns();
        },

        // Single calculation (no unknowns)
        async calculateSingle() {
            const request = {
                generation: parseInt(this.generation),
                attacker: {
                    ...this.buildPokemonRequest(this.attacker),
                    currentHPPercent: this.attackerHPPercent
                },
                defender: {
                    ...this.buildPokemonRequest(this.defender),
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
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(request)
                });

                if (!response.ok) {
                    console.error('Calculation failed:', await response.text());
                    return;
                }

                this.result = await response.json();
                this.result.hasRange = false;
            } catch (e) {
                console.error('Calculation failed:', e);
            }
        },

        // Calculate with unknowns - compute best and worst case
        async calculateWithUnknowns() {
            // Get the move to determine which stats matter
            const moveData = this.getMoveDetails(this.move.name);
            const isPhysical = moveData?.category === 'Physical';
            const attackStat = isPhysical ? 'atk' : 'spa';
            const defenseStat = isPhysical ? 'def' : 'spd';

            // Build EV/IV scenarios for attacker
            const attackerScenarios = this.getStatScenarios(this.attacker, attackStat, true);

            // Build EV/IV scenarios for defender (need HP + defense stat)
            const defenderScenarios = this.getStatScenarios(this.defender, defenseStat, false);

            // Calculate all combinations and find extremes
            let bestResult = null;  // Highest damage
            let worstResult = null; // Lowest damage

            for (const atkScenario of attackerScenarios) {
                for (const defScenario of defenderScenarios) {
                    const request = {
                        generation: parseInt(this.generation),
                        attacker: {
                            ...this.buildPokemonRequest(this.attacker, atkScenario.evs, atkScenario.ivs),
                            currentHPPercent: this.attackerHPPercent
                        },
                        defender: {
                            ...this.buildPokemonRequest(this.defender, defScenario.evs, defScenario.ivs),
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
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify(request)
                        });

                        if (!response.ok) continue;

                        const result = await response.json();

                        if (!bestResult || result.maxPercent > bestResult.maxPercent) {
                            bestResult = result;
                        }
                        if (!worstResult || result.minPercent < worstResult.minPercent) {
                            worstResult = result;
                        }
                    } catch (e) {
                        console.error('Calculation failed:', e);
                    }
                }
            }

            // Combine results into a range result
            if (bestResult && worstResult) {
                this.result = {
                    hasRange: true,
                    // Best case (highest damage)
                    bestMinDamage: bestResult.minDamage,
                    bestMaxDamage: bestResult.maxDamage,
                    bestMinPercent: bestResult.minPercent,
                    bestMaxPercent: bestResult.maxPercent,
                    // Worst case (lowest damage)
                    worstMinDamage: worstResult.minDamage,
                    worstMaxDamage: worstResult.maxDamage,
                    worstMinPercent: worstResult.minPercent,
                    worstMaxPercent: worstResult.maxPercent,
                    // For compatibility with existing display
                    minDamage: worstResult.minDamage,
                    maxDamage: bestResult.maxDamage,
                    minPercent: worstResult.minPercent,
                    maxPercent: bestResult.maxPercent,
                    description: bestResult.description
                };
            }
        },

        // Get EV/IV scenarios for a stat
        getStatScenarios(pokemon, stat, isAttacker) {
            const scenarios = [];

            // If nothing is unknown, just return current values
            if (!pokemon.unknownEvs && !pokemon.unknownIvs) {
                scenarios.push({ evs: pokemon.evs, ivs: pokemon.ivs });
                return scenarios;
            }

            // Generate min and max scenarios
            const minEvs = { ...pokemon.evs };
            const maxEvs = { ...pokemon.evs };
            const minIvs = { ...pokemon.ivs };
            const maxIvs = { ...pokemon.ivs };

            if (pokemon.unknownEvs) {
                // For attacker: 0 EV is worst, 252 is best
                // For defender: 0 EV is worst (takes more damage), 252 is best (takes less)
                minEvs[stat] = 0;
                maxEvs[stat] = 252;
                // Also consider HP for defender
                if (!isAttacker) {
                    minEvs.hp = 0;
                    maxEvs.hp = 252;
                }
            }

            if (pokemon.unknownIvs) {
                minIvs[stat] = 0;
                maxIvs[stat] = 31;
                if (!isAttacker) {
                    minIvs.hp = 0;
                    maxIvs.hp = 31;
                }
            }

            // For attacker: min scenario = min EVs/IVs, max scenario = max EVs/IVs
            // For defender: min scenario (more damage taken) = min EVs/IVs, max scenario (less damage taken) = max EVs/IVs
            scenarios.push({ evs: minEvs, ivs: minIvs });
            scenarios.push({ evs: maxEvs, ivs: maxIvs });

            return scenarios;
        },

        // Get critical hit chance based on generation
        getCritChance() {
            // Gen 6+: 1/24 (4.17%), Gen 2-5: 1/16 (6.25%), Gen 1: varies
            if (this.generation === '3') {
                return '6.25%';
            }
            return '4.17%';
        },

        // Calculate speed stat for a Pokemon
        calculateSpeed(pokemon, evOverride = null, ivOverride = null) {
            if (!pokemon.speciesData) return 0;

            const baseSpe = pokemon.speciesData.baseStats?.spe || 0;
            const level = pokemon.level || 100;
            const ev = evOverride !== null ? evOverride : (pokemon.evs?.spe || 0);
            const iv = ivOverride !== null ? ivOverride : (pokemon.ivs?.spe || 31);

            // Base stat calculation
            let speed = Math.floor((Math.floor((2 * baseSpe + iv + Math.floor(ev / 4)) * level / 100) + 5));

            // Nature modifier
            const nature = pokemon.nature?.toLowerCase() || 'hardy';
            const speedBoostNatures = ['jolly', 'timid', 'hasty', 'naive'];
            const speedDropNatures = ['brave', 'relaxed', 'quiet', 'sassy'];
            if (speedBoostNatures.includes(nature)) {
                speed = Math.floor(speed * 1.1);
            } else if (speedDropNatures.includes(nature)) {
                speed = Math.floor(speed * 0.9);
            }

            // Speed boost stages
            const boost = pokemon.boosts?.spe || 0;
            if (boost > 0) {
                speed = Math.floor(speed * (2 + boost) / 2);
            } else if (boost < 0) {
                speed = Math.floor(speed * 2 / (2 - boost));
            }

            // Item modifiers
            const item = pokemon.item?.toLowerCase() || '';
            if (item === 'choice scarf') {
                speed = Math.floor(speed * 1.5);
            } else if (item === 'iron ball' || item === 'macho brace' || item === 'power weight' ||
                       item === 'power bracer' || item === 'power belt' || item === 'power lens' ||
                       item === 'power band' || item === 'power anklet') {
                speed = Math.floor(speed * 0.5);
            }

            // Paralysis halves speed in Gen 3 (quarters in later gens but we use Gen 3 formula)
            if (pokemon.status === 'par') {
                speed = Math.floor(speed * 0.25); // Gen 3 is 25%
            }

            return speed;
        },

        // Get speed comparison result
        getSpeedComparison() {
            if (!this.attacker.speciesData || !this.defender.speciesData) {
                return null;
            }

            // Check for priority items
            const attackerItem = this.attacker.item?.toLowerCase() || '';
            const defenderItem = this.defender.item?.toLowerCase() || '';

            let attackerQuickClaw = attackerItem === 'quick claw';
            let defenderQuickClaw = defenderItem === 'quick claw';

            // Handle unknown EVs/IVs
            const attackerUnknown = this.attacker.unknownEvs || this.attacker.unknownIvs;
            const defenderUnknown = this.defender.unknownEvs || this.defender.unknownIvs;

            if (!attackerUnknown && !defenderUnknown) {
                // Both known - simple comparison
                const attackerSpeed = this.calculateSpeed(this.attacker);
                const defenderSpeed = this.calculateSpeed(this.defender);

                return {
                    attackerSpeed,
                    defenderSpeed,
                    attackerFirst: attackerSpeed > defenderSpeed,
                    defenderFirst: defenderSpeed > attackerSpeed,
                    speedTie: attackerSpeed === defenderSpeed,
                    attackerQuickClaw,
                    defenderQuickClaw,
                    isRange: false
                };
            }

            // Calculate speed ranges for unknown stats
            const attackerMinEv = this.attacker.unknownEvs ? 0 : (this.attacker.evs?.spe || 0);
            const attackerMaxEv = this.attacker.unknownEvs ? 252 : (this.attacker.evs?.spe || 0);
            const attackerMinIv = this.attacker.unknownIvs ? 0 : (this.attacker.ivs?.spe || 31);
            const attackerMaxIv = this.attacker.unknownIvs ? 31 : (this.attacker.ivs?.spe || 31);

            const defenderMinEv = this.defender.unknownEvs ? 0 : (this.defender.evs?.spe || 0);
            const defenderMaxEv = this.defender.unknownEvs ? 252 : (this.defender.evs?.spe || 0);
            const defenderMinIv = this.defender.unknownIvs ? 0 : (this.defender.ivs?.spe || 31);
            const defenderMaxIv = this.defender.unknownIvs ? 31 : (this.defender.ivs?.spe || 31);

            const attackerMinSpeed = this.calculateSpeed(this.attacker, attackerMinEv, attackerMinIv);
            const attackerMaxSpeed = this.calculateSpeed(this.attacker, attackerMaxEv, attackerMaxIv);
            const defenderMinSpeed = this.calculateSpeed(this.defender, defenderMinEv, defenderMinIv);
            const defenderMaxSpeed = this.calculateSpeed(this.defender, defenderMaxEv, defenderMaxIv);

            // Determine outcomes
            const alwaysAttackerFirst = attackerMinSpeed > defenderMaxSpeed;
            const alwaysDefenderFirst = defenderMinSpeed > attackerMaxSpeed;
            const canTie = !(alwaysAttackerFirst || alwaysDefenderFirst);

            return {
                attackerSpeed: attackerUnknown ? `${attackerMinSpeed}-${attackerMaxSpeed}` : attackerMinSpeed,
                defenderSpeed: defenderUnknown ? `${defenderMinSpeed}-${defenderMaxSpeed}` : defenderMinSpeed,
                attackerMinSpeed,
                attackerMaxSpeed,
                defenderMinSpeed,
                defenderMaxSpeed,
                alwaysAttackerFirst,
                alwaysDefenderFirst,
                canTie,
                attackerQuickClaw,
                defenderQuickClaw,
                isRange: true
            };
        },

        // Get formatted speed comparison text
        getSpeedText() {
            const comparison = this.getSpeedComparison();
            if (!comparison) return '';

            if (!comparison.isRange) {
                if (comparison.speedTie) {
                    return 'Speed tie (50/50)';
                } else if (comparison.attackerFirst) {
                    return 'Attacker moves first';
                } else {
                    return 'Defender moves first';
                }
            } else {
                if (comparison.alwaysAttackerFirst) {
                    return 'Attacker always moves first';
                } else if (comparison.alwaysDefenderFirst) {
                    return 'Defender always moves first';
                } else {
                    return 'Speed depends on EVs/IVs';
                }
            }
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
        },

        // Get KO info for range results
        getRangeKO(type) {
            if (!this.result) return '-';

            // For range results, we need to estimate HP from damage percentages
            // Best case uses best damage, worst case uses worst damage
            if (type === 'best') {
                const damage = this.result.bestMaxDamage;
                if (!damage) return '-';
                // Estimate HP from percentage
                const hp = Math.round(damage / this.result.bestMaxPercent * 100);
                const currentHP = Math.floor(hp * this.defenderHPPercent / 100);
                if (currentHP <= 0) return '-';
                const hitsNeeded = Math.ceil(currentHP / damage);
                return hitsNeeded === 1 ? 'OHKO' : hitsNeeded + 'HKO';
            } else {
                const damage = this.result.worstMinDamage;
                if (!damage) return '-';
                // Estimate HP from percentage
                const hp = Math.round(damage / this.result.worstMinPercent * 100);
                const currentHP = Math.floor(hp * this.defenderHPPercent / 100);
                if (currentHP <= 0) return '-';
                const hitsNeeded = Math.ceil(currentHP / damage);
                return hitsNeeded === 1 ? 'OHKO' : hitsNeeded + 'HKO';
            }
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
        boosts: { atk: 0, def: 0, spa: 0, spd: 0, spe: 0 },
        unknownEvs: false,
        unknownIvs: false
    };
}
