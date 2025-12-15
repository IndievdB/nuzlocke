#!/usr/bin/env python3
"""
Generate a mapping from pokeemerald-expansion internal species IDs to national dex numbers.
This fetches the species.h and pokedex.h from the expansion repo and creates a Go map.
"""

import re
import urllib.request

def fetch_url(url):
    with urllib.request.urlopen(url) as response:
        return response.read().decode('utf-8')

def main():
    # Fetch species.h (internal IDs)
    species_url = "https://raw.githubusercontent.com/rh-hideout/pokeemerald-expansion/master/include/constants/species.h"
    species_content = fetch_url(species_url)

    # Fetch pokedex.h (national dex numbers)
    pokedex_url = "https://raw.githubusercontent.com/rh-hideout/pokeemerald-expansion/master/include/constants/pokedex.h"
    pokedex_content = fetch_url(pokedex_url)

    # Parse species.h for SPECIES_* definitions
    species_map = {}  # name -> internal_id
    for match in re.finditer(r'#define\s+SPECIES_(\w+)\s+(\d+)', species_content):
        name = match.group(1)
        internal_id = int(match.group(2))
        if name not in ['NONE', 'EGG'] and internal_id > 0:
            species_map[name] = internal_id

    # Parse pokedex.h for NATIONAL_DEX_* definitions (enum values)
    # The enum starts at 0 for NATIONAL_DEX_NONE, then increments
    national_map = {}  # name -> national_dex
    current_value = 0
    for line in pokedex_content.split('\n'):
        # Look for enum entries like "NATIONAL_DEX_BULBASAUR,"
        match = re.match(r'\s*NATIONAL_DEX_(\w+)\s*,?', line)
        if match:
            name = match.group(1)
            if name != 'NONE':
                national_map[name] = current_value
            current_value += 1

    # Create mapping from internal_id -> national_dex
    mapping = {}
    for name, internal_id in species_map.items():
        # Skip forms (names containing underscore after the base name typically)
        if name in national_map:
            national_dex = national_map[name]
            if national_dex > 0:  # Skip NONE
                mapping[internal_id] = national_dex

    # Print as Go map
    print("// expansionToNationalDex maps pokeemerald-expansion internal species IDs to national dex numbers")
    print("// Generated from pokeemerald-expansion species.h and pokedex.h")
    print("var expansionToNationalDex = map[int]int{")

    # Sort by internal ID
    for internal_id in sorted(mapping.keys()):
        national_dex = mapping[internal_id]
        # Only print if they differ (save space)
        if internal_id != national_dex:
            print(f"\t{internal_id}: {national_dex},")

    print("}")
    print()
    print(f"// Total mappings where internal != national: {sum(1 for k, v in mapping.items() if k != v)}")
    print(f"// Total species: {len(mapping)}")

    # Also show some examples
    print("\n// Examples:")
    examples = ['FLORAGATO', 'SPRIGATITO', 'MEOWSCARADA', 'KLANG', 'EEVEE', 'TORKOAL']
    for name in examples:
        if name in species_map and name in national_map:
            print(f"// {name}: internal={species_map[name]}, national={national_map[name]}")

if __name__ == '__main__':
    main()
