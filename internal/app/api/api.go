package api

import (
	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
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

func hasOprPermission(c *echo.Context, uuid int, op eum.OprPermission) bool {
	if isAdmin(c) {
		return true
	}

	per := repository.PermissionRepository.GetPermission(
		getUserName(c),
		uuid,
	)
	if per == nil {
		return false
	}

	switch op {
	case eum.OperationLog:
		return per.Log
	case eum.OperationTerminal:
		return per.Terminal
	case eum.OperationStart:
		return per.Start
	case eum.OperationStop:
		return per.Stop
	case eum.OperationTerminalWrite:
		return per.Write
	}
	return false
}
