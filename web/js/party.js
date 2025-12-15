function partyApp() {
    return {
        party: [],
        fileName: '',
        lastFile: null,
        error: '',
        loading: false,

        init() {
            // Nothing to initialize
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
                this.party = result.party || [];

                if (this.party.length === 0) {
                    this.error = 'No party Pokemon found in save file.';
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
