package app

import (
	"github.com/lzh-1625/go_process_manager/internal/app/api"
	"github.com/lzh-1625/go_process_manager/internal/app/cli"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/middle"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/internal/app/route"
	"go.uber.org/fx"
)

var ApiModule = fx.Options(
	fx.Provide(
		api.NewWsApi,
		api.NewProcApi,
		api.NewTaskApi,
		api.NewUserApi,
		api.NewPushApi,
		api.NewEventApi,
		api.NewPermissionApi,
		api.NewLogApi,
		api.NewConfigApi,
		api.NewMetricApi,
		api.NewPermissionTool,
	),
)

var MiddlewareModule = fx.Options(
	fx.Provide(
		middle.NewAuthMiddleware,
		middle.NewEventLoggerMiddleware,
	),
)

var LogicModule = fx.Options(
	fx.Provide(
		logic.NewEventLogic,
		logic.NewPermissionLogic,
		logic.NewMetricLogic,
		logic.NewTaskLogic,
		logic.NewUserLogic,
		logic.NewPushLogic,
		logic.NewLogLogic,
		logic.NewConfigLogic,
		logic.NewProcessCtlLogic,
		logic.NewWsShareLogic,
		logic.NewLogHandler,
		logic.NewEventBus,
	),
)

var RouteModule = fx.Options(
	fx.Provide(
		route.NewRoute,
	),
)

var RepositoryModule = fx.Options(
	fx.Provide(
		repository.NewProcessRepository,
		repository.NewUserRepository,
		repository.NewPushRepository,
		repository.NewEventRepository,
		repository.NewPermissionRepository,
		repository.NewLogRepository,
		repository.NewWsShareRepository,
		repository.NewTaskRepository,
		repository.NewDB,
		repository.NewQuery,
	),
)

var CliModule = fx.Options(
	fx.Provide(
		cli.NewProcessCli,
		cli.NewTaskCli,
		cli.NewUserCli,
		cli.NewPushCli,
		cli.NewWSShareCli,
	),
)

var Module = fx.Options(
	ApiModule,
	MiddlewareModule,
	LogicModule,
	RepositoryModule,
	RouteModule,
	CliModule,
)
