package route

import (
	"mime"
	"net/http"
	"net/http/pprof"
	"path/filepath"
	"strings"

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

func Route() {

	r := echo.New()
	r.Use(middleware.Recover())
	r.Use(middle.Logger)
	r.HTTPErrorHandler = func(c *echo.Context, err error) {
		log.Logger.Errorw("HTTPErrorHandler", "err", err)
		c.JSON(http.StatusInternalServerError, model.Response[struct{}]{
			Code:    -1,
			Message: "error: " + err.Error(),
		})
	}
	r.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		HTML5:      true,
		Root:       "dist", // because files are located in `assets` directory in `webAssets` fs
		Index:      "index.html",
		Filesystem: resources.Templates,
		Skipper: func(c *echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.Path, "/api")
		},
		Browse: true,
	}))
	if config.CF.PprofEnable {
		pprofInit(r)
	}

	if config.CF.GZipEnable {
		r.Use(middleware.Gzip())
	}
	r.Use(middle.EventLogger)
	routePathInit(r)
	// pprofInit(r)
	// err := r.Start(config.CF.Listen)
	err := r.Start(":8081")
	log.Logger.Fatalw("服务器启动失败", "err", err)
}

func staticInit(r *echo.Echo) {
	r.Any("/*", func(c *echo.Context) error {
		path := "dist" + c.Request().URL.Path
		if data, err := resources.Templates.ReadFile(path); err == nil {
			return c.Blob(http.StatusOK, mime.TypeByExtension(filepath.Ext(path)), data)
		}
		data, _ := resources.Templates.ReadFile("dist/index.html")
		return c.HTMLBlob(http.StatusOK, data)
	})
}

func pprofInit(r *echo.Echo) {
	if config.CF.PprofEnable {
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
		log.Logger.Debug("启用 pprof")
	}
}

func routePathInit(r *echo.Echo) {

	ProcessWaitCond := middle.NewWaitCond(logic.ProcessWaitCond)
	TaskWaitCond := middle.NewWaitCond(logic.TaskWaitCond)

	apiGroup := r.Group("/api")
	apiGroup.Use(middle.Auth)
	// apiGroup.Use(middle.DemoMiddle())
	{
		wsGroup := apiGroup.Group("/ws")
		{
			wsGroup.GET("", api.WsApi.WebsocketHandle)
			wsGroup.GET("/share", api.WsApi.WebsocketShareHandle)
			wsGroup.GET("/token/list", api.GetWsShareList, middle.RolePermission(eum.RoleAdmin))
			wsGroup.DELETE("/token", api.DeleteWsShareById, middle.RolePermission(eum.RoleAdmin))
		}

		processGroup := apiGroup.Group("/process")
		{
			processGroup.DELETE("", api.ProcApi.KillProcess)
			processGroup.GET("", api.ProcApi.GetProcessList)
			processGroup.GET("/wait", api.ProcApi.GetProcessList, ProcessWaitCond.WaitGetMiddel)
			processGroup.PUT("", api.ProcApi.StartProcess)
			processGroup.PUT("/all", api.ProcApi.StartAllProcess)
			processGroup.DELETE("/all", api.ProcApi.KillAllProcess)
			processGroup.POST("/share", api.ProcApi.ProcessCreateShare, middle.RolePermission(eum.RoleAdmin))
			processGroup.GET("/control", api.ProcApi.ProcessControl, middle.RolePermission(eum.RoleRoot), ProcessWaitCond.WaitTriggerMiddel)

			proConfigGroup := processGroup.Group("/config")
			{
				proConfigGroup.POST("", api.ProcApi.CreateProcess, middle.RolePermission(eum.RoleRoot), ProcessWaitCond.WaitTriggerMiddel)
				proConfigGroup.DELETE("", api.ProcApi.DeleteProcess, middle.RolePermission(eum.RoleRoot), ProcessWaitCond.WaitTriggerMiddel)
				proConfigGroup.PUT("", api.ProcApi.UpdateProcessConfig, middle.RolePermission(eum.RoleRoot))
				proConfigGroup.GET("", api.ProcApi.GetProcessConfig, middle.RolePermission(eum.RoleAdmin))
			}
		}

		taskGroup := apiGroup.Group("/task")
		{
			taskGroup.GET("", api.TaskApi.GetTaskById, middle.RolePermission(eum.RoleAdmin))
			taskGroup.GET("/all", api.TaskApi.GetTaskList, middle.RolePermission(eum.RoleAdmin))
			taskGroup.GET("/all/wait", api.TaskApi.GetTaskList, middle.RolePermission(eum.RoleAdmin), TaskWaitCond.WaitGetMiddel)
			taskGroup.POST("", api.TaskApi.CreateTask, middle.RolePermission(eum.RoleAdmin), TaskWaitCond.WaitTriggerMiddel)
			taskGroup.DELETE("", api.TaskApi.DeleteTaskById, middle.RolePermission(eum.RoleAdmin), TaskWaitCond.WaitTriggerMiddel)
			taskGroup.PUT("", api.TaskApi.EditTask, middle.RolePermission(eum.RoleAdmin), TaskWaitCond.WaitTriggerMiddel)
			taskGroup.PUT("/enable", api.TaskApi.EditTaskEnable, middle.RolePermission(eum.RoleAdmin), TaskWaitCond.WaitTriggerMiddel)
			taskGroup.GET("/start", api.TaskApi.StartTask, middle.RolePermission(eum.RoleAdmin))
			taskGroup.GET("/stop", api.TaskApi.StopTask, middle.RolePermission(eum.RoleAdmin))
			taskGroup.POST("/key", api.TaskApi.CreateTaskApiKey, middle.RolePermission(eum.RoleAdmin))
			taskGroup.GET("/api-key/:key", api.TaskApi.RunTaskByKey)
		}

		userGroup := apiGroup.Group("/user")
		{
			userGroup.POST("/login", api.UserApi.LoginHandler)
			userGroup.POST("", api.UserApi.CreateUser, middle.RolePermission(eum.RoleRoot))
			userGroup.PUT("", api.UserApi.EditUser, middle.RolePermission(eum.RoleUser))
			userGroup.DELETE("", api.UserApi.DeleteUser, middle.RolePermission(eum.RoleRoot))
			userGroup.GET("", api.UserApi.GetUserList, middle.RolePermission(eum.RoleRoot))
		}

		pushGroup := apiGroup.Group("/push", middle.RolePermission(eum.RoleAdmin))
		{
			pushGroup.GET("/list", api.PushApi.GetPushList)
			pushGroup.GET("", api.PushApi.GetPushById)
			pushGroup.POST("", api.PushApi.AddPushConfig)
			pushGroup.PUT("", api.PushApi.UpdatePushConfig)
			pushGroup.DELETE("", api.PushApi.DeletePushConfig)
		}

		eventGroup := apiGroup.Group("/event", middle.RolePermission(eum.RoleAdmin))
		{
			eventGroup.GET("", api.EventApi.GetEventList)
		}

		permissionGroup := apiGroup.Group("/permission", middle.RolePermission(eum.RoleRoot))
		{
			permissionGroup.GET("/list", api.PermissionApi.GetPermissionList)
			permissionGroup.PUT("", api.PermissionApi.EditPermssion, ProcessWaitCond.WaitTriggerMiddel)
		}

		logGroup := apiGroup.Group("/log", middle.RolePermission(eum.RoleUser))
		{
			logGroup.POST("", api.LogApi.GetLog)
			logGroup.GET("/running", api.LogApi.GetRunningLog)
		}

		configGroup := apiGroup.Group("/config", middle.RolePermission(eum.RoleRoot))
		{
			configGroup.GET("", api.ConfigApi.GetSystemConfiguration)
			configGroup.PUT("", api.ConfigApi.SetSystemConfiguration)
			configGroup.PUT("/reload", api.ConfigApi.LogConfigReload)
		}
		metricGroup := apiGroup.Group("/metric", middle.RolePermission(eum.RoleAdmin))
		{
			metricGroup.GET("/log", api.MetricApi.GetLogicStatsticMetric)
			metricGroup.GET("/performce", api.MetricApi.GetPerformceUsage)
		}
	}
}
