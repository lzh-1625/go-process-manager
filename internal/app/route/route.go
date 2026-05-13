package route

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/pprof"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/api"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/middle"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
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
	r.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	r.HTTPErrorHandler = func(c *echo.Context, err error) {
		log.Logger.Errorw("HTTPErrorHandler", "err", err)
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
	r.Use(loggerMiddleware.EventLogger)

	ProcessWaitCond := middle.NewWaitCond(logic.ProcessWaitCond)
	TaskWaitCond := middle.NewWaitCond(logic.TaskWaitCond)

	apiGroup := r.Group("/api")
	apiGroup.Use(authMiddleware.Auth)
	// apiGroup.Use(middle.DemoMiddle())
	{
		wsGroup := apiGroup.Group("/ws")
		{
			wsGroup.GET("", wsApi.WebsocketHandle)
			wsGroup.GET("/share", wsApi.WebsocketShareHandle)
			wsGroup.GET("/token/list", wsApi.GetWsShareList, middle.RolePermission(eum.RoleAdmin))
			wsGroup.DELETE("/token", wsApi.DeleteWsShareByID, middle.RolePermission(eum.RoleAdmin))
		}

		processGroup := apiGroup.Group("/process")
		{
			processGroup.DELETE("", procApi.KillProcess)
			processGroup.GET("", procApi.GetProcessList)
			processGroup.GET("/wait", procApi.GetProcessList, ProcessWaitCond.WaitGetMiddel)
			processGroup.PUT("", procApi.StartProcess)
			processGroup.PUT("/all", procApi.StartAllProcess)
			processGroup.DELETE("/all", procApi.KillAllProcess)
			processGroup.POST("/share", procApi.ProcessCreateShare, middle.RolePermission(eum.RoleAdmin))
			processGroup.GET("/control", procApi.ProcessControl, middle.RolePermission(eum.RoleRoot), ProcessWaitCond.WaitTriggerMiddel)

			proConfigGroup := processGroup.Group("/config")
			{
				proConfigGroup.POST("", procApi.CreateProcess, middle.RolePermission(eum.RoleRoot), ProcessWaitCond.WaitTriggerMiddel)
				proConfigGroup.DELETE("", procApi.DeleteProcess, middle.RolePermission(eum.RoleRoot), ProcessWaitCond.WaitTriggerMiddel)
				proConfigGroup.PUT("", procApi.UpdateProcessConfig, middle.RolePermission(eum.RoleRoot))
				proConfigGroup.GET("", procApi.GetProcessConfig, middle.RolePermission(eum.RoleAdmin))
			}
		}

		taskGroup := apiGroup.Group("/task")
		{
			taskGroup.GET("", taskApi.GetTaskByID, middle.RolePermission(eum.RoleAdmin))
			taskGroup.GET("/all", taskApi.GetTaskList, middle.RolePermission(eum.RoleAdmin))
			taskGroup.GET("/all/wait", taskApi.GetTaskList, middle.RolePermission(eum.RoleAdmin), TaskWaitCond.WaitGetMiddel)
			taskGroup.POST("", taskApi.CreateTask, middle.RolePermission(eum.RoleAdmin), TaskWaitCond.WaitTriggerMiddel)
			taskGroup.DELETE("", taskApi.DeleteTaskByID, middle.RolePermission(eum.RoleAdmin), TaskWaitCond.WaitTriggerMiddel)
			taskGroup.PUT("", taskApi.EditTask, middle.RolePermission(eum.RoleAdmin), TaskWaitCond.WaitTriggerMiddel)
			taskGroup.GET("/start", taskApi.StartTask, middle.RolePermission(eum.RoleAdmin))
			taskGroup.GET("/stop", taskApi.StopTask, middle.RolePermission(eum.RoleAdmin))
			taskGroup.POST("/key", taskApi.CreateTaskApiKey, middle.RolePermission(eum.RoleAdmin))
			taskGroup.GET("/api-key/:key", taskApi.RunTaskByKey)
		}

		userGroup := apiGroup.Group("/user")
		{
			userGroup.POST("/login", userApi.LoginHandler)
			userGroup.POST("", userApi.CreateUser, middle.RolePermission(eum.RoleRoot))
			userGroup.PUT("", userApi.EditUser, middle.RolePermission(eum.RoleUser))
			userGroup.DELETE("", userApi.DeleteUser, middle.RolePermission(eum.RoleRoot))
			userGroup.GET("", userApi.GetUserList, middle.RolePermission(eum.RoleRoot))
		}

		pushGroup := apiGroup.Group("/push", middle.RolePermission(eum.RoleAdmin))
		{
			pushGroup.GET("/list", pushApi.GetPushList)
			pushGroup.GET("", pushApi.GetPushByID)
			pushGroup.POST("", pushApi.AddPushConfig)
			pushGroup.PUT("", pushApi.UpdatePushConfig)
			pushGroup.DELETE("", pushApi.DeletePushConfig)
		}

		eventGroup := apiGroup.Group("/event", middle.RolePermission(eum.RoleAdmin))
		{
			eventGroup.GET("", eventApi.GetEventList)
		}

		permissionGroup := apiGroup.Group("/permission", middle.RolePermission(eum.RoleRoot))
		{
			permissionGroup.GET("/list", permissionApi.GetPermissionList)
			permissionGroup.PUT("", permissionApi.EditPermssion, ProcessWaitCond.WaitTriggerMiddel)
		}

		logGroup := apiGroup.Group("/log", middle.RolePermission(eum.RoleUser))
		{
			logGroup.POST("", logApi.GetLog)
			logGroup.GET("/running", logApi.GetRunningLog)
		}

		configGroup := apiGroup.Group("/config", middle.RolePermission(eum.RoleRoot))
		{
			configGroup.GET("", configApi.GetSystemConfiguration)
			configGroup.PUT("", configApi.SetSystemConfiguration)
			configGroup.PUT("/reload", configApi.LogConfigReload)
		}
		metricGroup := apiGroup.Group("/metric", middle.RolePermission(eum.RoleAdmin))
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
