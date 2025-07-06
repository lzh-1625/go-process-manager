package route

import (
	"io/fs"
	"net/http"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/api"
	"github.com/lzh-1625/go_process_manager/internal/app/constants"
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
	routePathInit(r)
	staticInit(r)
	pprofInit(r)
	err := r.Run(config.CF.Listen)
	log.Logger.Fatalw("服务器启动失败", "err", err)
}

func staticInit(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		b, _ := resources.Templates.ReadFile("templates/index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", b)
	})
	r.StaticFS("/js", http.FS(utils.UnwarpIgnore(fs.Sub(resources.Templates, "templates/js"))))
	r.StaticFS("/css", http.FS(utils.UnwarpIgnore(fs.Sub(resources.Templates, "templates/css"))))
	r.StaticFS("/media", http.FS(utils.UnwarpIgnore(fs.Sub(resources.Templates, "templates/media"))))
	r.StaticFS("/fonts", http.FS(utils.UnwarpIgnore(fs.Sub(resources.Templates, "templates/fonts"))))
	r.GET("/favicon.ico", func(ctx *gin.Context) {
		ctx.Data(200, "image/x-icon", utils.UnwarpIgnore(resources.Templates.ReadFile("templates/favicon.ico")))
	})
}

func pprofInit(r *gin.Engine) {
	if config.CF.PprofEnable {
		pprof.Register(r)
		log.Logger.Debug("启用 pprof")
	}
}

func routePathInit(r *gin.Engine) {

	apiGroup := r.Group("/api")
	apiGroup.Use(middle.CheckToken())
	// apiGroup.Use(middle.DemoMiddle())
	{
		wsGroup := apiGroup.Group("/ws")
		{
			wsGroup.GET("", middle.OprPermission(constants.OPERATION_TERMINAL), bind(api.WsApi.WebsocketHandle, Query))
			wsGroup.GET("/share", bind(api.WsApi.WebsocketShareHandle, Query))
		}

		processGroup := apiGroup.Group("/process")
		{
			processGroup.DELETE("", middle.OprPermission(constants.OPERATION_STOP), bind(api.ProcApi.KillProcess, Query))
			processGroup.GET("", bind(api.ProcApi.GetProcessList, None))
			processGroup.GET("/wait", middle.ProcessWaitCond.WaitGetMiddel, bind(api.ProcApi.GetProcessList, None))
			processGroup.PUT("", middle.OprPermission(constants.OPERATION_START), bind(api.ProcApi.StartProcess, Body))
			processGroup.PUT("/all", bind(api.ProcApi.StartAllProcess, None))
			processGroup.DELETE("/all", bind(api.ProcApi.KillAllProcess, None))
			processGroup.POST("/share", middle.RolePermission(constants.ROLE_ADMIN), bind(api.ProcApi.ProcessCreateShare, Query))
			processGroup.GET("/control", middle.RolePermission(constants.ROLE_ROOT), middle.ProcessWaitCond.WaitTriggerMiddel, bind(api.ProcApi.ProcessControl, Query))

			proConfigGroup := processGroup.Group("/config")
			{
				proConfigGroup.POST("", middle.RolePermission(constants.ROLE_ROOT), middle.ProcessWaitCond.WaitTriggerMiddel, bind(api.ProcApi.CreateNewProcess, Body))
				proConfigGroup.DELETE("", middle.RolePermission(constants.ROLE_ROOT), middle.ProcessWaitCond.WaitTriggerMiddel, bind(api.ProcApi.DeleteNewProcess, Query))
				proConfigGroup.PUT("", middle.RolePermission(constants.ROLE_ROOT), bind(api.ProcApi.UpdateProcessConfig, Body))
				proConfigGroup.GET("", middle.RolePermission(constants.ROLE_ADMIN), bind(api.ProcApi.GetProcessConfig, Query))
			}
		}

		taskGroup := apiGroup.Group("/task")
		{
			taskGroup.GET("", middle.RolePermission(constants.ROLE_ADMIN), bind(api.TaskApi.GetTaskById, Query))
			taskGroup.GET("/all", middle.RolePermission(constants.ROLE_ADMIN), bind(api.TaskApi.GetTaskList, None))
			taskGroup.GET("/all/wait", middle.RolePermission(constants.ROLE_ADMIN), middle.TaskWaitCond.WaitGetMiddel, bind(api.TaskApi.GetTaskList, None))
			taskGroup.POST("", middle.RolePermission(constants.ROLE_ADMIN), middle.TaskWaitCond.WaitTriggerMiddel, bind(api.TaskApi.CreateTask, Body))
			taskGroup.DELETE("", middle.RolePermission(constants.ROLE_ADMIN), middle.TaskWaitCond.WaitTriggerMiddel, bind(api.TaskApi.DeleteTaskById, Query))
			taskGroup.PUT("", middle.RolePermission(constants.ROLE_ADMIN), middle.TaskWaitCond.WaitTriggerMiddel, bind(api.TaskApi.EditTask, Body))
			taskGroup.PUT("/enable", middle.RolePermission(constants.ROLE_ADMIN), middle.TaskWaitCond.WaitTriggerMiddel, bind(api.TaskApi.EditTaskEnable, Body))
			taskGroup.GET("/start", middle.RolePermission(constants.ROLE_ADMIN), bind(api.TaskApi.StartTask, Query))
			taskGroup.GET("/stop", middle.RolePermission(constants.ROLE_ADMIN), bind(api.TaskApi.StopTask, Query))
			taskGroup.POST("/key", middle.RolePermission(constants.ROLE_ADMIN), bind(api.TaskApi.CreateTaskApiKey, Body))
			taskGroup.GET("/api-key/:key", bind(api.TaskApi.RunTaskByKey, None))
		}

		userGroup := apiGroup.Group("/user")
		{
			userGroup.POST("/login", bind(api.UserApi.LoginHandler, Body))
			userGroup.POST("", middle.RolePermission(constants.ROLE_ROOT), bind(api.UserApi.CreateUser, Body))
			userGroup.PUT("/password", middle.RolePermission(constants.ROLE_USER), bind(api.UserApi.ChangePassword, Body))
			userGroup.DELETE("", middle.RolePermission(constants.ROLE_ROOT), bind(api.UserApi.DeleteUser, Query))
			userGroup.GET("", middle.RolePermission(constants.ROLE_ROOT), bind(api.UserApi.GetUserList, None))
		}

		pushGroup := apiGroup.Group("/push").Use(middle.RolePermission(constants.ROLE_ADMIN))
		{
			pushGroup.GET("/list", bind(api.PushApi.GetPushList, None))
			pushGroup.GET("", bind(api.PushApi.GetPushById, Query))
			pushGroup.POST("", bind(api.PushApi.AddPushConfig, Body))
			pushGroup.PUT("", bind(api.PushApi.UpdatePushConfig, Body))
			pushGroup.DELETE("", bind(api.PushApi.DeletePushConfig, Query))
		}

		fileGroup := apiGroup.Group("/file").Use(middle.RolePermission(constants.ROLE_ADMIN))
		{
			fileGroup.GET("/list", bind(api.FileApi.FilePathHandler, Query))
			fileGroup.PUT("", bind(api.FileApi.FileWriteHandler, None))
			fileGroup.GET("", bind(api.FileApi.FileReadHandler, Query))
		}

		permissionGroup := apiGroup.Group("/permission").Use(middle.RolePermission(constants.ROLE_ROOT))
		{
			permissionGroup.GET("/list", bind(api.PermissionApi.GetPermissionList, Query))
			permissionGroup.PUT("", middle.ProcessWaitCond.WaitTriggerMiddel, bind(api.PermissionApi.EditPermssion, Body))
		}

		logGroup := apiGroup.Group("/log").Use(middle.RolePermission(constants.ROLE_USER))
		{
			logGroup.POST("", bind(api.LogApi.GetLog, Body))
			logGroup.GET("/running", bind(api.LogApi.GetRunningLog, None))
		}

		configGroup := apiGroup.Group("/config").Use(middle.RolePermission(constants.ROLE_ROOT))
		{
			configGroup.GET("", bind(api.ConfigApi.GetSystemConfiguration, None))
			configGroup.PUT("", bind(api.ConfigApi.SetSystemConfiguration, None))
			configGroup.PUT("/es", bind(api.ConfigApi.EsConfigReload, None))
		}
	}
}

const (
	None   = 0
	Header = 1 << iota
	Body
	Query
)

func bind[T any](fn func(*gin.Context, T) error, bindOption int) func(*gin.Context) {
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
		err := fn(ctx, req)
		if err != nil {
			rErr(ctx, err)
			return
		}
		if !ctx.Writer.Written() {
			ctx.JSON(200, gin.H{
				"code":    0,
				"message": "success",
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
