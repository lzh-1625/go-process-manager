package middle

import (
	"net/http"
	"slices"

	"github.com/labstack/echo"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
)

// 演示模式
func DemoMiddle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			whiteListUri := []string{
				"/api/user/login",
				"/api/log",
			}
			if ctx.Request().Method == http.MethodGet || slices.Contains(whiteListUri, ctx.Request().URL.String()) {
				return next(ctx)
			} else {
				return ctx.JSON(http.StatusForbidden, model.Response[struct{}]{
					Code:    -1,
					Message: "当前处于演示模式",
				})
			}
		}
	}
}
