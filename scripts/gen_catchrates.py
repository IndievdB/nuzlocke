#!/usr/bin/env python3
"""
Fetch catch rates for all Pokemon from PokeAPI and generate catchrates.json
"""

import json
import urllib.request
import time
import sys

def fetch_pokemon_count():
    """Get total number of Pokemon species"""
    url = "https://pokeapi.co/api/v2/pokemon-species?limit=1"
    with urllib.request.urlopen(url) as response:
        data = json.loads(response.read().decode())
        return data['count']

def fetch_pokemon_species(pokemon_id):
    """Fetch a single Pokemon species data"""
    url = f"https://pokeapi.co/api/v2/pokemon-species/{pokemon_id}"
    try:
        with urllib.request.urlopen(url) as response:
            return json.loads(response.read().decode())
    except urllib.error.HTTPError as e:
        if e.code == 404:
            return None
        raise

def main():
    print("Fetching Pokemon count...")
    total = fetch_pokemon_count()
    print(f"Total Pokemon species: {total}")

    catch_rates = {}

    # Fetch all Pokemon (1 to total)
    for i in range(1, total + 1):
        try:
            data = fetch_pokemon_species(i)
            if data:
                catch_rate = data.get('capture_rate', 0)
                name = data.get('name', f'unknown-{i}')
                catch_rates[str(i)] = catch_rate
                print(f"[{i}/{total}] {name}: {catch_rate}")
            else:
                print(f"[{i}/{total}] Not found")
        except Exception as e:
            print(f"[{i}/{total}] Error: {e}")

        # Rate limit - be nice to the API
        if i % 100 == 0:
            time.sleep(1)

    # Write to JSON file
    output_path = "data/catchrates.json"
    with open(output_path, 'w') as f:
        json.dump(catch_rates, f, indent=2, sort_keys=lambda x: int(x))

    print(f"\nWrote {len(catch_rates)} catch rates to {output_path}")

if __name__ == "__main__":
    main()
