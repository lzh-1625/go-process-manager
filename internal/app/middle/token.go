package middle

import (
	"slices"
	"strings"

	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"

	"github.com/gin-gonic/gin"
)

// code -1为失败,-2为token失效
func rErr(ctx *gin.Context, code int, message string, err error) {
	var statusCode int
	switch code {
	case -1:
		statusCode = 500
	case -2:
		statusCode = 401
	default:
		statusCode = 200
	}
	if err != nil {
		log.Logger.Warn(err)
	}
	ctx.JSON(statusCode, map[string]any{
		"code": code,
		"message":  message,
	})
	ctx.Abort()
}

func CheckToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		whiteList := []string{
			"/api/user/login",
			"/api/user/register/admin",
			"/api/task/api-key/",
			"/api/ws/share",
		}
		if !slices.ContainsFunc(whiteList, func(s string) bool {
			return strings.HasPrefix(c.Request.URL.Path, s)
		}) {
			var token string
			if c.Request.Header.Get("Authorization") != "" {
				token = strings.TrimPrefix(c.Request.Header.Get("Authorization"), "bearer ")
			} else {
				token = c.Query("token")
			}
			if mc, err := utils.VerifyToken(token); err != nil {
				rErr(c, -2, "token校验失败", err)
				return
			} else {
				c.Set(eum.CtxUserName, mc.Username)
				c.Set(eum.CtxRole, repository.UserRepository.GetUserByName(mc.Username).Role)
			}
		}
		c.Next()
	}
}
