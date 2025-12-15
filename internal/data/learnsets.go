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

	// Track level-up moves with their generation to pick highest gen's level
	levelUpByMove := make(map[string]struct {
		level int
		gen   byte
	})

	tmSeen := make(map[string]bool)
	tutorSeen := make(map[string]bool)
	eggSeen := make(map[string]bool)
	eventSeen := make(map[string]bool)

	for moveName, sources := range raw.Learnset {
		for _, source := range sources {
			if len(source) < 2 {
				continue
			}

			// Check if this source is from the target generation or earlier
			sourceGen := source[0]
			if string(sourceGen) > genStr {
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
				// Keep the highest generation's level for each move
				if existing, ok := levelUpByMove[moveName]; !ok || sourceGen > existing.gen {
					levelUpByMove[moveName] = struct {
						level int
						gen   byte
					}{level, sourceGen}
				}
			case 'M':
				// TM/HM move - deduplicate
				if !tmSeen[moveName] {
					tmSeen[moveName] = true
					parsed.TM = append(parsed.TM, moveName)
				}
			case 'T':
				// Tutor move - deduplicate
				if !tutorSeen[moveName] {
					tutorSeen[moveName] = true
					parsed.Tutor = append(parsed.Tutor, moveName)
				}
			case 'E':
				// Egg move - deduplicate
				if !eggSeen[moveName] {
					eggSeen[moveName] = true
					parsed.Egg = append(parsed.Egg, moveName)
				}
			case 'S':
				// Event move - deduplicate
				if !eventSeen[moveName] {
					eventSeen[moveName] = true
					parsed.Event = append(parsed.Event, moveName)
				}
			}
		}
	}

	// Convert level-up map to slice
	for moveName, data := range levelUpByMove {
		parsed.LevelUp = append(parsed.LevelUp, LevelUpMove{
			Move:  moveName,
			Level: data.level,
		})
	}

	return parsed
}

