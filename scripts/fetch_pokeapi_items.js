const fs = require('fs');
const path = require('path');
const https = require('https');

const dataDir = path.join(__dirname, '..', 'data');

// Helper to fetch JSON from URL
function fetchJson(url) {
    return new Promise((resolve, reject) => {
        https.get(url, (res) => {
            let data = '';
            res.on('data', chunk => data += chunk);
            res.on('end', () => {
                try {
                    resolve(JSON.parse(data));
                } catch (e) {
                    reject(e);
                }
            });
        }).on('error', reject);
    });
}

// Sleep helper for rate limiting
function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

// Convert PokeAPI item name to ID (rare-candy -> rarecandy)
function toId(name) {
    return name.toLowerCase().replace(/[^a-z0-9]/g, '');
}

// Capitalize item name properly
function formatName(name) {
    return name.split('-').map(word =>
        word.charAt(0).toUpperCase() + word.slice(1)
    ).join(' ');
}

async function fetchAllItems() {
    console.log('Fetching item list from PokeAPI...');

    // Get all item URLs
    const listResponse = await fetchJson('https://pokeapi.co/api/v2/item?limit=3000');
    const itemUrls = listResponse.results;
    console.log(`Found ${itemUrls.length} items`);

    // Load existing Showdown items
    const showdownPath = path.join(dataDir, 'items.json');
    let showdownItems = {};
    if (fs.existsSync(showdownPath)) {
        showdownItems = JSON.parse(fs.readFileSync(showdownPath, 'utf8'));
        console.log(`Loaded ${Object.keys(showdownItems).length} existing Showdown items`);
    }

    const mergedItems = { ...showdownItems };
    let newCount = 0;
    let skipped = 0;

    // Fetch each item
    for (let i = 0; i < itemUrls.length; i++) {
        const itemUrl = itemUrls[i];
        const id = toId(itemUrl.name);

        // Skip if already have from Showdown (it has better battle descriptions)
        if (showdownItems[id]) {
            skipped++;
            continue;
        }

        try {
            // Rate limit: 100 requests per minute max for PokeAPI
            if (i > 0 && i % 50 === 0) {
                console.log(`Progress: ${i}/${itemUrls.length} (${newCount} new, ${skipped} skipped)`);
                await sleep(1000);
            }

            const item = await fetchJson(itemUrl.url);

            // Get English name
            const englishName = item.names.find(n => n.language.name === 'en');
            const name = englishName ? englishName.name : formatName(item.name);

            // Get description (prefer short_effect, fall back to flavor text)
            let desc = '';
            const effectEntry = item.effect_entries.find(e => e.language.name === 'en');
            if (effectEntry) {
                desc = effectEntry.short_effect || effectEntry.effect;
            }
            if (!desc) {
                const flavorText = item.flavor_text_entries.find(f => f.language.name === 'en');
                if (flavorText) {
                    desc = flavorText.text.replace(/\n/g, ' ').replace(/\s+/g, ' ');
                }
            }

            mergedItems[id] = {
                id: id,
                name: name,
                num: item.id,
                desc: desc || ''
            };
            newCount++;

        } catch (e) {
            console.error(`Failed to fetch ${itemUrl.name}: ${e.message}`);
        }
    }

    console.log(`\nMerge complete: ${Object.keys(mergedItems).length} total items`);
    console.log(`  - ${Object.keys(showdownItems).length} from Showdown`);
    console.log(`  - ${newCount} new from PokeAPI`);

    // Save merged items
    fs.writeFileSync(
        showdownPath,
        JSON.stringify(mergedItems, null, 2)
    );
    console.log('Saved to items.json');
}

fetchAllItems().catch(console.error);
