function matchupApp() {
    return {
        // Enemy Pokemon state
        enemy: { species: '', level: 50 },
        enemySearchQuery: '',
        enemySearchResults: [],
        showEnemyResults: false,
        enemyData: null,
        enemyLearnset: [],

        // Config
        generation: 3,
        evMode: 'unknown',  // 'min', 'unknown', 'max'
        ivMode: 'unknown',  // 'min', 'unknown', 'max'

        // Party data
        partyPokemon: [],

        // Results
        matchups: [],
        calculating: false,

        async init() {
            this.loadState();
            this.loadPartyData();
            if (this.enemy.species) {
                await this.fetchEnemyData();
                await this.fetchEnemyLearnset();
                await this.calculateAllMatchups();
            }
        },

        loadState() {
            try {
                const saved = localStorage.getItem('matchup_state');
                if (saved) {
                    const state = JSON.parse(saved);
                    this.enemy = state.enemy || { species: '', level: 50 };
                    this.generation = state.generation || 3;
                    this.evMode = state.evMode || 'unknown';
                    this.ivMode = state.ivMode || 'unknown';
                    this.enemySearchQuery = state.enemySearchQuery || '';
                }
            } catch (e) {
                console.error('Failed to load matchup state:', e);
            }
        },

        saveState() {
            try {
                const state = {
                    enemy: this.enemy,
                    generation: this.generation,
                    evMode: this.evMode,
                    ivMode: this.ivMode,
                    enemySearchQuery: this.enemySearchQuery
                };
                localStorage.setItem('matchup_state', JSON.stringify(state));
            } catch (e) {
                console.error('Failed to save matchup state:', e);
            }
        },

        loadPartyData() {
            try {
                const saved = localStorage.getItem('nuzlocke_party');
                if (saved) {
                    const state = JSON.parse(saved);
                    // Only use party Pokemon, not box Pokemon
                    const party = state.party || [];
                    // Deduplicate by species+level+nickname to avoid display issues
                    const seen = new Set();
                    this.partyPokemon = party.filter(p => {
                        const key = `${p.species}-${p.level}-${p.nickname || ''}`;
                        if (seen.has(key)) return false;
                        seen.add(key);
                        return true;
                    });
                }
            } catch (e) {
                console.error('Failed to load party data:', e);
            }
        },

        async searchEnemy() {
            if (this.enemySearchQuery.length < 2) {
                this.enemySearchResults = [];
                this.showEnemyResults = false;
                return;
            }

            try {
                const response = await fetch(`/api/search/pokemon?q=${encodeURIComponent(this.enemySearchQuery)}`);
                const data = await response.json();
                this.enemySearchResults = data.results || [];
                this.showEnemyResults = this.enemySearchResults.length > 0;
            } catch (e) {
                console.error('Failed to search Pokemon:', e);
            }
        },

        async selectEnemy(result) {
            this.enemy.species = result.id;
            this.enemySearchQuery = result.name;
            this.showEnemyResults = false;
            this.saveState();
            await this.fetchEnemyData();
            await this.fetchEnemyLearnset();
            await this.calculateAllMatchups();
        },

        async fetchEnemyData() {
            if (!this.enemy.species) return;

            try {
                const response = await fetch(`/api/pokemon/${this.enemy.species}`);
                this.enemyData = await response.json();
            } catch (e) {
                console.error('Failed to fetch enemy data:', e);
            }
        },

        async fetchEnemyLearnset() {
            if (!this.enemy.species) return;

            try {
                // Always use Gen 9 for learnset as per requirements
                const response = await fetch(`/api/pokemon/${this.enemy.species}/learnset?gen=9`);
                const data = await response.json();

                // Filter to level-up moves at or below current level, deduplicate
                const levelUpMoves = [...new Set((data.learnset?.levelup || [])
                    .filter(m => m.level <= this.enemy.level)
                    .map(m => m.move))];

                // Fetch move details for each
                this.enemyLearnset = [];
                for (const moveId of levelUpMoves) {
                    try {
                        const moveResponse = await fetch(`/api/moves/${moveId}`);
                        if (!moveResponse.ok) continue;
                        const moveData = await moveResponse.json();
                        // Only add if we got valid data
                        if (moveData && moveData.name && moveData.type) {
                            this.enemyLearnset.push({
                                id: moveId,
                                name: moveData.name,
                                type: moveData.type,
                                category: moveData.category || 'Physical',
                                power: moveData.basePower || 0,
                                accuracy: moveData.accuracy === true ? '-' : (moveData.accuracy || '-'),
                                description: moveData.shortDesc || moveData.desc || ''
                            });
                        }
                    } catch (e) {
                        console.error(`Failed to fetch move ${moveId}:`, e);
                    }
                }
            } catch (e) {
                console.error('Failed to fetch enemy learnset:', e);
            }
        },

        async onEnemyChange() {
            this.saveState();
            if (this.enemy.species) {
                await this.fetchEnemyLearnset();
                await this.calculateAllMatchups();
            }
        },

        async onConfigChange() {
            this.saveState();
            if (this.enemy.species && this.partyPokemon.length > 0) {
                await this.calculateAllMatchups();
            }
        },

        setEvMode(mode) {
            this.evMode = mode;
            this.onConfigChange();
        },

        setIvMode(mode) {
            this.ivMode = mode;
            this.onConfigChange();
        },

        // Get EVs based on current mode
        getEvsForMode(mode) {
            if (mode === 'min') return { hp: 0, atk: 0, def: 0, spa: 0, spd: 0, spe: 0 };
            if (mode === 'max') return { hp: 252, atk: 252, def: 252, spa: 252, spd: 252, spe: 252 };
            return null; // unknown
        },

        // Get IVs based on current mode
        getIvsForMode(mode) {
            if (mode === 'min') return { hp: 0, atk: 0, def: 0, spa: 0, spd: 0, spe: 0 };
            if (mode === 'max') return { hp: 31, atk: 31, def: 31, spa: 31, spd: 31, spe: 31 };
            return null; // unknown
        },

        async calculateAllMatchups() {
            if (!this.enemyData || this.partyPokemon.length === 0) return;
            if (this.calculating) return; // Prevent concurrent calculations

            this.calculating = true;
            this.matchups = [];

            try {
                for (const partyMember of this.partyPokemon) {
                    const matchup = {
                        partyMember: partyMember,
                        yourMoves: [],
                        enemyThreats: []
                    };

                    // Calculate damage for each of party member's moves vs enemy
                    for (const move of (partyMember.moves || [])) {
                        const damage = await this.calculateDamage(
                            partyMember,
                            this.buildEnemyPokemon(),
                            move.name,
                            false // party attacking enemy
                        );
                        matchup.yourMoves.push({
                            name: move.name,
                            type: move.type,
                            damage: damage
                        });
                    }

                    // Calculate damage for enemy's moves vs party member
                    const enemyMoveResults = [];
                    for (const move of this.enemyLearnset) {
                        // Skip status moves and moves with no power (case-insensitive check)
                        if (move.category?.toLowerCase() === 'status' || !move.power) continue;

                        const damage = await this.calculateDamage(
                            this.buildEnemyPokemon(),
                            partyMember,
                            move.id,
                            true // enemy attacking party
                        );
                        enemyMoveResults.push({
                            name: move.name,
                            type: move.type,
                            damage: damage,
                            maxDamage: this.extractMaxDamage(damage)
                        });
                    }

                    // Deduplicate by move name, keeping highest damage version
                    const uniqueMoves = new Map();
                    for (const m of enemyMoveResults) {
                        if (!uniqueMoves.has(m.name) || m.maxDamage > uniqueMoves.get(m.name).maxDamage) {
                            uniqueMoves.set(m.name, m);
                        }
                    }

                    // Sort by max damage and take top 4
                    const sortedMoves = [...uniqueMoves.values()].sort((a, b) => b.maxDamage - a.maxDamage);
                    matchup.enemyThreats = sortedMoves.slice(0, 4).map(m => ({
                        name: m.name,
                        type: m.type,
                        damage: m.damage
                    }));

                    this.matchups.push(matchup);
                }

                // Sort by worst case damage taken (lowest first = safest matchup)
                this.matchups.sort((a, b) => {
                    const aMaxDmg = this.getWorstCaseDamagePercent(a.enemyThreats);
                    const bMaxDmg = this.getWorstCaseDamagePercent(b.enemyThreats);
                    return aMaxDmg - bMaxDmg;
                });
            } catch (e) {
                console.error('Failed to calculate matchups:', e);
            }

            this.calculating = false;
        },

        buildEnemyPokemon(statOverride = null) {
            // Default to min values, scenarios will override as needed
            const evs = statOverride?.evs || this.getEvsForMode(this.evMode) || { hp: 0, atk: 0, def: 0, spa: 0, spd: 0, spe: 0 };
            const ivs = statOverride?.ivs || this.getIvsForMode(this.ivMode) || { hp: 0, atk: 0, def: 0, spa: 0, spd: 0, spe: 0 };

            return {
                species: this.enemy.species,
                level: this.enemy.level,
                evs: evs,
                ivs: ivs,
                nature: 'hardy', // Neutral nature
                ability: '',
                item: '',
                status: ''
            };
        },

        async calculateDamage(attacker, defender, moveName, isEnemyAttacking) {
            // Determine if we need to calculate multiple scenarios
            const scenarios = this.getScenarios(isEnemyAttacking);

            let minResult = null;
            let maxResult = null;

            for (const scenario of scenarios) {
                const request = {
                    generation: this.generation,
                    attacker: this.buildBattlePokemon(attacker, isEnemyAttacking ? scenario : null),
                    defender: this.buildBattlePokemon(defender, isEnemyAttacking ? null : scenario),
                    move: { name: moveName.toLowerCase().replace(/[^a-z0-9]/g, '') },
                    field: {}
                };

                try {
                    const response = await fetch('/api/calculate', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(request)
                    });
                    const result = await response.json();

                    if (!minResult || result.minDamage < minResult.minDamage) {
                        minResult = result;
                    }
                    if (!maxResult || result.maxDamage > maxResult.maxDamage) {
                        maxResult = result;
                    }
                } catch (e) {
                    console.error('Failed to calculate damage:', e);
                }
            }

            if (!minResult || !maxResult) return '--';
            if (minResult.maxDamage === 0) return '--';

            // Format the result
            const minDmg = minResult.minDamage;
            const maxDmg = maxResult.maxDamage;
            const minPct = minResult.minPercent?.toFixed(0) || '?';
            const maxPct = maxResult.maxPercent?.toFixed(0) || '?';

            if (minDmg === maxDmg) {
                return `${minDmg} (${minPct}%)`;
            }
            return `${minDmg}-${maxDmg} (${minPct}-${maxPct}%)`;
        },

        getScenarios(isEnemyAttacking) {
            const minEvs = { hp: 0, atk: 0, def: 0, spa: 0, spd: 0, spe: 0 };
            const maxEvs = { hp: 252, atk: 252, def: 252, spa: 252, spd: 252, spe: 252 };
            const minIvs = { hp: 0, atk: 0, def: 0, spa: 0, spd: 0, spe: 0 };
            const maxIvs = { hp: 31, atk: 31, def: 31, spa: 31, spd: 31, spe: 31 };

            // If both EVs and IVs are known (not 'unknown'), return single scenario
            if (this.evMode !== 'unknown' && this.ivMode !== 'unknown') {
                const evs = this.evMode === 'min' ? minEvs : maxEvs;
                const ivs = this.ivMode === 'min' ? minIvs : maxIvs;
                return [{ evs, ivs }];
            }

            // For unknown stats, calculate min and max scenarios
            // Use min values for min damage, max values for max damage
            const scenarios = [];

            // Min scenario: uses min EVs/IVs for whichever is unknown, or the set value
            const minScenarioEvs = this.evMode === 'unknown' ? minEvs : (this.evMode === 'min' ? minEvs : maxEvs);
            const minScenarioIvs = this.ivMode === 'unknown' ? minIvs : (this.ivMode === 'min' ? minIvs : maxIvs);
            scenarios.push({ evs: minScenarioEvs, ivs: minScenarioIvs });

            // Max scenario: uses max EVs/IVs for whichever is unknown, or the set value
            const maxScenarioEvs = this.evMode === 'unknown' ? maxEvs : (this.evMode === 'min' ? minEvs : maxEvs);
            const maxScenarioIvs = this.ivMode === 'unknown' ? maxIvs : (this.ivMode === 'min' ? minIvs : maxIvs);
            scenarios.push({ evs: maxScenarioEvs, ivs: maxScenarioIvs });

            return scenarios;
        },

        buildBattlePokemon(pokemon, statOverride) {
            const evs = statOverride?.evs || pokemon.evs || { hp: 0, atk: 0, def: 0, spa: 0, spd: 0, spe: 0 };
            const ivs = statOverride?.ivs || pokemon.ivs || { hp: 31, atk: 31, def: 31, spa: 31, spd: 31, spe: 31 };

            // Convert party Pokemon IVs/EVs format to API format
            const formattedEvs = {
                hp: evs.hp ?? evs.HP ?? 0,
                atk: evs.atk ?? evs.attack ?? 0,
                def: evs.def ?? evs.defense ?? 0,
                spa: evs.spa ?? evs.spAtk ?? 0,
                spd: evs.spd ?? evs.spDef ?? 0,
                spe: evs.spe ?? evs.speed ?? 0
            };

            const formattedIvs = {
                hp: ivs.hp ?? ivs.HP ?? 31,
                atk: ivs.atk ?? ivs.attack ?? 31,
                def: ivs.def ?? ivs.defense ?? 31,
                spa: ivs.spa ?? ivs.spAtk ?? 31,
                spd: ivs.spd ?? ivs.spDef ?? 31,
                spe: ivs.spe ?? ivs.speed ?? 31
            };

            return {
                species: pokemon.species?.toLowerCase().replace(/[^a-z0-9]/g, '') || pokemon.species,
                level: pokemon.level || 50,
                nature: pokemon.nature?.toLowerCase() || 'hardy',
                ability: pokemon.ability?.name?.toLowerCase().replace(/[^a-z0-9]/g, '') || '',
                item: pokemon.item?.name?.toLowerCase().replace(/[^a-z0-9]/g, '') || '',
                status: '',
                evs: formattedEvs,
                ivs: formattedIvs,
                boosts: { atk: 0, def: 0, spa: 0, spd: 0, spe: 0 }
            };
        },

        extractMaxDamage(damageStr) {
            if (damageStr === '--') return 0;
            // Extract the max damage number from strings like "45-78 (23-41%)" or "45 (23%)"
            const match = damageStr.match(/(\d+)-?(\d+)?/);
            if (match) {
                return parseInt(match[2] || match[1], 10);
            }
            return 0;
        },

        getWorstCaseDamagePercent(enemyThreats) {
            if (!enemyThreats || enemyThreats.length === 0) return 0;
            // Get the highest damage percentage from the top threat
            // Format is like "23-44 (39-75%)" - we want the max percent (75)
            const topThreat = enemyThreats[0];
            if (!topThreat || topThreat.damage === '--') return 0;
            const match = topThreat.damage.match(/\((\d+)-?(\d+)?%\)/);
            if (match) {
                return parseInt(match[2] || match[1], 10);
            }
            return 0;
        },

        getSpriteUrl(species) {
            if (!species) return '';
            const name = species.toLowerCase().replace(/[^a-z0-9-]/g, '');
            return `https://play.pokemonshowdown.com/sprites/gen5/${name}.png`;
        },

        handleSpriteError(event) {
            event.target.src = 'https://play.pokemonshowdown.com/sprites/gen5/0.png';
        },

        // Navigate to damage calculator with party Pokemon vs enemy
        goToCalculator(partyMember) {
            const minEvs = { hp: 0, atk: 0, def: 0, spa: 0, spd: 0, spe: 0 };
            const maxEvs = { hp: 252, atk: 252, def: 252, spa: 252, spd: 252, spe: 252 };
            const minIvs = { hp: 0, atk: 0, def: 0, spa: 0, spd: 0, spe: 0 };
            const maxIvs = { hp: 31, atk: 31, def: 31, spa: 31, spd: 31, spe: 31 };

            // For the calculator, use min values when unknown (damage calc will show range)
            const evs = this.evMode === 'max' ? maxEvs : minEvs;
            const ivs = this.ivMode === 'max' ? maxIvs : minIvs;

            // Store the matchup data for the calculator to pick up
            const matchupData = {
                attacker: {
                    species: partyMember.species,
                    level: partyMember.level,
                    nature: partyMember.nature,
                    ability: partyMember.ability?.name || '',
                    item: partyMember.item?.name || '',
                    ivs: partyMember.ivs,
                    evs: partyMember.evs,
                    moves: partyMember.moves
                },
                defender: {
                    species: this.enemy.species,
                    level: this.enemy.level,
                    evs: evs,
                    ivs: ivs,
                    evMode: this.evMode,
                    ivMode: this.ivMode
                },
                generation: this.generation
            };
            localStorage.setItem('matchup_to_calculator', JSON.stringify(matchupData));
            window.location.href = '/calculator';
        }
    };
}
