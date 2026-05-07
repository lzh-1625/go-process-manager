package middle

import (
	"net/http"
	"slices"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/utils"
)

var whiteList = []string{
	"/api/user/login",
	"/api/user/register/admin",
	"/api/task/api-key/",
	"/api/ws/share",
}

func Auth(next echo.HandlerFunc) echo.HandlerFunc {
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
				return err
			}
			c.Set(eum.CtxUserName, mc.Username)
			c.Set(
				eum.CtxRole,
				repository.UserRepository.GetUserByName(mc.Username).Role,
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
