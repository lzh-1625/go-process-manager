package middle

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lzh-1625/go_process_manager/internal/app/constants"
	"github.com/lzh-1625/go_process_manager/log"
)

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path

		if !strings.HasPrefix(path, "/api") {
			return
		}
		// Process request
		ctx.Next()
		logKv := []any{}
		logKv = append(logKv, "Method", ctx.Request.Method)
		logKv = append(logKv, "Status", ctx.Writer.Status())
		logKv = append(logKv, "Path", path)
		logKv = append(logKv, "耗时", fmt.Sprintf("%dms", time.Now().UnixMilli()-start.UnixMilli()))
		if user, ok := ctx.Get(constants.CTXFLG_USER_NAME); ok {
			logKv = append(logKv, "user", user)
		}
		switch {
		case ctx.Writer.Status() >= 200 && ctx.Writer.Status() < 300:
			log.Logger.Infow("\033[102mGIN\033[0m", logKv...)
		case ctx.Writer.Status() >= 500:
			log.Logger.Infow("\033[101mGIN\033[0m", logKv...)
		default:
			log.Logger.Infow("\033[103mGIN\033[0m", logKv...)
		}
	}

}
