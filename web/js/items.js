function itemsApp() {
    return {
        searchQuery: '',
        allItems: [],
        filteredItems: [],
        loading: true,

        async init() {
            await this.loadItems();
        },

        async loadItems() {
            this.loading = true;
            try {
                const response = await fetch('/api/items');
                this.allItems = await response.json();
                // Sort alphabetically by name
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
