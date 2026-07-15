package api

import (
	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/types"
)

func getRole(c *echo.Context) types.Role {
	if v := c.Get(types.CtxRole); v != nil {
		if role, ok := v.(types.Role); ok {
			return role
		}
	}
	return types.RoleGuest
}

func getUserName(c *echo.Context) string {
	if v := c.Get(types.CtxUserName); v != nil {
		if name, ok := v.(string); ok {
			return name
		}
	}
	return ""
}

func isAdmin(c *echo.Context) bool {
	return getRole(c) <= types.RoleAdmin
}
