package route

import (
	"mime"
	"net/http"
	"path/filepath"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/api"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/middle"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/resources"
	"github.com/lzh-1625/go_process_manager/utils"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func Route() {
	r := gin.New()
	r.Use(gin.Recovery())
	if !config.CF.Tui {
		r.Use(middle.Logger())
	}
	r.Use(middle.EventLogger())
	routePathInit(r)
	staticInit(r)
	pprofInit(r)
	err := r.Run(config.CF.Listen)
	log.Logger.Fatalw("服务器启动失败", "err", err)
}

func staticInit(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		path := "dist" + c.Request.URL.Path
		if data, err := resources.Templates.ReadFile(path); err == nil {
			c.Data(http.StatusOK, mime.TypeByExtension(filepath.Ext(path)), data)
		} else {
			c.Data(http.StatusOK, "text/html; charset=utf-8", utils.UnwarpIgnore(resources.Templates.ReadFile("dist/index.html")))
		}
	})
}

func pprofInit(r *gin.Engine) {
	if config.CF.PprofEnable {
		pprof.Register(r)
		log.Logger.Debug("启用 pprof")
	}
}

func routePathInit(r *gin.Engine) {

	ProcessWaitCond := middle.NewWaitCond(logic.ProcessWaitCond)
	TaskWaitCond := middle.NewWaitCond(logic.TaskWaitCond)

	apiGroup := r.Group("/api")
	apiGroup.Use(middle.CheckToken())
	// apiGroup.Use(middle.DemoMiddle())
	{
		wsGroup := apiGroup.Group("/ws")
		{
			wsGroup.GET("", bind(api.WsApi.WebsocketHandle, Query))
			wsGroup.GET("/share", bind(api.WsApi.WebsocketShareHandle, Query))
			wsGroup.GET("/token/list", middle.RolePermission(eum.RoleAdmin), bind(api.GetWsShareList, None))
			wsGroup.DELETE("/token", middle.RolePermission(eum.RoleAdmin), bind(api.DeleteWsShareById, Query))
		}

		processGroup := apiGroup.Group("/process")
		{
			processGroup.DELETE("", bind(api.ProcApi.KillProcess, Query))
			processGroup.GET("", bind(api.ProcApi.GetProcessList, None))
			processGroup.GET("/wait", ProcessWaitCond.WaitGetMiddel, bind(api.ProcApi.GetProcessList, None))
			processGroup.PUT("", bind(api.ProcApi.StartProcess, Body))
			processGroup.PUT("/all", bind(api.ProcApi.StartAllProcess, None))
			processGroup.DELETE("/all", bind(api.ProcApi.KillAllProcess, None))
			processGroup.POST("/share", middle.RolePermission(eum.RoleAdmin), bind(api.ProcApi.ProcessCreateShare, Body))
			processGroup.GET("/control", middle.RolePermission(eum.RoleRoot), ProcessWaitCond.WaitTriggerMiddel, bind(api.ProcApi.ProcessControl, Query))

			proConfigGroup := processGroup.Group("/config")
			{
				proConfigGroup.POST("", middle.RolePermission(eum.RoleRoot), ProcessWaitCond.WaitTriggerMiddel, bind(api.ProcApi.CreateNewProcess, Body))
				proConfigGroup.DELETE("", middle.RolePermission(eum.RoleRoot), ProcessWaitCond.WaitTriggerMiddel, bind(api.ProcApi.DeleteNewProcess, Query))
				proConfigGroup.PUT("", middle.RolePermission(eum.RoleRoot), bind(api.ProcApi.UpdateProcessConfig, Body))
				proConfigGroup.GET("", middle.RolePermission(eum.RoleAdmin), bind(api.ProcApi.GetProcessConfig, Query))
			}
		}

		taskGroup := apiGroup.Group("/task")
		{
			taskGroup.GET("", middle.RolePermission(eum.RoleAdmin), bind(api.TaskApi.GetTaskById, Query))
			taskGroup.GET("/all", middle.RolePermission(eum.RoleAdmin), bind(api.TaskApi.GetTaskList, None))
			taskGroup.GET("/all/wait", middle.RolePermission(eum.RoleAdmin), TaskWaitCond.WaitGetMiddel, bind(api.TaskApi.GetTaskList, None))
			taskGroup.POST("", middle.RolePermission(eum.RoleAdmin), TaskWaitCond.WaitTriggerMiddel, bind(api.TaskApi.CreateTask, Body))
			taskGroup.DELETE("", middle.RolePermission(eum.RoleAdmin), TaskWaitCond.WaitTriggerMiddel, bind(api.TaskApi.DeleteTaskById, Query))
			taskGroup.PUT("", middle.RolePermission(eum.RoleAdmin), TaskWaitCond.WaitTriggerMiddel, bind(api.TaskApi.EditTask, Body))
			taskGroup.PUT("/enable", middle.RolePermission(eum.RoleAdmin), TaskWaitCond.WaitTriggerMiddel, bind(api.TaskApi.EditTaskEnable, Body))
			taskGroup.GET("/start", middle.RolePermission(eum.RoleAdmin), bind(api.TaskApi.StartTask, Query))
			taskGroup.GET("/stop", middle.RolePermission(eum.RoleAdmin), bind(api.TaskApi.StopTask, Query))
			taskGroup.POST("/key", middle.RolePermission(eum.RoleAdmin), bind(api.TaskApi.CreateTaskApiKey, Body))
			taskGroup.GET("/api-key/:key", bind(api.TaskApi.RunTaskByKey, None))
		}

		userGroup := apiGroup.Group("/user")
		{
			userGroup.POST("/login", bind(api.UserApi.LoginHandler, Body))
			userGroup.POST("", middle.RolePermission(eum.RoleRoot), bind(api.UserApi.CreateUser, Body))
			userGroup.PUT("", middle.RolePermission(eum.RoleUser), bind(api.UserApi.EditUser, Body))
			userGroup.DELETE("", middle.RolePermission(eum.RoleRoot), bind(api.UserApi.DeleteUser, Query))
			userGroup.GET("", middle.RolePermission(eum.RoleRoot), bind(api.UserApi.GetUserList, None))
		}

		pushGroup := apiGroup.Group("/push").Use(middle.RolePermission(eum.RoleAdmin))
		{
			pushGroup.GET("/list", bind(api.PushApi.GetPushList, None))
			pushGroup.GET("", bind(api.PushApi.GetPushById, Query))
			pushGroup.POST("", bind(api.PushApi.AddPushConfig, Body))
			pushGroup.PUT("", bind(api.PushApi.UpdatePushConfig, Body))
			pushGroup.DELETE("", bind(api.PushApi.DeletePushConfig, Query))
		}

		fileGroup := apiGroup.Group("/file").Use(middle.RolePermission(eum.RoleAdmin))
		{
			fileGroup.GET("/list", bind(api.FileApi.FilePathHandler, Query))
			fileGroup.PUT("", bind(api.FileApi.FileWriteHandler, None))
			fileGroup.GET("", bind(api.FileApi.FileReadHandler, Query))
		}

		eventGroup := apiGroup.Group("/event").Use(middle.RolePermission(eum.RoleAdmin))
		{
			eventGroup.GET("", bind(api.EventApi.GetEventList, Query))
		}

		permissionGroup := apiGroup.Group("/permission").Use(middle.RolePermission(eum.RoleRoot))
		{
			permissionGroup.GET("/list", bind(api.PermissionApi.GetPermissionList, Query))
			permissionGroup.PUT("", ProcessWaitCond.WaitTriggerMiddel, bind(api.PermissionApi.EditPermssion, Body))
		}

		logGroup := apiGroup.Group("/log").Use(middle.RolePermission(eum.RoleUser))
		{
			logGroup.POST("", bind(api.LogApi.GetLog, Body))
			logGroup.GET("/running", bind(api.LogApi.GetRunningLog, None))
		}

		configGroup := apiGroup.Group("/config").Use(middle.RolePermission(eum.RoleRoot))
		{
			configGroup.GET("", bind(api.ConfigApi.GetSystemConfiguration, None))
			configGroup.PUT("", bind(api.ConfigApi.SetSystemConfiguration, None))
			configGroup.PUT("/reload", bind(api.ConfigApi.LogConfigReload, None))
		}
		metricGroup := apiGroup.Group("/metric").Use(middle.RolePermission(eum.RoleAdmin))
		{
			metricGroup.GET("/log", bind(api.MetricApi.GetLogicStatsticMetric, Query))
			metricGroup.GET("/performce", bind(api.MetricApi.GetPerformceUsage, None))
		}
	}
}

const (
	None   = 0
	Header = 1 << iota
	Body
	Query
)

func bind[T any, R any](fn func(*gin.Context, T) R, bindOption int) func(*gin.Context) {
	return func(ctx *gin.Context) {
		var req T
		if bindOption&Body != 0 {
			if err := ctx.BindJSON(&req); err != nil {
				rErr(ctx, err)
				return
			}
		}
		if bindOption&Header != 0 {
			if err := ctx.BindHeader(&req); err != nil {
				rErr(ctx, err)
				return
			}
		}
		if bindOption&Query != 0 {
			if err := ctx.BindQuery(&req); err != nil {
				rErr(ctx, err)
				return
			}
		}
		result := fn(ctx, req)
		switch v := any(result).(type) {
		case error:
			if v != nil {
				rErr(ctx, v)
				return
			} else {
				ctx.JSON(200, gin.H{
					"code":    0,
					"message": "success",
				})
				return
			}
		case *api.Response:
			ctx.JSON(v.StatusCode, gin.H{
				"data": v.Data,
				"msg":  v.Msg,
				"code": v.Code,
			})
			return
		default:
			ctx.JSON(200, gin.H{
				"code":    0,
				"message": "success",
				"data":    v,
			})
		}
	}
}

func rErr(ctx *gin.Context, err error) {
	log.Logger.Warn(err)
	ctx.JSON(500, gin.H{
		"code":    -1,
		"message": err.Error(),
	})
}
