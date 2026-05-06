package middle

import (
	"github.com/labstack/echo"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/log"
)

func RolePermission(needPermission eum.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if v, ok := c.Get(eum.CtxRole).(eum.Role); !ok || v > needPermission {
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
