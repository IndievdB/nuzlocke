#!/usr/bin/env node
/**
 * Convert Pokemon Showdown TypeScript data files to JSON
 * Usage: node convert_showdown_data.js
 */

const fs = require('fs');
const path = require('path');

const dataDir = path.join(__dirname, '..', 'data');

/**
 * Parse a TypeScript data file and extract the object
 */
function parseDataFile(content, varName) {
    // Remove TypeScript type annotations
    // Match: export const VarName: SomeType = { ... }
    const exportRegex = new RegExp(`export\\s+const\\s+${varName}\\s*:\\s*[^=]+=\\s*`);
    content = content.replace(exportRegex, '');

    // Remove trailing semicolon
    content = content.trim();
    if (content.endsWith(';')) {
        content = content.slice(0, -1);
    }

    // The content should now be a valid JavaScript object literal
    // We need to evaluate it safely
    try {
        // Use Function constructor to evaluate (safer than eval)
        const fn = new Function(`return ${content}`);
        return fn();
    } catch (e) {
        console.error(`Error parsing ${varName}:`, e.message);
        return null;
    }
}

/**
 * Convert natures.ts to natures.json
 */
function convertNatures() {
    const content = fs.readFileSync(path.join(dataDir, 'natures.ts'), 'utf8');
    const data = parseDataFile(content, 'Natures');
    if (data) {
        fs.writeFileSync(
            path.join(dataDir, 'natures.json'),
            JSON.stringify(data, null, 2)
        );
        console.log('Converted natures.json');
    }
}

/**
 * Convert typechart.ts to typechart.json
 */
function convertTypechart() {
    const content = fs.readFileSync(path.join(dataDir, 'typechart.ts'), 'utf8');
    const data = parseDataFile(content, 'TypeChart');
    if (data) {
        fs.writeFileSync(
            path.join(dataDir, 'typechart.json'),
            JSON.stringify(data, null, 2)
        );
        console.log('Converted typechart.json');
    }
}

/**
 * Convert items.ts to items.json
 * Items file has more complex structure with functions - extract just the data
 */
function convertItems() {
    const content = fs.readFileSync(path.join(dataDir, 'items.ts'), 'utf8');

    // Load descriptions from text file
    const descriptions = parseItemDescriptions();

    // Parse item entries - extract key properties only
    const items = {};

    // Match item blocks: itemname: { ... }
    const itemRegex = /\t(\w+):\s*\{([^}]+(?:\{[^}]*\}[^}]*)*)\}/g;
    let match;

    while ((match = itemRegex.exec(content)) !== null) {
        const [, id, body] = match;
        const item = { id };

        // Extract simple properties
        // Handle names with apostrophes - match based on delimiter type
        const nameMatchDouble = body.match(/name:\s*"([^"]+)"/);
        const nameMatchSingle = body.match(/name:\s*'([^']+)'/);
        if (nameMatchDouble) item.name = nameMatchDouble[1];
        else if (nameMatchSingle) item.name = nameMatchSingle[1];

        const numMatch = body.match(/num:\s*(\d+)/);
        if (numMatch) item.num = parseInt(numMatch[1]);

        const genMatch = body.match(/gen:\s*(\d+)/);
        if (genMatch) item.gen = parseInt(genMatch[1]);

        const descMatch = body.match(/desc:\s*["']([^"']+)["']/);
        if (descMatch) item.desc = descMatch[1];

        const flingPowerMatch = body.match(/fling:\s*\{\s*basePower:\s*(\d+)/);
        if (flingPowerMatch) item.flingBasePower = parseInt(flingPowerMatch[1]);

        // Check for key item effects
        if (body.includes('onModifyDamage')) item.onModifyDamage = true;
        if (body.includes('onModifyAtk')) item.onModifyAtk = true;
        if (body.includes('onModifyDef')) item.onModifyDef = true;
        if (body.includes('onModifySpA')) item.onModifySpA = true;
        if (body.includes('onModifySpD')) item.onModifySpD = true;
        if (body.includes('onModifySpe')) item.onModifySpe = true;
        if (body.includes('onBasePower')) item.onBasePower = true;

        // Natural Gift type and power
        const naturalGiftMatch = body.match(/naturalGift:\s*\{\s*basePower:\s*(\d+),\s*type:\s*["'](\w+)["']/);
        if (naturalGiftMatch) {
            item.naturalGift = {
                basePower: parseInt(naturalGiftMatch[1]),
                type: naturalGiftMatch[2]
            };
        }

        // Check for boosts
        const boostsMatch = body.match(/boosts:\s*\{([^}]+)\}/);
        if (boostsMatch) {
            const boosts = {};
            const statMatches = boostsMatch[1].matchAll(/(\w+):\s*(-?\d+)/g);
            for (const [, stat, val] of statMatches) {
                boosts[stat] = parseInt(val);
            }
            item.boosts = boosts;
        }

        // Add descriptions from text file
        if (descriptions[id]) {
            if (descriptions[id].desc) {
                item.desc = descriptions[id].desc;
            } else if (descriptions[id].shortDesc) {
                item.desc = descriptions[id].shortDesc;
            }
        }

        if (item.name) {
            items[id] = item;
        }
    }

    fs.writeFileSync(
        path.join(dataDir, 'items.json'),
        JSON.stringify(items, null, 2)
    );
    console.log(`Converted items.json (${Object.keys(items).length} items)`);
}

/**
 * Parse ability descriptions from abilities-text.ts
 */
function parseAbilityDescriptions() {
    const textPath = path.join(dataDir, 'abilities-text.ts');
    if (!fs.existsSync(textPath)) {
        console.log('abilities-text.ts not found, skipping descriptions');
        return {};
    }

    const content = fs.readFileSync(textPath, 'utf8');
    const descriptions = {};

    // Find each ability text block
    const lines = content.split('\n');
    let currentAbilityId = null;
    let braceCount = 0;
    let inAbility = false;
    let abilityContent = '';

    for (const line of lines) {
        const abilityStart = line.match(/^\t(\w+):\s*\{/);

        if (abilityStart && !inAbility) {
            currentAbilityId = abilityStart[1];
            inAbility = true;
            braceCount = 1;
            abilityContent = line;
            continue;
        }

        if (inAbility) {
            abilityContent += '\n' + line;

            for (const char of line) {
                if (char === '{') braceCount++;
                if (char === '}') braceCount--;
            }

            if (braceCount === 0) {
                // Extract desc and shortDesc - handle escaped quotes and apostrophes
                const descMatch = abilityContent.match(/\tdesc:\s*"((?:[^"\\]|\\.)*)"/);
                const shortDescMatch = abilityContent.match(/\tshortDesc:\s*"((?:[^"\\]|\\.)*)"/);

                descriptions[currentAbilityId] = {
                    desc: descMatch ? descMatch[1].replace(/\\'/g, "'") : null,
                    shortDesc: shortDescMatch ? shortDescMatch[1].replace(/\\'/g, "'") : null
                };

                inAbility = false;
                currentAbilityId = null;
                abilityContent = '';
            }
        }
    }

    return descriptions;
}

/**
 * Parse item descriptions from text/items.ts
 */
function parseItemDescriptions() {
    const textPath = path.join(dataDir, 'text', 'items.ts');
    if (!fs.existsSync(textPath)) {
        console.log('text/items.ts not found, skipping item descriptions');
        return {};
    }

    const content = fs.readFileSync(textPath, 'utf8');
    const descriptions = {};

    // Find each item text block
    const lines = content.split('\n');
    let currentItemId = null;
    let braceCount = 0;
    let inItem = false;
    let itemContent = '';

    for (const line of lines) {
        const itemStart = line.match(/^\t(\w+):\s*\{/);

        if (itemStart && !inItem) {
            currentItemId = itemStart[1];
            inItem = true;
            braceCount = 1;
            itemContent = line;
            continue;
        }

        if (inItem) {
            itemContent += '\n' + line;

            for (const char of line) {
                if (char === '{') braceCount++;
                if (char === '}') braceCount--;
            }

            if (braceCount === 0) {
                // Extract desc and shortDesc - handle escaped quotes and apostrophes
                const descMatch = itemContent.match(/\tdesc:\s*"((?:[^"\\]|\\.)*)"/);
                const shortDescMatch = itemContent.match(/\tshortDesc:\s*"((?:[^"\\]|\\.)*)"/);

                descriptions[currentItemId] = {
                    desc: descMatch ? descMatch[1].replace(/\\'/g, "'") : null,
                    shortDesc: shortDescMatch ? shortDescMatch[1].replace(/\\'/g, "'") : null
                };

                inItem = false;
                currentItemId = null;
                itemContent = '';
            }
        }
    }

    return descriptions;
}

/**
 * Convert abilities.ts to abilities.json
 * Handles complex nested structures with functions
 */
function convertAbilities() {
    const content = fs.readFileSync(path.join(dataDir, 'abilities.ts'), 'utf8');

    // Load descriptions from text file
    const descriptions = parseAbilityDescriptions();

    const abilities = {};

    // Find the start of the export
    const exportStart = content.indexOf('export const Abilities');
    if (exportStart === -1) {
        console.error('Could not find Abilities export');
        return;
    }

    // Find each ability block by looking for pattern: \n\tabilityname: {
    const lines = content.slice(exportStart).split('\n');
    let currentAbilityId = null;
    let braceCount = 0;
    let inAbility = false;
    let abilityContent = '';

    for (const line of lines) {
        // Check for start of new ability (single tab + identifier + colon + brace)
        const abilityStart = line.match(/^\t(\w+):\s*\{/);

        if (abilityStart && !inAbility) {
            currentAbilityId = abilityStart[1];
            inAbility = true;
            braceCount = 1; // Opening brace
            abilityContent = line;
            continue;
        }

        if (inAbility) {
            abilityContent += '\n' + line;

            // Count braces
            for (const char of line) {
                if (char === '{') braceCount++;
                if (char === '}') braceCount--;
            }

            // End of ability block
            if (braceCount === 0) {
                // Extract properties from abilityContent
                const ability = { id: currentAbilityId };

                const nameMatch = abilityContent.match(/name:\s*["']([^"']+)["']/);
                if (nameMatch) ability.name = nameMatch[1];

                const numMatch = abilityContent.match(/num:\s*(\d+)/);
                if (numMatch) ability.num = parseInt(numMatch[1]);

                const ratingMatch = abilityContent.match(/rating:\s*(-?\d+(?:\.\d+)?)/);
                if (ratingMatch) ability.rating = parseFloat(ratingMatch[1]);

                // Add descriptions from text file
                if (descriptions[currentAbilityId]) {
                    if (descriptions[currentAbilityId].desc) {
                        ability.desc = descriptions[currentAbilityId].desc;
                    }
                    if (descriptions[currentAbilityId].shortDesc) {
                        ability.shortDesc = descriptions[currentAbilityId].shortDesc;
                    }
                }

                // Check for key ability effects (for damage calculation)
                if (abilityContent.includes('onModifyDamage')) ability.onModifyDamage = true;
                if (abilityContent.includes('onModifyAtk')) ability.onModifyAtk = true;
                if (abilityContent.includes('onModifyDef')) ability.onModifyDef = true;
                if (abilityContent.includes('onModifySpA')) ability.onModifySpA = true;
                if (abilityContent.includes('onModifySpD')) ability.onModifySpD = true;
                if (abilityContent.includes('onModifySpe')) ability.onModifySpe = true;
                if (abilityContent.includes('onBasePower')) ability.onBasePower = true;
                if (abilityContent.includes('onSourceModifyDamage')) ability.onSourceModifyDamage = true;
                if (abilityContent.includes('onSourceBasePower')) ability.onSourceBasePower = true;
                if (abilityContent.includes('onModifySTAB')) ability.onModifySTAB = true;
                if (abilityContent.includes('suppressWeather')) ability.suppressWeather = true;
                if (abilityContent.includes('onImmunity')) ability.onImmunity = true;
                if (abilityContent.includes('onModifyType')) ability.onModifyType = true;

                // Extract chainModify values if present (these are the actual multipliers)
                const chainModifyMatches = abilityContent.matchAll(/chainModify\(\s*\[(\d+),\s*(\d+)\]\s*\)/g);
                for (const match of chainModifyMatches) {
                    if (!ability.modifiers) ability.modifiers = [];
                    ability.modifiers.push({
                        numerator: parseInt(match[1]),
                        denominator: parseInt(match[2])
                    });
                }

                if (ability.name) {
                    abilities[currentAbilityId] = ability;
                }

                inAbility = false;
                currentAbilityId = null;
                abilityContent = '';
            }
        }
    }

    fs.writeFileSync(
        path.join(dataDir, 'abilities.json'),
        JSON.stringify(abilities, null, 2)
    );
    console.log(`Converted abilities.json (${Object.keys(abilities).length} abilities)`);
}

/**
 * Convert learnsets.ts to learnsets.json
 * Extracts the learnset data for each Pokemon
 */
function convertLearnsets() {
    const content = fs.readFileSync(path.join(dataDir, 'learnsets.ts'), 'utf8');

    const learnsets = {};

    // Find the start of the export
    const exportStart = content.indexOf('export const Learnsets');
    if (exportStart === -1) {
        console.error('Could not find Learnsets export');
        return;
    }

    // Find each learnset block by looking for pattern: \n\tpokemonname: {
    const lines = content.slice(exportStart).split('\n');
    let currentPokemonId = null;
    let braceCount = 0;
    let inPokemon = false;
    let pokemonContent = '';

    for (const line of lines) {
        // Check for start of new pokemon (single tab + identifier + colon + brace)
        const pokemonStart = line.match(/^\t(\w+):\s*\{/);

        if (pokemonStart && !inPokemon) {
            currentPokemonId = pokemonStart[1];
            inPokemon = true;
            braceCount = 1; // Opening brace
            pokemonContent = line;
            continue;
        }

        if (inPokemon) {
            pokemonContent += '\n' + line;

            // Count braces
            for (const char of line) {
                if (char === '{') braceCount++;
                if (char === '}') braceCount--;
            }

            // End of pokemon block
            if (braceCount === 0) {
                // Extract learnset from pokemonContent
                const learnsetMatch = pokemonContent.match(/learnset:\s*\{([^}]+(?:\{[^}]*\}[^}]*)*)\}/s);
                if (learnsetMatch) {
                    const learnsetBody = learnsetMatch[1];
                    const learnset = {};

                    // Match each move: movename: ["source1", "source2", ...]
                    const moveRegex = /(\w+):\s*\[([^\]]+)\]/g;
                    let moveMatch;
                    while ((moveMatch = moveRegex.exec(learnsetBody)) !== null) {
                        const [, moveName, sourcesStr] = moveMatch;
                        // Parse the sources array
                        const sources = [];
                        const sourceMatches = sourcesStr.matchAll(/"([^"]+)"/g);
                        for (const sm of sourceMatches) {
                            sources.push(sm[1]);
                        }
                        if (sources.length > 0) {
                            learnset[moveName] = sources;
                        }
                    }

                    if (Object.keys(learnset).length > 0) {
                        learnsets[currentPokemonId] = { learnset };
                    }
                }

                inPokemon = false;
                currentPokemonId = null;
                pokemonContent = '';
            }
        }
    }

    fs.writeFileSync(
        path.join(dataDir, 'learnsets.json'),
        JSON.stringify(learnsets, null, 2)
    );
    console.log(`Converted learnsets.json (${Object.keys(learnsets).length} Pokemon)`);
}

// Run conversions
console.log('Converting Pokemon Showdown data files...');

// Only convert files that exist (skip already-converted ones)
if (fs.existsSync(path.join(dataDir, 'natures.ts'))) {
    convertNatures();
}
if (fs.existsSync(path.join(dataDir, 'typechart.ts'))) {
    convertTypechart();
}
if (fs.existsSync(path.join(dataDir, 'items.ts'))) {
    convertItems();
}
if (fs.existsSync(path.join(dataDir, 'abilities.ts'))) {
    convertAbilities();
}
if (fs.existsSync(path.join(dataDir, 'learnsets.ts'))) {
    convertLearnsets();
}

console.log('Done!');
