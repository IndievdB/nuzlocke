function partyApp() {
    return {
        party: [],
        boxes: [], // 14 boxes, each an array of Pokemon
        bag: null, // Bag pockets with items
        fileName: '',
        lastFile: null,
        error: '',
        loading: false,

        init() {
            this.loadState();
        },

        loadState() {
            try {
                const saved = localStorage.getItem('nuzlocke_party');
                if (saved) {
                    const state = JSON.parse(saved);
                    this.party = state.party || [];
                    this.boxes = state.boxes || [];
                    this.bag = state.bag || null;
                    this.fileName = state.fileName || '';
                }
            } catch (e) {
                console.error('Failed to load party state:', e);
            }
        },

        saveState() {
            try {
                const state = {
                    party: this.party,
                    boxes: this.boxes,
                    bag: this.bag,
                    fileName: this.fileName
                };
                localStorage.setItem('nuzlocke_party', JSON.stringify(state));
            } catch (e) {
                console.error('Failed to save party state:', e);
            }
        },

        discardParty() {
            this.party = [];
            this.boxes = [];
            this.bag = null;
            this.fileName = '';
            this.lastFile = null;
            localStorage.removeItem('nuzlocke_party');
        },

        hasBoxPokemon() {
            return this.boxes && this.boxes.some(box => box && box.length > 0);
        },

        getBoxCount(box) {
            return box ? box.length : 0;
        },

        async handleFileUpload(event) {
            const file = event.target.files[0];
            if (!file) return;

            this.lastFile = file;
            await this.parseFile(file);
        },

        async refreshFile() {
            if (!this.lastFile) return;
            await this.parseFile(this.lastFile);
        },

        async parseFile(file) {
            this.fileName = file.name;
            this.error = '';
            this.loading = true;
            this.party = [];
            this.boxes = [];
            this.bag = null;

            try {
                const arrayBuffer = await file.arrayBuffer();
                const response = await fetch('/api/party/parse', {
                    method: 'POST',
                    body: arrayBuffer,
                    headers: {
                        'Content-Type': 'application/octet-stream'
                    }
                });

                if (!response.ok) {
                    const errorText = await response.text();
                    throw new Error(errorText);
                }

                const result = await response.json();
                console.log('Parsed save data:', result);
                this.party = result.party || [];
                this.boxes = result.boxes || [];
                this.bag = result.bag || null;

                // Count total box Pokemon
                const boxCount = this.boxes.reduce((sum, box) => sum + (box ? box.length : 0), 0);
                console.log(`Found ${this.party.length} party Pokemon and ${boxCount} box Pokemon`);
                if (this.bag) {
                    const bagCount = (this.bag.items?.length || 0) + (this.bag.keyItems?.length || 0) +
                        (this.bag.pokeBalls?.length || 0) + (this.bag.tmsHms?.length || 0) +
                        (this.bag.berries?.length || 0) + (this.bag.pcItems?.length || 0);
                    console.log(`Found ${bagCount} unique items in bag`);
                }

                if (this.party.length === 0) {
                    this.error = 'No party Pokemon found in save file.';
                } else {
                    this.saveState();
                }
            } catch (e) {
                console.error('Failed to parse save file:', e);
                this.error = 'Failed to parse save file: ' + e.message;
            } finally {
                this.loading = false;
            }
        },

        getSpriteUrl(species) {
            if (!species) return '';
            // Convert species name to sprite filename (lowercase, no spaces/special chars)
            const spriteId = species.toLowerCase().replace(/[^a-z0-9]/g, '');
            return `https://play.pokemonshowdown.com/sprites/gen5/${spriteId}.png`;
        },

        handleSpriteError(event) {
            // Fallback to a placeholder or default sprite
            event.target.src = 'https://play.pokemonshowdown.com/sprites/gen5/0.png';
        },

        getNatureTooltip(natureEffect) {
            if (!natureEffect || (!natureEffect.plus && !natureEffect.minus)) {
                return 'Neutral nature (no stat changes)';
            }
            return `+${natureEffect.plus} / -${natureEffect.minus}`;
        }
    };
}
