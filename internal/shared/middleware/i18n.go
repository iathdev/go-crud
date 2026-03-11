package middleware

import (
	"learning-go/internal/shared/i18n"

	"github.com/gin-gonic/gin"
)

func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Query("lang")
		if lang == "" {
			lang = c.GetHeader("X-Lang")
		}
		if lang == "" {
			lang = i18n.FromAcceptLanguage(c.GetHeader("Accept-Language"))
		}
		lang = i18n.Normalize(lang)
		c.Set("lang", lang)
		c.Header("Content-Language", lang)
		c.Next()
	}
}
