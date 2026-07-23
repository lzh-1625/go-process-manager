package middle

import (
	"slices"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/types"
	"github.com/lzh-1625/go_process_manager/utils"
)

// whiteList contains paths excluded from JWT validation.
var whiteList = []string{
	"/api/user/login",
	"/api/user/register/admin",
	"/api/task/api-key/",
	"/api/ws/share",
}

func NewAuthMiddleware(userLogic *logic.UserLogic) *AuthMiddleware {
	return &AuthMiddleware{
		userLogic: userLogic,
	}
}

type AuthMiddleware struct {
	userLogic *logic.UserLogic
}

func (a *AuthMiddleware) Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		path := c.Request().URL.Path

		if !slices.ContainsFunc(whiteList, func(s string) bool {
			return strings.HasPrefix(path, s)
		}) {
			token := c.Request().Header.Get("Authorization")
			if token == "" {
				if cookie, err := c.Cookie("Authorization"); err == nil && cookie != nil {
					token = cookie.Value
				}
			}
			token = strings.TrimPrefix(token, "bearer ")
			mc, err := utils.VerifyToken(token, config.CF.SecretKey)
			if err != nil {
				return c.JSON(401, model.Response[struct{}]{
					Code:    -1,
					Message: "invalid token",
				})
			}
			c.Set(types.CtxUserName, mc.Username)
			c.Set(
				types.CtxRole,
				a.userLogic.GetUserByName(mc.Username).Role,
			)
		}
		return next(c)
	}
}
