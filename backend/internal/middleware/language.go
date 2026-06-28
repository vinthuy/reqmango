package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend/internal/i18n"
)

// LanguageMiddleware detects and sets the user's preferred language.
func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := i18n.DetectLanguage(c.GetHeader("Accept-Language"))
		// Also allow override via ?lang= query param
		if ql := c.Query("lang"); ql != "" {
			lang = ql
		}
		c.Set("lang", lang)
		c.Next()
	}
}

// GetLang extracts the language from Gin context.
func GetLang(c *gin.Context) string {
	if l, exists := c.Get("lang"); exists {
		return l.(string)
	}
	return "zh"
}
