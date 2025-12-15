package data

// Learnset represents a Pokemon's learnset data
type Learnset struct {
	Learnset map[string][]string `json:"learnset"` // moveid -> source codes (e.g. "9L25", "9M", "9T", "9E")
}

// ParsedLearnset represents a learnset with parsed move sources
type ParsedLearnset struct {
	LevelUp []LevelUpMove `json:"levelup"`
	TM      []string      `json:"tm"`
	Tutor   []string      `json:"tutor"`
	Egg     []string      `json:"egg"`
	Event   []string      `json:"event"`
}

// LevelUpMove represents a move learned at a specific level
type LevelUpMove struct {
	Move  string `json:"move"`
	Level int    `json:"level"`
}

// ParseLearnset converts raw source codes into categorized moves
// Source codes: L=levelup, M=TM/HM, T=tutor, E=egg, S=event
func ParseLearnset(raw *Learnset, generation int) *ParsedLearnset {
	if raw == nil || raw.Learnset == nil {
		return &ParsedLearnset{}
	}

	parsed := &ParsedLearnset{
		LevelUp: []LevelUpMove{},
		TM:      []string{},
		Tutor:   []string{},
		Egg:     []string{},
		Event:   []string{},
	}

	genStr := string('0' + generation)

	for moveName, sources := range raw.Learnset {
		for _, source := range sources {
			if len(source) < 2 {
				continue
			}

			// Check if this source is from the target generation or earlier
			sourceGen := string(source[0])
			if sourceGen > genStr {
				continue
			}

			sourceType := source[1]
			switch sourceType {
			case 'L':
				// Level-up move: "9L25" means level 25
				level := 0
				if len(source) > 2 {
					for _, c := range source[2:] {
						if c >= '0' && c <= '9' {
							level = level*10 + int(c-'0')
						}
					}
				}
				// Only add if this is the highest generation available
				parsed.LevelUp = append(parsed.LevelUp, LevelUpMove{
					Move:  moveName,
					Level: level,
				})
			case 'M':
				// TM/HM move
				parsed.TM = append(parsed.TM, moveName)
			case 'T':
				// Tutor move
				parsed.Tutor = append(parsed.Tutor, moveName)
			case 'E':
				// Egg move
				parsed.Egg = append(parsed.Egg, moveName)
			case 'S':
				// Event move
				parsed.Event = append(parsed.Event, moveName)
			}
		}
	}

	// Remove duplicates from each category
	parsed.LevelUp = deduplicateLevelUp(parsed.LevelUp)
	parsed.TM = deduplicate(parsed.TM)
	parsed.Tutor = deduplicate(parsed.Tutor)
	parsed.Egg = deduplicate(parsed.Egg)
	parsed.Event = deduplicate(parsed.Event)

	return parsed
}

// deduplicateLevelUp removes duplicate moves, keeping the lowest level for each
func deduplicateLevelUp(moves []LevelUpMove) []LevelUpMove {
	seen := make(map[string]int) // moveName -> index in result
	result := []LevelUpMove{}

	for _, m := range moves {
		if idx, ok := seen[m.Move]; ok {
			// Keep the lower level
			if m.Level < result[idx].Level {
				result[idx].Level = m.Level
			}
		} else {
			seen[m.Move] = len(result)
			result = append(result, m)
		}
	}

	return result
}

// deduplicate removes duplicate strings from a slice
func deduplicate(items []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
