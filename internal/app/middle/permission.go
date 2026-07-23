package middle

import (
	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/types"
	"github.com/lzh-1625/go_process_manager/log"
)

// RolePermission validates role permissions.
func RolePermission(needPermission types.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if v, ok := c.Get(types.CtxRole).(types.Role); !ok || v > needPermission {
				log.Logger.Errorw("Insufficient permissions", "needPermission", needPermission, "role", v)
				return c.JSON(500, model.Response[struct{}]{
					Message: "Insufficient permissions",
					Code:    -1,
				})
			}
			return next(c)
		}
	}

}
