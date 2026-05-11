package middle

import (
	"net/http"
	"slices"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/utils"
)

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
		// 白名单放行
		if !slices.ContainsFunc(whiteList, func(s string) bool {
			return strings.HasPrefix(path, s)
		}) {

			var token string
			auth := c.Request().Header.Get("Authorization")

			if auth != "" {
				token = strings.TrimPrefix(auth, "bearer ")
			} else {
				token = c.QueryParam("token")
			}

			mc, err := utils.VerifyToken(token)
			if err != nil {
				return c.JSON(401, model.Response[struct{}]{
					Code:    -1,
					Message: "invalid token",
				})
			}
			c.Set(eum.CtxUserName, mc.Username)
			c.Set(
				eum.CtxRole,
				a.userLogic.GetUserByName(mc.Username).Role,
			)
		}
		err := next(c)
		if err != nil {
			return err
		}
		if resp, err := echo.UnwrapResponse(c.Response()); err == nil && !resp.Committed {
			return c.JSON(http.StatusOK, model.Response[any]{
				Code:    0,
				Message: "success",
			})
		}

		return nil
	}
}
