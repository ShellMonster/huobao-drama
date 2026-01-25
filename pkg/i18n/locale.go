package i18n

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	LangZhCN = "zh-CN"
	LangEnUS = "en-US"
)

func Normalize(lang string) string {
	normalized := strings.ToLower(strings.TrimSpace(lang))
	switch {
	case normalized == "zh" || strings.HasPrefix(normalized, "zh-") || strings.HasPrefix(normalized, "zh_"):
		return LangZhCN
	case normalized == "en" || strings.HasPrefix(normalized, "en-") || strings.HasPrefix(normalized, "en_"):
		return LangEnUS
	default:
		return LangZhCN
	}
}

func ParseAcceptLanguage(header string, fallback string) string {
	fallbackLang := Normalize(fallback)
	if strings.TrimSpace(header) == "" {
		return fallbackLang
	}
	parts := strings.Split(header, ",")
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		if idx := strings.Index(token, ";"); idx >= 0 {
			token = strings.TrimSpace(token[:idx])
		}
		if token == "" {
			continue
		}
		return Normalize(token)
	}
	return fallbackLang
}

func GetLang(c *gin.Context, fallback string) string {
	if c != nil {
		if value, ok := c.Get("lang"); ok {
			if lang, ok := value.(string); ok && strings.TrimSpace(lang) != "" {
				return Normalize(lang)
			}
		}
	}
	return Normalize(fallback)
}
