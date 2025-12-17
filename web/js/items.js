function itemsApp() {
    return {
        searchQuery: '',
        allItems: [],
        filteredItems: [],
        loading: true,
        bag: null,

        async init() {
            this.loadBagFromStorage();
            await this.loadItems();
        },

        loadBagFromStorage() {
            try {
                const saved = localStorage.getItem('nuzlocke_party');
                console.log('Loading from localStorage:', saved ? 'found' : 'not found');
                if (saved) {
                    const state = JSON.parse(saved);
                    this.bag = state.bag || null;
                    console.log('Bag data:', this.bag);
                    console.log('hasBagItems:', this.hasBagItems());
                }
            } catch (e) {
                console.error('Failed to load bag state:', e);
            }
        },

        hasBagItems() {
            if (!this.bag) return false;
            return (this.bag.items?.length > 0) ||
                   (this.bag.keyItems?.length > 0) ||
                   (this.bag.pokeBalls?.length > 0) ||
                   (this.bag.tmsHms?.length > 0) ||
                   (this.bag.berries?.length > 0) ||
                   (this.bag.pcItems?.length > 0);
        },

        async loadItems() {
            this.loading = true;
            try {
                const response = await fetch('/api/items');
                this.allItems = await response.json();
                this.allItems.sort((a, b) => a.name.localeCompare(b.name));
            } catch (e) {
                console.error('Failed to load items:', e);
                this.allItems = [];
            }
            this.loading = false;
        },

        searchItems() {
            if (!this.searchQuery || this.searchQuery.length < 1) {
                this.filteredItems = [];
                return;
            }

            const query = this.searchQuery.toLowerCase();
            this.filteredItems = this.allItems.filter(item =>
                item.name.toLowerCase().includes(query)
            );
        }
    };
}
