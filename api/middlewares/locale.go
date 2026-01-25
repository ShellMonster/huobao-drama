package middlewares

import (
	"strings"

	"github.com/drama-generator/backend/pkg/i18n"
	"github.com/gin-gonic/gin"
)

func LocaleMiddleware(defaultLang string) gin.HandlerFunc {
	return func(c *gin.Context) {
		headerLang := c.GetHeader("Accept-Language")
		queryLang := strings.TrimSpace(c.Query("lang"))
		if queryLang == "" {
			queryLang = strings.TrimSpace(c.GetHeader("X-Language"))
		}
		lang := headerLang
		if queryLang != "" {
			lang = queryLang
		}
		c.Set("lang", i18n.ParseAcceptLanguage(lang, defaultLang))
		c.Next()
	}
}
