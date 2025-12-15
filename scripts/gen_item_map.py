#!/usr/bin/env python3
"""
Generate a mapping from pokeemerald-expansion item IDs to Showdown item IDs.
"""

import re
import urllib.request
import json

def fetch_url(url):
    with urllib.request.urlopen(url) as response:
        return response.read().decode('utf-8')

def normalize_name(name):
    """Convert ITEM_IRON_PLATE to ironplate"""
    return name.lower().replace('_', '')

def main():
    # Load Showdown items data
    with open('/Users/jarrett/Documents/nuzlocke/data/items.json') as f:
        showdown_items = json.load(f)

    # Create lookup by normalized name
    showdown_by_name = {}
    for item_id, item in showdown_items.items():
        showdown_by_name[item_id] = item.get('num', 0)

    # Fetch pokeemerald-expansion items.h
    items_url = "https://raw.githubusercontent.com/rh-hideout/pokeemerald-expansion/master/include/constants/items.h"
    items_content = fetch_url(items_url)

    # Parse item definitions
    expansion_items = {}  # expansion_id -> normalized_name
    for match in re.finditer(r'#define\s+ITEM_(\w+)\s+(\d+)', items_content):
        name = match.group(1)
        expansion_id = int(match.group(2))
        if name not in ['NONE', 'USE_MAIL_EDIT'] and expansion_id > 0:
            normalized = normalize_name(name)
            expansion_items[expansion_id] = normalized

    # Create mapping where expansion ID differs from Showdown ID
    mapping = {}
    misses = []
    for exp_id, norm_name in expansion_items.items():
        if norm_name in showdown_by_name:
            showdown_id = showdown_by_name[norm_name]
            if showdown_id != exp_id and showdown_id > 0:
                mapping[exp_id] = showdown_id
        else:
            misses.append((exp_id, norm_name))

    # Print as Go map
    print("// expansionItemToShowdown maps pokeemerald-expansion item IDs to Showdown item IDs")
    print("// Only items where IDs differ are included")
    print("var expansionItemToShowdown = map[int]int{")

    for exp_id in sorted(mapping.keys()):
        showdown_id = mapping[exp_id]
        print(f"\t{exp_id}: {showdown_id},")

    print("}")
    print()
    print(f"// Total mappings: {len(mapping)}")
    print(f"// Items not found in Showdown: {len(misses)}")

    # Verify specific items
    print("\n// Verification:")
    test_items = [('IRON_PLATE', 265), ('GRASS_GEM', 343)]
    for name, exp_id in test_items:
        norm = normalize_name(name)
        showdown_id = showdown_by_name.get(norm, 'NOT FOUND')
        mapped = mapping.get(exp_id, exp_id)
        print(f"// {name}: expansion={exp_id}, showdown={showdown_id}, mapped={mapped}")

if __name__ == '__main__':
    main()
