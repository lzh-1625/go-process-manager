package middle

import (
	"slices"
	"strings"

	"github.com/labstack/echo/v5"
)

func CacheMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			contentType := c.Response().Header().Get("content-type")
			if slices.ContainsFunc([]string{"text/javascript", "text/javascript", "font/woff2"}, func(s string) bool {
				return strings.Contains(contentType, s)
			}) {
				c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			return next(c)
		}
	}
}
