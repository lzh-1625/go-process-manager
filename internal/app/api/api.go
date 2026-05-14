package api

import (
	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
)

func getRole(c *echo.Context) eum.Role {
	if v := c.Get(eum.CtxRole); v != nil {
		if role, ok := v.(eum.Role); ok {
			return role
		}
	}
	return eum.RoleGuest
}

func getUserName(c *echo.Context) string {
	if v := c.Get(eum.CtxUserName); v != nil {
		if name, ok := v.(string); ok {
			return name
		}
	}
	return ""
}

func isAdmin(c *echo.Context) bool {
	return getRole(c) <= eum.RoleAdmin
}
