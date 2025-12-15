function catchApp() {
    return {
        // Search state
        searchQuery: '',
        searchResults: [],
        showResults: false,
        selectedPokemon: null,

        // Catch parameters
        hpPercent: 100,
        status: 'none',
        level: 50,
        turns: 0,
        alreadyCaught: false,
        inCaveOrNight: false,
        underwater: false,
        throwCount: 10,

        // Status multipliers
        statusMultipliers: {
            'none': 1.0,
            'sleep': 2.0,
            'freeze': 2.0,
            'paralysis': 1.5,
            'burn': 1.5,
            'poison': 1.5
        },

        // Ball definitions
        balls: [
            { id: 'pokeball', name: 'Poke Ball', getMultiplier: () => 1.0 },
            { id: 'greatball', name: 'Great Ball', getMultiplier: () => 1.5 },
            { id: 'ultraball', name: 'Ultra Ball', getMultiplier: () => 2.0 },
            { id: 'masterball', name: 'Master Ball', getMultiplier: () => 255 },
            { id: 'premierball', name: 'Premier Ball', getMultiplier: () => 1.0 },
            { id: 'luxuryball', name: 'Luxury Ball', getMultiplier: () => 1.0 },
            {
                id: 'repeatball',
                name: 'Repeat Ball',
                getMultiplier: function() { return this.alreadyCaught ? 3.0 : 1.0; }.bind(this)
            },
            {
                id: 'timerball',
                name: 'Timer Ball',
                getMultiplier: function() { return Math.min(4.0, (this.turns + 10) / 10); }.bind(this)
            },
            {
                id: 'netball',
                name: 'Net Ball',
                getMultiplier: function() {
                    if (!this.selectedPokemon) return 1.0;
                    const types = this.selectedPokemon.types || [];
                    return types.includes('Water') || types.includes('Bug') ? 3.0 : 1.0;
                }.bind(this)
            },
            {
                id: 'nestball',
                name: 'Nest Ball',
                getMultiplier: function() {
                    return Math.max(1.0, Math.min(4.0, (41 - this.level) / 10));
                }.bind(this)
            },
            {
                id: 'diveball',
                name: 'Dive Ball',
                getMultiplier: function() { return this.underwater ? 3.5 : 1.0; }.bind(this)
            },
            {
                id: 'duskball',
                name: 'Dusk Ball',
                getMultiplier: function() { return this.inCaveOrNight ? 3.5 : 1.0; }.bind(this)
            }
        ],

        init() {
            // Bind ball multiplier functions to this context
            this.balls = this.balls.map(ball => ({
                ...ball,
                getMultiplier: ball.getMultiplier.bind(this)
            }));
        },

        async searchPokemon() {
            if (this.searchQuery.length < 2) {
                this.searchResults = [];
                this.showResults = false;
                return;
            }

            try {
                const response = await fetch(`/api/search/pokemon?q=${encodeURIComponent(this.searchQuery)}`);
                const data = await response.json();
                this.searchResults = data.results || [];
                this.showResults = this.searchResults.length > 0;
            } catch (e) {
                console.error('Search failed:', e);
                this.searchResults = [];
            }
        },

        async selectPokemon(result) {
            this.showResults = false;
            this.searchQuery = result.name;

            try {
                const response = await fetch(`/api/pokemon/${result.id}`);
                this.selectedPokemon = await response.json();
            } catch (e) {
                console.error('Failed to load Pokemon:', e);
            }
        },

        getSpriteUrl() {
            if (!this.selectedPokemon) return '';
            const spriteId = this.selectedPokemon.name.toLowerCase().replace(/[^a-z0-9]/g, '');
            return `https://play.pokemonshowdown.com/sprites/gen5/${spriteId}.png`;
        },

        handleSpriteError(event) {
            event.target.src = 'https://play.pokemonshowdown.com/sprites/gen5/0.png';
        },

        getHPBarClass() {
            if (this.hpPercent > 50) return 'hp-green';
            if (this.hpPercent > 20) return 'hp-yellow';
            return 'hp-red';
        },

        getHPZoneText() {
            if (this.hpPercent > 50) return 'Green zone';
            if (this.hpPercent > 20) return 'Yellow zone';
            return 'Red zone';
        },

        getStatusMultiplier() {
            return this.statusMultipliers[this.status] || 1.0;
        },

        // Gen 3 catch rate formula
        calculateCatchProbability(ballMultiplier) {
            if (!this.selectedPokemon || !this.selectedPokemon.catchRate) return 0;

            // Master Ball is always guaranteed
            if (ballMultiplier >= 255) return 1.0;

            const catchRate = this.selectedPokemon.catchRate;
            const statusBonus = this.getStatusMultiplier();

            // HP factor: (3*HPmax - 2*HPcurrent) / (3*HPmax)
            // With percentage: (3 - 2*hpPercent/100) / 3
            const hpFactor = (3 - 2 * this.hpPercent / 100) / 3;

            // Calculate 'a' value
            const a = hpFactor * catchRate * ballMultiplier * statusBonus;

            // If a >= 255, guaranteed catch
            if (a >= 255) return 1.0;

            // Calculate shake check value 'b'
            // b = 1048560 / sqrt(sqrt(16711680 / a))
            const b = Math.floor(1048560 / Math.sqrt(Math.sqrt(16711680 / a)));

            // Probability = (b / 65536)^4
            const prob = Math.pow(b / 65536, 4);

            return Math.min(1.0, prob);
        },

        // Cumulative probability after N throws
        cumulativeProbability(singleProb, throws) {
            if (singleProb >= 1) return 1.0;
            return 1 - Math.pow(1 - singleProb, throws);
        },

        // Get expected number of throws
        expectedThrows(singleProb) {
            if (singleProb >= 1) return 1;
            if (singleProb <= 0) return Infinity;
            return 1 / singleProb;
        },

        getBallResults() {
            if (!this.selectedPokemon) return [];

            return this.balls.map(ball => {
                const multiplier = ball.getMultiplier();
                const perThrow = this.calculateCatchProbability(multiplier);
                const cumulative = this.cumulativeProbability(perThrow, this.throwCount);

                return {
                    id: ball.id,
                    name: ball.name,
                    multiplier: multiplier,
                    perThrow: perThrow,
                    cumulative: cumulative
                };
            });
        },

        getExpectedThrows() {
            if (!this.selectedPokemon) return [];

            // Only show a few key balls
            const keyBalls = ['pokeball', 'greatball', 'ultraball', 'duskball'];

            return this.balls
                .filter(ball => keyBalls.includes(ball.id))
                .map(ball => {
                    const multiplier = ball.getMultiplier();
                    const perThrow = this.calculateCatchProbability(multiplier);
                    const expected = this.expectedThrows(perThrow);

                    return {
                        name: ball.name,
                        expected: expected >= 1000 ? '1000+' :
                                  expected === 1 ? '1 (Guaranteed)' :
                                  '~' + Math.ceil(expected)
                    };
                });
        },

        getCatchClass(probability) {
            if (probability >= 1) return 'catch-guaranteed';
            if (probability >= 0.5) return 'catch-high';
            if (probability >= 0.2) return 'catch-medium';
            if (probability >= 0.05) return 'catch-low';
            return 'catch-very-low';
        }
    };
}
