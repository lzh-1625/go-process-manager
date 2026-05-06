package middle

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/log"
)

func Logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		path := c.Request().URL.Path

		// 非 /api 直接放行
		if !strings.HasPrefix(path, "/api") {
			return next(c)
		}

		// 执行 handler
		err := next(c)

		logKv := []any{}
		logKv = append(logKv, "Method", c.Request().Method)
		logKv = append(logKv, "Status", c.Response().Status)
		logKv = append(logKv, "Path", path)
		logKv = append(logKv, "耗时", fmt.Sprintf("%dms", time.Now().UnixMilli()-start.UnixMilli()))

		if user, ok := c.Get(eum.CtxUserName).(string); ok && user != "" {
			logKv = append(logKv, "user", user)
		}

		switch {
		case c.Response().Status >= 200 && c.Response().Status < 300:
			log.Logger.Infow("\033[102mHTTP\033[0m", logKv...)
		case c.Response().Status >= 500:
			log.Logger.Infow("\033[101mHTTP\033[0m", logKv...)
		default:
			log.Logger.Infow("\033[103mHTTP\033[0m", logKv...)
		}

		return err
	}
}

func EventLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)

		if !slices.Contains([]string{
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
		}, c.Request().Method) {
			return err
		}

		user, _ := c.Get(eum.CtxUserName).(string)
		if user == "" {
			return err
		}

		logic.EventLogic.Create(
			c.Request().Method,
			eum.EventApiRequest,
			"uri", c.Request().URL.Path,
			"method", c.Request().Method,
			"user", user,
			"status", strconv.Itoa(c.Response().Status),
		)
		return err
	}
}
