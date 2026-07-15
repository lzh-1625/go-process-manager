package route

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/api"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/middle"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/types"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/resources"
)

func NewRoute(
	wsApi *api.WsApi,
	procApi *api.ProcApi,
	taskApi *api.TaskApi,
	userApi *api.UserApi,
	pushApi *api.PushApi,
	eventApi *api.EventApi,
	permissionApi *api.PermissionApi,
	logApi *api.LogApi,
	configApi *api.ConfigApi,
	metricApi *api.MetricApi,
	loggerMiddleware *middle.EventLoggerMiddleware,
	authMiddleware *middle.AuthMiddleware,

) *echo.Echo {

	r := echo.New()
	r.Use(middleware.Recover())
	r.Use(middle.Logger)

	// close echo default log print
	r.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	r.HTTPErrorHandler = func(c *echo.Context, err error) {
		log.Logger.Warn(err)
		c.JSON(http.StatusInternalServerError, model.Response[struct{}]{
			Code:    -1,
			Message: "error: " + err.Error(),
		})
	}
	r.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			next(c)
			if resp, err := echo.UnwrapResponse(c.Response()); err == nil && !resp.Committed && !c.IsWebSocket() {
				return c.JSON(http.StatusOK, model.Response[any]{
					Code:    0,
					Message: "success",
				})
			}
			return nil
		}
	})
	if config.CF.StaticResourceCahce {
		r.Use(middle.CacheMiddleware())
	}
	// static file
	r.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		HTML5:      true,
		Root:       "dist",
		Index:      "index.html",
		Filesystem: resources.Templates,
		Browse:     true,
	}))
	if config.CF.PprofEnable {
		pprofInit(r)
	}

	if config.CF.GZipEnable {
		r.Use(middleware.Gzip())
	}

	if os.Getenv("GPM_DEMO") == "true" {
		r.Use(middle.DemoMiddle())
	}
	r.Use(loggerMiddleware.EventLogger)

	ProcessWaitCond := middle.NewWaitCond(logic.ProcessWaitCond())
	TaskWaitCond := middle.NewWaitCond(logic.TaskWaitCond())

	apiGroup := r.Group("/api")
	apiGroup.Use(authMiddleware.Auth)
	{
		wsGroup := apiGroup.Group("/ws")
		{
			wsGroup.GET("", wsApi.WebsocketHandle)
			wsGroup.GET("/share", wsApi.WebsocketShareHandle)
			wsGroup.GET("/token/list", wsApi.GetWsShareList, middle.RolePermission(types.RoleAdmin))
			wsGroup.DELETE("/token", wsApi.DeleteWsShareByID, middle.RolePermission(types.RoleAdmin))
		}

		processGroup := apiGroup.Group("/process")
		{
			processGroup.DELETE("", procApi.KillProcess)
			processGroup.GET("", procApi.GetProcessList)
			processGroup.GET("/wait", procApi.GetProcessList, ProcessWaitCond.WaitGetMiddel)
			processGroup.PUT("", procApi.StartProcess)
			processGroup.PUT("/all", procApi.StartAllProcess)
			processGroup.DELETE("/all", procApi.KillAllProcess)
			processGroup.POST("/share", procApi.ProcessCreateShare, middle.RolePermission(types.RoleAdmin))
			processGroup.GET("/control", procApi.ProcessControl, middle.RolePermission(types.RoleRoot), ProcessWaitCond.WaitTriggerMiddel)

			proConfigGroup := processGroup.Group("/config")
			{
				proConfigGroup.POST("", procApi.CreateProcess, middle.RolePermission(types.RoleRoot), ProcessWaitCond.WaitTriggerMiddel)
				proConfigGroup.DELETE("", procApi.DeleteProcess, middle.RolePermission(types.RoleRoot), ProcessWaitCond.WaitTriggerMiddel)
				proConfigGroup.PUT("", procApi.UpdateProcessConfig, middle.RolePermission(types.RoleRoot))
				proConfigGroup.GET("", procApi.GetProcessConfig, middle.RolePermission(types.RoleAdmin))
			}
		}

		taskGroup := apiGroup.Group("/task")
		{
			taskGroup.GET("", taskApi.GetTaskByID, middle.RolePermission(types.RoleAdmin))
			taskGroup.GET("/all", taskApi.GetTaskList, middle.RolePermission(types.RoleAdmin))
			taskGroup.GET("/all/wait", taskApi.GetTaskList, middle.RolePermission(types.RoleAdmin), TaskWaitCond.WaitGetMiddel)
			taskGroup.POST("", taskApi.CreateTask, middle.RolePermission(types.RoleAdmin), TaskWaitCond.WaitTriggerMiddel)
			taskGroup.DELETE("", taskApi.DeleteTaskByID, middle.RolePermission(types.RoleAdmin), TaskWaitCond.WaitTriggerMiddel)
			taskGroup.PUT("", taskApi.EditTask, middle.RolePermission(types.RoleAdmin), TaskWaitCond.WaitTriggerMiddel)
			taskGroup.GET("/start", taskApi.StartTask, middle.RolePermission(types.RoleAdmin))
			taskGroup.GET("/stop", taskApi.StopTask, middle.RolePermission(types.RoleAdmin))
			taskGroup.POST("/key", taskApi.CreateTaskApiKey, middle.RolePermission(types.RoleAdmin))
			taskGroup.GET("/api-key/:key", taskApi.RunTaskByKey)
		}

		userGroup := apiGroup.Group("/user")
		{
			userGroup.POST("/login", userApi.LoginHandler)
			userGroup.POST("", userApi.CreateUser, middle.RolePermission(types.RoleRoot))
			userGroup.PUT("", userApi.EditUser, middle.RolePermission(types.RoleUser))
			userGroup.DELETE("", userApi.DeleteUser, middle.RolePermission(types.RoleRoot))
			userGroup.GET("", userApi.GetUserList, middle.RolePermission(types.RoleRoot))
		}

		pushGroup := apiGroup.Group("/push", middle.RolePermission(types.RoleAdmin))
		{
			pushGroup.GET("/list", pushApi.GetPushList)
			pushGroup.GET("", pushApi.GetPushByID)
			pushGroup.POST("", pushApi.AddPushConfig)
			pushGroup.PUT("", pushApi.UpdatePushConfig)
			pushGroup.DELETE("", pushApi.DeletePushConfig)
		}

		eventGroup := apiGroup.Group("/event", middle.RolePermission(types.RoleAdmin))
		{
			eventGroup.GET("", eventApi.GetEventList)
		}

		permissionGroup := apiGroup.Group("/permission", middle.RolePermission(types.RoleRoot))
		{
			permissionGroup.GET("/list", permissionApi.GetPermissionList)
			permissionGroup.PUT("", permissionApi.EditPermssion, ProcessWaitCond.WaitTriggerMiddel)
		}

		logGroup := apiGroup.Group("/log", middle.RolePermission(types.RoleUser))
		{
			logGroup.POST("", logApi.GetLog)
			logGroup.GET("/running", logApi.GetRunningLog)
		}

		configGroup := apiGroup.Group("/config", middle.RolePermission(types.RoleRoot))
		{
			configGroup.GET("", configApi.GetSystemConfiguration)
			configGroup.PUT("", configApi.SetSystemConfiguration)
			configGroup.PUT("/reload", configApi.LogConfigReload)
		}
		metricGroup := apiGroup.Group("/metric", middle.RolePermission(types.RoleAdmin))
		{
			metricGroup.GET("/log", metricApi.GetLogicStatsticMetric)
			metricGroup.GET("/performce", metricApi.GetPerformceUsage)
		}
	}
	return r
}

func pprofInit(r *echo.Echo) {
	g := r.Group("/debug/pprof")
	g.GET("/", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
	g.GET("/cmdline", echo.WrapHandler(http.HandlerFunc(pprof.Cmdline)))
	g.GET("/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)))
	g.GET("/symbol", echo.WrapHandler(http.HandlerFunc(pprof.Symbol)))
	g.GET("/trace", echo.WrapHandler(http.HandlerFunc(pprof.Trace)))
	g.GET("/heap", echo.WrapHandler(pprof.Handler("heap")))
	g.GET("/goroutine", echo.WrapHandler(pprof.Handler("goroutine")))
	g.GET("/threadcreate", echo.WrapHandler(pprof.Handler("threadcreate")))
	g.GET("/block", echo.WrapHandler(pprof.Handler("block")))
	log.Logger.Debug("enable pprof")
}
