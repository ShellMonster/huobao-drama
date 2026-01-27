package services

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/drama-generator/backend/domain/models"
)

type styleDefinition struct {
	Key        string
	Name       string
	PromptZh   string
	PromptEn   string
	PreviewURL string
	SortOrder  int
	IsDefault  bool
}

const (
	fallbackDefaultStyleKey = "c65ac50bb5eba9a88cde3dd51919440e"
	realisticStyleKey       = "4e8ccfcdeb7abda4f35eea94759e00df"
	stylePlaceholder        = "{{STYLE}}"
)

var (
	styleCatalogMu  sync.RWMutex
	styleCatalog    = buildDefaultStyleCatalog()
	defaultStyleKey = fallbackDefaultStyleKey
)

var styleAliases = map[string]string{
	"realistic":       realisticStyleKey,
	"realistic_urban": realisticStyleKey,
	"anime_cinematic": realisticStyleKey,
	"anime_isekai":    realisticStyleKey,
	"fantasy_cartoon": realisticStyleKey,
	"ink_wash":        realisticStyleKey,
}

func buildDefaultStyleCatalog() map[string]styleDefinition {
	seeds := DefaultStyleSeeds()
	catalog := make(map[string]styleDefinition, len(seeds))
	for idx, seed := range seeds {
		key := hashStyleKey(seed.ImageURL)
		if key == "" || seed.Name == "" {
			continue
		}
		promptZh := strings.TrimSpace(seed.PromptZh)
		if promptZh == "" {
			promptZh = seed.Name
		}
		promptEn := strings.TrimSpace(seed.PromptEn)
		if promptEn == "" {
			promptEn = seed.Name
		}
		catalog[key] = styleDefinition{
			Key:        key,
			Name:       seed.Name,
			PromptZh:   promptZh,
			PromptEn:   promptEn,
			PreviewURL: fmt.Sprintf("/static/styles/%s.webp", key),
			SortOrder:  idx,
			IsDefault:  idx == 0,
		}
	}
	return catalog
}

func normalizeStyleKey(styleKey string) string {
	key, ok := resolveStyleKey(styleKey)
	if ok && key != "" {
		return key
	}
	return defaultStyleKey
}

func getStylePrompt(styleKey string, isEnglish bool) string {
	key, ok := resolveStyleKey(styleKey)
	if !ok || key == "" {
		return ""
	}
	styleCatalogMu.RLock()
	style, ok := styleCatalog[key]
	styleCatalogMu.RUnlock()
	if !ok {
		return ""
	}
	if isEnglish {
		return style.PromptEn
	}
	return style.PromptZh
}

func resolveStyleKey(styleKey string) (string, bool) {
	key := strings.TrimSpace(strings.ToLower(styleKey))
	if key == "" {
		return defaultStyleKey, true
	}
	if alias, ok := styleAliases[key]; ok {
		key = alias
	}
	styleCatalogMu.RLock()
	_, ok := styleCatalog[key]
	styleCatalogMu.RUnlock()
	if ok {
		return key, true
	}
	return "", false
}

func LoadStyleCatalog(styles []models.Style) {
	styleCatalogMu.Lock()
	defer styleCatalogMu.Unlock()

	updated := make(map[string]styleDefinition)
	defaultKey := fallbackDefaultStyleKey

	for _, style := range styles {
		if !style.IsActive {
			continue
		}
		def := styleDefinition{
			Key:        style.Key,
			Name:       style.Name,
			PromptZh:   style.PromptZh,
			PromptEn:   style.PromptEn,
			PreviewURL: style.PreviewURL,
			SortOrder:  style.SortOrder,
			IsDefault:  style.IsDefault,
		}
		updated[style.Key] = def
		if style.IsDefault {
			defaultKey = style.Key
		}
	}

	if len(updated) > 0 {
		styleCatalog = updated
		defaultStyleKey = defaultKey
		if _, ok := styleCatalog[defaultStyleKey]; !ok {
			for key := range styleCatalog {
				defaultStyleKey = key
				break
			}
		}
	}
}

func DefaultStyles() []models.Style {
	styleCatalogMu.RLock()
	definitions := make([]styleDefinition, 0, len(styleCatalog))
	for _, def := range styleCatalog {
		definitions = append(definitions, def)
	}
	styleCatalogMu.RUnlock()

	sort.Slice(definitions, func(i, j int) bool {
		if definitions[i].SortOrder == definitions[j].SortOrder {
			return definitions[i].Key < definitions[j].Key
		}
		return definitions[i].SortOrder < definitions[j].SortOrder
	})

	styles := make([]models.Style, 0, len(definitions))
	for _, def := range definitions {
		styles = append(styles, models.Style{
			Key:        def.Key,
			Name:       def.Name,
			PromptZh:   def.PromptZh,
			PromptEn:   def.PromptEn,
			PreviewURL: def.PreviewURL,
			SortOrder:  def.SortOrder,
			IsActive:   true,
			IsDefault:  def.IsDefault,
			IsSystem:   true,
		})
	}
	return styles
}

var legacyEnglishStyleTokens = []string{
	"cinematic anime-style",
	"cinematic anime style",
	"anime-style",
	"anime style",
}

var legacyChineseStyleTokens = []string{
	"电影感的动漫风格",
	"电影感动漫风格",
	"动漫风格",
}

var legacyEnglishStyleRegexps = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bcinematic anime[- ]style\b`),
	regexp.MustCompile(`(?i)\banime[- ]style\b`),
}

func isEnglishPrompt(prompt string) bool {
	for _, r := range prompt {
		if r >= 0x4E00 && r <= 0x9FFF {
			return false
		}
	}
	return true
}

func containsLegacyStyleTokens(prompt string) bool {
	lower := strings.ToLower(prompt)
	for _, token := range legacyEnglishStyleTokens {
		if strings.Contains(lower, token) {
			return true
		}
	}
	for _, token := range legacyChineseStyleTokens {
		if strings.Contains(prompt, token) {
			return true
		}
	}
	return false
}

func replaceStylePrompt(prompt, stylePrompt string) string {
	if stylePrompt == "" {
		return prompt
	}
	if !strings.Contains(prompt, stylePlaceholder) {
		return prompt
	}
	return strings.ReplaceAll(prompt, stylePlaceholder, stylePrompt)
}

func replaceLegacyStyleTokensWithPlaceholder(prompt string) string {
	updated := prompt
	for _, re := range legacyEnglishStyleRegexps {
		updated = re.ReplaceAllString(updated, stylePlaceholder)
	}
	for _, token := range legacyChineseStyleTokens {
		updated = strings.ReplaceAll(updated, token, stylePlaceholder)
	}
	return updated
}

func applyStyleToPrompt(prompt, styleKey string) string {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return prompt
	}
	isEnglish := isEnglishPrompt(prompt)
	stylePrompt := getStylePrompt(styleKey, isEnglish)
	if strings.Contains(prompt, stylePlaceholder) {
		if stylePrompt == "" {
			stylePrompt = getStylePrompt(defaultStyleKey, isEnglish)
		}
		if stylePrompt == "" {
			return strings.ReplaceAll(prompt, stylePlaceholder, "")
		}
		return replaceStylePrompt(prompt, stylePrompt)
	}
	if stylePrompt == "" {
		return prompt
	}
	if strings.Contains(prompt, stylePrompt) || containsLegacyStyleTokens(prompt) {
		return prompt
	}
	if isEnglish {
		return strings.TrimSpace(fmt.Sprintf("%s, %s", prompt, stylePrompt))
	}
	return strings.TrimSpace(fmt.Sprintf("%s，%s", prompt, stylePrompt))
}
