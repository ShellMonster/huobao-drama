package services

import (
	"encoding/json"
	"sort"
	"strings"

	models "github.com/drama-generator/backend/domain/models"
)

func hasCharacterImage(character models.Character) bool {
	if character.ImageURL == nil {
		return false
	}
	return strings.TrimSpace(*character.ImageURL) != ""
}

func hasCharacterReferenceImages(character models.Character) bool {
	if len(character.ReferenceImages) == 0 {
		return false
	}
	raw := strings.TrimSpace(string(character.ReferenceImages))
	if raw == "" || raw == "null" || raw == "[]" {
		return false
	}
	var values []string
	if err := json.Unmarshal(character.ReferenceImages, &values); err != nil {
		return true
	}
	return len(values) > 0
}

func hasCharacterDetails(character models.Character) bool {
	if character.Role != nil && strings.TrimSpace(*character.Role) != "" {
		return true
	}
	if character.Description != nil && strings.TrimSpace(*character.Description) != "" {
		return true
	}
	if character.Appearance != nil && strings.TrimSpace(*character.Appearance) != "" {
		return true
	}
	if character.Personality != nil && strings.TrimSpace(*character.Personality) != "" {
		return true
	}
	if character.VoiceStyle != nil && strings.TrimSpace(*character.VoiceStyle) != "" {
		return true
	}
	if character.SeedValue != nil && strings.TrimSpace(*character.SeedValue) != "" {
		return true
	}
	return false
}

func characterQualityScore(character models.Character) int {
	score := 0
	if hasCharacterImage(character) {
		score += 100
	}
	if hasCharacterReferenceImages(character) {
		score += 40
	}
	if character.Description != nil && strings.TrimSpace(*character.Description) != "" {
		score += 10
	}
	if character.Appearance != nil && strings.TrimSpace(*character.Appearance) != "" {
		score += 6
	}
	if character.Personality != nil && strings.TrimSpace(*character.Personality) != "" {
		score += 4
	}
	if character.VoiceStyle != nil && strings.TrimSpace(*character.VoiceStyle) != "" {
		score += 2
	}
	return score
}

func preferCharacter(a models.Character, b models.Character) models.Character {
	scoreA := characterQualityScore(a)
	scoreB := characterQualityScore(b)
	if scoreA != scoreB {
		if scoreA > scoreB {
			return a
		}
		return b
	}
	if a.UpdatedAt.After(b.UpdatedAt) {
		return a
	}
	if b.UpdatedAt.After(a.UpdatedAt) {
		return b
	}
	if a.ID >= b.ID {
		return a
	}
	return b
}

func pickBestCharacter(characters []models.Character) models.Character {
	best := characters[0]
	for i := 1; i < len(characters); i++ {
		best = preferCharacter(best, characters[i])
	}
	return best
}

func dedupeCharactersByName(characters []models.Character) []models.Character {
	if len(characters) == 0 {
		return nil
	}
	grouped := make(map[string][]models.Character)
	var unnamed []models.Character
	for _, character := range characters {
		if character.Name == "" {
			unnamed = append(unnamed, character)
			continue
		}
		grouped[character.Name] = append(grouped[character.Name], character)
	}

	names := make([]string, 0, len(grouped))
	for name := range grouped {
		names = append(names, name)
	}
	sort.Strings(names)

	var result []models.Character
	for _, name := range names {
		group := grouped[name]
		if len(group) == 1 {
			result = append(result, group[0])
			continue
		}

		var withVisual []models.Character
		for _, character := range group {
			if hasCharacterImage(character) || hasCharacterReferenceImages(character) {
				withVisual = append(withVisual, character)
			}
		}

		if len(withVisual) == 1 {
			result = append(result, withVisual[0])
			continue
		}

		sort.Slice(group, func(i, j int) bool {
			return group[i].ID < group[j].ID
		})
		result = append(result, group...)
	}

	if len(unnamed) > 0 {
		sort.Slice(unnamed, func(i, j int) bool {
			return unnamed[i].ID < unnamed[j].ID
		})
		result = append(result, unnamed...)
	}

	return result
}

func buildBestCharacterByName(characters []models.Character) map[string]models.Character {
	bestByName := make(map[string]models.Character)
	for _, character := range characters {
		if character.Name == "" {
			continue
		}
		if existing, ok := bestByName[character.Name]; ok {
			bestByName[character.Name] = preferCharacter(existing, character)
			continue
		}
		bestByName[character.Name] = character
	}
	return bestByName
}
