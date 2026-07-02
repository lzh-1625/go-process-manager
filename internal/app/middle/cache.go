package middle

import (
	"strings"

	"github.com/labstack/echo/v5"
)

func CacheMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if !strings.Contains(c.Request().Header.Get("accept"), "text/html") && !strings.HasPrefix(c.Request().URL.Path, "/api") {
				c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			return next(c)
		}
	}
}
